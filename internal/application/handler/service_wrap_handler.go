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
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
)

// ServiceHandler defines the interface for handlers that require MigrationService.
type ServiceHandler interface {
	Handle(cmd *Command, svc MigrationService) error
}

// ServiceWrapHandler wraps a ServiceHandler, managing database connection
// and MigrationService lifecycle for each command execution.
type ServiceWrapHandler struct {
	options *Options
	logger  Logger
	handler ServiceHandler
}

// NewServiceWrapHandler creates a new ServiceWrapHandler instance.
func NewServiceWrapHandler(
	options *Options,
	logger Logger,
	handler ServiceHandler,
) *ServiceWrapHandler {
	return &ServiceWrapHandler{
		options: options,
		logger:  logger,
		handler: handler,
	}
}

// Handle executes the command by creating database connection and MigrationService,
// then delegating to the wrapped handler.
func (w *ServiceWrapHandler) Handle(cmd *Command) error {
	if w.options.DryRun {
		w.options.Interactive = false
		w.logger.Warnf("[DRY RUN] No changes will be applied to the database.")
	}
	
	// Create database connection
	conn, err := connection.Try(w.options.DSN, w.options.MaxConnAttempts)
	if err != nil {
		return errors.WithMessage(err, "failed to connect to database")
	}
	defer conn.Close()

	// Create MigrationService
	svc, err := NewMigrationService(w.options, w.logger, conn)
	if err != nil {
		return errors.WithMessage(err, "failed to create migration service")
	}

	return w.handler.Handle(cmd, svc)
}
