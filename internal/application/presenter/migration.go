/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package presenter

import (
	"fmt"

	"github.com/raoptimus/db-migrator.go/internal/domain/log"
	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/raoptimus/db-migrator.go/internal/helper/plural"
)

// Logger defines the interface for logging migration presentation messages.
//
//go:generate mockery
type Logger = log.Logger

// MigrationPresenter handles presentation logic for migration operations.
// It formats and displays migration-related information to the user.
type MigrationPresenter struct {
	logger Logger
}

// NewMigrationPresenter creates a new MigrationPresenter instance.
func NewMigrationPresenter(logger Logger) *MigrationPresenter {
	return &MigrationPresenter{
		logger: logger,
	}
}

// ShowUpgradePlan displays the plan for applying migrations.
// It shows the number of migrations to be applied and prints their list.
func (p *MigrationPresenter) ShowUpgradePlan(migrations model.Migrations, total int) {
	if migrations.Len() == total {
		p.logger.Warnf(
			"Total %d new %s to be applied: \n",
			migrations.Len(),
			plural.Migration(migrations.Len()),
		)
	} else {
		p.logger.Warnf(
			"Total %d out of %d new %s to be applied: \n",
			migrations.Len(),
			total,
			plural.Migration(total),
		)
	}

	p.PrintMigrations(migrations, false)
}

// ShowDowngradePlan displays the plan for reverting migrations.
// It shows the number of migrations to be reverted and prints their list.
func (p *MigrationPresenter) ShowDowngradePlan(migrations model.Migrations) {
	p.logger.Warnf("Total %d %s to be reverted: \n",
		migrations.Len(),
		plural.Migration(migrations.Len()),
	)

	p.PrintMigrations(migrations, false)
}

// PrintMigrations prints a list of migrations.
// If withTime is true, it includes the apply time for each migration.
func (p *MigrationPresenter) PrintMigrations(migrations model.Migrations, withTime bool) {
	for _, migration := range migrations {
		if withTime {
			p.logger.Infof("\t(%s) %s\n", migration.ApplyTimeFormat(), migration.Version)
			continue
		}

		p.logger.Infof("\t%s\n", migration.Version)
	}
}

// AskUpgradeConfirmation returns a confirmation question for applying migrations.
func (p *MigrationPresenter) AskUpgradeConfirmation(count int) string {
	return fmt.Sprintf("Apply the above %s?", plural.Migration(count))
}

// AskDowngradeConfirmation returns a confirmation question for reverting migrations.
func (p *MigrationPresenter) AskDowngradeConfirmation(count int) string {
	return fmt.Sprintf("Revert the above %d %s?", count, plural.Migration(count))
}

// ShowUpgradeError displays a message when some migrations were applied before an error occurred.
func (p *MigrationPresenter) ShowUpgradeError(applied, total int) {
	p.logger.Errorf("%d from %d %s applied.\n",
		applied,
		total,
		plural.MigrationWas(applied),
	)
	p.logger.Error("The rest of the migrations are canceled.")
}

// ShowDowngradeError displays a message when some migrations were reverted before an error occurred.
func (p *MigrationPresenter) ShowDowngradeError(reverted, total int) {
	p.logger.Errorf(
		"%d from %d %s reverted.\n"+
			"Migration failed. The rest of the migrations are canceled.\n",
		reverted,
		total,
		plural.MigrationWas(reverted),
	)
}

// ShowUpgradeSuccess displays a success message after all migrations have been applied.
func (p *MigrationPresenter) ShowUpgradeSuccess(count int) {
	p.logger.Successf("%d %s applied\n", count, plural.MigrationWas(count))
	p.logger.Success("Migrated up successfully")
}

// ShowDowngradeSuccess displays a success message after all migrations have been reverted.
func (p *MigrationPresenter) ShowDowngradeSuccess(count int) {
	p.logger.Successf("%d %s reverted\n", count, plural.MigrationWas(count))
	p.logger.Success("Migrated down successfully")
}

// ShowNoNewMigrations displays a message when there are no new migrations to apply.
func (p *MigrationPresenter) ShowNoNewMigrations() {
	p.logger.Success("No new migrations found. Your system is up-to-date.")
}

// ShowNoMigrationsToRevert displays a message when there are no migrations to revert.
func (p *MigrationPresenter) ShowNoMigrationsToRevert() {
	p.logger.Success("No migration has been done before.")
}

// ShowRedoPlan displays the plan for redoing migrations.
func (p *MigrationPresenter) ShowRedoPlan(migrations model.Migrations) {
	p.logger.Warnf(
		"Total %d %s to be redone: \n",
		migrations.Len(),
		plural.Migration(migrations.Len()),
	)

	p.PrintMigrations(migrations, false)
}

// AskRedoConfirmation returns a confirmation question for redoing migrations.
func (p *MigrationPresenter) AskRedoConfirmation(count int) string {
	return fmt.Sprintf("Redo the above %d %s?\n", count, plural.Migration(count))
}

// ShowRedoError displays a message when redo operation failed.
func (p *MigrationPresenter) ShowRedoError() {
	p.logger.Error("Migration failed. The rest of the migrations are canceled.\n")
}

// ShowRedoSuccess displays a success message after all migrations have been redone.
func (p *MigrationPresenter) ShowRedoSuccess(count int) {
	p.logger.Warnf("%d %s redone.\n", count, plural.MigrationWas(count))
	p.logger.Success("Migration redone successfully.\n")
}

// ShowHistoryHeader displays the header for migration history with limit.
func (p *MigrationPresenter) ShowHistoryHeader(count int) {
	p.logger.Warnf(
		"Showing the last %d %s: \n",
		count,
		plural.Migration(count),
	)
}

// ShowAllHistoryHeader displays the header for all applied migrations.
func (p *MigrationPresenter) ShowAllHistoryHeader(count int) {
	p.logger.Warnf(
		"Total %d %s been applied before: \n",
		count,
		plural.MigrationHas(count),
	)
}

// ShowNewMigrationsHeader displays the header for new migrations.
func (p *MigrationPresenter) ShowNewMigrationsHeader(count int) {
	p.logger.Warnf(
		"Found %d new %s \n",
		count,
		plural.Migration(count),
	)
}

// ShowNewMigrationsLimitedHeader displays the header when showing limited new migrations.
func (p *MigrationPresenter) ShowNewMigrationsLimitedHeader(shown, total int) {
	p.logger.Warnf(
		"Showing %d out of %d new %s \n",
		shown,
		total,
		plural.Migration(total),
	)
}

// ShowReleasePlan displays the plan for applying all migrations in a release.
func (p *MigrationPresenter) ShowReleasePlan(migrations model.Migrations) {
	p.logger.Warnf(
		"Total %d %s to be released: \n",
		migrations.Len(),
		plural.Migration(migrations.Len()),
	)

	p.PrintMigrations(migrations, false)
}

// AskReleaseConfirmation returns a confirmation question for applying a release.
func (p *MigrationPresenter) AskReleaseConfirmation(count int) string {
	return fmt.Sprintf("Release the above %d %s?", count, plural.Migration(count))
}

// ShowReleaseSuccess displays a success message after all migrations have been released.
func (p *MigrationPresenter) ShowReleaseSuccess(count int) {
	p.logger.Successf("%d %s released\n", count, plural.MigrationWas(count))
	p.logger.Success("Release applied successfully")
}

// ShowReleaseError displays a message when release operation failed.
func (p *MigrationPresenter) ShowReleaseError() {
	p.logger.Error("Release failed. All changes have been rolled back.\n")
}

// ShowRollbackPlan displays the plan for reverting a release.
func (p *MigrationPresenter) ShowRollbackPlan(migrations model.Migrations) {
	p.logger.Warnf(
		"Total %d %s to be rolled back: \n",
		migrations.Len(),
		plural.Migration(migrations.Len()),
	)

	p.PrintMigrations(migrations, true)
}

// AskRollbackConfirmation returns a confirmation question for reverting a release.
func (p *MigrationPresenter) AskRollbackConfirmation(count int) string {
	return fmt.Sprintf("Rollback the above %d %s?", count, plural.Migration(count))
}

// ShowRollbackSuccess displays a success message after release has been rolled back.
func (p *MigrationPresenter) ShowRollbackSuccess(count int) {
	p.logger.Successf("%d %s rolled back\n", count, plural.MigrationWas(count))
	p.logger.Success("Rollback completed successfully")
}

// ShowRollbackError displays a message when rollback operation failed.
func (p *MigrationPresenter) ShowRollbackError() {
	p.logger.Error("Rollback failed. Some changes may have been partially reverted.\n")
}

// ShowNoMigrationsToRollback displays a message when there are no migrations to roll back.
func (p *MigrationPresenter) ShowNoMigrationsToRollback() {
	p.logger.Success("No migrations to roll back.")
}

// ShowMissingDownFiles displays a message about missing down migration files.
func (p *MigrationPresenter) ShowMissingDownFiles(versions []string) {
	p.logger.Error("Cannot rollback: missing down migration files for versions:\n")
	for _, v := range versions {
		p.logger.Errorf("\t%s\n", v)
	}
}
