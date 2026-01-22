/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package config

// Options contains configuration parameters for database migration operations.
type Options struct {
	DSN                string
	MaxConnAttempts    int
	Directory          string
	TableName          string
	ClusterName        string
	Replicated         bool
	Compact            bool
	Interactive        bool
	MaxSQLOutputLength int
}
