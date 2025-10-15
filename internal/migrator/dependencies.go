/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package migrator

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/sqlex"
)

//go:generate mockery
type Connection interface {
	DSN() string
	Driver() connection.Driver
	Ping() error
	QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error)
	Transaction(ctx context.Context, txFn func(ctx context.Context) error) error
}

//go:generate mockery
type FileNameBuilder interface {
	// Up builds a file name for migration update
	Up(version string, forceSafely bool) (fname string, safely bool)
	// Down builds a file name for migration downgrade
	Down(version string, forceSafely bool) (fname string, safely bool)
}
