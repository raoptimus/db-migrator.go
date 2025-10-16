/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package connection

import (
	"context"
	"database/sql"
	"io"

	"github.com/raoptimus/db-migrator.go/internal/sqlex"
)

//go:generate mockery
type DBPinger interface {
	Ping() error
}

//go:generate mockery
type DBQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error)
}

//go:generate mockery
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error)
}

//go:generate mockery
type DBTransactor interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (sqlex.Tx, error)
}

//go:generate mockery
type SQLDB interface {
	DBPinger
	DBQuerier
	DBExecutor
	DBTransactor
	io.Closer
}
