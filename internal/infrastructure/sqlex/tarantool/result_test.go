/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package tarantool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDone_LastInsertId_ReturnsError(t *testing.T) {
	done := Done(true)

	id, err := done.LastInsertId()

	assert.Equal(t, int64(0), id)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIsNotSupportedByThisDriver)
	assert.Contains(t, err.Error(), "LastInsertId")
}

func TestDone_RowsAffected_ReturnsError(t *testing.T) {
	done := Done(true)

	rows, err := done.RowsAffected()

	assert.Equal(t, int64(0), rows)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIsNotSupportedByThisDriver)
	assert.Contains(t, err.Error(), "RowsAffected")
}

func TestDone_FalseValue_LastInsertId_ReturnsError(t *testing.T) {
	done := Done(false)

	id, err := done.LastInsertId()

	assert.Equal(t, int64(0), id)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIsNotSupportedByThisDriver)
}

func TestDone_FalseValue_RowsAffected_ReturnsError(t *testing.T) {
	done := Done(false)

	rows, err := done.RowsAffected()

	assert.Equal(t, int64(0), rows)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrIsNotSupportedByThisDriver)
}
