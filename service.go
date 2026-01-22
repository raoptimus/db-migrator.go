/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package dbmigrator

import (
	"context"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/application/handler"
	"github.com/raoptimus/db-migrator.go/internal/domain/validator"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/log"
)

type (
	// Options configures the database migration service.
	// It contains connection settings and database-specific parameters.
	Options struct {
		DSN string
		// table name to history of migrations
		TableName string
		// cluster name to clickhouse
		ClusterName string
		// is replicated used to clickhouse?
		Replicated bool
	}

	// DBService provides high-level operations for database migrations.
	// It orchestrates the migration process using the underlying migration service.
	DBService struct {
		opts   *handler.Options
		logger Logger
		conn   Connection
	}
)

// NewDBService creates a new database migration service with the provided options, connection, and logger.
// If logger is nil, a no-op logger will be used.
func NewDBService(opts *Options, conn Connection, logger Logger) (*DBService, error) {
	if logger == nil {
		logger = &log.NopLogger{}
	}
	options := &handler.Options{
		DSN:                opts.DSN,
		MaxConnAttempts:    1,
		Directory:          "",
		TableName:          opts.TableName,
		ClusterName:        opts.ClusterName,
		Replicated:         opts.Replicated,
		Compact:            true,
		Interactive:        true,
		MaxSQLOutputLength: 0,
	}
	if err := options.Validate(); err != nil {
		return nil, err
	}

	return &DBService{
		opts:   options,
		conn:   conn,
		logger: logger,
	}, nil
}

// Upgrade apply changes to db. apply specific version of migration.
func (d *DBService) Upgrade(ctx context.Context, version, sql string, safety bool) error {
	if err := validator.ValidateVersion(version); err != nil {
		return err
	}

	serviceMigration, err := handler.NewMigrationService(d.opts, d.logger, d.conn)
	if err != nil {
		return err
	}

	exists, err := serviceMigration.Exists(ctx, version)
	if err != nil {
		return err
	}

	if exists {
		return errors.WithStack(ErrMigrationAlreadyExists)
	}

	return serviceMigration.ApplySQL(ctx, safety, version, sql)
}

// Downgrade revert changes to db. revert specific version of migration.
func (d *DBService) Downgrade(ctx context.Context, version, sql string, safety bool) error {
	if err := validator.ValidateVersion(version); err != nil {
		return err
	}

	serviceMigration, err := handler.NewMigrationService(d.opts, d.logger, d.conn)
	if err != nil {
		return err
	}

	exists, err := serviceMigration.Exists(ctx, version)
	if err != nil {
		return err
	}

	if !exists {
		return errors.WithStack(ErrAppliedMigrationNotFound)
	}

	return serviceMigration.RevertSQL(ctx, safety, version, sql)
}
