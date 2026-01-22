/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package urfavecli

import (
	"context"

	"github.com/raoptimus/db-migrator.go/internal/application/handler"
	"github.com/urfave/cli/v3"
)

func Adapt(h handler.Handler) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		internalCmd := &handler.Command{
			Args: cmd.Args(),
		}

		return h.Handle(internalCmd.WithContext(ctx))
	}
}
