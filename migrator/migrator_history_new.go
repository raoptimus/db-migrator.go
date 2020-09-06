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
)

func (s *MigrateController) HistoryNew(limit string) error {
	limitInt, err := parseLimit(limit, 10)
	if err != nil {
		return err
	}
	hist, err := s.migration.GetNewMigrations(limitInt)
	if err != nil {
		return err
	}
	n := hist.Len()
	if n == 0 {
		fmt.Println(console.Green("No new migrations found. Your system is up-to-date."))
		return nil
	}

	if limitInt > 0 && n > limitInt {
		hist = hist[0:limitInt]
		fmt.Printf(console.Yellow("Showing %d out of %d new %s: \n"),
			limitInt, n, console.NumberPlural(n, "migration", "migrations"))
	} else {
		fmt.Printf(console.Yellow("Found %d new %s: \n"),
			n, console.NumberPlural(n, "migration", "migrations"))
	}

	printAllMigrations(hist, false)

	return nil
}
