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
	ErrMigrationAlreadyExists   = errors.New("migration already exists")
	ErrAppliedMigrationNotFound = errors.New("applied migration not found")
)
