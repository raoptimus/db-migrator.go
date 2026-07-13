/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

// Package iceberg provides a thin SQLDB adapter for Apache Iceberg REST catalog.
// Ping/Close are backed by the catalog client; SQL-level methods (QueryContext, ExecContext,
// BeginTx) are not on the critical migration path and return ErrNotSupported.
// Actual DDL execution is handled by the repository layer.
package iceberg

import (
	"context"
	"database/sql"

	"github.com/raoptimus/db-migrator.go/internal/helper/dsn"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/iceberg/catalog"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

// DB wraps a catalog.Client and implements connection.SQLDB.
// Ping and Close are delegated to the REST catalog; SQL-level operations are unsupported.
type DB struct {
	cat *catalog.Client
}

// Open parses the DSN and creates a new DB backed by an Iceberg REST catalog client.
func Open(dsnStr string) (*DB, error) {
	parsed, err := dsn.Parse(dsnStr)
	if err != nil {
		return nil, err
	}

	cat, err := catalog.New(parsed)
	if err != nil {
		return nil, err
	}

	return &DB{cat: cat}, nil
}

// Ping verifies connectivity to the Iceberg REST catalog.
func (db *DB) Ping() error {
	return db.cat.Ping(context.Background())
}

// Close closes the catalog client (no-op for HTTP REST client).
func (db *DB) Close() error {
	return db.cat.Close()
}

// QueryContext is not supported; use the repository layer for Iceberg operations.
//
//nolint:ireturn,nolintlint // its ok
func (db *DB) QueryContext(_ context.Context, _ string, _ ...any) (sqlex.Rows, error) {
	return nil, ErrNotSupported
}

// ExecContext is not supported; use the repository layer for Iceberg operations.
//
//nolint:ireturn,nolintlint // its ok
func (db *DB) ExecContext(_ context.Context, _ string, _ ...any) (sqlex.Result, error) {
	return nil, ErrNotSupported
}

// BeginTx is not supported; Iceberg REST catalog does not use SQL transactions.
//
//nolint:ireturn,nolintlint // its ok
func (db *DB) BeginTx(_ context.Context, _ *sql.TxOptions) (sqlex.Tx, error) {
	return nil, ErrNotSupported
}
