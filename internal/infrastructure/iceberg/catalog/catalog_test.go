/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package catalog_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
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

// brokenHeadCatalog emulates an Iceberg REST server (older apache/iceberg-rest builds on
// JdbcCatalog) that does NOT implement HEAD /v1/namespaces/{ns} and rejects it with 400 for
// any namespace, while GET works correctly. NamespaceExists must rely on GET only, so that a
// missing namespace yields (false, nil) — not a bad-request error — even though HEAD returns 400.
func brokenHeadCatalog(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// HEAD on any namespace is unsupported and rejected with 400.
		if r.Method == http.MethodHead && strings.HasPrefix(r.URL.Path, "/v1/namespaces/") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch r.URL.Path {
		case "/v1/config":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"defaults":{},"overrides":{}}`))
		case "/v1/namespaces/exists":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"namespace":["exists"],"properties":{}}`))
		case "/v1/namespaces/missing":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"namespace does not exist",` +
				`"type":"NoSuchNamespaceException","code":404}}`))
		case "/v1/namespaces/boom":
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":{"message":"boom","type":"ServerError","code":500}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"not found","type":"NotFound","code":404}}`))
		}
	}))
}

// TestNamespaceExists_HeadUnsupported is a regression test for the dev bug where db-migrator up
// died with "check namespace exists: REST error: bad request" against a REST server that rejects
// HEAD /v1/namespaces/{ns} with 400. NamespaceExists must probe via GET, so it works despite HEAD
// returning 400 — the passing test proves HEAD is no longer used.
func TestNamespaceExists_HeadUnsupported(t *testing.T) {
	ts := brokenHeadCatalog(t)
	defer ts.Close()

	parsed, err := dsn.Parse("iceberg://" + strings.TrimPrefix(ts.URL, "http://") + "/warehouse")
	require.NoError(t, err)

	client, err := catalog.New(parsed)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx := context.Background()

	exists, err := client.NamespaceExists(ctx, []string{"exists"})
	require.NoError(t, err)
	assert.True(t, exists, "existing namespace must be reported as existing")

	exists, err = client.NamespaceExists(ctx, []string{"missing"})
	require.NoError(t, err, "missing namespace must not surface HEAD 400 as an error")
	assert.False(t, exists, "missing namespace must be reported as not existing")

	_, err = client.NamespaceExists(ctx, []string{"boom"})
	require.Error(t, err, "a non-404 error on GET must be propagated")
	assert.Contains(t, err.Error(), "check namespace exists")
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
