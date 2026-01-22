//#go:build integration

package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/repository"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_UpDown_Successfully(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := godotenv.Load("../../../.env"); err != nil {
		require.NoError(t, err, "Load environments")
	}

	// region data provider
	tests := []struct {
		name                      string
		selectQueryToRecordsCount string
		wantRecordsCount          int
		options                   *Options
	}{
		{
			name:                      "tarantool",
			selectQueryToRecordsCount: "return box.space.test:len()",
			wantRecordsCount:          1,
			options: &Options{
				DSN:       os.Getenv("TARANTOOL_DSN"),
				Directory: migrationsPathAbs(os.Getenv("TARANTOOL_MIGRATIONS_PATH")),
				TableName: "migration",
				//Compact:     true,
				Interactive: false,
			},
		},
		{
			name:                      "postgres",
			selectQueryToRecordsCount: "select count(*) from test",
			wantRecordsCount:          1,
			options: &Options{
				DSN:         os.Getenv("POSTGRES_DSN"),
				Directory:   migrationsPathAbs(os.Getenv("POSTGRES_MIGRATIONS_PATH")),
				TableName:   "migration",
				Compact:     true,
				Interactive: false,
			},
		},
		{
			name:                      "mysql",
			selectQueryToRecordsCount: "select count(*) from test",
			wantRecordsCount:          1,
			options: &Options{
				DSN:         os.Getenv("MYSQL_DSN"),
				Directory:   migrationsPathAbs(os.Getenv("MYSQL_MIGRATIONS_PATH")),
				TableName:   "migration",
				Compact:     true,
				Interactive: false,
			},
		},
		{
			name:                      "clickhouse",
			selectQueryToRecordsCount: "select count() from test",
			wantRecordsCount:          1,
			options: &Options{
				DSN:         os.Getenv("CLICKHOUSE_DSN"),
				Directory:   migrationsPathAbs(os.Getenv("CLICKHOUSE_MIGRATIONS_PATH")),
				TableName:   "migration",
				Compact:     true,
				Interactive: false,
			},
		},
		{
			name:                      "clickhouse_cluster",
			selectQueryToRecordsCount: "select count() from raw.test",
			wantRecordsCount:          1,
			options: &Options{
				DSN:         os.Getenv("CLICKHOUSE_CLUSTER_DSN"),
				Directory:   migrationsPathAbs(os.Getenv("CLICKHOUSE_CLUSTER_MIGRATIONS_PATH")),
				TableName:   "migration",
				ClusterName: os.Getenv("MIGRATION_CLUSTER_NAME"),
				Compact:     true,
				Interactive: false,
			},
		},
		{
			name:                      "clickhouse_cluster_replicated",
			selectQueryToRecordsCount: "select count() from raw.test",
			wantRecordsCount:          1,
			options: &Options{
				DSN:         os.Getenv("CLICKHOUSE_CLUSTER_R_DSN"),
				Replicated:  true,
				Directory:   migrationsPathAbs(os.Getenv("CLICKHOUSE_CLUSTER_R_MIGRATIONS_PATH")),
				TableName:   "migration",
				Compact:     true,
				Interactive: false,
			},
		},
	}
	// endregion

	logger := &log.NopLogger{}

	createCommand := func(arg string) *Command {
		args := NewMockArgs(t)
		args.EXPECT().First().Return(arg).Once()
		args.EXPECT().Present().Return(true).Once()

		return &Command{Args: args}
	}

	for _, tt := range tests {
		handlers := NewHandlers(tt.options, logger)
		t.Cleanup(func() {
			_ = handlers.Downgrade.Handle(createCommand("all"))
		})

		t.Run(tt.name, func(t *testing.T) {
			conn, err := connection.Try(tt.options.DSN, 1)
			require.NoError(t, err)

			repo, err := repository.New(
				conn,
				&repository.Options{
					TableName:   tt.options.TableName,
					ClusterName: tt.options.ClusterName,
					Replicated:  tt.options.Replicated,
				},
			)
			require.NoError(t, err)

			ctx := context.Background()

			err = handlers.Upgrade.Handle(createCommand("2"))
			require.NoError(t, err)

			assertEqualMigrationsCount(t, ctx, repo, 3) // basic + 2 migrations

			err = handlers.Upgrade.Handle(createCommand("1")) // migration with error
			require.Error(t, err)
			assertEqualMigrationsCount(t, ctx, repo, 3) // basic + 2 migrations

			assertEqualRecordsCount(t, ctx, repo, tt.selectQueryToRecordsCount, tt.wantRecordsCount)

			err = handlers.Downgrade.Handle(createCommand("all"))
			require.NoError(t, err)
			assertEqualMigrationsCount(t, ctx, repo, 1) // basic
		})
	}
}

func assertEqualMigrationsCount(
	t *testing.T,
	ctx context.Context,
	repo repository.Repository,
	expected int,
) {
	count, err := repo.MigrationsCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, expected, count)
}

func assertEqualRecordsCount(
	t *testing.T,
	ctx context.Context,
	repo repository.Repository,
	query string,
	wantRecordsCount int,
) {
	var gotRecordsCount int
	err := repo.QueryScalar(ctx, query, &gotRecordsCount)
	require.NoError(t, err)
	assert.Equal(t, wantRecordsCount, gotRecordsCount)
}

func migrationsPathAbs(basePath string) string {
	path, _ := filepath.Abs("../../../" + basePath)
	return path
}
