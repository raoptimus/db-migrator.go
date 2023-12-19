/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package action

import (
	"github.com/raoptimus/db-migrator.go/internal/console"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

func printMigrations(migrations entity.Migrations, withTime bool) {
	for _, migration := range migrations {
		//todo: check len of version name
		if withTime {
			console.Infof("\t(%s) %s\n", migration.ApplyTimeFormat(), migration.Version)
		} else {
			console.Infof("\t%s\n", migration.Version)
		}
	}

	console.InfoLn("")
}
