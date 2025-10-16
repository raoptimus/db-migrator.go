/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package connection

type Driver string

const (
	DriverClickhouse Driver = "clickhouse"
	DriverMySQL      Driver = "mysql"
	DriverPostgres   Driver = "postgres"
	DriverTarantool  Driver = "tarantool"
)

func (d Driver) String() string {
	return string(d)
}
