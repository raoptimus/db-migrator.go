package action

import (
	"github.com/raoptimus/db-migrator.go/internal/console"
	"github.com/urfave/cli/v2"
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

func (t *To) Run(ctx *cli.Context) error {
	// version string from args
	console.Info("coming soon")
	return nil
}
