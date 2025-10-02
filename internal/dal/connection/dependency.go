package connection

import (
	"context"
	"database/sql"

	"github.com/raoptimus/db-migrator.go/internal/sqlex"
)

type SQLDB interface {
	Ping() error
	QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}
