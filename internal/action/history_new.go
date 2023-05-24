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

	migrations, err := h.service.NewMigrations(ctx.Context, limit)
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		console.SuccessLn(noNewMigrationsFound)
		return nil
	}

	if limit > 0 {
		console.Warnf(
			"Showing the last %d %s: \n",
			migrationsCount,
			console.NumberPlural(migrationsCount, "migration", "migrations"),
		)
	} else {
		console.Warnf(
			"Total %d %s been applied before: \n",
			migrationsCount,
			console.NumberPlural(migrationsCount, "migration has", "migrations have"),
		)
	}

	printMigrations(migrations, true)
	return nil
}
