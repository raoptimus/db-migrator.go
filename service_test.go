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

func TestNewDBService_ValidOptions_Success(t *testing.T) {
	conn := NewMockConnection(t)
	logger := NewMockLogger(t)

	opts := &Options{
		DSN:       "postgres://user:pass@localhost:5432/db",
		TableName: "migration",
	}

	svc, err := NewDBService(opts, conn, logger)
	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestNewDBService_NilLogger_UsesNopLogger(t *testing.T) {
	conn := NewMockConnection(t)

	opts := &Options{
		DSN:       "postgres://user:pass@localhost:5432/db",
		TableName: "migration",
	}

	svc, err := NewDBService(opts, conn, nil)
	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestNewDBService_InvalidOptions_ReturnsError(t *testing.T) {
	tests := []struct {
		name       string
		opts       *Options
		wantErrMsg string
	}{
		{
			name: "invalid table name with special chars",
			opts: &Options{
				DSN:       "postgres://user:pass@localhost:5432/db",
				TableName: "table; DROP TABLE users;--",
			},
			wantErrMsg: "tableName",
		},
		{
			name: "invalid cluster name",
			opts: &Options{
				DSN:         "postgres://user:pass@localhost:5432/db",
				TableName:   "migration",
				ClusterName: "cluster; DROP TABLE users;--",
			},
			wantErrMsg: "clusterName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := NewMockConnection(t)
			logger := NewMockLogger(t)

			svc, err := NewDBService(tt.opts, conn, logger)
			require.Error(t, err)
			assert.Nil(t, svc)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestUpgrade_InvalidVersion_ReturnsError(t *testing.T) {
	conn := NewMockConnection(t)
	opts := &Options{
		DSN:       "postgres://user:pass@localhost:5432/db",
		TableName: "migration",
	}

	svc, err := NewDBService(opts, conn, nil)
	require.NoError(t, err)

	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "empty version",
			version: "",
		},
		{
			name:    "invalid format",
			version: "not-a-version",
		},
		{
			name:    "missing name part",
			version: "210328_221600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			err := svc.Upgrade(ctx, tt.version, "CREATE TABLE test;", false)
			require.Error(t, err)
		})
	}
}

func TestDowngrade_InvalidVersion_ReturnsError(t *testing.T) {
	conn := NewMockConnection(t)
	opts := &Options{
		DSN:       "postgres://user:pass@localhost:5432/db",
		TableName: "migration",
	}

	svc, err := NewDBService(opts, conn, nil)
	require.NoError(t, err)

	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "empty version",
			version: "",
		},
		{
			name:    "invalid format",
			version: "invalid",
		},
		{
			name:    "missing name",
			version: "210328_221600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			err := svc.Downgrade(ctx, tt.version, "DROP TABLE test;", false)
			require.Error(t, err)
		})
	}
}

func TestOptions_Defaults(t *testing.T) {
	opts := &Options{}
	assert.Empty(t, opts.DSN)
	assert.Empty(t, opts.TableName)
	assert.Empty(t, opts.ClusterName)
	assert.False(t, opts.Replicated)
}
