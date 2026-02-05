/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/raoptimus/db-migrator.go/internal/helper/console"
)

// To handles the migration to a specific version.
type To struct {
	options         *Options
	presenter       Presenter
	fileNameBuilder FileNameBuilder
}

// NewTo creates a new To handler instance.
func NewTo(
	options *Options,
	presenter Presenter,
	fileNameBuilder FileNameBuilder,
) *To {
	return &To{
		options:         options,
		presenter:       presenter,
		fileNameBuilder: fileNameBuilder,
	}
}

// Handle processes the to command to migrate to a specific version.
func (t *To) Handle(cmd *Command, svc MigrationService) error {
	// Get and parse target version
	if !cmd.Args.Present() {
		return ErrTargetVersionRequired
	}

	targetInput := cmd.Args.First()
	targetVersion, err := parseTargetVersion(targetInput)
	if err != nil {
		return err
	}

	// Check if target version is already applied by getting all applied migrations
	// and checking if any migration's timestamp matches the target
	allAppliedMigrations, err := svc.Migrations(cmd.Context(), 0)
	if err != nil {
		return err
	}

	targetApplied := false
	for _, m := range allAppliedMigrations {
		if extractTimestamp(m.Version) == targetVersion {
			targetApplied = true
			break
		}
	}

	// Determine migration direction
	if !targetApplied {
		// Direction: UP (apply migrations)
		return t.handleUpgrade(cmd, svc, targetVersion)
	}

	// Direction: DOWN (revert migrations)
	return t.handleDowngrade(cmd, svc, targetVersion)
}

// handleUpgrade applies all migrations up to and including the target version.
func (t *To) handleUpgrade(cmd *Command, svc MigrationService, targetVersion string) error {
	// Get all new migrations
	allNewMigrations, err := svc.NewMigrations(cmd.Context())
	if err != nil {
		return err
	}

	// Filter: keep only migrations <= targetVersion
	migrationsToApply := make(model.Migrations, 0)
	for _, m := range allNewMigrations {
		// Extract timestamp part for comparison (e.g., "251002_184510_change_scheme" -> "251002_184510")
		migrationTimestamp := extractTimestamp(m.Version)
		if migrationTimestamp <= targetVersion {
			migrationsToApply = append(migrationsToApply, m)
		}
	}

	// Check if there are migrations to apply
	if len(migrationsToApply) == 0 {
		t.presenter.ShowNoNewMigrations()
		return nil
	}

	// Sort by version (ASC)
	migrationsToApply.SortByVersion()

	// Show migration plan
	t.presenter.ShowUpgradePlan(migrationsToApply, len(migrationsToApply))

	// User confirmation
	question := t.presenter.AskUpgradeConfirmation(len(migrationsToApply))
	if t.options.Interactive && !console.Confirm(question) {
		return nil
	}

	// Apply migrations using helper function
	applied, err := applyMigrations(cmd, svc, t.presenter, t.fileNameBuilder, migrationsToApply)
	if err != nil {
		return err
	}

	t.presenter.ShowUpgradeSuccess(applied)
	return nil
}

// handleDowngrade reverts all migrations after the target version.
func (t *To) handleDowngrade(cmd *Command, svc MigrationService, targetVersion string) error {
	// Get all applied migrations (unlimited)
	allAppliedMigrations, err := svc.Migrations(cmd.Context(), 0)
	if err != nil {
		return err
	}

	// Filter: keep only migrations > targetVersion
	migrationsToRevert := make(model.Migrations, 0)
	for _, m := range allAppliedMigrations {
		// Extract timestamp part for comparison (e.g., "251002_184510_change_scheme" -> "251002_184510")
		migrationTimestamp := extractTimestamp(m.Version)
		if migrationTimestamp > targetVersion {
			migrationsToRevert = append(migrationsToRevert, m)
		}
	}

	// Check if there are migrations to revert
	if len(migrationsToRevert) == 0 {
		// Already at target version
		t.presenter.ShowNoMigrationsToRevert()
		return nil
	}

	// Migrations are already sorted DESC from DB (correct order for rollback)

	// Show migration plan
	t.presenter.ShowDowngradePlan(migrationsToRevert)

	// User confirmation
	question := t.presenter.AskDowngradeConfirmation(len(migrationsToRevert))
	if t.options.Interactive && !console.Confirm(question) {
		return nil
	}

	// Revert migrations using helper function
	reverted, err := revertMigrations(cmd, svc, t.presenter, t.fileNameBuilder, migrationsToRevert)
	if err != nil {
		return err
	}

	t.presenter.ShowDowngradeSuccess(reverted)
	return nil
}
