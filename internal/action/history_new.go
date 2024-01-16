/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package action

import (
	"github.com/raoptimus/db-migrator.go/internal/args"
	"github.com/raoptimus/db-migrator.go/internal/console"
	"github.com/urfave/cli/v2"
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

func (h *HistoryNew) Run(ctx *cli.Context) error {
	limit, err := args.ParseStepStringOrDefault(ctx.Args().Get(0), defaultGetHistoryLimit)
	if err != nil {
		return err
	}

	migrations, err := h.service.NewMigrations(ctx.Context)
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
