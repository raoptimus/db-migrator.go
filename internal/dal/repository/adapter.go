package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

type adapter interface {
	Migrations(ctx context.Context, limit int) (entity.Migrations, error)
	HasMigrationHistoryTable(ctx context.Context) (exists bool, err error)
	InsertMigration(ctx context.Context, version string) error
	RemoveMigration(ctx context.Context, version string) error
	ExecQuery(ctx context.Context, query string, args ...any) error
	ExecQueryTransaction(ctx context.Context, fnTx func(ctx context.Context) error) error
	DropMigrationHistoryTable(ctx context.Context) error
	CreateMigrationHistoryTable(ctx context.Context) error
	MigrationsCount(ctx context.Context) (int, error)
	TableNameWithSchema() string
	ForceSafely() bool
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

func (r *Repository) TableNameWithSchema() string {
	return r.adapter.TableNameWithSchema()
}

func (r *Repository) ForceSafely() bool {
	return r.adapter.ForceSafely()
}

// create creates repository adapter
//
//nolint:ireturn,nolintlint // its ok
func create(conn Connection, options *Options) (adapter, error) {
	switch conn.Driver() {
	case connection.DriverMySQL:
		cfg, err := mysql.ParseDSN(conn.DSN())
		if err != nil {
			return nil, err
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
		repo := NewClickhouse(conn, &Options{
			TableName:   options.TableName,
			SchemaName:  options.SchemaName,
			ClusterName: options.ClusterName,
			ShardName:   options.ShardName,
		})
		return repo, nil
	default:
		return nil, fmt.Errorf("driver \"%s\" doesn't support", conn.Driver())
	}
}
