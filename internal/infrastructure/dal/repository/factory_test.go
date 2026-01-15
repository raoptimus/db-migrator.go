/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFactoryRegistry(t *testing.T) {
	registry := NewFactoryRegistry()
	require.NotNil(t, registry)
	require.Len(t, registry.factories, 4)
}

func TestFactoryRegistry_Create_Postgres(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)

	registry := NewFactoryRegistry()
	repo, err := registry.Create(conn, &Options{
		TableName: "migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	_, ok := repo.(*Postgres)
	assert.True(t, ok)
}

func TestFactoryRegistry_Create_Postgres_WithSchema(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)

	registry := NewFactoryRegistry()
	repo, err := registry.Create(conn, &Options{
		TableName: "myschema.migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	pgRepo, ok := repo.(*Postgres)
	assert.True(t, ok)
	assert.Equal(t, "myschema", pgRepo.options.SchemaName)
	assert.Equal(t, "migration", pgRepo.options.TableName)
}

func TestFactoryRegistry_Create_MySQL(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverMySQL)
	conn.EXPECT().DSN().Return("user:pass@tcp(localhost:3306)/testdb")

	registry := NewFactoryRegistry()
	repo, err := registry.Create(conn, &Options{
		TableName: "migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	_, ok := repo.(*MySQL)
	assert.True(t, ok)
}

func TestFactoryRegistry_Create_Clickhouse(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverClickhouse)
	conn.EXPECT().DSN().Return("clickhouse://user:pass@localhost:9000/default")

	registry := NewFactoryRegistry()
	repo, err := registry.Create(conn, &Options{
		TableName:   "migration",
		ClusterName: "test_cluster",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	_, ok := repo.(*Clickhouse)
	assert.True(t, ok)
}

func TestFactoryRegistry_Create_Tarantool(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverTarantool)

	registry := NewFactoryRegistry()
	repo, err := registry.Create(conn, &Options{
		TableName: "migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	_, ok := repo.(*Tarantool)
	assert.True(t, ok)
}

func TestFactoryRegistry_Create_UnsupportedDriver(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.Driver("unsupported"))

	registry := NewFactoryRegistry()
	repo, err := registry.Create(conn, &Options{
		TableName: "migration",
	})

	require.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "doesn't support")
}

// PostgresFactory tests

func TestPostgresFactory_Supports(t *testing.T) {
	factory := &PostgresFactory{}

	assert.True(t, factory.Supports(connection.DriverPostgres))
	assert.False(t, factory.Supports(connection.DriverMySQL))
	assert.False(t, factory.Supports(connection.DriverClickhouse))
	assert.False(t, factory.Supports(connection.DriverTarantool))
}

func TestPostgresFactory_Create_InvalidTableName(t *testing.T) {
	factory := &PostgresFactory{}

	_, err := factory.Create(nil, &Options{
		TableName: "table;DROP TABLE users;--",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestPostgresFactory_Create_InvalidSchemaInTableName(t *testing.T) {
	factory := &PostgresFactory{}

	_, err := factory.Create(nil, &Options{
		TableName: "schema;DROP.migration",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// MySQLFactory tests

func TestMySQLFactory_Supports(t *testing.T) {
	factory := &MySQLFactory{}

	assert.True(t, factory.Supports(connection.DriverMySQL))
	assert.False(t, factory.Supports(connection.DriverPostgres))
	assert.False(t, factory.Supports(connection.DriverClickhouse))
	assert.False(t, factory.Supports(connection.DriverTarantool))
}

func TestMySQLFactory_Create_InvalidDSN(t *testing.T) {
	factory := &MySQLFactory{}
	conn := NewMockConnection(t)
	conn.EXPECT().DSN().Return("invalid_dsn")

	_, err := factory.Create(conn, &Options{
		TableName: "migration",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing dsn")
}

func TestMySQLFactory_Create_InvalidTableName(t *testing.T) {
	factory := &MySQLFactory{}
	conn := NewMockConnection(t)
	conn.EXPECT().DSN().Return("user:pass@tcp(localhost:3306)/testdb")

	_, err := factory.Create(conn, &Options{
		TableName: "table;DROP TABLE users;--",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestMySQLFactory_Create_InvalidDatabaseName(t *testing.T) {
	factory := &MySQLFactory{}
	conn := NewMockConnection(t)
	// DSN with invalid database name containing special chars
	conn.EXPECT().DSN().Return("user:pass@tcp(localhost:3306)/test;db")

	_, err := factory.Create(conn, &Options{
		TableName: "migration",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid database name")
}

// ClickhouseFactory tests

func TestClickhouseFactory_Supports(t *testing.T) {
	factory := &ClickhouseFactory{}

	assert.True(t, factory.Supports(connection.DriverClickhouse))
	assert.False(t, factory.Supports(connection.DriverPostgres))
	assert.False(t, factory.Supports(connection.DriverMySQL))
	assert.False(t, factory.Supports(connection.DriverTarantool))
}

func TestClickhouseFactory_Create_InvalidDSN(t *testing.T) {
	factory := &ClickhouseFactory{}
	conn := NewMockConnection(t)
	conn.EXPECT().DSN().Return("invalid_dsn")

	_, err := factory.Create(conn, &Options{
		TableName:   "migration",
		ClusterName: "test_cluster",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing dsn")
}

func TestClickhouseFactory_Create_InvalidTableName(t *testing.T) {
	factory := &ClickhouseFactory{}
	conn := NewMockConnection(t)
	conn.EXPECT().DSN().Return("clickhouse://user:pass@localhost:9000/default")

	_, err := factory.Create(conn, &Options{
		TableName:   "table;DROP TABLE users;--",
		ClusterName: "test_cluster",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestClickhouseFactory_Create_InvalidClusterName(t *testing.T) {
	factory := &ClickhouseFactory{}
	conn := NewMockConnection(t)
	conn.EXPECT().DSN().Return("clickhouse://user:pass@localhost:9000/default")

	_, err := factory.Create(conn, &Options{
		TableName:   "migration",
		ClusterName: "cluster;DROP DATABASE;--",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cluster name")
}

// TarantoolFactory tests

func TestTarantoolFactory_Supports(t *testing.T) {
	factory := &TarantoolFactory{}

	assert.True(t, factory.Supports(connection.DriverTarantool))
	assert.False(t, factory.Supports(connection.DriverPostgres))
	assert.False(t, factory.Supports(connection.DriverMySQL))
	assert.False(t, factory.Supports(connection.DriverClickhouse))
}

func TestTarantoolFactory_Create_Success(t *testing.T) {
	factory := &TarantoolFactory{}

	repo, err := factory.Create(nil, &Options{
		TableName: "migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	tRepo, ok := repo.(*Tarantool)
	assert.True(t, ok)
	assert.Equal(t, "migration", tRepo.options.TableName)
	assert.Equal(t, "", tRepo.options.SchemaName)
}

func TestTarantoolFactory_Create_InvalidTableName(t *testing.T) {
	factory := &TarantoolFactory{}

	_, err := factory.Create(nil, &Options{
		TableName: "table;DROP SPACE users;--",
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid table name")
}

// parseAndValidateTableName tests

func TestParseAndValidateTableName(t *testing.T) {
	tests := []struct {
		name          string
		tableName     string
		wantTableName string
		wantErr       bool
	}{
		{
			name:          "simple table name",
			tableName:     "migration",
			wantTableName: "migration",
			wantErr:       false,
		},
		{
			name:          "table with underscore",
			tableName:     "my_table",
			wantTableName: "my_table",
			wantErr:       false,
		},
		{
			name:          "schema.table format",
			tableName:     "myschema.migration",
			wantTableName: "migration",
			wantErr:       false,
		},
		{
			name:          "invalid simple table name",
			tableName:     "table;DROP",
			wantTableName: "",
			wantErr:       true,
		},
		{
			name:          "invalid schema in schema.table",
			tableName:     "schema;DROP.migration",
			wantTableName: "",
			wantErr:       true,
		},
		{
			name:          "invalid table in schema.table",
			tableName:     "schema.table;DROP",
			wantTableName: "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAndValidateTableName(tt.tableName)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantTableName, got)
			}
		})
	}
}

// parseAndValidateSchemaTableName tests

func TestParseAndValidateSchemaTableName(t *testing.T) {
	tests := []struct {
		name          string
		tableName     string
		defaultSchema string
		wantSchema    string
		wantTable     string
		wantErr       bool
	}{
		{
			name:          "simple table name uses default schema",
			tableName:     "migration",
			defaultSchema: "public",
			wantSchema:    "public",
			wantTable:     "migration",
			wantErr:       false,
		},
		{
			name:          "schema.table format",
			tableName:     "myschema.migration",
			defaultSchema: "public",
			wantSchema:    "myschema",
			wantTable:     "migration",
			wantErr:       false,
		},
		{
			name:          "invalid simple table name",
			tableName:     "table;DROP",
			defaultSchema: "public",
			wantErr:       true,
		},
		{
			name:          "invalid schema in schema.table",
			tableName:     "schema;DROP.migration",
			defaultSchema: "public",
			wantErr:       true,
		},
		{
			name:          "invalid table in schema.table",
			tableName:     "schema.table;DROP",
			defaultSchema: "public",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, table, err := parseAndValidateSchemaTableName(tt.tableName, tt.defaultSchema)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantSchema, schema)
				assert.Equal(t, tt.wantTable, table)
			}
		})
	}
}
