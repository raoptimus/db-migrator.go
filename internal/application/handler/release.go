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
	"time"

	"github.com/raoptimus/db-migrator.go/internal/helper/console"
)

// Release handles the application of all pending migrations atomically in a single transaction.
// All migrations in a release share the same apply_time, enabling batch rollback.
type Release struct {
	options         *Options
	presenter       Presenter
	fileNameBuilder FileNameBuilder
}

// NewRelease creates a new Release handler instance.
func NewRelease(
	options *Options,
	presenter Presenter,
	fileNameBuilder FileNameBuilder,
) *Release {
	return &Release{
		options:         options,
		presenter:       presenter,
		fileNameBuilder: fileNameBuilder,
	}
}

// Handle processes the release command to apply all pending migrations atomically.
func (r *Release) Handle(cmd *Command, svc MigrationService) error {
	migrations, err := svc.NewMigrations(cmd.Context())
	if err != nil {
		return err
	}

	if migrations.Len() == 0 {
		r.presenter.ShowNoNewMigrations()
		return nil
	}

	r.presenter.ShowReleasePlan(migrations)

	question := r.presenter.AskReleaseConfirmation(migrations.Len())
	if r.options.Interactive && !console.Confirm(question) {
		return nil
	}

	applyTime := time.Now().Unix()

	err = svc.ExecInTransaction(cmd.Context(), func(ctx context.Context) error {
		for i := range migrations {
			migration := &migrations[i]
			fileName, _ := r.fileNameBuilder.Up(migration.Version, false)

			if err := svc.ApplyFileWithApplyTime(ctx, migration, fileName, applyTime); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		r.presenter.ShowReleaseError()
		return err
	}

	r.presenter.ShowReleaseSuccess(migrations.Len())
	return nil
}
