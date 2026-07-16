/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

// Package catalog provides a thin wrapper over github.com/apache/iceberg-go/catalog/rest.
// It constructs a REST catalog client from a parsed DSN and maps auth options from
// DSN query parameters.
//
// Auth mapping (by presence of DSN parameters):
//   - ?token=<t>                                   → rest.WithOAuthToken(t)
//   - ?credential=<c>&oauth2_server_uri=<u>[&scope=<s>] → rest.WithCredential(c)+rest.WithAuthURI(u)[+rest.WithScope(s)]
//   - &prefix=<p>                                  → rest.WithPrefix(p) (combined with any auth branch)
//
// Namespace-property methods and table/schema/spec methods are implemented here,
// fulfilling the repository.IcebergCatalog interface.
package catalog

import (
	"context"
	"net/url"
	"strings"

	iceberg "github.com/apache/iceberg-go"
	"github.com/apache/iceberg-go/catalog/rest"
	_ "github.com/apache/iceberg-go/io/gocloud" // register s3/gcs/azure schemes
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/helper/dsn"
)

// Client is a thin wrapper over *rest.Catalog that exposes the subset of
// catalog operations needed by the Iceberg repository driver.
type Client struct {
	cat       *rest.Catalog
	warehouse string // warehouse name extracted from DSN path (used for catalog-prefix stripping)
}

// New constructs a REST catalog client from a parsed DSN.
// It maps auth parameters from DSN.Options:
//   - token         → bearer token (WithOAuthToken)
//   - credential + oauth2_server_uri (+ scope) → OAuth2 client-credentials
//   - prefix        → catalog path prefix
func New(d *dsn.DSN) (*Client, error) {
	scheme := "http"
	if d.Options.Get("secure") == "true" || d.Options.Get("sslmode") == "require" {
		scheme = "https"
	}
	uri := scheme + "://" + d.PrimaryAddress()

	opts := make([]rest.Option, 0)

	token := d.Options.Get("token")
	credential := d.Options.Get("credential")
	oauthServerURI := d.Options.Get("oauth2_server_uri")

	switch {
	case token != "":
		opts = append(opts, rest.WithOAuthToken(token))
	case credential != "" && oauthServerURI != "":
		authURI, err := url.Parse(oauthServerURI)
		if err != nil {
			return nil, errors.WithMessage(err, "parsing oauth2_server_uri")
		}
		opts = append(opts, rest.WithCredential(credential), rest.WithAuthURI(authURI))
		if scope := d.Options.Get("scope"); scope != "" {
			opts = append(opts, rest.WithScope(scope))
		}
	}

	if prefix := d.Options.Get("prefix"); prefix != "" {
		opts = append(opts, rest.WithPrefix(prefix))
	}

	// Collect all DSN parameters with prefix "s3." and forward them as additional
	// catalog properties so the REST catalog can configure S3-compatible storage.
	if s3Props := extractS3Props(d.Options); len(s3Props) > 0 {
		opts = append(opts, rest.WithAdditionalProps(s3Props))
	}

	cat, err := rest.NewCatalog(context.Background(), d.Database, uri, opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "creating iceberg REST catalog")
	}

	return &Client{cat: cat, warehouse: d.Database}, nil
}

// Warehouse returns the warehouse name from the DSN path.
func (c *Client) Warehouse() string {
	return c.warehouse
}

// Ping verifies connectivity to the REST catalog by listing top-level namespaces.
func (c *Client) Ping(ctx context.Context) error {
	// Disable pagination: the reference Iceberg REST server (iceberg-rest-fixture) throws
	// NumberFormatException on listNamespaces when a pageSize is sent without a pageToken.
	// iceberg-go sends pageSize=20 unconditionally, so suppress it to avoid the server bug.
	ctx = c.cat.SetPageSize(ctx, 0)
	_, err := c.cat.ListNamespaces(ctx, nil)
	if err != nil {
		return errors.WithMessage(err, "ping iceberg catalog")
	}
	return nil
}

// Close is a no-op for the REST catalog client (HTTP connections are not pooled at this level).
func (c *Client) Close() error {
	return nil
}

// CreateNamespace creates a namespace with the given properties.
func (c *Client) CreateNamespace(ctx context.Context, ns []string, props map[string]string) error {
	if err := c.cat.CreateNamespace(ctx, ns, iceberg.Properties(props)); err != nil {
		return errors.WithMessage(err, "create namespace")
	}
	return nil
}

// DropNamespace drops the given namespace.
func (c *Client) DropNamespace(ctx context.Context, ns []string) error {
	if err := c.cat.DropNamespace(ctx, ns); err != nil {
		return errors.WithMessage(err, "drop namespace")
	}
	return nil
}

// NamespaceExists checks whether the given namespace exists.
func (c *Client) NamespaceExists(ctx context.Context, ns []string) (bool, error) {
	exists, err := c.cat.CheckNamespaceExists(ctx, ns)
	if err != nil {
		return false, errors.WithMessage(err, "check namespace exists")
	}
	return exists, nil
}

// LoadNamespaceProperties returns the properties of the given namespace.
func (c *Client) LoadNamespaceProperties(ctx context.Context, ns []string) (map[string]string, error) {
	props, err := c.cat.LoadNamespaceProperties(ctx, ns)
	if err != nil {
		return nil, errors.WithMessage(err, "load namespace properties")
	}
	return map[string]string(props), nil
}

// UpdateNamespaceProperties updates namespace properties by removing and setting keys.
func (c *Client) UpdateNamespaceProperties(
	ctx context.Context, ns []string, removals []string, updates map[string]string,
) error {
	_, err := c.cat.UpdateNamespaceProperties(ctx, ns, removals, iceberg.Properties(updates))
	if err != nil {
		return errors.WithMessage(err, "update namespace properties")
	}
	return nil
}

// extractS3Props scans DSN query parameters and returns all entries whose key
// starts with "s3." as an iceberg.Properties map. This makes the DSN the single
// source of S3/MinIO storage configuration — callers pass the result directly to
// rest.WithAdditionalProps so the REST catalog can configure its IO layer.
func extractS3Props(opts url.Values) iceberg.Properties {
	props := make(iceberg.Properties)
	for key, vals := range opts {
		if strings.HasPrefix(key, "s3.") && len(vals) > 0 {
			props[key] = vals[0]
		}
	}
	if len(props) == 0 {
		return nil
	}
	return props
}
