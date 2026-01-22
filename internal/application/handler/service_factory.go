/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/domain/service"
	"github.com/raoptimus/db-migrator.go/internal/helper/dsn"
	iohelp "github.com/raoptimus/db-migrator.go/internal/helper/io"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/repository"
)

// NewMigrationService creates a new migration service instance with the specified logger, connection, and options.
// This function is used by the public API (service.go) to create migration service instances.
//
//nolint:ireturn // Returns interface by design for dependency inversion
func NewMigrationService(
	options *Options,
	logger Logger,
	conn Connection,
) (MigrationService, error) {
	// Parse DSN to extract credentials
	parsed, err := dsn.Parse(options.DSN)
	if err != nil {
		return nil, errors.WithMessage(err, "parsing DSN")
	}

	// Create repository
	var serviceRepo service.Repository
	repo, err := repository.New(
		conn,
		&repository.Options{
			TableName:   options.TableName,
			ClusterName: options.ClusterName,
			Replicated:  options.Replicated,
		},
	)
	if err != nil {
		return nil, err
	}
	
	if options.DryRun {
		serviceRepo = service.NewDryRunRepository(repo)
	} else {
		serviceRepo = repo
	}

	// Create and return service
	return service.NewMigration(
		&service.Options{
			MaxSQLOutputLength: options.MaxSQLOutputLength,
			Directory:          options.Directory,
			Compact:            options.Compact,
			PlaceholderCustom:  options.PlaceholderCustom,
			ClusterName:        options.ClusterName,
			Username:           parsed.Username,
			Password:           parsed.Password,
		},
		logger,
		iohelp.StdFile,
		serviceRepo,
	), nil
}
