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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTx_Commit_WhenAlreadyClosed_ReturnsError(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: true,
	}

	err := tx.Commit()

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionAlreadyClosed)
}

func TestTx_Rollback_WhenAlreadyClosed_ReturnsError(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: true,
	}

	err := tx.Rollback()

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionAlreadyClosed)
}

func TestTx_ExecContext_WhenAlreadyClosed_ReturnsError(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: true,
	}

	result, err := tx.ExecContext(context.Background(), "return true")

	assert.Nil(t, result)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionAlreadyClosed)
}

func TestTx_PrepareContext_WhenAlreadyClosed_ReturnsError(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: true,
	}

	stmt, err := tx.PrepareContext(context.Background(), "SELECT * FROM migration")

	assert.Nil(t, stmt)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrTransactionAlreadyClosed)
}

func TestNewTx_ReturnsNotNil(t *testing.T) {
	tx := NewTx(nil)

	require.NotNil(t, tx)
}

func TestTx_InitialState_NotClosed(t *testing.T) {
	tx := &tx{
		stream: nil,
		closed: false,
	}

	assert.False(t, tx.closed)
}
