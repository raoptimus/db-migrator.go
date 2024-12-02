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

	if os.Getenv("CLICKHOUSE_CLUSTER_DSN1") == "" {
		if err := godotenv.Load("../../.env"); err != nil {
			require.NoError(t, err, "Load environments")
		}
	}

	// region data provider
	tests := []struct {
		name               string
		countMigrationsSQL string
		options            *Options
	}{
		{
			name: "postgres",
			options: &Options{
				DSN:         os.Getenv("POSTGRES_DSN"),
				Directory:   migrationsPathAbs(os.Getenv("POSTGRES_MIGRATIONS_PATH")),
				TableName:   "migration",
				Compact:     true,
				Interactive: false,
			},
		},
		{
			name: "mysql",
			options: &Options{
				DSN:         os.Getenv("MYSQL_DSN"),
				Directory:   migrationsPathAbs(os.Getenv("MYSQL_MIGRATIONS_PATH")),
				TableName:   "migration",
				Compact:     true,
				Interactive: false,
			},
		},
		{
			name: "clickhouse",
			options: &Options{
				DSN:         os.Getenv("CLICKHOUSE_DSN"),
				Directory:   migrationsPathAbs(os.Getenv("CLICKHOUSE_MIGRATIONS_PATH")),
				TableName:   "migration",
				Compact:     true,
				Interactive: false,
			},
		},
		{
			name: "clickhouse_cluster",
			options: &Options{
				DSN:         os.Getenv("CLICKHOUSE_CLUSTER_DSN1"),
				Directory:   migrationsPathAbs(os.Getenv("CLICKHOUSE_CLUSTER_MIGRATIONS_PATH")),
				TableName:   "migration",
				ClusterName: os.Getenv("CLICKHOUSE_CLUSTER_NAME"),
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
			assert.NoError(t, err)
			up, err := dbServ.Upgrade()
			assert.NoError(t, err)

			defer func() {
				_ = down.Run(cliContext(t, "all"))
			}()

			err = up.Run(cliContext(t, "2"))
			assert.NoError(t, err)
			assertEqualRowsCount(t, ctx, dbServ.repo, 3)

			err = up.Run(cliContext(t, "1")) // migration with error
			assert.Error(t, err)
			assertEqualRowsCount(t, ctx, dbServ.repo, 3)
			err = dbServ.repo.ExecQuery(ctx, "select * from test") // checks table exists
			assert.NoError(t, err)

			err = down.Run(cliContext(t, "all"))
			assert.NoError(t, err)
			assertEqualRowsCount(t, ctx, dbServ.repo, 1)
		})
	}
}

func TestIntegrationDBService_Upgrade_AlreadyExistsMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	opts := Options{
		DSN:         os.Getenv("POSTGRES_DSN"),
		Directory:   migrationsPathAbs(os.Getenv("POSTGRES_MIGRATIONS_PATH")),
		TableName:   "migration",
		Compact:     true,
		Interactive: false,
	}
	dbServ := New(&opts)

	down, err := dbServ.Downgrade()
	assert.NoError(t, err)
	err = down.Run(cliContext(t, "all"))
	assert.NoError(t, err)

	up, err := dbServ.Upgrade()
	assert.NoError(t, err)
	// apply first migration
	err = up.Run(cliContext(t, "1"))
	assert.NoError(t, err)
	// apply second migration
	err = up.Run(cliContext(t, "1"))
	assert.NoError(t, err)
	// apply third broken migration
	err = up.Run(cliContext(t, "1"))
	assert.Error(t, err)
}

func assertEqualRowsCount(
	t *testing.T,
	ctx context.Context,
	repo *repository.Repository,
	expected int,
) {
	count, err := repo.MigrationsCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expected, count)
}

func migrationsPathAbs(basePath string) string {
	path, _ := filepath.Abs("../../" + basePath)
	return path
}
