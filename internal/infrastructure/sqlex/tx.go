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

// Tx abstracts database transaction operations including commit, rollback,
// query execution, and statement preparation.
type Tx interface {
	Rollback() error
	Commit() error
	ExecContext(ctx context.Context, query string, args ...any) (Result, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
}

// sqlTx wraps the standard database/sql.Tx to implement the custom Tx interface.
type sqlTx struct {
	*sql.Tx
}

// NewTx creates a new Tx wrapper around the provided sql.Tx.
//
//nolint:ireturn,nolintlint // its ok
func NewTx(tx *sql.Tx) Tx {
	return &sqlTx{Tx: tx}
}

// PrepareContext creates a prepared statement for use within the transaction.
//
//nolint:ireturn,nolintlint // its ok
func (s *sqlTx) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	prepared, err := s.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &stmt{Stmt: prepared}, nil
}

// ExecContext executes a query within the transaction without returning any rows.
//
//nolint:ireturn,nolintlint // its ok
func (s *sqlTx) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	return s.Tx.ExecContext(ctx, query, args...)
}
