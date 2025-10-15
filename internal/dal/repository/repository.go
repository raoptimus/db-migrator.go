/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

type Repository interface {
	Migrations(ctx context.Context, limit int) (entity.Migrations, error)
	HasMigrationHistoryTable(ctx context.Context) (exists bool, err error)
	InsertMigration(ctx context.Context, version string) error
	RemoveMigration(ctx context.Context, version string) error
	ExecQuery(ctx context.Context, query string, args ...any) error
	QueryScalar(ctx context.Context, query string, ptr any) error
	ExecQueryTransaction(ctx context.Context, fnTx func(ctx context.Context) error) error
	DropMigrationHistoryTable(ctx context.Context) error
	CreateMigrationHistoryTable(ctx context.Context) error
	MigrationsCount(ctx context.Context) (int, error)
	ExistsMigration(ctx context.Context, version string) (bool, error)
	TableNameWithSchema() string
}

// New creates repository by connection
//
//nolint:ireturn,nolintlint // it's ok
func New(conn Connection, options *Options) (Repository, error) {
	registry := NewFactoryRegistry()

	return registry.Create(conn, options)
}
