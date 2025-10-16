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

	"github.com/raoptimus/db-migrator.go/internal/console"
)

type To struct {
	service         MigrationService
	fileNameBuilder FileNameBuilder
	interactive     bool
}

func NewTo(
	service MigrationService,
	fileNameBuilder FileNameBuilder,
	interactive bool,
) *To {
	return &To{
		service:         service,
		fileNameBuilder: fileNameBuilder,
		interactive:     interactive,
	}
}

func (t *To) Run(_ context.Context, _ ...string) error {
	// version string from args
	console.Info("coming soon")
	return nil
}
