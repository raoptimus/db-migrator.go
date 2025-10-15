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
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

type adapter interface {
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

type Repository struct {
	adapter
}

func (r *Repository) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	return r.adapter.Migrations(ctx, limit)
}

func (r *Repository) HasMigrationHistoryTable(ctx context.Context) (exists bool, err error) {
	return r.adapter.HasMigrationHistoryTable(ctx)
}

func (r *Repository) InsertMigration(ctx context.Context, version string) error {
	return r.adapter.InsertMigration(ctx, version)
}

func (r *Repository) RemoveMigration(ctx context.Context, version string) error {
	return r.adapter.RemoveMigration(ctx, version)
}

func (r *Repository) ExecQuery(ctx context.Context, query string, args ...any) error {
	return r.adapter.ExecQuery(ctx, query, args...)
}

func (r *Repository) QueryScalar(ctx context.Context, query string, ptr any) error {
	return r.adapter.QueryScalar(ctx, query, ptr)
}

func (r *Repository) ExecQueryTransaction(ctx context.Context, fnTx func(ctx context.Context) error) error {
	return r.adapter.ExecQueryTransaction(ctx, fnTx)
}

func (r *Repository) DropMigrationHistoryTable(ctx context.Context) error {
	return r.adapter.DropMigrationHistoryTable(ctx)
}

func (r *Repository) CreateMigrationHistoryTable(ctx context.Context) error {
	return r.adapter.CreateMigrationHistoryTable(ctx)
}

func (r *Repository) MigrationsCount(ctx context.Context) (int, error) {
	return r.adapter.MigrationsCount(ctx)
}

func (r *Repository) ExistsMigration(ctx context.Context, version string) (bool, error) {
	return r.adapter.ExistsMigration(ctx, version)
}

func (r *Repository) TableNameWithSchema() string {
	return r.adapter.TableNameWithSchema()
}

// create creates repository adapter
//
//nolint:ireturn,nolintlint // it's ok
func create(conn Connection, options *Options) (adapter, error) {
	switch conn.Driver() {
	case connection.DriverTarantool:
		return NewTarantool(conn, &Options{
			TableName:  options.TableName,
			SchemaName: "",
		}), nil
	case connection.DriverMySQL:
		cfg, err := mysql.ParseDSN(conn.DSN())
		if err != nil {
			return nil, errors.WithMessage(err, "parsing dsn")
		}
		repo := NewMySQL(conn, &Options{
			TableName:  options.TableName,
			SchemaName: cfg.DBName,
		})
		return repo, err
	case connection.DriverPostgres:
		var tableName, schemaName string
		if strings.Contains(options.TableName, ".") {
			parts := strings.Split(options.TableName, ".")
			schemaName = parts[0]
			tableName = parts[1]
		} else {
			schemaName = postgresDefaultSchema
			tableName = options.TableName
		}

		repo := NewPostgres(conn, &Options{
			TableName:  tableName,
			SchemaName: schemaName,
		})
		return repo, nil
	case connection.DriverClickhouse:
		opts, err := clickhouse.ParseDSN(conn.DSN())
		if err != nil {
			return nil, errors.WithMessage(err, "parsing dsn")
		}
		var tableName string
		if strings.Contains(options.TableName, ".") {
			parts := strings.Split(options.TableName, ".")
			tableName = parts[1]
		} else {
			tableName = options.TableName
		}
		repo := NewClickhouse(conn, &Options{
			SchemaName:  opts.Auth.Database,
			TableName:   tableName,
			ClusterName: options.ClusterName,
		})
		return repo, nil
	default:
		return nil, fmt.Errorf("driver \"%s\" doesn't support", conn.Driver())
	}
}
