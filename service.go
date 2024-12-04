package dbmigrator

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/migrator"
)

type (
	Options struct {
		DSN string
		// table name to history of migrations
		TableName string
		// cluster name to clickhouse
		ClusterName string
		// is replicated used to clickhouse?
		Replicated bool
	}
	DBService struct {
		dbs  *migrator.DBService
		opts *Options
	}
)

func NewDBService(opts *Options) *DBService {
	return &DBService{
		dbs: migrator.New(&migrator.Options{
			DSN:                opts.DSN,
			Directory:          "",
			TableName:          opts.TableName,
			ClusterName:        opts.ClusterName,
			Replicated:         opts.Replicated,
			Compact:            true,
			Interactive:        true,
			MaxSQLOutputLength: 0,
		}),
		opts: opts,
	}
}

// Upgrade apply changes to db. apply specific version of migration.
func (d *DBService) Upgrade(ctx context.Context, version, sql string, safety bool) error {
	ms, err := d.dbs.MigrationService()
	if err != nil {
		return err
	}

	exists, err := ms.Exists(ctx, version)
	if err != nil {
		return err
	}

	if exists {
		return ErrMigrationAlreadyExists
	}

	return ms.ApplySQL(ctx, safety, version, sql)
}

// Downgrade revert changes to db. revert specific version of migration.
func (d *DBService) Downgrade(ctx context.Context, version, sql string, safety bool) error {
	ms, err := d.dbs.MigrationService()
	if err != nil {
		return err
	}

	exists, err := ms.Exists(ctx, version)
	if err != nil {
		return err
	}

	if !exists {
		return ErrAppliedMigrationNotFound
	}

	return ms.RevertSQL(ctx, safety, version, sql)
}
