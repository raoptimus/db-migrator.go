/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"github.com/raoptimus/db-migrator.go/internal/helper/console"
)

const (
	defaultUpgradeLimit    = 0
	migratedUpSuccessfully = "Migrated up successfully"
)

// Upgrade handles the application of pending migrations to the database.
type Upgrade struct {
	options         *Options
	presenter       Presenter
	fileNameBuilder FileNameBuilder
}

// NewUpgrade creates a new Upgrade handler instance.
func NewUpgrade(
	options *Options,
	presenter Presenter,
	fileNameBuilder FileNameBuilder,
) *Upgrade {
	return &Upgrade{
		options:         options,
		presenter:       presenter,
		fileNameBuilder: fileNameBuilder,
	}
}

// Handle processes the upgrade command to apply pending migrations.
func (u *Upgrade) Handle(cmd *Command, svc MigrationService) error {
	limit, err := stepOrDefault(cmd, defaultUpgradeLimit)
	if err != nil {
		return err
	}

	migrations, err := svc.NewMigrations(cmd.Context())
	if err != nil {
		return err
	}

	totalNewMigrations := migrations.Len()
	if totalNewMigrations == 0 {
		u.presenter.ShowNoNewMigrations()
		return nil
	}

	if limit > 0 && migrations.Len() > limit {
		migrations = migrations[:limit]
	}

	u.presenter.ShowUpgradePlan(migrations, totalNewMigrations)

	question := u.presenter.AskUpgradeConfirmation(migrations.Len())
	if u.options.Interactive && !console.Confirm(question) {
		return nil
	}

	var applied int
	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := u.fileNameBuilder.Up(migration.Version, false)

		if err := svc.ApplyFile(cmd.Context(), migration, fileName, safely); err != nil {
			u.presenter.ShowUpgradeError(applied, migrations.Len())
			return err
		}

		applied++
	}

	u.presenter.ShowUpgradeSuccess(migrations.Len())
	return nil
}
