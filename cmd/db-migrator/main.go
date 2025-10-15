/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/migrator"
	"github.com/raoptimus/db-migrator.go/internal/util/console"
	"github.com/urfave/cli/v3"
)

const maxIdentifierLen = 65000

var (
	Version   string
	GitCommit string
	dbService *migrator.DBService
)

func main() {
	options := migrator.Options{}

	cmd := &cli.Command{
		Name:           "DB Service",
		Usage:          "up/down/redo command for migrates the different db",
		Version:        fmt.Sprintf("%s.rev[%s]", Version, GitCommit),
		Commands:       commands(&options),
		DefaultCommand: "help",
		Flags:          flags(&options, false),
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			dbService = migrator.New(&options)

			return ctx, nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		console.Std.Fatal(err)
	}
}

func commands(options *migrator.Options) []*cli.Command {
	return []*cli.Command{
		{
			Name: "up",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if a, err := dbService.Upgrade(); err != nil {
					return err
				} else {
					return a.Run(ctx, cmd.Args().Get(0))
				}
			},
			Flags: flags(options, true),
		},
		{
			Name: "down",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if a, err := dbService.Downgrade(); err != nil {
					return err
				} else {
					return a.Run(ctx, cmd.Args().Get(0))
				}
			},
			Flags: flags(options, true),
		},
		{
			Name: "redo",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if a, err := dbService.Redo(); err != nil {
					return err
				} else {
					return a.Run(ctx, cmd.Args().Get(0))
				}
			},
			Flags: flags(options, true),
		},
		{
			Name: "create",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return dbService.Create().Run(ctx, cmd.Args().Get(0))
			},
			Flags: flags(options, false),
		},
		{
			Name: "history",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if a, err := dbService.History(); err != nil {
					return err
				} else {
					return a.Run(ctx, cmd.Args().Get(0))
				}
			},
			Flags: flags(options, true),
		},
		{
			Name: "new",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if a, err := dbService.HistoryNew(); err != nil {
					return err
				} else {
					return a.Run(ctx, cmd.Args().Get(0))
				}
			},
			Flags: flags(options, true),
		},
		{
			Name: "to",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if a, err := dbService.To(); err != nil {
					return err
				} else {
					return a.Run(ctx, cmd.Args().Get(0))
				}
			},
			Flags: flags(options, true),
		},
	}
}

func flags(options *migrator.Options, dsnIsRequired bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "dsn",
			Sources:     cli.EnvVars("DSN"),
			Aliases:     []string{"d"},
			Usage:       "DB connection string",
			Destination: &options.DSN,
			Required:    dsnIsRequired,
		},
		&cli.IntFlag{
			Name:        "maxConnAttempts",
			Sources:     cli.EnvVars("MAX_CONN_ATTEMPTS"),
			Aliases:     []string{"ma"},
			Usage:       "Maximum number of connection attempts",
			Destination: &options.MaxConnAttempts,
			Value:       1,
		},
		&cli.StringFlag{
			Name:        "migrationPath",
			Sources:     cli.EnvVars("MIGRATION_PATH"),
			Aliases:     []string{"p"},
			Value:       "./migrations",
			Usage:       "Directory for migrated files",
			Destination: &options.Directory,
		},
		&cli.StringFlag{
			Name:        "migrationTable",
			Sources:     cli.EnvVars("MIGRATION_TABLE"),
			Aliases:     []string{"t"},
			Value:       "migration",
			Usage:       "Table name for history of migrates",
			Destination: &options.TableName,
			Action: func(ctx context.Context, command *cli.Command, s string) error {
				if !isValidIdentifier(s) {
					return errors.New("invalid value of variable MIGRATION_TABLE")
				}

				return nil
			},
		},
		&cli.StringFlag{
			Name:        "migrationClusterName",
			Sources:     cli.EnvVars("MIGRATION_CLUSTER_NAME"),
			Aliases:     []string{"cn"},
			Value:       "",
			Usage:       "Cluster name for history of migrates",
			Destination: &options.ClusterName,
			Action: func(ctx context.Context, command *cli.Command, s string) error {
				if !isValidIdentifier(s) {
					return errors.New("invalid value of variable MIGRATION_CLUSTER_NAME")
				}

				return nil
			},
		},
		&cli.BoolFlag{
			Name:        "migrationReplicated",
			Sources:     cli.EnvVars("MIGRATION_REPLICATED"),
			Aliases:     []string{"cr"},
			Value:       false,
			Usage:       "Using replicated experimental function to clickhouse for history table of migrates",
			Destination: &options.Replicated,
		},
		&cli.BoolFlag{
			Name:        "compact",
			Sources:     cli.EnvVars("COMPACT"),
			Aliases:     []string{"c"},
			Usage:       "Indicates whether the console output should be compacted.",
			Value:       false,
			Destination: &options.Compact,
		},
		&cli.BoolFlag{
			Name:        "interactive",
			Sources:     cli.EnvVars("INTERACTIVE"),
			Aliases:     []string{"i"},
			Usage:       "Whether to run the command interactively",
			Value:       true,
			Destination: &options.Interactive,
		},
	}
}

func isValidIdentifier(name string) bool {
	if len(name) > maxIdentifierLen {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}
