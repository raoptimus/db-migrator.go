/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"context"
	"io"

	"github.com/raoptimus/db-migrator.go/internal/domain/log"
	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

// Logger defines the interface for logging migration operations with different severity levels.
//
//go:generate mockery
type Logger = log.Logger

// File defines the interface for file system operations.
//
//go:generate mockery
type File interface {
	Create(filename string) error
	Exists(path string) (bool, error)
}

// FileNameBuilder defines the interface for building migration file names.
//
//go:generate mockery
type FileNameBuilder interface {
	// Up builds a file name for migration update
	Up(version string, forceSafely bool) (fname string, safely bool)
	// Down builds a file name for migration downgrade
	Down(version string, forceSafely bool) (fname string, safely bool)
}

// MigrationService defines the interface for managing database migration operations.
//
//go:generate mockery
type MigrationService interface {
	// Migrations returns migrations from the database
	Migrations(ctx context.Context, limit int) (model.Migrations, error)
	// NewMigrations returns new migrations that have not been applied yet
	NewMigrations(ctx context.Context) (model.Migrations, error)
	// ApplyFile applies new migration
	ApplyFile(ctx context.Context, migration *model.Migration, fileName string, safely bool) error
	// RevertFile reverts the migration
	RevertFile(ctx context.Context, migration *model.Migration, fileName string, safely bool) error
	// Exists checks whether a migration with the specified version has been applied
	Exists(ctx context.Context, version string) (bool, error)
	// ApplySQL applies a migration by executing the provided SQL statements
	ApplySQL(ctx context.Context, safely bool, version, upSQL string) error
	// RevertSQL reverts a migration by executing the provided SQL statements
	RevertSQL(ctx context.Context, safely bool, version, downSQL string) error
}

// Connection defines the interface for database connection operations.
//
//go:generate mockery
type Connection interface {
	DSN() string
	Driver() connection.Driver
	Ping() error
	QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error)
	Transaction(ctx context.Context, txFn func(ctx context.Context) error) error
	io.Closer
}

// Presenter defines the interface for presenting migration information to the user.
//
//go:generate mockery
type Presenter interface {
	// PrintMigrations prints a list of migrations with optional time information.
	PrintMigrations(migrations model.Migrations, withTime bool)
	// ShowNoNewMigrations displays a message when there are no new migrations to apply.
	ShowNoNewMigrations()
	// ShowNoMigrationsToRevert displays a message when there are no migrations to revert.
	ShowNoMigrationsToRevert()
	// ShowHistoryHeader displays the header for migration history with limit.
	ShowHistoryHeader(count int)
	// ShowAllHistoryHeader displays the header for all applied migrations.
	ShowAllHistoryHeader(count int)
	// ShowNewMigrationsHeader displays the header for new migrations.
	ShowNewMigrationsHeader(count int)
	// ShowNewMigrationsLimitedHeader displays the header when showing limited new migrations.
	ShowNewMigrationsLimitedHeader(shown, total int)
	// ShowUpgradePlan displays the plan for applying migrations.
	ShowUpgradePlan(migrations model.Migrations, total int)
	// ShowDowngradePlan displays the plan for reverting migrations.
	ShowDowngradePlan(migrations model.Migrations)
	// ShowRedoPlan displays the plan for redoing migrations.
	ShowRedoPlan(migrations model.Migrations)
	// AskUpgradeConfirmation returns a confirmation question for applying migrations.
	AskUpgradeConfirmation(count int) string
	// AskDowngradeConfirmation returns a confirmation question for reverting migrations.
	AskDowngradeConfirmation(count int) string
	// AskRedoConfirmation returns a confirmation question for redoing migrations.
	AskRedoConfirmation(count int) string
	// ShowUpgradeError displays a message when some migrations failed during upgrade.
	ShowUpgradeError(applied, total int)
	// ShowDowngradeError displays a message when some migrations failed during downgrade.
	ShowDowngradeError(reverted, total int)
	// ShowRedoError displays a message when redo operation failed.
	ShowRedoError()
	// ShowUpgradeSuccess displays a success message after all migrations have been applied.
	ShowUpgradeSuccess(count int)
	// ShowDowngradeSuccess displays a success message after all migrations have been reverted.
	ShowDowngradeSuccess(count int)
	// ShowRedoSuccess displays a success message after all migrations have been redone.
	ShowRedoSuccess(count int)
}
