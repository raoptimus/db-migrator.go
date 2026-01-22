/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package service

// Options configures the Migration service behavior.
// It contains settings for SQL output formatting, migration file location, and database credentials.
type Options struct {
	// MaxSQLOutputLength limits the length of SQL statements in log output. Zero means no limit.
	MaxSQLOutputLength int
	// Directory is the path to the directory containing migration files.
	Directory string
	// Compact enables compact logging mode with reduced output verbosity.
	Compact bool

	// PlaceholderCustom used for placeholder replacement in migration SQL.
	PlaceholderCustom string
	// ClusterName is the database ClusterName used for placeholder replacement in migration SQL.
	ClusterName string
	// Username is the database username used for placeholder replacement in migration SQL.
	Username string
	// Password is the database password used for placeholder replacement in migration SQL.
	Password string
}
