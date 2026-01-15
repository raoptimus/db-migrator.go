/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package tarantool

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
	"github.com/tarantool/go-tarantool/v2"
)

var (
	// ErrTransactionAlreadyClosed indicates that an operation was attempted on a transaction that has already been committed or rolled back.
	ErrTransactionAlreadyClosed = errors.New("transaction already closed")
)

// tx wraps a Tarantool stream and implements the sqlex.Tx interface.
// It provides transaction operations including commit, rollback, and query execution.
type tx struct {
	stream *tarantool.Stream
	closed bool
	mu     sync.RWMutex
}

// NewTx creates a new transaction wrapper around the provided Tarantool stream.
//
//nolint:ireturn,nolintlint // its ok
func NewTx(stream *tarantool.Stream) sqlex.Tx {
	return &tx{
		stream: stream,
	}
}

// Commit commits the transaction, making all changes permanent.
// It returns ErrTransactionAlreadyClosed if the transaction has already been closed.
func (tx *tx) Commit() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.closed {
		return errors.WithStack(ErrTransactionAlreadyClosed)
	}
	_, err := tx.stream.Do(tarantool.NewCommitRequest()).Get()
	tx.closed = true

	return err
}

// Rollback aborts the transaction, discarding all changes.
// It returns ErrTransactionAlreadyClosed if the transaction has already been closed.
func (tx *tx) Rollback() error {
	tx.mu.Lock()
	defer tx.mu.Unlock()

	if tx.closed {
		return errors.WithStack(ErrTransactionAlreadyClosed)
	}

	_, err := tx.stream.Do(tarantool.NewRollbackRequest()).Get()
	tx.closed = true

	return err
}

// ExecContext executes a Lua script within the transaction without returning rows.
// It returns ErrTransactionAlreadyClosed if the transaction has already been closed.
//
//nolint:ireturn,nolintlint // its ok
func (tx *tx) ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if tx.closed {
		return nil, errors.WithStack(ErrTransactionAlreadyClosed)
	}

	req := tarantool.NewEvalRequest(query).Context(ctx)
	if len(args) > 0 {
		req = req.Args(args)
	}

	if _, err := tx.stream.Do(req).Get(); err != nil {
		return nil, err
	}

	return Done(true), nil
}

// PrepareContext creates a prepared statement for use within the transaction.
// It returns ErrTransactionAlreadyClosed if the transaction has already been closed.
//
//nolint:ireturn,nolintlint // its ok
func (tx *tx) PrepareContext(ctx context.Context, query string) (sqlex.Stmt, error) {
	tx.mu.RLock()
	defer tx.mu.RUnlock()

	if tx.closed {
		return nil, errors.WithStack(ErrTransactionAlreadyClosed)
	}

	resp, err := tx.stream.Do(tarantool.NewPrepareRequest(query).Context(ctx)).GetResponse()
	if err != nil {
		return nil, err
	}

	stmt, err := tarantool.NewPreparedFromResponse(tx.stream.Conn, resp)
	if err != nil {
		return nil, err
	}

	return &Stmt{stmt: stmt}, nil
}
