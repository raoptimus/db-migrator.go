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
	"github.com/urfave/cli/v3"
)

const defaultGetHistoryLimit = 10

type History struct {
	service MigrationService
}

func NewHistory(
	service MigrationService,
) *History {
	return &History{
		service: service,
	}
}

func (h *History) Run(ctx context.Context, cmdArgs cli.Args) error {
	limit, err := args.ParseStepStringOrDefault(cmdArgs.Get(0), defaultGetHistoryLimit)
	if err != nil {
		return err
	}

	migrations, err := h.service.Migrations(ctx, limit)
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		console.SuccessLn("No migration has been done before.")
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
