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

func (s *MigrateController) History(limit string) error {
	limitInt, err := parseLimit(limit, 10)
	if err != nil {
		return err
	}
	hist, err := s.migration.GetMigrationHistory(limitInt)
	if err != nil {
		return err
	}
	n := hist.Len()
	if n == 0 {
		fmt.Println(console.Green("No migration has been done before."))
		return nil
	}

	if limitInt > 0 {
		fmt.Printf(console.Yellow("Showing the last %d %s: \n"),
			n, console.NumberPlural(n, "migration", "migrations"))
	} else {
		fmt.Printf(console.Yellow("Total %d %s been applied before: \n"),
			n, console.NumberPlural(n, "migration has", "migrations have"))
	}

	printAllMigrations(hist, true)

	return nil
}
