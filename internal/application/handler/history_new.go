/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

// HistoryNew handles the display of pending migrations that have not been applied yet.
type HistoryNew struct {
	options   *Options
	presenter Presenter
}

// NewHistoryNew creates a new HistoryNew handler instance.
func NewHistoryNew(
	options *Options,
	presenter Presenter,
) *HistoryNew {
	return &HistoryNew{
		options:   options,
		presenter: presenter,
	}
}

// Handle processes the new command to display pending migrations.
func (h *HistoryNew) Handle(cmd *Command, svc MigrationService) error {
	limit, err := stepOrDefault(cmd, defaultGetHistoryLimit)
	if err != nil {
		return err
	}

	migrations, err := svc.NewMigrations(cmd.Context())
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		h.presenter.ShowNoNewMigrations()
		return nil
	}

	if limit > 0 && migrationsCount > limit {
		h.presenter.ShowNewMigrationsLimitedHeader(limit, migrationsCount)
		migrations = migrations[:limit]
	} else {
		h.presenter.ShowNewMigrationsHeader(migrationsCount)
	}

	h.presenter.PrintMigrations(migrations, false)

	return nil
}
