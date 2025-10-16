package migrator

import (
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/migrator/mockmigrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBService_Create_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{})
	action := dbServ.Create()
	assert.NotNil(t, action)
}

func TestDBService_Upgrade_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	conn := mockmigrator.NewConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)
	dbServ.conn = conn

	action, err := dbServ.Upgrade()
	require.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_Downgrade_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	conn := mockmigrator.NewConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)
	dbServ.conn = conn

	action, err := dbServ.Downgrade()
	require.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_To_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	conn := mockmigrator.NewConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)
	dbServ.conn = conn

	action, err := dbServ.To()
	require.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_History_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	conn := mockmigrator.NewConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)
	dbServ.conn = conn

	action, err := dbServ.History()
	require.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_HistoryNew_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	conn := mockmigrator.NewConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)
	dbServ.conn = conn

	action, err := dbServ.HistoryNew()
	require.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_Redo_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	conn := mockmigrator.NewConnection(t)
	conn.EXPECT().Driver().Return(connection.DriverPostgres)
	dbServ.conn = conn

	action, err := dbServ.Redo()
	require.NoError(t, err)
	assert.NotNil(t, action)
}
