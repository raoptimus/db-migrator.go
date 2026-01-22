package dbmigrator

import (
	"context"
	"io"

	"github.com/raoptimus/db-migrator.go/internal/domain/log"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

// Logger defines the interface for logging operations in the migration system.
// It provides methods for different log levels including info, success, warning, error, and fatal.
//
//go:generate mockery
type Logger = log.Logger

// Connection defines the interface for database connection operations.
// It provides methods for querying, executing commands, and managing transactions.
//
//go:generate mockery
type Connection interface {
	io.Closer

	DSN() string
	Driver() connection.Driver
	Ping() error
	QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error)
	Transaction(ctx context.Context, txFn func(ctx context.Context) error) error
}
