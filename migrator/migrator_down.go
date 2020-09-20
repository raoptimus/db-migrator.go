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
	"log"
)

func (s *Service) Down(limit string) error {
	limitInt, err := parseLimit(limit, 1)
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

	fmt.Printf(console.Yellow("Total %d %s to be reverted: \n"),
		n, console.NumberPlural(n, "migration", "migrations"))

	printAllMigrations(hist, false)

	reverted := 0
	question := fmt.Sprintf("Revert the above %d %s?",
		n, console.NumberPlural(n, "migration", "migrations"),
	)
	if s.options.Interactive && !console.Confirm(question) {
		return nil
	}

	for _, item := range hist {
		if err := s.migration.MigrateDown(item); err != nil {
			return fmt.Errorf(
				"%v\n%d from %d %s reverted.\nMigration failed. "+
					"Migration failed. The rest of the migrations are canceled.",
				err, reverted, n, console.NumberPlural(reverted, "migration was", "migrations were"),
			)
		}

		reverted++
	}

	log.Printf(console.Green("%d %s reverted"),
		n, console.NumberPlural(n, "migration was", "migrations were"))
	fmt.Println(console.Green("Migrated down successfully"))

	return nil
}
