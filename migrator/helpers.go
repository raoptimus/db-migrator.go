/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package migrator

import (
	"errors"
	"fmt"
	"github.com/raoptimus/db-migrator.go/migrator/db"
	"strconv"
)

func parseLimit(limit string, defaults int) (int, error) {
	switch limit {
	case "":
		return defaults, nil
	case "all":
		return 0, nil
	default:
		i, err := strconv.Atoi(limit)
		if err != nil {
			return -1, fmt.Errorf("The step argument %v is not valid", limit)
		}

		if i < 1 {
			return -1, errors.New("The step argument must be greater than 0.")
		}

		return i, nil
	}
}

func printAllMigrations(hist db.HistoryItems, withTime bool) {
	for _, item := range hist {
		//todo check len of version name
		if withTime {
			fmt.Printf("\t(%s) %s", item.ApplyTimeFormat(), item.Version)
		} else {
			fmt.Printf("\t%s", item.Version)
		}

	}
	fmt.Println("")
}
