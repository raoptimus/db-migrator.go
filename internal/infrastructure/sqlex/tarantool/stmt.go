/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package tarantool

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
	"github.com/tarantool/go-tarantool/v2"
)

// Stmt wraps a Tarantool prepared statement and implements the sqlex.Stmt interface.
type Stmt struct {
	stmt *tarantool.Prepared
}

// ExecContext executes a prepared statement with the given arguments.
//
//nolint:ireturn,nolintlint // its ok
func (s *Stmt) ExecContext(ctx context.Context, args ...any) (sqlex.Result, error) {
	req := tarantool.NewExecutePreparedRequest(s.stmt).Context(ctx)
	if len(args) > 0 {
		req = req.Args(args)
	}

	_, err := s.stmt.Conn.Do(req).Get()
	if err != nil {
		return nil, err
	}

	return Done(true), nil
}
