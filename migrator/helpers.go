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
	"github.com/raoptimus/db-migrator/migrator/db"
	"strconv"
)

func parseLimit(limit string) (int, error) {
	switch limit {
	case "":
		return 1, nil
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

func printAllMigrations(hist db.HistoryItems) {
	for _, item := range hist {
		//todo check len of version name
		fmt.Printf("\t%s", item.Version)
	}
	fmt.Println("")
}
