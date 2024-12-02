/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package action

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/args"
	"github.com/raoptimus/db-migrator.go/internal/console"
)

type HistoryNew struct {
	service MigrationService
}

func NewHistoryNew(
	service MigrationService,
) *HistoryNew {
	return &HistoryNew{
		service: service,
	}
}

func (h *HistoryNew) Run(ctx context.Context, cmdArgs ...string) error {
	limit, err := args.ParseStepStringOrDefault(cmdArgs[0], defaultGetHistoryLimit)
	if err != nil {
		return err
	}

	migrations, err := h.service.NewMigrations(ctx)
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		console.SuccessLn(noNewMigrationsFound)
		return nil
	}

	if limit > 0 && migrationsCount > limit {
		migrations = migrations[:limit]
		console.Warnf(
			"Showing %d out of %d new %s \n",
			limit,
			migrationsCount,
			console.NumberPlural(migrationsCount, "migration", "migrations"),
		)
	} else {
		console.Warnf(
			"Found %d new %s \n",
			migrationsCount,
			console.NumberPlural(migrationsCount, "migration", "migrations"),
		)
	}

	printMigrations(migrations, true)

	return nil
}
