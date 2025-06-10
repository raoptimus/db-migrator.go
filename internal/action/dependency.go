/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package action

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

//go:generate mockery
type Console interface {
	Confirm(s string) bool
	Info(message string)
	InfoLn(message string)
	Infof(message string, a ...any)
	Success(message string)
	SuccessLn(message string)
	Successf(message string, a ...any)
	Warn(message string)
	WarnLn(message string)
	Warnf(message string, a ...any)
	Error(message string)
	ErrorLn(message string)
	Errorf(message string, a ...any)
	Fatal(err error)
	NumberPlural(count int, one, many string) string
}

//go:generate mockery
type File interface {
	Create(filename string) error
	Exists(path string) (bool, error)
}

//go:generate mockery
type FileNameBuilder interface {
	// Up builds a file name for migration update
	Up(version string, forceSafely bool) (fname string, safely bool)
	// Down builds a file name for migration downgrade
	Down(version string, forceSafely bool) (fname string, safely bool)
}

//go:generate mockery
type MigrationService interface {
	// Migrations returns an entities of migrations
	Migrations(ctx context.Context, limit int) (entity.Migrations, error)
	// NewMigrations returns an entities of new migrations
	//todo: domain.Migrations
	NewMigrations(ctx context.Context) (entity.Migrations, error)
	// ApplyFile applies new migration
	//todo: domain.Migration
	ApplyFile(ctx context.Context, entity *entity.Migration, fileName string, safely bool) error
	// RevertFile reverts the migration
	RevertFile(ctx context.Context, entity *entity.Migration, fileName string, safely bool) error
}
