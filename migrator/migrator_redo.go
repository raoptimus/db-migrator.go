/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package migrator

import (
	"fmt"
	"github.com/raoptimus/db-migrator.go/console"
	"github.com/raoptimus/db-migrator.go/migrator/db"
	"log"
)

func (s *Service) Redo(limit string) error {
	limitInt, err := parseLimit(limit, 1)
	if err != nil {
		return err
	}
	entityList, err := s.migration.GetMigrationHistory(limitInt)
	if err != nil {
		return err
	}
	n := entityList.Len()
	if n == 0 {
		log.Println(console.Green("No migration has been done before."))
		return nil
	}

	fmt.Printf(console.Yellow("Total %d %s to be redone: \n"),
		n, console.NumberPlural(n, "migration", "migrations"))

	printAllMigrations(entityList, false)

	question := fmt.Sprintf("Redo the above %d %s?",
		n, console.NumberPlural(n, "migration", "migrations"),
	)
	if s.options.Interactive && !console.Confirm(question) {
		return nil
	}

	var reverseHist db.MigrationEntityList
	for _, entity := range entityList {
		fileName, safely := s.fileBuilder.BuildDownFileName(entity.Version, true)
		if err := s.migration.MigrateDown(entity, fileName, safely); err != nil {
			return fmt.Errorf(
				"%v\nMigration failed. The rest of the migrations are canceled.", err,
			)
		}

		reverseHist = append(reverseHist, entity)
	}

	for _, entity := range reverseHist {
		fileName, safely := s.fileBuilder.BuildUpFileName(entity.Version, true)
		if err := s.migration.MigrateUp(entity, fileName, safely); err != nil {
			return fmt.Errorf(
				"%v\nMigration failed. The rest of the migrations are canceled.", err,
			)
		}
	}

	log.Printf(console.Green("%d %s redone."),
		n, console.NumberPlural(n, "migration was", "migrations were"))
	fmt.Println(console.Green("Migration redone successfully."))

	return nil
}
