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

	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

//go:generate mockery
type Console interface {
	Confirm(s string) bool
	Info(message string)
	InfoLn(message string)
	Infof(message string, a ...any)
	Success(message string)
	SuccessLn(message string)
	Successf(message string, a ...any)
	Warn(message string)
	WarnLn(message string)
	Warnf(message string, a ...any)
	Error(message string)
	ErrorLn(message string)
	Errorf(message string, a ...any)
	Fatal(err error)
	NumberPlural(count int, one, many string) string
}

//go:generate mockery
type File interface {
	Exists(fileName string) (bool, error)
	Open(filename string) (io.ReadCloser, error)
	ReadAll(filename string) ([]byte, error)
}

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
	ExecQueryTransaction(ctx context.Context, fnTx func(ctx context.Context) error) error
	DropMigrationHistoryTable(ctx context.Context) error
	CreateMigrationHistoryTable(ctx context.Context) error
	MigrationsCount(ctx context.Context) (int, error)
	TableNameWithSchema() string
	ForceSafely() bool
}
