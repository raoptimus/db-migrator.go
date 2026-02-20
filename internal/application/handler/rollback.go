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

	"github.com/raoptimus/db-migrator.go/internal/helper/console"
)

// Rollback handles the reverting of the latest release batch atomically in a single transaction.
// It identifies the batch by MAX(apply_time) and reverts all migrations in that batch.
type Rollback struct {
	options         *Options
	presenter       Presenter
	fileNameBuilder FileNameBuilder
}

// NewRollback creates a new Rollback handler instance.
func NewRollback(
	options *Options,
	presenter Presenter,
	fileNameBuilder FileNameBuilder,
) *Rollback {
	return &Rollback{
		options:         options,
		presenter:       presenter,
		fileNameBuilder: fileNameBuilder,
	}
}

// Handle processes the rollback command to revert the latest release batch.
func (r *Rollback) Handle(cmd *Command, svc MigrationService) error {
	migrations, err := svc.LatestReleaseMigrations(cmd.Context())
	if err != nil {
		return err
	}

	if migrations.Len() == 0 {
		r.presenter.ShowNoMigrationsToRevert()
		return nil
	}

	// Check all down files exist before proceeding
	var missingVersions []string
	for i := range migrations {
		fileName, _ := r.fileNameBuilder.Down(migrations[i].Version, false)
		exists, err := svc.FileExists(fileName)
		if err != nil {
			return err
		}
		if !exists {
			missingVersions = append(missingVersions, migrations[i].Version)
		}
	}

	if len(missingVersions) > 0 {
		r.presenter.ShowMissingDownFiles(missingVersions)
		return ErrMissingDownFiles
	}

	r.presenter.ShowDowngradePlan(migrations)

	question := r.presenter.AskDowngradeConfirmation(migrations.Len())
	if r.options.Interactive && !console.Confirm(question) {
		return nil
	}

	err = svc.ExecInTransaction(cmd.Context(), func(ctx context.Context) error {
		for i := range migrations {
			migration := &migrations[i]
			fileName, _ := r.fileNameBuilder.Down(migration.Version, false)

			if err := svc.RevertFile(ctx, migration, fileName, false); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		r.presenter.ShowRollbackError()
		return err
	}

	r.presenter.ShowDowngradeSuccess(migrations.Len())
	return nil
}
