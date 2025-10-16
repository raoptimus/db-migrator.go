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
	"github.com/raoptimus/db-migrator.go/internal/sqlex"
	"github.com/tarantool/go-tarantool/v2"
)

var (
	ErrTransactionAlreadyClosed = errors.New("transaction already closed")
)

type tx struct {
	stream *tarantool.Stream
	closed bool
	mu     sync.RWMutex
}

//nolint:ireturn,nolintlint // its ok
func NewTx(stream *tarantool.Stream) sqlex.Tx {
	return &tx{
		stream: stream,
	}
}

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
