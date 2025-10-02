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

type DB struct {
	*sql.DB
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (Rows, error) {
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return NewRowsBySQLRows(rows), nil
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (Result, error) {
	return db.DB.ExecContext(ctx, query, args...)
}
