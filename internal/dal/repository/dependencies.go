package repository

import (
	"context"
	"io"

	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/sqlex"
)

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
