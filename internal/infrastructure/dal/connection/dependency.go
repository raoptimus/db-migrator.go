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

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

//go:generate mockery

// DBPinger defines the interface for database connectivity verification.
type DBPinger interface {
	Ping() error
}

//go:generate mockery

// DBQuerier defines the interface for executing database queries that return rows.
type DBQuerier interface {
	QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error)
}

//go:generate mockery

// DBExecutor defines the interface for executing database commands that modify data.
type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error)
}

//go:generate mockery

// DBTransactor defines the interface for beginning database transactions.
type DBTransactor interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (sqlex.Tx, error)
}

//go:generate mockery

// SQLDB combines all database operation interfaces into a single interface.
// It provides a complete abstraction for database interactions including queries, execution, transactions, and lifecycle management.
type SQLDB interface {
	DBPinger
	DBQuerier
	DBExecutor
	DBTransactor
	io.Closer
}
