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

func (s *Service) Up(limit string) error {
	limitInt, err := parseLimit(limit, 0)
	if err != nil {
		return err
	}
	hist, err := s.migration.GetNewMigrations(limitInt)
	if err != nil {
		return err
	}
	total := hist.Len()
	if total == 0 {
		fmt.Println(console.Green("No new migrations found. Your system is up-to-date."))
		return nil
	}
	if limitInt > 0 && len(hist) > limitInt {
		hist = hist[:limitInt]
	}
	n := hist.Len()
	if n == total {
		fmt.Printf(console.Yellow("Total %d new %s to be applied: \n"),
			n, console.NumberPlural(n, "migration", "migrations"))
	} else {
		fmt.Printf(console.Yellow("Total %d out of %d new %s to be applied: \n"),
			n, total, console.NumberPlural(total, "migration", "migrations"))
	}

	printAllMigrations(hist, false)

	applied := 0
	question := fmt.Sprintf("Apply the above %s?",
		console.NumberPlural(n, "migration", "migrations"),
	)
	if s.options.Interactive && !console.Confirm(question) {
		return nil
	}

	for _, item := range hist {
		if err := s.migration.MigrateUp(item); err != nil {
			return fmt.Errorf(
				"%v\n%d from %d %s applied.\nMigration failed. The rest of the migrations are canceled.",
				err, applied, n, console.NumberPlural(applied, "migration was", "migrations were"),
			)
		}

		applied++
	}

	log.Printf(console.Green("%d %s applied"),
		n, console.NumberPlural(n, "migration was", "migrations were"))
	fmt.Println(console.Green("Migrated up successfully"))

	return nil
}
