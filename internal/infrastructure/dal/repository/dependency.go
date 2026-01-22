package repository

import (
	"context"
	"io"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

//go:generate mockery

// Connection defines the interface for database connection operations used by repositories.
// It provides methods for executing queries, transactions, and managing the connection lifecycle.
type Connection interface {
	io.Closer

	// DSN returns the Data Source Name string for the connection.
	DSN() string
	// Driver returns the database driver type for this connection.
	Driver() connection.Driver
	// Ping verifies the connection to the database is alive.
	Ping() error
	// QueryContext executes a query that returns rows with the provided context and arguments.
	QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error)
	// ExecContext executes a query that doesn't return rows with the provided context and arguments.
	ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error)
	// Transaction executes a function within a database transaction.
	Transaction(ctx context.Context, txFn func(ctx context.Context) error) error
}
