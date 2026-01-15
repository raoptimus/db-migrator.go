/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dbmigrator

import "github.com/pkg/errors"

var (
	// ErrMigrationAlreadyExists indicates that a migration with the specified version already exists in the database.
	ErrMigrationAlreadyExists = errors.New("migration already exists")

	// ErrAppliedMigrationNotFound indicates that the requested migration was not found in the migration history.
	ErrAppliedMigrationNotFound = errors.New("applied migration not found")
)
