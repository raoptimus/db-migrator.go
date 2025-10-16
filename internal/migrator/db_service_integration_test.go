package migrator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/raoptimus/db-migrator.go/internal/dal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationDBService_UpDown_Successfully(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	//if os.Getenv("CLICKHOUSE_CLUSTER_DSN") == "" {
	if err := godotenv.Load("../../.env"); err != nil {
		require.NoError(t, err, "Load environments")
	}
	//}

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			dbServ := New(tt.options)
			down, err := dbServ.Downgrade()
			require.NoError(t, err)
			up, err := dbServ.Upgrade()
			require.NoError(t, err)

			defer func() {
				_ = down.Run(ctx, "all")
			}()

			err = up.Run(ctx, "2")
			require.NoError(t, err)
			if err != nil {
				return
			}
			assertEqualMigrationsCount(t, ctx, dbServ.repo, 3) // basic + 2 migrations

			err = up.Run(ctx, "1") // migration with error
			assert.Error(t, err)
			assertEqualMigrationsCount(t, ctx, dbServ.repo, 3) // basic + 2 migrations

			assertEqualRecordsCount(t, ctx, dbServ.repo, tt.selectQueryToRecordsCount, tt.wantRecordsCount)

			err = down.Run(ctx, "all")
			require.NoError(t, err)
			assertEqualMigrationsCount(t, ctx, dbServ.repo, 1) // basic
		})
	}
}

func TestIntegrationDBService_Upgrade_AlreadyExistsMigration(t *testing.T) {
	//if testing.Short() {
	t.Skip("skipping integration test")
	//}
	ctx := context.Background()
	opts := Options{
		DSN:         os.Getenv("POSTGRES_DSN"),
		Directory:   migrationsPathAbs(os.Getenv("POSTGRES_MIGRATIONS_PATH")),
		TableName:   "migration",
		Compact:     true,
		Interactive: false,
	}
	dbServ := New(&opts)

	down, err := dbServ.Downgrade()
	require.NoError(t, err)
	err = down.Run(ctx, "all")
	require.NoError(t, err)

	up, err := dbServ.Upgrade()
	require.NoError(t, err)
	// apply first migration
	err = up.Run(ctx, "1")
	require.NoError(t, err)
	// apply second migration
	err = up.Run(ctx, "1")
	require.NoError(t, err)
	// apply third broken migration
	err = up.Run(ctx, "1")
	assert.Error(t, err)
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
	path, _ := filepath.Abs("../../" + basePath)
	return path
}
