/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package connection

// Driver represents a supported database driver type.
type Driver string

const (
	// DriverClickhouse represents the ClickHouse database driver.
	DriverClickhouse Driver = "clickhouse"

	// DriverMySQL represents the MySQL database driver.
	DriverMySQL Driver = "mysql"

	// DriverPostgres represents the PostgreSQL database driver.
	DriverPostgres Driver = "postgres"

	// DriverTarantool represents the Tarantool database driver.
	DriverTarantool Driver = "tarantool"
)

// String returns the string representation of the driver.
func (d Driver) String() string {
	return string(d)
}
