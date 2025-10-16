package repository

import (
	"context"
	"testing"

	"github.com/lib/pq"
	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/dal/repository/mockrepository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgres_ExecQuery_Successfully(t *testing.T) {
	ctx := context.Background()
	conn := mockrepository.NewConnection(t)
	conn.EXPECT().
		Driver().
		Return(connection.DriverPostgres)
	conn.EXPECT().
		ExecContext(ctx, "SELECT 1").
		Return(nil, nil)

	repo, err := New(conn, &Options{})
	require.NoError(t, err)
	err = repo.ExecQuery(ctx, "SELECT 1")
	require.NoError(t, err)
}

func TestPostgres_ExecQuery_Failure(t *testing.T) {
	ctx := context.Background()
	conn := mockrepository.NewConnection(t)
	conn.EXPECT().
		Driver().
		Return(connection.DriverPostgres)
	conn.EXPECT().
		ExecContext(ctx, "SELECT 1").
		Return(nil, &pq.Error{Severity: pq.Efatal})

	repo, err := New(conn, &Options{})
	require.NoError(t, err)
	err = repo.ExecQuery(ctx, "SELECT 1")
	require.Error(t, err)
	var dbErr *DBError
	require.ErrorAs(t, err, &dbErr)
	assert.Equal(t, pq.Efatal, dbErr.Severity)
	assert.Equal(t, "SELECT 1", dbErr.InternalQuery)
}
