/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dbmigrator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnection_InvalidDSN_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		wantErrMsg  string
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
			dsn:        "unsupported://user:pass@localhost:5432/db",
			wantErrMsg: "doesn't support",
		},
		{
			name:       "missing host",
			dsn:        "postgres:///db",
			wantErrMsg: "parsing DSN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := NewConnection(tt.dsn)
			require.Error(t, err)
			assert.Nil(t, conn)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestTryConnection_InvalidDSN_ReturnsError(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		maxAttempts int
		wantErrMsg  string
	}{
		{
			name:        "empty dsn with single attempt",
			dsn:         "",
			maxAttempts: 1,
			wantErrMsg:  "parsing DSN",
		},
		{
			name:        "invalid dsn with multiple attempts",
			dsn:         "invalid://",
			maxAttempts: 2,
			wantErrMsg:  "parsing DSN",
		},
		{
			name:        "unsupported driver",
			dsn:         "oracle://user:pass@localhost:1521/db",
			maxAttempts: 1,
			wantErrMsg:  "doesn't support",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := TryConnection(tt.dsn, tt.maxAttempts)
			require.Error(t, err)
			assert.Nil(t, conn)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestTryConnection_ZeroAttempts_DefaultsToOne(t *testing.T) {
	// With zero attempts, it should default to 1 attempt
	// and still return error for invalid DSN
	conn, err := TryConnection("", 0)
	require.Error(t, err)
	assert.Nil(t, conn)
}

func TestTryConnection_NegativeAttempts_DefaultsToOne(t *testing.T) {
	// With negative attempts, it should default to 1 attempt
	// and still return error for invalid DSN
	conn, err := TryConnection("", -5)
	require.Error(t, err)
	assert.Nil(t, conn)
}
