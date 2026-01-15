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

// Downgrade handles the reverting of previously applied migrations.
type Downgrade struct {
	options         *Options
	presenter       Presenter
	fileNameBuilder FileNameBuilder
}

// NewDowngrade creates a new Downgrade handler instance.
func NewDowngrade(options *Options, presenter Presenter, fileNameBuilder FileNameBuilder) *Downgrade {
	return &Downgrade{
		options:         options,
		presenter:       presenter,
		fileNameBuilder: fileNameBuilder,
	}
}

// Handle processes the downgrade command to revert applied migrations.
func (d *Downgrade) Handle(cmd *Command, svc MigrationService) error {
	limit, err := stepOrDefault(cmd, minLimit)
	if err != nil {
		return err
	}

	migrations, err := svc.Migrations(cmd.Context(), limit)
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		d.presenter.ShowNoMigrationsToRevert()
		return nil
	}

	d.presenter.ShowDowngradePlan(migrations)

	question := d.presenter.AskDowngradeConfirmation(migrationsCount)
	if d.options.Interactive && !console.Confirm(question) {
		return nil
	}

	reverted := 0

	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := d.fileNameBuilder.Down(migration.Version, false)

		if err := svc.RevertFile(cmd.Context(), migration, fileName, safely); err != nil {
			d.presenter.ShowDowngradeError(reverted, migrationsCount)
			return err
		}

		reverted++
	}

	d.presenter.ShowDowngradeSuccess(migrationsCount)
	return nil
}
