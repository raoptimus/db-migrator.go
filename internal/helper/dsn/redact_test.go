/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dsn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedact(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "iceberg DSN with token and s3.secret-access-key",
			input:    "iceberg://localhost:8181/iceberg?token=SUPERSECRET123&s3.secret-access-key=SECRETKEY456&s3.access-key-id=minioadmin&s3.region=us-east-1",
			expected: "iceberg://localhost:8181/iceberg?token=****&s3.secret-access-key=****&s3.access-key-id=minioadmin&s3.region=us-east-1",
		},
		{
			name:     "iceberg DSN with credential and s3.session-token",
			input:    "iceberg://localhost:8181/iceberg?credential=CLIENT:SECRET&s3.session-token=MYSESSIONTOKEN&s3.region=us-east-1",
			expected: "iceberg://localhost:8181/iceberg?credential=****&s3.session-token=****&s3.region=us-east-1",
		},
		{
			name:     "iceberg DSN with all sensitive params",
			input:    "iceberg://localhost:8181/iceberg?token=T123&credential=C456&s3.secret-access-key=SK789&s3.session-token=ST012&s3.access-key-id=AK",
			expected: "iceberg://localhost:8181/iceberg?token=****&credential=****&s3.secret-access-key=****&s3.session-token=****&s3.access-key-id=AK",
		},
		{
			name:     "postgres DSN with password in userinfo",
			input:    "postgres://user:mysecretpassword@localhost:5432/mydb?sslmode=disable",
			expected: "postgres://user:****@localhost:5432/mydb?sslmode=disable",
		},
		{
			name:     "mysql DSN with password in userinfo",
			input:    "mysql://dbuser:topsecret@localhost:3306/testdb",
			expected: "mysql://dbuser:****@localhost:3306/testdb",
		},
		{
			name:     "clickhouse DSN with password in userinfo and no query params",
			input:    "clickhouse://default:pass123@host1:9000,host2:9000/default?compress=true",
			expected: "clickhouse://default:****@host1:9000,host2:9000/default?compress=true",
		},
		{
			name:     "DSN without password in userinfo unchanged",
			input:    "postgres://user@localhost:5432/mydb",
			expected: "postgres://user@localhost:5432/mydb",
		},
		{
			name:     "DSN without credentials unchanged",
			input:    "tarantool://localhost:3301/testdb",
			expected: "tarantool://localhost:3301/testdb",
		},
		{
			name:     "DSN without any secrets unchanged",
			input:    "postgres://user:pass@localhost:5432/db?sslmode=disable",
			expected: "postgres://user:****@localhost:5432/db?sslmode=disable",
		},
		{
			name:     "DSN with empty password (trailing colon) not masked",
			input:    "clickhouse://default:@localhost:9000/default",
			expected: "clickhouse://default:@localhost:9000/default",
		},
		{
			name:     "query parameter password key masked",
			input:    "iceberg://localhost:8181/ns?password=secret123&other=value",
			expected: "iceberg://localhost:8181/ns?password=****&other=value",
		},
		{
			name:     "token at end of string (no trailing delimiter)",
			input:    "iceberg://localhost:8181/ns?other=val&token=ENDTOKEN",
			expected: "iceberg://localhost:8181/ns?other=val&token=****",
		},
		{
			name:     "multiple occurrences of sensitive param",
			input:    "iceberg://localhost:8181/ns?token=FIRST&token=SECOND",
			expected: "iceberg://localhost:8181/ns?token=****&token=****",
		},
		{
			name:     "empty DSN returns empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "malformed DSN without scheme does not panic",
			input:    "not-a-valid-dsn",
			expected: "not-a-valid-dsn",
		},
		{
			name:     "malformed DSN with query params still masks params",
			input:    "not-a-url?token=LEAKED&other=safe",
			expected: "not-a-url?token=****&other=safe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Redact(tt.input)
			assert.Equal(t, tt.expected, got)
			// Ensure no raw secret is present in the output for iceberg test cases.
			if tt.name == "iceberg DSN with token and s3.secret-access-key" {
				assert.NotContains(t, got, "SUPERSECRET123")
				assert.NotContains(t, got, "SECRETKEY456")
			}
		})
	}
}

func TestRedact_NoPanic(t *testing.T) {
	// These should never panic regardless of input.
	inputs := []string{
		"",
		"://",
		"://user:pass@",
		"notaurl",
		"token=naked",
		"http://",
		"://:::@??##",
	}
	for _, input := range inputs {
		assert.NotPanics(t, func() {
			Redact(input)
		}, "Redact panicked on input: %q", input)
	}
}
