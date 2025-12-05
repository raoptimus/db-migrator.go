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
	"fmt"

	"github.com/raoptimus/db-migrator.go/internal/args"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

type Redo struct {
	console         Console
	service         MigrationService
	fileNameBuilder FileNameBuilder
	interactive     bool
}

func NewRedo(
	console Console,
	service MigrationService,
	fileNameBuilder FileNameBuilder,
	interactive bool,
) *Redo {
	return &Redo{
		console:         console,
		service:         service,
		fileNameBuilder: fileNameBuilder,
		interactive:     interactive,
	}
}

func (r *Redo) Run(ctx context.Context, cmdArgs ...string) error {
	arg := ""
	if len(cmdArgs) > 0 {
		arg = cmdArgs[0]
	}
	limit, err := args.ParseStepStringOrDefault(arg, minLimit)
	if err != nil {
		return err
	}

	migrations, err := r.service.Migrations(ctx, limit)
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		r.console.SuccessLn("No migration has been done before.")
		return nil
	}

	r.console.Warnf(
		"Total %d %s to be redone: \n",
		migrationsCount,
		r.console.NumberPlural(migrationsCount, "migration", "migrations"),
	)

	printMigrations(migrations, false)

	question := fmt.Sprintf("Redo the above %d %s?",
		migrationsCount, r.console.NumberPlural(migrationsCount, "migration", "migrations"),
	)
	if r.interactive && !r.console.Confirm(question) {
		return nil
	}

	reversedMigrations := make(entity.Migrations, 0, migrationsCount)
	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := r.fileNameBuilder.Down(migration.Version, false)

		if err := r.service.RevertFile(ctx, migration, fileName, safely); err != nil {
			r.console.ErrorLn("Migration failed. The rest of the migrations are canceled.")
			return err
		}

		reversedMigrations = append(reversedMigrations, migrations[i])
	}

	for i := migrationsCount - 1; i >= 0; i-- {
		migration := &reversedMigrations[i]
		fileName, safely := r.fileNameBuilder.Up(migration.Version, false)

		if err := r.service.ApplyFile(ctx, migration, fileName, safely); err != nil {
			r.console.ErrorLn("Migration failed. The rest of the migrations are canceled.\n")
			return err
		}
	}

	r.console.Warnf(
		"%d %s redone.",
		migrationsCount,
		r.console.NumberPlural(migrationsCount, migrationWas, migrationsWere),
	)
	r.console.SuccessLn("Migration redone successfully.\n")

	return nil
}
