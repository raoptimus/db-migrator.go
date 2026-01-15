/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package sqlex

import (
	"database/sql"
)

// Result extends the standard database/sql.Result interface to provide
// a common abstraction for query execution results across different database drivers.
type Result interface {
	sql.Result
}
