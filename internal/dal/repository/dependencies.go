package repository

import (
	"context"
	"database/sql"

	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
)

//go:generate mockery --name=Connection --outpkg=mockrepository --output=./mockrepository
type Connection interface {
	DSN() string
	Driver() connection.Driver
	Ping() error
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Transaction(ctx context.Context, txFn func(ctx context.Context) error) error
}
