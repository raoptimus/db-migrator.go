/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

const defaultGetHistoryLimit = 10

// History handles the display of applied migration history.
type History struct {
	options   *Options
	presenter Presenter
}

// NewHistory creates a new History handler instance.
func NewHistory(
	options *Options,
	presenter Presenter,
) *History {
	return &History{
		options:   options,
		presenter: presenter,
	}
}

// Handle processes the history command to display applied migrations.
func (h *History) Handle(cmd *Command, svc MigrationService) error {
	limit, err := stepOrDefault(cmd, defaultGetHistoryLimit)
	if err != nil {
		return err
	}

	migrations, err := svc.Migrations(cmd.Context(), limit)
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		h.presenter.ShowNoMigrationsToRevert()
		return nil
	}

	if limit > 0 {
		h.presenter.ShowHistoryHeader(migrationsCount)
	} else {
		h.presenter.ShowAllHistoryHeader(migrationsCount)
	}

	h.presenter.PrintMigrations(migrations, true)

	return nil
}
