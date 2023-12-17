/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package action

import (
	"fmt"

	"github.com/raoptimus/db-migrator.go/internal/args"
	"github.com/raoptimus/db-migrator.go/internal/console"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
	"github.com/urfave/cli/v2"
)

type Redo struct {
	service         MigrationService
	fileNameBuilder FileNameBuilder
	interactive     bool
}

func NewRedo(
	service MigrationService,
	fileNameBuilder FileNameBuilder,
	interactive bool,
) *Redo {
	return &Redo{
		service:         service,
		fileNameBuilder: fileNameBuilder,
		interactive:     interactive,
	}
}

func (r *Redo) Run(ctx *cli.Context) error {
	limit, err := args.ParseStepStringOrDefault(ctx.Args().Get(0), 1)
	if err != nil {
		return err
	}

	migrations, err := r.service.Migrations(ctx.Context, limit)
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		console.SuccessLn("No migration has been done before.")
		return nil
	}

	console.Warnf(
		"Total %d %s to be redone: \n",
		migrationsCount,
		console.NumberPlural(migrationsCount, "migration", "migrations"),
	)

	printMigrations(migrations, false)

	question := fmt.Sprintf("Redo the above %d %s?",
		migrationsCount, console.NumberPlural(migrationsCount, "migration", "migrations"),
	)
	if r.interactive && !console.Confirm(question) {
		return nil
	}

	reversedMigrations := make(entity.Migrations, 0, len(migrations))
	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := r.fileNameBuilder.Down(migration.Version, false)

		if err := r.service.RevertFile(ctx.Context, migration, fileName, safely); err != nil {
			console.ErrorLn("Migration failed. The rest of the migrations are canceled.")
			return err
		}

		reversedMigrations = append(reversedMigrations, migrations[i])
	}

	for i := range reversedMigrations {
		migration := &reversedMigrations[i]
		fileName, safely := r.fileNameBuilder.Up(migration.Version, false)

		if err := r.service.ApplyFile(ctx.Context, migration, fileName, safely); err != nil {
			console.ErrorLn("Migration failed. The rest of the migrations are canceled.\n")
			return err
		}
	}

	console.Warnf(
		"%d %s redone.",
		migrationsCount,
		console.NumberPlural(migrationsCount, migrationWas, migrationsWere),
	)
	console.SuccessLn("Migration redone successfully.\n")
	return nil
}
