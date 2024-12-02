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
	"github.com/raoptimus/db-migrator.go/internal/console"
	"github.com/urfave/cli/v3"
)

type Downgrade struct {
	service         MigrationService
	fileNameBuilder FileNameBuilder
	interactive     bool
}

func NewDowngrade(
	service MigrationService,
	fileNameBuilder FileNameBuilder,
	interactive bool,
) *Downgrade {
	return &Downgrade{
		service:         service,
		fileNameBuilder: fileNameBuilder,
		interactive:     interactive,
	}
}

func (d *Downgrade) Run(ctx context.Context, cmdArgs cli.Args) error {
	limit, err := args.ParseStepStringOrDefault(cmdArgs.Get(0), minLimit)
	if err != nil {
		return err
	}

	migrations, err := d.service.Migrations(ctx, limit)
	if err != nil {
		return err
	}

	migrationsCount := migrations.Len()
	if migrationsCount == 0 {
		console.SuccessLn("No migration has been done before.")
		return nil
	}

	console.Warnf(
		"Total %d %s to be reverted: \n",
		migrationsCount,
		console.NumberPlural(migrationsCount, "migration", "migrations"),
	)

	printMigrations(migrations, false)

	reverted := 0
	question := fmt.Sprintf("RevertFile the above %d %s?",
		migrationsCount,
		console.NumberPlural(migrationsCount, "migration", "migrations"),
	)
	if d.interactive && !console.Confirm(question) {
		return nil
	}

	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := d.fileNameBuilder.Down(migration.Version, false)

		if err := d.service.RevertFile(ctx, migration, fileName, safely); err != nil {
			console.Errorf(
				"%d from %d %s reverted.\n"+
					"Migration failed. The rest of the migrations are canceled.\n",
				reverted,
				migrationsCount,
				console.NumberPlural(reverted, migrationWas, migrationsWere),
			)
			return err
		}

		reverted++
	}

	console.Successf(
		"%d %s reverted\n",
		migrationsCount,
		console.NumberPlural(migrationsCount, migrationWas, migrationsWere),
	)
	console.SuccessLn("Migrated down successfully\n")
	return nil
}
