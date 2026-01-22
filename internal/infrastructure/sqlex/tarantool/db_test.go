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

// Test sentinel errors
var (
	errConnectionFailed = errors.New("connection failed")
	errQueryFailed      = errors.New("query execution failed")
	errStreamFailed     = errors.New("stream creation failed")
	errCloseFailed      = errors.New("close connection failed")
)

func TestDB_Ping_Successfully(t *testing.T) {
	conn := NewMockConnection(t)

	pingReq := tarantool.NewPingRequest()
	fut := tarantool.NewFuture(pingReq)
	fut.SetResponse(tarantool.Header{}, bytes.NewBuffer([]byte("")))

	conn.EXPECT().
		Do(mock.AnythingOfType("*tarantool.PingRequest")).
		Return(fut)

	db := &DB{conn: conn}

	err := db.Ping()

	require.NoError(t, err)
}

func TestDB_Ping_ConnectionError_Failure(t *testing.T) {
	conn := NewMockConnection(t)

	pingReq := tarantool.NewPingRequest()
	fut := tarantool.NewFuture(pingReq)
	fut.SetError(errConnectionFailed)

	conn.EXPECT().
		Do(mock.AnythingOfType("*tarantool.PingRequest")).
		Return(fut)

	db := &DB{conn: conn}

	err := db.Ping()

	require.Error(t, err)
	require.ErrorIs(t, err, errConnectionFailed)
}

func TestDB_QueryContext_SimpleQuery_Successfully(t *testing.T) {
	tests := []struct {
		name  string
		query string
		args  []any
	}{
		{
			name:  "query without arguments",
			query: "return box.space.test:select()",
			args:  nil,
		},
		{
			name:  "query with single argument",
			query: "return box.space.test:select({...})",
			args:  []any{"key1"},
		},
		{
			name:  "query with multiple arguments",
			query: "return box.space.test:select({...})",
			args:  []any{"key1", int64(123), "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			conn := NewMockConnection(t)

			req := tarantool.NewEvalRequest(tt.query).Context(ctx)
			if len(tt.args) > 0 {
				req = req.Args(tt.args)
			}
			fut := tarantool.NewFuture(req)
			fut.SetResponse(tarantool.Header{}, bytes.NewBuffer([]byte("")))

			conn.EXPECT().
				Do(req).
				Return(fut)

			db := &DB{conn: conn}

			rows, err := db.QueryContext(ctx, tt.query, tt.args...)

			require.NoError(t, err)
			require.NotNil(t, rows)
		})
	}
}

func TestDB_QueryContext_ConnectionError_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)

	query := "return box.space.test:select()"
	req := tarantool.NewEvalRequest(query).Context(ctx)
	fut := tarantool.NewFuture(req)
	fut.SetError(errQueryFailed)

	conn.EXPECT().
		Do(req).
		Return(fut)

	db := &DB{conn: conn}

	rows, err := db.QueryContext(ctx, query)

	require.Error(t, err)
	require.Nil(t, rows)
	require.ErrorIs(t, err, errQueryFailed)
}

func TestDB_ExecContext_SimpleExec_Successfully(t *testing.T) {
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
			conn := NewMockConnection(t)

			req := tarantool.NewEvalRequest(tt.query).Context(ctx)
			if len(tt.args) > 0 {
				req = req.Args(tt.args)
			}
			fut := tarantool.NewFuture(req)
			fut.SetResponse(tarantool.Header{}, bytes.NewBuffer([]byte("")))

			conn.EXPECT().
				Do(req).
				Return(fut)

			db := &DB{conn: conn}

			result, err := db.ExecContext(ctx, tt.query, tt.args...)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, Done(true), result)
		})
	}
}

func TestDB_ExecContext_ConnectionError_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)

	query := "box.schema.space.create('test')"
	req := tarantool.NewEvalRequest(query).Context(ctx)
	fut := tarantool.NewFuture(req)
	fut.SetError(errQueryFailed)

	conn.EXPECT().
		Do(req).
		Return(fut)

	db := &DB{conn: conn}

	result, err := db.ExecContext(ctx, query)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, errQueryFailed)
}

func TestDB_ExecContext_WithArgsConnectionError_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)

	query := "box.space.test:insert({...})"
	args := []any{"key1", int64(123)}
	req := tarantool.NewEvalRequest(query).Context(ctx).Args(args)
	fut := tarantool.NewFuture(req)
	fut.SetError(errQueryFailed)

	conn.EXPECT().
		Do(req).
		Return(fut)

	db := &DB{conn: conn}

	result, err := db.ExecContext(ctx, query, args...)

	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, err, errQueryFailed)
}

func TestDB_BeginTx_NewStreamError_Failure(t *testing.T) {
	ctx := context.Background()
	conn := NewMockConnection(t)

	conn.EXPECT().
		NewStream().
		Return(nil, errStreamFailed)

	db := &DB{conn: conn}

	tx, err := db.BeginTx(ctx, nil)

	require.Error(t, err)
	require.Nil(t, tx)
	require.ErrorIs(t, err, errStreamFailed)
}

func TestDB_Close_WithConnection_Successfully(t *testing.T) {
	conn := NewMockConnection(t)

	conn.EXPECT().
		Close().
		Return(nil)

	db := &DB{conn: conn}

	err := db.Close()

	require.NoError(t, err)
}

func TestDB_Close_WithConnectionError_Failure(t *testing.T) {
	conn := NewMockConnection(t)

	conn.EXPECT().
		Close().
		Return(errCloseFailed)

	db := &DB{conn: conn}

	err := db.Close()

	require.Error(t, err)
	require.ErrorIs(t, err, errCloseFailed)
}

func TestDB_Close_WithNilConnection_Successfully(t *testing.T) {
	db := &DB{conn: nil}

	err := db.Close()

	require.NoError(t, err)
}

func TestDB_Open_InvalidDSN_Failure(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{
			name: "empty DSN",
			dsn:  "",
		},
		{
			name: "invalid scheme",
			dsn:  "://invalid",
		},
		{
			name: "missing host",
			dsn:  "tarantool://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Open(tt.dsn)

			require.Error(t, err)
			require.Nil(t, db)
		})
	}
}

func TestDB_Open_ConnectionFailed_Failure(t *testing.T) {
	// Valid DSN format but unreachable host (localhost with unlikely port)
	dsn := "tarantool://user:pass@127.0.0.1:59999"

	db, err := Open(dsn)

	require.Error(t, err)
	require.Nil(t, db)
	require.Contains(t, err.Error(), "connect to tarantool")
}
