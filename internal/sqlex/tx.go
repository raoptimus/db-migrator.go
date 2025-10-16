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

type Tx interface {
	Rollback() error
	Commit() error
	ExecContext(ctx context.Context, query string, args ...any) (Result, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
}

type sqlTx struct {
	*sql.Tx
}

//nolint:ireturn,nolintlint // its ok
func NewTx(tx *sql.Tx) Tx {
	return &sqlTx{Tx: tx}
}

//nolint:ireturn,nolintlint // its ok
func (s *sqlTx) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	prepared, err := s.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	return &stmt{Stmt: prepared}, nil
}

//nolint:ireturn,nolintlint // its ok
func (s *sqlTx) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	return s.Tx.ExecContext(ctx, query, args...)
}
