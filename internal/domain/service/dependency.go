/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package service

import (
	"context"
	"io"

	"github.com/raoptimus/db-migrator.go/internal/domain/log"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
)

// Logger defines the interface for logging migration operations.
// It provides methods for different log levels including info, success, warning, error, and fatal.
//
//go:generate mockery
type Logger = log.Logger

// File defines the interface for file system operations.
// It provides methods to check file existence, open files, and read file contents.
//
//go:generate mockery
type File interface {
	Exists(fileName string) (bool, error)
	Open(filename string) (io.ReadCloser, error)
	ReadAll(filename string) ([]byte, error)
}

// Repository defines the interface for database migration operations.
// It provides methods to manage migration history, execute queries, and interact with the migration table.
//
//go:generate mockery
type Repository interface {
	// ExistsMigration returns true if version of migration exists
	ExistsMigration(ctx context.Context, version string) (bool, error)
	// Migrations returns applied migrations history.
	Migrations(ctx context.Context, limit int) (entity.Migrations, error)
	// HasMigrationHistoryTable returns true if migration history table exists.
	HasMigrationHistoryTable(ctx context.Context) (exists bool, err error)
	// InsertMigration inserts the new migration record.
	InsertMigration(ctx context.Context, version string) error
	// RemoveMigration removes the migration record.
	RemoveMigration(ctx context.Context, version string) error
	// ExecQuery executes a query without returning any rows.
	// The args are for any placeholder parameters in the query.
	ExecQuery(ctx context.Context, query string, args ...any) error
	// ExecQueryTransaction executes a function within a database transaction.
	ExecQueryTransaction(ctx context.Context, fnTx func(ctx context.Context) error) error
	// DropMigrationHistoryTable drops the migration history table from the database.
	DropMigrationHistoryTable(ctx context.Context) error
	// CreateMigrationHistoryTable creates the migration history table in the database.
	CreateMigrationHistoryTable(ctx context.Context) error
	// MigrationsCount returns the total number of applied migrations.
	MigrationsCount(ctx context.Context) (int, error)
	// TableNameWithSchema returns the full table name including schema if applicable.
	TableNameWithSchema() string
}
