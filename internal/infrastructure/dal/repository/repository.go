/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
)

// Repository defines the interface for database-specific migration history operations.
// It provides methods for managing migration history table and executing database queries.
type Repository interface {
	// Migrations retrieves the list of applied migrations from the database, limited to the specified count.
	Migrations(ctx context.Context, limit int) (entity.Migrations, error)
	// HasMigrationHistoryTable checks if the migration history table exists in the database.
	HasMigrationHistoryTable(ctx context.Context) (exists bool, err error)
	// InsertMigration inserts a new migration version into the migration history table.
	InsertMigration(ctx context.Context, version string) error
	// RemoveMigration removes a migration version from the migration history table.
	RemoveMigration(ctx context.Context, version string) error
	// ExecQuery executes a query that doesn't return rows with the provided arguments.
	ExecQuery(ctx context.Context, query string, args ...any) error
	// QueryScalar executes a query that returns a single scalar value into the provided pointer.
	QueryScalar(ctx context.Context, query string, ptr any) error
	// ExecQueryTransaction executes a function within a database transaction.
	ExecQueryTransaction(ctx context.Context, fnTx func(ctx context.Context) error) error
	// DropMigrationHistoryTable drops the migration history table from the database.
	DropMigrationHistoryTable(ctx context.Context) error
	// CreateMigrationHistoryTable creates the migration history table in the database.
	CreateMigrationHistoryTable(ctx context.Context) error
	// MigrationsCount returns the total number of applied migrations in the database.
	MigrationsCount(ctx context.Context) (int, error)
	// ExistsMigration checks if a specific migration version exists in the migration history.
	ExistsMigration(ctx context.Context, version string) (bool, error)
	// TableNameWithSchema returns the fully qualified table name including schema.
	TableNameWithSchema() string
}

// New creates repository by connection
//
//nolint:ireturn,nolintlint // it's ok
func New(conn Connection, options *Options) (Repository, error) {
	registry := NewFactoryRegistry()

	return registry.Create(conn, options)
}
