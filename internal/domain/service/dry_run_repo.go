package service

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
)

type DryRunRepository struct {
	repo                Repository
	virtualTableCreated bool
}

func NewDryRunRepository(repo Repository) *DryRunRepository {
	return &DryRunRepository{
		repo: repo,
	}
}

// ExistsMigration returns true if version of migration exists
func (d *DryRunRepository) ExistsMigration(ctx context.Context, version string) (bool, error) {
	if d.virtualTableCreated {
		return false, nil
	}

	return d.repo.ExistsMigration(ctx, version)
}

// Migrations returns applied migrations history.
func (d *DryRunRepository) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	if d.virtualTableCreated {
		return nil, nil
	}

	return d.repo.Migrations(ctx, limit)
}

// HasMigrationHistoryTable returns true if migration history table exists.
func (d *DryRunRepository) HasMigrationHistoryTable(ctx context.Context) (exists bool, err error) {
	if d.virtualTableCreated {
		return true, nil
	}

	return d.repo.HasMigrationHistoryTable(ctx)
}

// InsertMigration inserts the new migration record.
func (d *DryRunRepository) InsertMigration(ctx context.Context, version string) error {
	return nil
}

// RemoveMigration removes the migration record.
func (d *DryRunRepository) RemoveMigration(ctx context.Context, version string) error {
	return nil
}

// ExecQuery executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (d *DryRunRepository) ExecQuery(ctx context.Context, query string, args ...any) error {
	return nil
}

// ExecQueryTransaction executes a function within a database transaction.
func (d *DryRunRepository) ExecQueryTransaction(ctx context.Context, fnTx func(ctx context.Context) error) error {
	return nil
}

// DropMigrationHistoryTable drops the migration history table from the database.
func (d *DryRunRepository) DropMigrationHistoryTable(ctx context.Context) error {
	return nil
}

// CreateMigrationHistoryTable creates the migration history table in the database.
func (d *DryRunRepository) CreateMigrationHistoryTable(ctx context.Context) error {
	d.virtualTableCreated = true

	return nil
}

// MigrationsCount returns the total number of applied migrations.
func (d *DryRunRepository) MigrationsCount(ctx context.Context) (int, error) {
	if d.virtualTableCreated {
		return 0, nil
	}

	return d.repo.MigrationsCount(ctx)
}

// TableNameWithSchema returns the full table name including schema if applicable.
func (d *DryRunRepository) TableNameWithSchema() string {
	if d.virtualTableCreated {
		return ""
	}

	return d.repo.TableNameWithSchema()
}
