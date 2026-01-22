/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package sqlex

import (
	"context"
	"database/sql"
)

// Stmt abstracts a prepared statement for executing queries with bound parameters.
type Stmt interface {
	ExecContext(ctx context.Context, args ...any) (Result, error)
}

// stmt wraps the standard database/sql.Stmt to implement the custom Stmt interface.
type stmt struct {
	*sql.Stmt
}

// ExecContext executes a prepared statement with the given arguments and returns a Result.
//
//nolint:ireturn,nolintlint // its ok
func (s *stmt) ExecContext(ctx context.Context, args ...any) (Result, error) {
	return s.Stmt.ExecContext(ctx, args...)
}
