/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package console

import (
	"bufio"
	"os"
	"strings"
)

func Confirm(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		Infof("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}
