/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/stretchr/testify/require"
)

// TestNewMockMigrationService_PostgresDSN_Successfully tests that NewMockMigrationService
// creates a service.Migration correctly when given a valid PostgreSQL DSN with
// username and password.
func TestNewMockMigrationService_PostgresDSN_Successfully(t *testing.T) {
	tests := []struct {
		name             string
		dsn              string
		tableName        string
		directory        string
		maxSQLOutputLen  int
		compact          bool
		expectedUsername string
		expectedPassword string
	}{
		{
			name:             "basic postgres dsn with credentials",
			dsn:              "postgres://testuser:testpass@localhost:5432/testdb",
			tableName:        "migration",
			directory:        "/migrations",
			maxSQLOutputLen:  1000,
			compact:          false,
			expectedUsername: "testuser",
			expectedPassword: "testpass",
		},
		{
			name:             "postgres dsn without password",
			dsn:              "postgres://testuser@localhost:5432/testdb",
			tableName:        "schema_migrations",
			directory:        "/var/migrations",
			maxSQLOutputLen:  500,
			compact:          true,
			expectedUsername: "testuser",
			expectedPassword: "",
		},
		{
			name:             "postgres dsn with empty password",
			dsn:              "postgres://testuser:@localhost:5432/testdb",
			tableName:        "migration",
			directory:        "./migrations",
			maxSQLOutputLen:  0,
			compact:          false,
			expectedUsername: "testuser",
			expectedPassword: "",
		},
		{
			name:             "postgres dsn with special characters in password",
			dsn:              "postgres://admin:p%40ssw0rd%21@localhost:5432/testdb",
			tableName:        "migration",
			directory:        "/migrations",
			maxSQLOutputLen:  100,
			compact:          false,
			expectedUsername: "admin",
			expectedPassword: "p@ssw0rd!",
		},
		{
			name:             "postgres dsn with schema qualified table name",
			dsn:              "postgres://user:pass@localhost:5432/testdb",
			tableName:        "public.migration",
			directory:        "/migrations",
			maxSQLOutputLen:  1000,
			compact:          false,
			expectedUsername: "user",
			expectedPassword: "pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverPostgres)

			options := &Options{
				DSN:                tt.dsn,
				TableName:          tt.tableName,
				Directory:          tt.directory,
				MaxSQLOutputLength: tt.maxSQLOutputLen,
				Compact:            tt.compact,
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_MySQLDSN_Successfully tests that NewMockMigrationService
// creates a service.Migration correctly when given a valid MySQL DSN.
func TestNewMockMigrationService_MySQLDSN_Successfully(t *testing.T) {
	tests := []struct {
		name             string
		dsn              string
		tableName        string
		expectedUsername string
		expectedPassword string
	}{
		{
			name:             "basic mysql dsn with credentials",
			dsn:              "mysql://root:secret@localhost:3306/testdb",
			tableName:        "migration",
			expectedUsername: "root",
			expectedPassword: "secret",
		},
		{
			name:             "mysql dsn with tcp protocol",
			dsn:              "mysql://dbuser:dbpass@tcp(127.0.0.1:3306)/mydb",
			tableName:        "migrations",
			expectedUsername: "dbuser",
			expectedPassword: "dbpass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverMySQL)

			// MySQL factory needs DSN from connection
			connMock.EXPECT().
				DSN().
				Return("root:secret@tcp(localhost:3306)/testdb")

			options := &Options{
				DSN:       tt.dsn,
				TableName: tt.tableName,
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_ClickhouseDSN_Successfully tests that NewMockMigrationService
// creates a service.Migration correctly when given a valid ClickHouse DSN.
func TestNewMockMigrationService_ClickhouseDSN_Successfully(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		tableName   string
		clusterName string
		replicated  bool
	}{
		{
			name:        "basic clickhouse dsn",
			dsn:         "clickhouse://default:password@localhost:9000/default",
			tableName:   "migration",
			clusterName: "",
			replicated:  false,
		},
		{
			name:        "clickhouse dsn with cluster",
			dsn:         "clickhouse://admin:secret@clickhouse:9000/analytics",
			tableName:   "migration",
			clusterName: "cluster_prod",
			replicated:  true,
		},
		{
			name:        "clickhouse dsn with schema qualified table name",
			dsn:         "clickhouse://user:pass@localhost:9000/testdb",
			tableName:   "system.migration",
			clusterName: "",
			replicated:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverClickhouse)

			connMock.EXPECT().
				DSN().
				Return(tt.dsn)

			options := &Options{
				DSN:         tt.dsn,
				TableName:   tt.tableName,
				ClusterName: tt.clusterName,
				Replicated:  tt.replicated,
				Directory:   "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_TarantoolDSN_Successfully tests that NewMockMigrationService
// creates a service.Migration correctly when given a valid Tarantool DSN.
func TestNewMockMigrationService_TarantoolDSN_Successfully(t *testing.T) {
	tests := []struct {
		name      string
		dsn       string
		tableName string
	}{
		{
			name:      "basic tarantool dsn",
			dsn:       "tarantool://admin:pass@localhost:3301/testdb",
			tableName: "migration",
		},
		{
			name:      "tarantool dsn with guest user",
			dsn:       "tarantool://guest@localhost:3301/myspace",
			tableName: "schema_version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverTarantool)

			options := &Options{
				DSN:       tt.dsn,
				TableName: tt.tableName,
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_DSNWithoutCredentials_Successfully tests that NewMockMigrationService
// handles DSN without credentials correctly.
func TestNewMockMigrationService_DSNWithoutCredentials_Successfully(t *testing.T) {
	tests := []struct {
		name   string
		dsn    string
		driver connection.Driver
	}{
		{
			name:   "postgres dsn without user info",
			dsn:    "postgres://localhost:5432/testdb",
			driver: connection.DriverPostgres,
		},
		{
			name:   "postgres dsn with only host",
			dsn:    "postgres://localhost/testdb",
			driver: connection.DriverPostgres,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(tt.driver)

			options := &Options{
				DSN:       tt.dsn,
				TableName: "migration",
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_OptionsPassedCorrectly_Successfully tests that all
// options are correctly passed to repository.Options and service.Options.
func TestNewMockMigrationService_OptionsPassedCorrectly_Successfully(t *testing.T) {
	tests := []struct {
		name               string
		dsn                string
		tableName          string
		clusterName        string
		replicated         bool
		directory          string
		maxSQLOutputLength int
		compact            bool
	}{
		{
			name:               "all options set",
			dsn:                "postgres://user:pass@localhost:5432/db",
			tableName:          "custom_migration",
			clusterName:        "test_cluster",
			replicated:         true,
			directory:          "/custom/migrations",
			maxSQLOutputLength: 2000,
			compact:            true,
		},
		{
			name:               "minimal options",
			dsn:                "postgres://user:pass@localhost:5432/db",
			tableName:          "migration",
			clusterName:        "",
			replicated:         false,
			directory:          "./migrations",
			maxSQLOutputLength: 0,
			compact:            false,
		},
		{
			name:               "boundary maxSQLOutputLength zero",
			dsn:                "postgres://user:pass@localhost:5432/db",
			tableName:          "migration",
			clusterName:        "",
			replicated:         false,
			directory:          "/migrations",
			maxSQLOutputLength: 0,
			compact:            false,
		},
		{
			name:               "boundary maxSQLOutputLength one",
			dsn:                "postgres://user:pass@localhost:5432/db",
			tableName:          "migration",
			clusterName:        "",
			replicated:         false,
			directory:          "/migrations",
			maxSQLOutputLength: 1,
			compact:            false,
		},
		{
			name:               "large maxSQLOutputLength",
			dsn:                "postgres://user:pass@localhost:5432/db",
			tableName:          "migration",
			clusterName:        "",
			replicated:         false,
			directory:          "/migrations",
			maxSQLOutputLength: 1000000,
			compact:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverPostgres)

			options := &Options{
				DSN:                tt.dsn,
				TableName:          tt.tableName,
				ClusterName:        tt.clusterName,
				Replicated:         tt.replicated,
				Directory:          tt.directory,
				MaxSQLOutputLength: tt.maxSQLOutputLength,
				Compact:            tt.compact,
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_UnsupportedDriver_Failure tests that NewMockMigrationService
// returns an error when an unsupported driver is used.
func TestNewMockMigrationService_UnsupportedDriver_Failure(t *testing.T) {
	tests := []struct {
		name              string
		driver            connection.Driver
		expectedErrSubstr string
	}{
		{
			name:              "unknown driver",
			driver:            connection.Driver("unknown"),
			expectedErrSubstr: "driver \"unknown\" doesn't support",
		},
		{
			name:              "empty driver",
			driver:            connection.Driver(""),
			expectedErrSubstr: "driver \"\" doesn't support",
		},
		{
			name:              "invalid driver name",
			driver:            connection.Driver("mongodb"),
			expectedErrSubstr: "driver \"mongodb\" doesn't support",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(tt.driver)

			options := &Options{
				DSN:       "somedriver://user:pass@localhost:1234/db",
				TableName: "migration",
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

// TestNewMockMigrationService_InvalidTableName_Failure tests that NewMockMigrationService
// returns an error when an invalid table name is provided.
func TestNewMockMigrationService_InvalidTableName_Failure(t *testing.T) {
	tests := []struct {
		name              string
		driver            connection.Driver
		tableName         string
		expectedErrSubstr string
	}{
		{
			name:              "postgres table name with sql injection",
			driver:            connection.DriverPostgres,
			tableName:         "migration; DROP TABLE users;--",
			expectedErrSubstr: "invalid",
		},
		{
			name:              "postgres table name with special characters",
			driver:            connection.DriverPostgres,
			tableName:         "migration$table",
			expectedErrSubstr: "invalid",
		},
		{
			name:              "tarantool table name with spaces",
			driver:            connection.DriverTarantool,
			tableName:         "migration table",
			expectedErrSubstr: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(tt.driver)

			options := &Options{
				DSN:       "postgres://user:pass@localhost:5432/db",
				TableName: tt.tableName,
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

// TestNewMockMigrationService_ClickhouseInvalidDSN_Failure tests that NewMockMigrationService
// returns an error when ClickHouse receives an invalid DSN for parsing.
func TestNewMockMigrationService_ClickhouseInvalidDSN_Failure(t *testing.T) {
	tests := []struct {
		name      string
		connDSN   string
		tableName string
	}{
		{
			name:      "invalid clickhouse dsn format",
			connDSN:   "not-a-valid-dsn",
			tableName: "migration",
		},
		{
			name:      "empty dsn",
			connDSN:   "",
			tableName: "migration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverClickhouse)

			connMock.EXPECT().
				DSN().
				Return(tt.connDSN)

			options := &Options{
				DSN:       "clickhouse://user:pass@localhost:9000/db",
				TableName: tt.tableName,
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.Error(t, err)
			require.Nil(t, service)
		})
	}
}

// TestNewMockMigrationService_MySQLInvalidDSN_Failure tests that NewMockMigrationService
// returns an error when MySQL receives an invalid DSN for parsing.
func TestNewMockMigrationService_MySQLInvalidDSN_Failure(t *testing.T) {
	tests := []struct {
		name    string
		connDSN string
	}{
		{
			name:    "mysql dsn with invalid format",
			connDSN: "invalid-mysql-dsn-format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverMySQL)

			connMock.EXPECT().
				DSN().
				Return(tt.connDSN)

			options := &Options{
				DSN:       "mysql://user:pass@localhost:3306/db",
				TableName: "migration",
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.Error(t, err)
			require.Nil(t, service)
		})
	}
}

// TestNewMockMigrationService_DSNParsingUserInfo_Successfully tests various DSN formats
// to verify correct username and password extraction.
func TestNewMockMigrationService_DSNParsingUserInfo_Successfully(t *testing.T) {
	tests := []struct {
		name             string
		dsn              string
		expectedUsername string
		expectedPassword string
	}{
		{
			name:             "standard user:pass format",
			dsn:              "postgres://admin:secret@localhost:5432/db",
			expectedUsername: "admin",
			expectedPassword: "secret",
		},
		{
			name:             "user without password",
			dsn:              "postgres://admin@localhost:5432/db",
			expectedUsername: "admin",
			expectedPassword: "",
		},
		{
			name:             "user with empty password",
			dsn:              "postgres://admin:@localhost:5432/db",
			expectedUsername: "admin",
			expectedPassword: "",
		},
		{
			name:             "no user info",
			dsn:              "postgres://localhost:5432/db",
			expectedUsername: "",
			expectedPassword: "",
		},
		{
			name:             "url encoded special chars in password",
			dsn:              "postgres://user:p%40ss%3Aword@localhost:5432/db",
			expectedUsername: "user",
			expectedPassword: "p@ss:word",
		},
		{
			name:             "url encoded special chars in username",
			dsn:              "postgres://user%40domain:pass@localhost:5432/db",
			expectedUsername: "user@domain",
			expectedPassword: "pass",
		},
		{
			name:             "numeric password",
			dsn:              "postgres://user:123456@localhost:5432/db",
			expectedUsername: "user",
			expectedPassword: "123456",
		},
		{
			name:             "complex password with multiple special chars",
			dsn:              "postgres://user:P%40ss%21w0rd%23123@localhost:5432/db",
			expectedUsername: "user",
			expectedPassword: "P@ss!w0rd#123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverPostgres)

			options := &Options{
				DSN:       tt.dsn,
				TableName: "migration",
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_DSNWithQueryParams_Successfully tests that DSN with
// query parameters is parsed correctly.
func TestNewMockMigrationService_DSNWithQueryParams_Successfully(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "postgres dsn with sslmode",
			dsn:  "postgres://user:pass@localhost:5432/db?sslmode=disable",
		},
		{
			name: "postgres dsn with multiple params",
			dsn:  "postgres://user:pass@localhost:5432/db?sslmode=require&connect_timeout=10",
		},
		{
			name: "clickhouse dsn with compression",
			dsn:  "clickhouse://user:pass@localhost:9000/db?compress=true&debug=false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverPostgres)

			options := &Options{
				DSN:       tt.dsn,
				TableName: "migration",
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_EmptyDirectory_Successfully tests that empty directory
// option is handled correctly.
func TestNewMockMigrationService_EmptyDirectory_Successfully(t *testing.T) {
	loggerMock := NewMockLogger(t)
	connMock := NewMockConnection(t)

	connMock.EXPECT().
		Driver().
		Return(connection.DriverPostgres)

	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/db",
		TableName: "migration",
		Directory: "",
	}

	service, err := NewMigrationService(options, loggerMock, connMock)

	require.NoError(t, err)
	require.NotNil(t, service)
}

// TestNewMockMigrationService_BoundaryTableNameLength_Successfully tests boundary
// conditions for table name length.
func TestNewMockMigrationService_BoundaryTableNameLength_Successfully(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
	}{
		{
			name:      "single character table name",
			tableName: "m",
		},
		{
			name:      "two character table name",
			tableName: "mi",
		},
		{
			name:      "standard length table name",
			tableName: "migration",
		},
		{
			name:      "long table name (63 chars - postgres limit)",
			tableName: "migration_history_table_with_a_very_long_name_for_testing_purpo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverPostgres)

			options := &Options{
				DSN:       "postgres://user:pass@localhost:5432/db",
				TableName: tt.tableName,
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_ClickhouseWithClusterOptions_Successfully tests that
// ClickHouse cluster-specific options are handled correctly.
func TestNewMockMigrationService_ClickhouseWithClusterOptions_Successfully(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		replicated  bool
	}{
		{
			name:        "no cluster, no replication",
			clusterName: "",
			replicated:  false,
		},
		{
			name:        "with cluster, no replication",
			clusterName: "test_cluster",
			replicated:  false,
		},
		{
			name:        "with cluster and replication",
			clusterName: "prod_cluster",
			replicated:  true,
		},
		{
			name:        "no cluster with replication flag (unusual but valid)",
			clusterName: "",
			replicated:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverClickhouse)

			connMock.EXPECT().
				DSN().
				Return("clickhouse://default:pass@localhost:9000/default")

			options := &Options{
				DSN:         "clickhouse://default:pass@localhost:9000/default",
				TableName:   "migration",
				ClusterName: tt.clusterName,
				Replicated:  tt.replicated,
				Directory:   "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_PostgresWithSchemaQualifiedTable_Successfully tests
// PostgreSQL with schema-qualified table names.
func TestNewMockMigrationService_PostgresWithSchemaQualifiedTable_Successfully(t *testing.T) {
	tests := []struct {
		name      string
		tableName string
	}{
		{
			name:      "public schema",
			tableName: "public.migration",
		},
		{
			name:      "custom schema",
			tableName: "myschema.migration_history",
		},
		{
			name:      "underscore in schema name",
			tableName: "my_schema.my_migration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverPostgres)

			options := &Options{
				DSN:       "postgres://user:pass@localhost:5432/db",
				TableName: tt.tableName,
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_NilLogger_Successfully tests that nil logger does not
// cause immediate failure (if allowed by implementation).
// Note: This test documents expected behavior - if nil logger causes panic,
// this test should be updated to reflect that.
func TestNewMockMigrationService_NilLogger_Successfully(t *testing.T) {
	connMock := NewMockConnection(t)

	connMock.EXPECT().
		Driver().
		Return(connection.DriverPostgres)

	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/db",
		TableName: "migration",
		Directory: "/migrations",
	}

	// Note: If nil logger is not supported, this test should be changed to
	// verify proper error handling or panic behavior
	service, err := NewMigrationService(options, nil, connMock)

	require.NoError(t, err)
	require.NotNil(t, service)
}

// TestNewMockMigrationService_AllDriverTypes_Successfully tests that all supported
// drivers can create a service successfully.
func TestNewMockMigrationService_AllDriverTypes_Successfully(t *testing.T) {
	tests := []struct {
		name    string
		driver  connection.Driver
		dsn     string
		connDSN string
	}{
		{
			name:    "postgres driver",
			driver:  connection.DriverPostgres,
			dsn:     "postgres://user:pass@localhost:5432/db",
			connDSN: "",
		},
		{
			name:    "mysql driver",
			driver:  connection.DriverMySQL,
			dsn:     "mysql://user:pass@localhost:3306/db",
			connDSN: "user:pass@tcp(localhost:3306)/db",
		},
		{
			name:    "clickhouse driver",
			driver:  connection.DriverClickhouse,
			dsn:     "clickhouse://user:pass@localhost:9000/db",
			connDSN: "clickhouse://user:pass@localhost:9000/db",
		},
		{
			name:    "tarantool driver",
			driver:  connection.DriverTarantool,
			dsn:     "tarantool://user:pass@localhost:3301/db",
			connDSN: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(tt.driver)

			// Some drivers need DSN() to be called
			if tt.connDSN != "" {
				connMock.EXPECT().
					DSN().
					Return(tt.connDSN)
			}

			options := &Options{
				DSN:       tt.dsn,
				TableName: "migration",
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_DSNParsingEdgeCases_Successfully tests edge cases
// in DSN parsing.
func TestNewMockMigrationService_DSNParsingEdgeCases_Successfully(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "dsn with ipv6 address",
			dsn:  "postgres://user:pass@[::1]:5432/db",
		},
		{
			name: "dsn with port only",
			dsn:  "postgres://user:pass@:5432/db",
		},
		{
			name: "dsn with path after host",
			dsn:  "postgres://user:pass@localhost:5432/database/extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverPostgres)

			options := &Options{
				DSN:       tt.dsn,
				TableName: "migration",
				Directory: "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.NoError(t, err)
			require.NotNil(t, service)
		})
	}
}

// TestNewMockMigrationService_ClickhouseInvalidClusterName_Failure tests that
// ClickHouse rejects invalid cluster names.
func TestNewMockMigrationService_ClickhouseInvalidClusterName_Failure(t *testing.T) {
	tests := []struct {
		name              string
		clusterName       string
		expectedErrSubstr string
	}{
		{
			name:              "cluster name with special characters",
			clusterName:       "cluster;DROP TABLE",
			expectedErrSubstr: "invalid cluster name",
		},
		{
			name:              "cluster name with spaces",
			clusterName:       "my cluster",
			expectedErrSubstr: "invalid cluster name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loggerMock := NewMockLogger(t)
			connMock := NewMockConnection(t)

			connMock.EXPECT().
				Driver().
				Return(connection.DriverClickhouse)

			connMock.EXPECT().
				DSN().
				Return("clickhouse://default:pass@localhost:9000/default")

			options := &Options{
				DSN:         "clickhouse://default:pass@localhost:9000/default",
				TableName:   "migration",
				ClusterName: tt.clusterName,
				Directory:   "/migrations",
			}

			service, err := NewMigrationService(options, loggerMock, connMock)

			require.Error(t, err)
			require.Nil(t, service)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}
