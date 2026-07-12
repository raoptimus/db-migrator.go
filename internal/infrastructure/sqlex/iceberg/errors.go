/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package iceberg

import "github.com/pkg/errors"

// ErrNotSupported is returned by SQL-level methods (QueryContext, ExecContext, BeginTx)
// that are not on the critical path for Iceberg migrations.
// Actual DDL execution is handled by the repository layer (task 04).
var ErrNotSupported = errors.New("operation is not supported by iceberg sqlex driver; use repository layer")
