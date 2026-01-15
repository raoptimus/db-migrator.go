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

// Redo handles the reverting and reapplying of previously applied migrations.
type Redo struct {
	options         *Options
	presenter       Presenter
	fileNameBuilder FileNameBuilder
}

// NewRedo creates a new Redo handler instance.
func NewRedo(
	options *Options,
	presenter Presenter,
	fileNameBuilder FileNameBuilder,
) *Redo {
	return &Redo{
		options:         options,
		presenter:       presenter,
		fileNameBuilder: fileNameBuilder,
	}
}

// Handle processes the redo command to revert and reapply migrations.
func (r *Redo) Handle(cmd *Command, svc MigrationService) error {
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
		r.presenter.ShowNoMigrationsToRevert()
		return nil
	}

	r.presenter.ShowRedoPlan(migrations)

	question := r.presenter.AskRedoConfirmation(migrationsCount)
	if r.options.Interactive && !console.Confirm(question) {
		return nil
	}

	reversedMigrations := make(model.Migrations, 0, migrationsCount)
	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := r.fileNameBuilder.Down(migration.Version, false)

		if err := svc.RevertFile(cmd.Context(), migration, fileName, safely); err != nil {
			r.presenter.ShowRedoError()
			return err
		}

		reversedMigrations = append(reversedMigrations, migrations[i])
	}

	for i := migrationsCount - 1; i >= 0; i-- {
		migration := &reversedMigrations[i]
		fileName, safely := r.fileNameBuilder.Up(migration.Version, false)

		if err := svc.ApplyFile(cmd.Context(), migration, fileName, safely); err != nil {
			r.presenter.ShowRedoError()
			return err
		}
	}

	r.presenter.ShowRedoSuccess(migrationsCount)
	return nil
}
