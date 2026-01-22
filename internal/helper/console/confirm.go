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
	"fmt"
	"os"
	"strings"
)

// Confirm prompts the user for confirmation with the given format string.
// It returns true if the user responds with "y" or "yes", false for "n" or "no".
func Confirm(format string) bool {
	return Confirmf(format)
}

// Confirmf prompts the user for confirmation with a formatted message.
// It accepts format and arguments similar to fmt.Printf.
// It returns true for affirmative responses ("y", "yes") and false otherwise.
func Confirmf(format string, args ...any) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf(format+" [y/n]: ", args...)

	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	response = strings.ToLower(strings.TrimSpace(response))

	return response == "y" || response == "yes"
}
