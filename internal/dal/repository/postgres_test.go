package repository

import (
	"context"
	"testing"

	"github.com/lib/pq"
	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/dal/repository/mockrepository"
	"github.com/stretchr/testify/assert"
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
	assert.NoError(t, err)
	err = repo.ExecQuery(ctx, "SELECT 1")
	assert.NoError(t, err)
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
	assert.NoError(t, err)
	err = repo.ExecQuery(ctx, "SELECT 1")
	assert.Error(t, err)
	assert.Equal(t, err, &DBError{Severity: pq.Efatal, InternalQuery: "SELECT 1"})
}
