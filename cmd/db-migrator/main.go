/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/raoptimus/db-migrator.go/internal/migrator"
	"github.com/raoptimus/db-migrator.go/pkg/console"
	"github.com/urfave/cli/v2"
)

var (
	Version   string
	GitCommit string
	dbService *migrator.DBService
)

func main() {
	options := &migrator.Options{}

	app := cli.NewApp()
	app.Name = "DB Service"
	app.Usage = "up/down/redo command for migrates the different db"
	app.Version = fmt.Sprintf("v%s.rev[%s]", Version, GitCommit)
	app.Flags = flags(options)
	app.Commands = commands()
	app.Before = func(context *cli.Context) error {
		dbService = migrator.New(options)
		return nil
	}
	app.Action = func(ctx *cli.Context) error {
		if a, err := dbService.Upgrade(); err != nil {
			return err
		} else {
			return a.Run(ctx)
		}
	}

	if err := app.Run(os.Args); err != nil {
		console.Std.Fatal(err)
	}
}

func commands() []*cli.Command {
	return []*cli.Command{
		{
			Name: "up",
			Action: func(ctx *cli.Context) error {
				if a, err := dbService.Upgrade(); err != nil {
					return err
				} else {
					return a.Run(ctx)
				}
			},
		},
		{
			Name: "down",
			Action: func(ctx *cli.Context) error {
				if a, err := dbService.Downgrade(); err != nil {
					return err
				} else {
					return a.Run(ctx)
				}
			},
		},
		{
			Name: "redo",
			Action: func(ctx *cli.Context) error {
				if a, err := dbService.Redo(); err != nil {
					return err
				} else {
					return a.Run(ctx)
				}
			},
		},
		{
			Name: "create",
			Action: func(ctx *cli.Context) error {
				return dbService.Create().Run(ctx)
			},
		},
		{
			Name: "history",
			Action: func(ctx *cli.Context) error {
				if a, err := dbService.History(); err != nil {
					return err
				} else {
					return a.Run(ctx)
				}
			},
		},
		{
			Name: "new",
			Action: func(ctx *cli.Context) error {
				if a, err := dbService.HistoryNew(); err != nil {
					return err
				} else {
					return a.Run(ctx)
				}
			},
		},
		{
			Name: "to",
			Action: func(ctx *cli.Context) error {
				if a, err := dbService.To(); err != nil {
					return err
				} else {
					return a.Run(ctx)
				}
			},
		},
	}
}

func flags(options *migrator.Options) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "dsn",
			EnvVars: []string{"DSN"},
			Aliases: []string{"d"},
			//Value: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
			//Value: "clickhouse://default:@clickhouse:9000/docker?sslmode=disable&compress=true&debug=false",
			Usage:       "DB connection string",
			Destination: &options.DSN,
		},
		&cli.StringFlag{
			Name:    "migrationPath",
			EnvVars: []string{"MIGRATION_PATH"},
			Aliases: []string{"p"},
			//Value: "./migrator/db/clickhouseMigration/test_migrates",
			//Value: "./migrator/db/postgresMigration/test_migrates",
			Usage:       "Directory for migrated files",
			Required:    true,
			Destination: &options.Directory,
		},
		&cli.StringFlag{
			Name:        "migrationTable",
			EnvVars:     []string{"MIGRATION_TABLE"},
			Aliases:     []string{"t"},
			Value:       "migration",
			Usage:       "Table name for history of migrates",
			Destination: &options.TableName,
		},
		&cli.StringFlag{
			Name:        "migrationClusterName",
			EnvVars:     []string{"MIGRATION_CLUSTER_NAME"},
			Aliases:     []string{"cn"},
			Value:       "",
			Usage:       "Cluster name for history of migrates",
			Destination: &options.ClusterName,
		},
		&cli.BoolFlag{
			Name:        "compact",
			EnvVars:     []string{"COMPACT"},
			Aliases:     []string{"c"},
			Usage:       "Indicates whether the console output should be compacted.",
			Value:       false,
			Destination: &options.Compact,
		},
		&cli.BoolFlag{
			Name:        "interactive",
			EnvVars:     []string{"INTERACTIVE"},
			Aliases:     []string{"i"},
			Usage:       "Whether to run the command interactively",
			Value:       true,
			Destination: &options.Interactive,
		},
	}
}
