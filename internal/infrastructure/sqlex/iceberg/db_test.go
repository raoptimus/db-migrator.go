/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package iceberg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen_InvalidDSN_ReturnsError(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{name: "empty dsn", dsn: ""},
		{name: "missing driver prefix", dsn: "localhost:8181/warehouse"},
		{name: "no hosts specified", dsn: "iceberg://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := Open(tt.dsn)
			require.Error(t, err)
			assert.Nil(t, db)
		})
	}
}

// SQL-level methods are intentionally unsupported by the Iceberg sqlex adapter;
// DDL goes through the repository layer. They must return ErrNotSupported without
// touching the catalog client, so a zero-value DB is sufficient.
func TestDB_SQLLevelMethods_ReturnErrNotSupported(t *testing.T) {
	ctx := t.Context()
	db := &DB{}

	rows, err := db.QueryContext(ctx, "SELECT 1")
	require.ErrorIs(t, err, ErrNotSupported)
	assert.Nil(t, rows)

	res, err := db.ExecContext(ctx, "SELECT 1")
	require.ErrorIs(t, err, ErrNotSupported)
	assert.Nil(t, res)

	tx, err := db.BeginTx(ctx, nil)
	require.ErrorIs(t, err, ErrNotSupported)
	assert.Nil(t, tx)
}
