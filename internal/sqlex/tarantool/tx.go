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

	"github.com/raoptimus/db-migrator.go/internal/sqlex"
	"github.com/tarantool/go-tarantool/v2"
)

type tx struct {
	stream *tarantool.Stream
}

//nolint:ireturn,nolintlint // its ok
func NewTx(stream *tarantool.Stream) sqlex.Tx {
	return &tx{
		stream: stream,
	}
}

func (tx *tx) Commit() error {
	if _, err := tx.stream.Do(tarantool.NewCommitRequest()).Get(); err != nil {
		return err
	}

	return nil
}

func (tx *tx) Rollback() error {
	if _, err := tx.stream.Do(tarantool.NewRollbackRequest()).Get(); err != nil {
		return err
	}

	return nil
}

//nolint:ireturn,nolintlint // its ok
func (tx *tx) ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error) {
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
