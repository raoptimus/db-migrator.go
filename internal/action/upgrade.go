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
)

const (
	defaultUpgradeLimit = 0
)

type Upgrade struct {
	console         Console
	service         MigrationService
	fileNameBuilder FileNameBuilder
	interactive     bool
}

func NewUpgrade(
	console Console,
	service MigrationService,
	fileNameBuilder FileNameBuilder,
	interactive bool,
) *Upgrade {
	return &Upgrade{
		console:         console,
		service:         service,
		fileNameBuilder: fileNameBuilder,
		interactive:     interactive,
	}
}

func (u *Upgrade) Run(ctx context.Context, cmdArgs ...string) error {
	limit, err := args.ParseStepStringOrDefault(cmdArgs[0], defaultUpgradeLimit)
	if err != nil {
		return err
	}

	migrations, err := u.service.NewMigrations(ctx)
	if err != nil {
		return err
	}

	totalNewMigrations := migrations.Len()
	if totalNewMigrations == 0 {
		u.console.SuccessLn(noNewMigrationsFound)
		return nil
	}

	if limit > 0 && migrations.Len() > limit {
		migrations = migrations[:limit]
	}

	if migrations.Len() == totalNewMigrations {
		u.console.Warnf(
			"Total %d new %s to be applied: \n",
			migrations.Len(),
			u.console.NumberPlural(migrations.Len(), "migration", "migrations"),
		)
	} else {
		u.console.Warnf(
			"Total %d out of %d new %s to be applied: \n",
			migrations.Len(),
			totalNewMigrations,
			u.console.NumberPlural(totalNewMigrations, "migration", "migrations"),
		)
	}

	printMigrations(migrations, false)

	question := fmt.Sprintf("Apply the above %s?",
		u.console.NumberPlural(migrations.Len(), "migration", "migrations"),
	)
	if u.interactive && !u.console.Confirm(question) {
		return nil
	}

	var applied int
	for i := range migrations {
		migration := &migrations[i]
		fileName, safely := u.fileNameBuilder.Up(migration.Version, false)

		if err := u.service.ApplyFile(ctx, migration, fileName, safely); err != nil {
			u.console.Errorf("%d from %d %s applied.\n",
				applied,
				migrations.Len(),
				u.console.NumberPlural(applied, migrationWas, migrationsWere),
			)
			u.console.Error("The rest of the migrations are canceled.\n")

			return err
		}

		applied++
	}

	u.console.Successf(
		"%d %s applied\n",
		migrations.Len(),
		u.console.NumberPlural(migrations.Len(), migrationWas, migrationsWere),
	)
	u.console.SuccessLn(migratedUpSuccessfully)

	return nil
}
