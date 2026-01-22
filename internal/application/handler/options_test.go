/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptions_Validate_ValidOptions_Success(t *testing.T) {
	tests := []struct {
		name    string
		options Options
	}{
		{
			name: "all valid fields",
			options: Options{
				PlaceholderCustom:  "custom_value",
				DSN:                "postgres://user:pass@localhost:5432/testdb",
				MaxConnAttempts:    5,
				Directory:          "/migrations",
				TableName:          "migration_history",
				ClusterName:        "my_cluster",
				Replicated:         true,
				Compact:            false,
				Interactive:        true,
				MaxSQLOutputLength: 1000,
			},
		},
		{
			name: "minimal valid options",
			options: Options{
				PlaceholderCustom: "",
				DSN:               "mysql://localhost/db",
				MaxConnAttempts:   1,
				Directory:         "",
				TableName:         "migration",
				ClusterName:       "",
				Replicated:        false,
				Compact:           false,
				Interactive:       false,
			},
		},
		{
			name: "max conn attempts at boundary",
			options: Options{
				MaxConnAttempts: 100,
				TableName:       "migrations",
			},
		},
		{
			name: "table name with underscore",
			options: Options{
				MaxConnAttempts: 1,
				TableName:       "my_migration_table",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			require.NoError(t, err)
		})
	}
}

func TestOptions_Validate_InvalidPlaceholderCustom_Failure(t *testing.T) {
	tests := []struct {
		name              string
		placeholderCustom string
		expectedErrSubstr string
	}{
		{
			name:              "contains semicolon",
			placeholderCustom: "value;drop",
			expectedErrSubstr: "placeholderCustom",
		},
		{
			name:              "contains quotes",
			placeholderCustom: "value'test",
			expectedErrSubstr: "placeholderCustom",
		},
		{
			name:              "contains double quotes",
			placeholderCustom: `value"test`,
			expectedErrSubstr: "placeholderCustom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				PlaceholderCustom: tt.placeholderCustom,
				MaxConnAttempts:   1,
				TableName:         "migration",
			}

			err := opts.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

func TestOptions_Validate_InvalidMaxConnAttempts_Failure(t *testing.T) {
	tests := []struct {
		name            string
		maxConnAttempts int
		expectedErrMsg  string
	}{
		{
			name:            "zero attempts",
			maxConnAttempts: 0,
			expectedErrMsg:  "maxConnAttempts",
		},
		{
			name:            "negative attempts",
			maxConnAttempts: -1,
			expectedErrMsg:  "maxConnAttempts",
		},
		{
			name:            "exceeds max",
			maxConnAttempts: 101,
			expectedErrMsg:  "maxConnAttempts",
		},
		{
			name:            "far exceeds max",
			maxConnAttempts: 1000,
			expectedErrMsg:  "maxConnAttempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				MaxConnAttempts: tt.maxConnAttempts,
				TableName:       "migration",
			}

			err := opts.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

func TestOptions_Validate_InvalidTableName_Failure(t *testing.T) {
	tests := []struct {
		name           string
		tableName      string
		expectedErrMsg string
	}{
		{
			name:           "contains semicolon",
			tableName:      "table;drop",
			expectedErrMsg: "tableName",
		},
		{
			name:           "contains space",
			tableName:      "table name",
			expectedErrMsg: "tableName",
		},
		{
			name:           "sql injection attempt",
			tableName:      "table'; DROP TABLE users;--",
			expectedErrMsg: "tableName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				MaxConnAttempts: 1,
				TableName:       tt.tableName,
			}

			err := opts.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}

func TestOptions_Validate_InvalidClusterName_Failure(t *testing.T) {
	tests := []struct {
		name           string
		clusterName    string
		expectedErrMsg string
	}{
		{
			name:           "contains semicolon",
			clusterName:    "cluster;drop",
			expectedErrMsg: "clusterName",
		},
		{
			name:           "contains quotes",
			clusterName:    "cluster'test",
			expectedErrMsg: "clusterName",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				MaxConnAttempts: 1,
				TableName:       "migration",
				ClusterName:     tt.clusterName,
			}

			err := opts.Validate()

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErrMsg)
		})
	}
}
