/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptions_ZeroValue_Successfully(t *testing.T) {
	var opts Options

	require.Equal(t, "", opts.DSN)
	require.Equal(t, 0, opts.MaxConnAttempts)
	require.Equal(t, "", opts.Directory)
	require.Equal(t, "", opts.TableName)
	require.Equal(t, "", opts.ClusterName)
	require.False(t, opts.Replicated)
	require.False(t, opts.Compact)
	require.False(t, opts.Interactive)
	require.Equal(t, 0, opts.MaxSQLOutputLength)
}

func TestOptions_SetAndGetFields_Successfully(t *testing.T) {
	tests := []struct {
		name               string
		dsn                string
		maxConnAttempts    int
		directory          string
		tableName          string
		clusterName        string
		replicated         bool
		compact            bool
		interactive        bool
		maxSQLOutputLength int
	}{
		{
			name:               "all fields set with typical values",
			dsn:                "postgres://user:pass@localhost:5432/db",
			maxConnAttempts:    3,
			directory:          "./migrations",
			tableName:          "migration",
			clusterName:        "cluster1",
			replicated:         true,
			compact:            true,
			interactive:        true,
			maxSQLOutputLength: 1000,
		},
		{
			name:               "minimal configuration",
			dsn:                "mysql://root@localhost/test",
			maxConnAttempts:    1,
			directory:          "/var/migrations",
			tableName:          "schema_version",
			clusterName:        "",
			replicated:         false,
			compact:            false,
			interactive:        false,
			maxSQLOutputLength: 0,
		},
		{
			name:               "clickhouse with cluster",
			dsn:                "clickhouse://default:@localhost:9000/default",
			maxConnAttempts:    5,
			directory:          "/opt/app/migrations",
			tableName:          "db_migrations",
			clusterName:        "production_cluster",
			replicated:         true,
			compact:            false,
			interactive:        true,
			maxSQLOutputLength: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				DSN:                tt.dsn,
				MaxConnAttempts:    tt.maxConnAttempts,
				Directory:          tt.directory,
				TableName:          tt.tableName,
				ClusterName:        tt.clusterName,
				Replicated:         tt.replicated,
				Compact:            tt.compact,
				Interactive:        tt.interactive,
				MaxSQLOutputLength: tt.maxSQLOutputLength,
			}

			require.Equal(t, tt.dsn, opts.DSN)
			require.Equal(t, tt.maxConnAttempts, opts.MaxConnAttempts)
			require.Equal(t, tt.directory, opts.Directory)
			require.Equal(t, tt.tableName, opts.TableName)
			require.Equal(t, tt.clusterName, opts.ClusterName)
			require.Equal(t, tt.replicated, opts.Replicated)
			require.Equal(t, tt.compact, opts.Compact)
			require.Equal(t, tt.interactive, opts.Interactive)
			require.Equal(t, tt.maxSQLOutputLength, opts.MaxSQLOutputLength)
		})
	}
}

func TestOptions_TypicalConfigurations_Successfully(t *testing.T) {
	tests := []struct {
		name string
		opts Options
	}{
		{
			name: "postgresql production config",
			opts: Options{
				DSN:                "postgres://migrator:secret@db.example.com:5432/production?sslmode=require",
				MaxConnAttempts:    3,
				Directory:          "/app/migrations",
				TableName:          "migration",
				ClusterName:        "",
				Replicated:         false,
				Compact:            true,
				Interactive:        false,
				MaxSQLOutputLength: 200,
			},
		},
		{
			name: "clickhouse cluster config",
			opts: Options{
				DSN:                "clickhouse://admin:password@ch-node1.example.com:9000/analytics?compress=true",
				MaxConnAttempts:    5,
				Directory:          "./db/migrations",
				TableName:          "schema_migrations",
				ClusterName:        "analytics_cluster",
				Replicated:         true,
				Compact:            false,
				Interactive:        true,
				MaxSQLOutputLength: 1000,
			},
		},
		{
			name: "mysql local development config",
			opts: Options{
				DSN:                "mysql://root:root@localhost:3306/devdb",
				MaxConnAttempts:    1,
				Directory:          "./migrations",
				TableName:          "migration",
				ClusterName:        "",
				Replicated:         false,
				Compact:            false,
				Interactive:        true,
				MaxSQLOutputLength: 0,
			},
		},
		{
			name: "tarantool in-memory config",
			opts: Options{
				DSN:                "tarantool://admin:admin@localhost:3301/myspace",
				MaxConnAttempts:    2,
				Directory:          "/etc/tarantool/migrations",
				TableName:          "_migration",
				ClusterName:        "",
				Replicated:         false,
				Compact:            true,
				Interactive:        false,
				MaxSQLOutputLength: 100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEmpty(t, tt.opts.DSN)
			require.GreaterOrEqual(t, tt.opts.MaxConnAttempts, 1)
			require.NotEmpty(t, tt.opts.Directory)
			require.NotEmpty(t, tt.opts.TableName)
		})
	}
}

func TestOptions_BoundaryValues_Successfully(t *testing.T) {
	tests := []struct {
		name               string
		maxConnAttempts    int
		maxSQLOutputLength int
	}{
		{
			name:               "zero values",
			maxConnAttempts:    0,
			maxSQLOutputLength: 0,
		},
		{
			name:               "minimum positive values",
			maxConnAttempts:    1,
			maxSQLOutputLength: 1,
		},
		{
			name:               "typical maximum conn attempts",
			maxConnAttempts:    100,
			maxSQLOutputLength: 10000,
		},
		{
			name:               "large sql output length",
			maxConnAttempts:    50,
			maxSQLOutputLength: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				MaxConnAttempts:    tt.maxConnAttempts,
				MaxSQLOutputLength: tt.maxSQLOutputLength,
			}

			require.Equal(t, tt.maxConnAttempts, opts.MaxConnAttempts)
			require.Equal(t, tt.maxSQLOutputLength, opts.MaxSQLOutputLength)
		})
	}
}
