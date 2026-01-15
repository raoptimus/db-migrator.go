/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

// To handles the migration to a specific version.
type To struct {
	options         *Options
	logger          Logger
	fileNameBuilder FileNameBuilder
}

// NewTo creates a new To handler instance.
func NewTo(
	options *Options,
	logger Logger,
	fileNameBuilder FileNameBuilder,
) *To {
	return &To{
		options:         options,
		logger:          logger,
		fileNameBuilder: fileNameBuilder,
	}
}

// Handle processes the to command to migrate to a specific version.
func (t *To) Handle(_ *Command, _ MigrationService) error {
	// version string from args
	t.logger.Info("coming soon")

	return nil
}
