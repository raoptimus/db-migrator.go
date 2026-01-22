/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dbmigrator

import (
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
)

// NewConnection creates a new database connection from the provided DSN string.
// Supported drivers: clickhouse, postgres, mysql, tarantool.
//
// DSN format: driver://username:password@host:port/dbname?options
//
// Examples:
//   - PostgreSQL: postgres://user:pass@localhost:5432/mydb?sslmode=disable
//   - MySQL: mysql://user:pass@localhost:3306/mydb
//   - ClickHouse: clickhouse://user:pass@localhost:9000/mydb?compress=true
//   - Tarantool: tarantool://user:pass@localhost:3301/mydb
//
//nolint:ireturn // intentionally returns interface to hide internal implementation
func NewConnection(dsn string) (Connection, error) {
	return connection.New(dsn)
}

// TryConnection attempts to create and ping a database connection with retries.
// It will retry up to maxAttempts times with a 1-second delay between attempts.
// If maxAttempts is less than 1, it defaults to 1.
//
//nolint:ireturn // intentionally returns interface to hide internal implementation
func TryConnection(dsn string, maxAttempts int) (Connection, error) {
	return connection.Try(dsn, maxAttempts)
}
