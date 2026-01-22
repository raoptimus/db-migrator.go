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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// nilTxOpts is used for BeginTx mock expectations
var nilTxOpts *sql.TxOptions

func TestDriver_String(t *testing.T) {
	tests := []struct {
		name   string
		driver Driver
		want   string
	}{
		{
			name:   "clickhouse driver",
			driver: DriverClickhouse,
			want:   "clickhouse",
		},
		{
			name:   "mysql driver",
			driver: DriverMySQL,
			want:   "mysql",
		},
		{
			name:   "postgres driver",
			driver: DriverPostgres,
			want:   "postgres",
		},
		{
			name:   "tarantool driver",
			driver: DriverTarantool,
			want:   "tarantool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.driver.String())
		})
	}
}

func TestNew_InvalidDSN_ReturnsError(t *testing.T) {
	tests := []struct {
		name       string
		dsn        string
		wantErrMsg string
	}{
		{
			name:       "empty dsn",
			dsn:        "",
			wantErrMsg: "parsing DSN",
		},
		{
			name:       "invalid dsn format",
			dsn:        "not-a-valid-dsn",
			wantErrMsg: "parsing DSN",
		},
		{
			name:       "unsupported driver",
			dsn:        "oracle://user:pass@localhost:1521/db",
			wantErrMsg: "doesn't support",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := New(tt.dsn)
			require.Error(t, err)
			assert.Nil(t, conn)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestConnection_DSN(t *testing.T) {
	dsn := "postgres://user:pass@localhost:5432/db"
	conn := &Connection{
		driver: DriverPostgres,
		dsn:    dsn,
	}

	assert.Equal(t, dsn, conn.DSN())
}

func TestConnection_Driver(t *testing.T) {
	conn := &Connection{
		driver: DriverMySQL,
		dsn:    "mysql://user:pass@localhost:3306/db",
	}

	assert.Equal(t, DriverMySQL, conn.Driver())
}

func TestConnection_Ping_Success(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockDB.EXPECT().Ping().Return(nil).Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
		ping:   false,
	}

	err := conn.Ping()
	require.NoError(t, err)
	assert.True(t, conn.ping)
}

func TestConnection_Ping_AlreadyPinged(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	// Ping should not be called because connection is already pinged

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
		ping:   true,
	}

	err := conn.Ping()
	require.NoError(t, err)
}

func TestConnection_Ping_Error(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockDB.EXPECT().Ping().Return(errors.New("connection refused")).Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
		ping:   false,
	}

	err := conn.Ping()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ping")
	assert.False(t, conn.ping)
}

func TestConnection_Close(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockDB.EXPECT().Close().Return(nil).Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	err := conn.Close()
	require.NoError(t, err)
}

func TestConnection_QueryContext(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockRows := NewMockRows(t)

	mockDB.EXPECT().
		QueryContext(mock.Anything, "SELECT * FROM users", mock.Anything).
		Return(mockRows, nil).
		Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	ctx := context.Background()
	rows, err := conn.QueryContext(ctx, "SELECT * FROM users")
	require.NoError(t, err)
	assert.NotNil(t, rows)
}

func TestConnection_ExecContext_WithoutTransaction(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockResult := NewMockResult(t)

	mockDB.EXPECT().
		ExecContext(mock.Anything, "INSERT INTO users (name) VALUES (?)", "John").
		Return(mockResult, nil).
		Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	ctx := context.Background()
	result, err := conn.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "John")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestConnection_ExecContext_WithTransaction(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockTx := NewMockTx(t)
	mockResult := NewMockResult(t)

	// Expect ExecContext to be called on transaction, not on db
	mockTx.EXPECT().
		ExecContext(mock.Anything, "INSERT INTO users (name) VALUES (?)", "John").
		Return(mockResult, nil).
		Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	ctx := ContextWithTx(context.Background(), mockTx)
	result, err := conn.ExecContext(ctx, "INSERT INTO users (name) VALUES (?)", "John")
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestConnection_Transaction_Success(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockTx := NewMockTx(t)

	mockDB.EXPECT().BeginTx(mock.Anything, nilTxOpts).Return(mockTx, nil).Once()
	mockTx.EXPECT().Commit().Return(nil).Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	ctx := context.Background()
	err := conn.Transaction(ctx, func(ctx context.Context) error {
		// Transaction body - verify we can get tx from context
		tx, txErr := TxFromContext(ctx)
		require.NoError(t, txErr)
		require.NotNil(t, tx)
		return nil
	})
	require.NoError(t, err)
}

func TestConnection_Transaction_RollbackOnError(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockTx := NewMockTx(t)

	mockDB.EXPECT().BeginTx(mock.Anything, nilTxOpts).Return(mockTx, nil).Once()
	mockTx.EXPECT().Rollback().Return(nil).Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	ctx := context.Background()
	txErr := errors.New("some error in transaction")
	err := conn.Transaction(ctx, func(ctx context.Context) error {
		return txErr
	})
	require.Error(t, err)
	assert.Equal(t, txErr, err)
}

func TestConnection_Transaction_NestedTransactionError(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockTx := NewMockTx(t)

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	// Create context with existing transaction
	ctx := ContextWithTx(context.Background(), mockTx)

	err := conn.Transaction(ctx, func(ctx context.Context) error {
		return nil
	})
	require.Error(t, err)
	assert.Equal(t, ErrTransactionAlreadyOpened, err)
}

func TestConnection_Transaction_BeginTxError(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	beginErr := errors.New("begin failed")

	mockDB.EXPECT().BeginTx(mock.Anything, nilTxOpts).Return(nil, beginErr).Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	ctx := context.Background()
	err := conn.Transaction(ctx, func(ctx context.Context) error {
		return nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction")
}

func TestConnection_Transaction_CommitError(t *testing.T) {
	mockDB := NewMockSQLDB(t)
	mockTx := NewMockTx(t)

	mockDB.EXPECT().BeginTx(mock.Anything, nilTxOpts).Return(mockTx, nil).Once()
	mockTx.EXPECT().Commit().Return(errors.New("commit failed")).Once()
	mockTx.EXPECT().Rollback().Return(nil).Once()

	conn := &Connection{
		driver: DriverPostgres,
		dsn:    "postgres://user:pass@localhost:5432/db",
		db:     mockDB,
	}

	ctx := context.Background()
	err := conn.Transaction(ctx, func(ctx context.Context) error {
		return nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "commit transaction")
}

func TestContextWithTx_And_TxFromContext(t *testing.T) {
	mockTx := NewMockTx(t)
	ctx := context.Background()

	// Test storing and retrieving tx
	ctxWithTx := ContextWithTx(ctx, mockTx)
	retrievedTx, err := TxFromContext(ctxWithTx)
	require.NoError(t, err)
	assert.Equal(t, mockTx, retrievedTx)
}

func TestTxFromContext_NoTransaction(t *testing.T) {
	ctx := context.Background()

	tx, err := TxFromContext(ctx)
	require.Error(t, err)
	assert.Nil(t, tx)
	assert.Equal(t, ErrNoTransaction, err)
}

func TestSentinelErrors(t *testing.T) {
	assert.NotEqual(t, ErrTransactionAlreadyOpened, ErrNoTransaction)
	assert.Equal(t, "transaction already opened", ErrTransactionAlreadyOpened.Error())
	assert.Equal(t, "no transaction", ErrNoTransaction.Error())
}
