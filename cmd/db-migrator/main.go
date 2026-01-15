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
	"github.com/raoptimus/db-migrator.go/internal/application/handler"
	"github.com/raoptimus/db-migrator.go/internal/domain/validator"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/adapter/urfavecli"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/log"
	"github.com/urfave/cli/v3"
)

const maxConnAttempts = 100

var (
	Version   string
	GitCommit string
)

func main() {
	options := handler.Options{}
	logger := log.Std
	handlers := handler.NewHandlers(&options, logger)

	cmd := &cli.Command{
		Name:      "Database Migration Tool",
		Usage:     "Database migration tool for different databases",
		UsageText: "db-migrator [command] [command options]\n\n Command-line options override environment variables.",
		Description: `This tool helps to manage database migrations. 
			It supports PostgreSQL and other databases that support SQL commands. 
			The tool can perform up, down, redo, create, history, new, and to operations on the database. 
			For more information, please refer to the documentation. 
			More details about the tool can be found at https://github.com/raoptimus/db-migrator.go`,
		Version: fmt.Sprintf("%s.rev[%s]", Version, GitCommit),
		Commands: []*cli.Command{
			{Name: "up", Action: urfavecli.Adapt(handlers.Upgrade), Flags: flags(&options, true)},
			{Name: "down", Action: urfavecli.Adapt(handlers.Downgrade), Flags: flags(&options, true)},
			{Name: "redo", Action: urfavecli.Adapt(handlers.Redo), Flags: flags(&options, true)},
			{Name: "to", Action: urfavecli.Adapt(handlers.To), Flags: flags(&options, true)},
			{Name: "create", Action: urfavecli.Adapt(handlers.Create), Flags: flags(&options, false)},
			{Name: "history", Action: urfavecli.Adapt(handlers.History), Flags: flags(&options, true)},
			{Name: "new", Action: urfavecli.Adapt(handlers.HistoryNew), Flags: flags(&options, true)},
		},
		DefaultCommand: "help",
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logger.Fatal(err)
	}
}

func flags(options *handler.Options, dsnIsRequired bool) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "placeholderCustom",
			Sources:     cli.EnvVars("PLACEHOLDER_CUSTOM"),
			Aliases:     []string{"phc"},
			Usage:       "PLACEHOLDER_CUSTOM",
			Destination: &options.PlaceholderCustom,
			Value:       "",
			Validator:   validator.ValidateIdentifier,
		},
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
			Validator: func(i int) error {
				return validator.ValidateStringLen(1, maxConnAttempts, i)
			},
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
				return validator.ValidateIdentifier(s)
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
				return validator.ValidateIdentifier(s)
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
