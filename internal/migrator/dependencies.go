package migrator

import (
	"context"
	"database/sql"

	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
)

//go:generate mockery
type Connection interface {
	DSN() string
	Driver() connection.Driver
	Ping() error
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Transaction(ctx context.Context, txFn func(ctx context.Context) error) error
}

//go:generate mockery
type FileNameBuilder interface {
	// Up builds a file name for migration update
	Up(version string, forceSafely bool) (fname string, safely bool)
	// Down builds a file name for migration downgrade
	Down(version string, forceSafely bool) (fname string, safely bool)
}
