/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package action

import (
	"errors"
)

var (
	ErrInvalidFileName = errors.New("the migration name should contain letters, digits, underscore and/or backslash characters only")
)
