/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package tarantool

import (
	"bytes"
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/tarantool/go-tarantool/v2"
)

// Test sentinel errors for tx tests
var (
	errCommitFailed   = errors.New("commit failed")
	errRollbackFailed = errors.New("rollback failed")
	errExecFailed     = errors.New("exec failed")
	errPrepareFailed  = errors.New("prepare failed")
)

func TestTx_Commit_Successfully(t *testing.T) {
	stream := NewMockStreamDoer(t)

	commitReq := tarantool.NewCommitRequest()
	fut := tarantool.NewFuture(commitReq)
	fut.SetResponse(tarantool.Header{}, bytes.NewBuffer([]byte("")))

	stream.EXPECT().
		Do(mock.AnythingOfType("*tarantool.CommitRequest")).
		Return(fut)

	tx := newTxWithStreamDoer(stream)

	err := tx.Commit()

	require.NoError(t, err)
	require.True(t, tx.closed)
}

func TestTx_Commit_StreamError_Failure(t *testing.T) {
	stream := NewMockStreamDoer(t)

	commitReq := tarantool.NewCommitRequest()
	fut := tarantool.NewFuture(commitReq)
	fut.SetError(errCommitFailed)

	stream.EXPECT().
		Do(mock.AnythingOfType("*tarantool.CommitRequest")).
		Return(fut)

	tx := newTxWithStreamDoer(stream)

	err := tx.Commit()

	require.Error(t, err)
	require.ErrorIs(t, err, errCommitFailed)
	require.True(t, tx.closed) // closed is set to true even on error
}

func TestTx_Commit_AlreadyClosed_Failure(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: true,
	}

	err := tx.Commit()

	require.Error(t, err)
	require.ErrorIs(t, err, ErrTransactionAlreadyClosed)
}

func TestTx_Rollback_Successfully(t *testing.T) {
	stream := NewMockStreamDoer(t)

	rollbackReq := tarantool.NewRollbackRequest()
	fut := tarantool.NewFuture(rollbackReq)
	fut.SetResponse(tarantool.Header{}, bytes.NewBuffer([]byte("")))

	stream.EXPECT().
		Do(mock.AnythingOfType("*tarantool.RollbackRequest")).
		Return(fut)

	tx := newTxWithStreamDoer(stream)

	err := tx.Rollback()

	require.NoError(t, err)
	require.True(t, tx.closed)
}

func TestTx_Rollback_StreamError_Failure(t *testing.T) {
	stream := NewMockStreamDoer(t)

	rollbackReq := tarantool.NewRollbackRequest()
	fut := tarantool.NewFuture(rollbackReq)
	fut.SetError(errRollbackFailed)

	stream.EXPECT().
		Do(mock.AnythingOfType("*tarantool.RollbackRequest")).
		Return(fut)

	tx := newTxWithStreamDoer(stream)

	err := tx.Rollback()

	require.Error(t, err)
	require.ErrorIs(t, err, errRollbackFailed)
	require.True(t, tx.closed) // closed is set to true even on error
}

func TestTx_Rollback_AlreadyClosed_Failure(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: true,
	}

	err := tx.Rollback()

	require.Error(t, err)
	require.ErrorIs(t, err, ErrTransactionAlreadyClosed)
}

func TestTx_ExecContext_Successfully(t *testing.T) {
	tests := []struct {
		name  string
		query string
		args  []any
	}{
		{
			name:  "exec without arguments",
			query: "box.schema.space.create('test')",
			args:  nil,
		},
		{
			name:  "exec with single argument",
			query: "box.space.test:insert({...})",
			args:  []any{"value1"},
		},
		{
			name:  "exec with multiple arguments",
			query: "box.space.test:insert({...})",
			args:  []any{"key1", int64(123), "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			stream := NewMockStreamDoer(t)

			req := tarantool.NewEvalRequest(tt.query).Context(ctx)
			if len(tt.args) > 0 {
				req = req.Args(tt.args)
			}
			fut := tarantool.NewFuture(req)
			fut.SetResponse(tarantool.Header{}, bytes.NewBuffer([]byte("")))

			stream.EXPECT().
				Do(req).
				Return(fut)

			tx := newTxWithStreamDoer(stream)

			result, err := tx.ExecContext(ctx, tt.query, tt.args...)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, Done(true), result)
		})
	}
}

func TestTx_ExecContext_StreamError_Failure(t *testing.T) {
	ctx := context.Background()
	stream := NewMockStreamDoer(t)

	query := "box.schema.space.create('test')"
	req := tarantool.NewEvalRequest(query).Context(ctx)
	fut := tarantool.NewFuture(req)
	fut.SetError(errExecFailed)

	stream.EXPECT().
		Do(req).
		Return(fut)

	tx := newTxWithStreamDoer(stream)

	result, err := tx.ExecContext(ctx, query)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, errExecFailed)
}

func TestTx_ExecContext_AlreadyClosed_Failure(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: true,
	}

	result, err := tx.ExecContext(context.Background(), "return true")

	require.Nil(t, result)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrTransactionAlreadyClosed)
}

func TestTx_PrepareContext_AlreadyClosed_Failure(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: true,
	}

	stmt, err := tx.PrepareContext(context.Background(), "SELECT * FROM migration")

	require.Nil(t, stmt)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrTransactionAlreadyClosed)
}

func TestTx_PrepareContext_DoError_Failure(t *testing.T) {
	ctx := context.Background()
	stream := NewMockStreamDoer(t)

	query := "SELECT * FROM migration"
	req := tarantool.NewPrepareRequest(query).Context(ctx)
	fut := tarantool.NewFuture(req)
	fut.SetError(errPrepareFailed)

	stream.EXPECT().
		Do(mock.AnythingOfType("*tarantool.PrepareRequest")).
		Return(fut)

	tx := newTxWithStreamDoer(stream)

	stmt, err := tx.PrepareContext(ctx, query)

	require.Nil(t, stmt)
	require.Error(t, err)
	require.ErrorIs(t, err, errPrepareFailed)
}

func TestNewTx_ReturnsNotNil(t *testing.T) {
	tx := NewTx(nil)

	require.NotNil(t, tx)
}

func TestNewTxWithStreamDoer_ReturnsNotNil(t *testing.T) {
	stream := NewMockStreamDoer(t)
	tx := newTxWithStreamDoer(stream)

	require.NotNil(t, tx)
	require.False(t, tx.closed)
}

func TestStreamWrapper_ImplementsStreamDoer(t *testing.T) {
	// Verify that streamWrapper correctly implements the StreamDoer interface
	var _ StreamDoer = &streamWrapper{Stream: nil}
}

func TestStreamWrapper_Conn_ReturnsStreamConnection(t *testing.T) {
	// Create a real tarantool.Stream with nil Conn to test Conn() method
	stream := &tarantool.Stream{
		Id:   123,
		Conn: nil, // nil connection is acceptable for this test
	}
	wrapper := &streamWrapper{Stream: stream}

	conn := wrapper.Conn()

	// The connection should be nil as we set it to nil
	require.Nil(t, conn)
}
