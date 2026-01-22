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

func TestNew_Postgres_Successfully(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)

	repo, err := New(conn, &Options{
		TableName: "migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	_, ok := repo.(*Postgres)
	assert.True(t, ok)
}

func TestNew_MySQL_Successfully(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverMySQL)
	conn.EXPECT().DSN().Return("user:pass@tcp(localhost:3306)/testdb")

	repo, err := New(conn, &Options{
		TableName: "migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	_, ok := repo.(*MySQL)
	assert.True(t, ok)
}

func TestNew_Clickhouse_Successfully(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverClickhouse)
	conn.EXPECT().DSN().Return("clickhouse://user:pass@localhost:9000/default")

	repo, err := New(conn, &Options{
		TableName: "migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	_, ok := repo.(*Clickhouse)
	assert.True(t, ok)
}

func TestNew_Tarantool_Successfully(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverTarantool)

	repo, err := New(conn, &Options{
		TableName: "migration",
	})

	require.NoError(t, err)
	require.NotNil(t, repo)
	_, ok := repo.(*Tarantool)
	assert.True(t, ok)
}

func TestNew_UnsupportedDriver_Failure(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.Driver("unsupported"))

	repo, err := New(conn, &Options{
		TableName: "migration",
	})

	require.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "doesn't support")
}

func TestNew_InvalidTableName_Failure(t *testing.T) {
	conn := NewMockConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)

	repo, err := New(conn, &Options{
		TableName: "table;DROP TABLE users;--",
	})

	require.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "invalid")
}
