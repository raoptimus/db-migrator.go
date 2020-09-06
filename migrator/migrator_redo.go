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
	"github.com/raoptimus/db-migrator/console"
	"github.com/raoptimus/db-migrator/migrator/db"
	"log"
)

func (s *MigrateController) Redo(limit string) error {
	limitInt, err := parseLimit(limit)
	if err != nil {
		return err
	}
	hist, err := s.migration.GetMigrationHistory(limitInt)
	if err != nil {
		return err
	}
	n := hist.Len()
	if n == 0 {
		log.Println(console.Green("No migration has been done before."))
		return nil
	}

	fmt.Printf(console.Yellow("Total %d %s to be redone: \n"),
		n, console.NumberPlural(n, "migration", "migrations"))

	printAllMigrations(hist)

	question := fmt.Sprintf("Redo the above %d %s?",
		n, console.NumberPlural(n, "migration", "migrations"),
	)
	if s.options.Interactive && !console.Confirm(question) {
		return nil
	}

	var reverseHist db.HistoryItems
	for _, item := range hist {
		if err := s.migration.MigrateDown(item); err != nil {
			return fmt.Errorf(
				"%v\nMigration failed. The rest of the migrations are canceled.", err,
			)
		}

		reverseHist = append(reverseHist, item)
	}
	for _, item := range reverseHist {
		if err := s.migration.MigrateUp(item); err != nil {
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
