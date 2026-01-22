/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

// Options represents configuration options for creating a repository instance.
type Options struct {
	// TableName is the name of the migration history table.
	TableName string
	// SchemaName is the database schema name where the migration table resides.
	SchemaName string
	// ClusterName is the ClickHouse cluster name for cluster-aware migrations.
	ClusterName string
	// Replicated indicates whether to use replicated tables in ClickHouse.
	Replicated bool
}
