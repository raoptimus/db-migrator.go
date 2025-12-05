/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package connection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveSslModeFromDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		expected string
	}{
		{
			name:     "DSN with sslmode parameter",
			dsn:      "clickhouse://default:@localhost:9000/docker?sslmode=disable&compress=true",
			expected: "clickhouse://default:@localhost:9000/docker?compress=true",
		},
		{
			name:     "DSN with only sslmode parameter",
			dsn:      "clickhouse://default:@localhost:9000/docker?sslmode=disable",
			expected: "clickhouse://default:@localhost:9000/docker",
		},
		{
			name:     "DSN without sslmode parameter",
			dsn:      "clickhouse://default:@localhost:9000/docker?compress=true",
			expected: "clickhouse://default:@localhost:9000/docker?compress=true",
		},
		{
			name:     "DSN without any parameters",
			dsn:      "clickhouse://default:@localhost:9000/docker",
			expected: "clickhouse://default:@localhost:9000/docker",
		},
		{
			name:     "DSN with sslmode in the middle",
			dsn:      "clickhouse://default:@localhost:9000/docker?compress=true&sslmode=disable&debug=true",
			expected: "clickhouse://default:@localhost:9000/docker?compress=true&debug=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := removeSslModeFromDSN(tt.dsn)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveSslModeFromDSN_InvalidDSN(t *testing.T) {
	_, err := removeSslModeFromDSN("://invalid-dsn")
	assert.Error(t, err)
}
