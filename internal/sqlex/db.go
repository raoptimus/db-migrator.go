/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package sqlex

import (
	"context"
	"database/sql"
)

type DB struct {
	*sql.DB
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
//
//nolint:ireturn,nolintlint // its ok
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return NewRowsBySQLRows(rows), nil
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
//
//nolint:ireturn,nolintlint // its ok
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}

// BeginTx starts a transaction.
//
// The provided context is used until the transaction is committed or rolled back.
// If the context is canceled, the sql package will roll back
// the transaction. [Tx.Commit] will return an error if the context provided to
// BeginTx is canceled.
//
// The provided [TxOptions] is optional and may be nil if defaults should be used.
// If a non-default isolation level is used that the driver doesn't support,
// an error will be returned.
//
//nolint:ireturn,nolintlint // its ok
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return NewTx(tx), nil
}
