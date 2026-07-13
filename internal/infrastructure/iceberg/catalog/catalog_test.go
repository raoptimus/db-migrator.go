/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package catalog_test

import (
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/helper/dsn"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/iceberg/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew_DSNParsing verifies that New parses the DSN correctly and builds a Client
// without panicking. It does NOT make network calls; actual connectivity is verified
// in integration tests.
func TestNew_DSNParsing(t *testing.T) {
	tests := []struct {
		name      string
		dsnStr    string
		wantErr   bool
		warehouse string
	}{
		{
			name:      "bearer token auth",
			dsnStr:    "iceberg://localhost:8181/warehouse?token=mytoken",
			wantErr:   false,
			warehouse: "warehouse",
		},
		{
			name:      "oauth2 client credentials",
			dsnStr:    "iceberg://localhost:8181/mywarehouse?credential=client:secret&oauth2_server_uri=http://auth.example.com/token",
			wantErr:   false,
			warehouse: "mywarehouse",
		},
		{
			name:      "oauth2 with scope",
			dsnStr:    "iceberg://localhost:8181/mywarehouse?credential=client:secret&oauth2_server_uri=http://auth.example.com/token&scope=catalog",
			wantErr:   false,
			warehouse: "mywarehouse",
		},
		{
			name:      "no auth",
			dsnStr:    "iceberg://localhost:8181/warehouse",
			wantErr:   false,
			warehouse: "warehouse",
		},
		{
			name:      "bearer token with prefix",
			dsnStr:    "iceberg://localhost:8181/warehouse?token=mytoken&prefix=myprefix",
			wantErr:   false,
			warehouse: "warehouse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := dsn.Parse(tt.dsnStr)
			require.NoError(t, err, "DSN should parse without error")

			// NewCatalog makes an HTTP call to fetch config, which will fail without a live server.
			// We only assert that the function handles the DSN correctly and returns a known error
			// (network error, not a panic or DSN parsing error).
			client, err := catalog.New(parsed)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			// Client may be nil if REST catalog connection fails (no live server).
			// The important thing is no panic, and the DSN was parsed.
			if err != nil {
				// Accept only network/connection errors (not DSN errors).
				assert.NotContains(t, err.Error(), "parsing oauth2_server_uri",
					"oauth2_server_uri should not cause parse error for valid URLs")
				return
			}

			require.NotNil(t, client)
			assert.Equal(t, tt.warehouse, client.Warehouse())
		})
	}
}

// TestNew_InvalidOAuth2ServerURI verifies that a malformed oauth2_server_uri returns an error.
func TestNew_InvalidOAuth2ServerURI(t *testing.T) {
	parsed, err := dsn.Parse("iceberg://localhost:8181/warehouse?credential=c:s&oauth2_server_uri=://bad")
	require.NoError(t, err)

	_, err = catalog.New(parsed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing oauth2_server_uri")
}

// TestNew_S3Params verifies that New parses DSN containing s3.* parameters without panicking.
// The s3.* params are forwarded as additional catalog properties to configure S3-compatible storage.
// No network call assertions are made; a live server is required for connectivity checks.
func TestNew_S3Params(t *testing.T) {
	tests := []struct {
		name   string
		dsnStr string
	}{
		{
			name: "s3 endpoint and credentials",
			dsnStr: "iceberg://localhost:8181/iceberg" +
				"?s3.endpoint=http://localhost:9000" +
				"&s3.access-key-id=minioadmin" +
				"&s3.secret-access-key=minioadmin" +
				"&s3.region=us-east-1" +
				"&s3.force-virtual-addressing=false",
		},
		{
			name: "s3 with session token",
			dsnStr: "iceberg://localhost:8181/iceberg" +
				"?s3.endpoint=http://localhost:9000" +
				"&s3.access-key-id=ak" +
				"&s3.secret-access-key=sk" +
				"&s3.session-token=st" +
				"&s3.region=eu-west-1",
		},
		{
			name:   "no s3 params",
			dsnStr: "iceberg://localhost:8181/iceberg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := dsn.Parse(tt.dsnStr)
			require.NoError(t, err, "DSN must parse without error")

			// New may return a network error (no live server in unit tests).
			// The test asserts no panic and no DSN-parsing error.
			_, err = catalog.New(parsed)
			if err != nil {
				assert.NotContains(t, err.Error(), "s3",
					"s3.* params must not cause DSN-level errors (only network errors are expected)")
			}
		})
	}
}
