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

type Stmt interface {
	ExecContext(ctx context.Context, args ...any) (Result, error)
}

type stmt struct {
	*sql.Stmt
}

//nolint:ireturn,nolintlint // its ok
func (s *stmt) ExecContext(ctx context.Context, args ...any) (Result, error) {
	return s.Stmt.ExecContext(ctx, args...)
}
