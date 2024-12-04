package dbmigrator

import "github.com/pkg/errors"

var (
	ErrMigrationAlreadyExists   = errors.New("migration already exists")
	ErrAppliedMigrationNotFound = errors.New("applied migration not found")
)
