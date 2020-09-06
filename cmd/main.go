/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package main

import (
	"fmt"
	"github.com/raoptimus/db-migrator/console"
	"github.com/raoptimus/db-migrator/migrator"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
)

var (
	Version    string
	GitCommit  string
	controller *migrator.MigrateController
)

func main() {
	app := cli.NewApp()
	app.Name = "DB MigrateController"
	app.Usage = "up/down/redo command for migrates the different db"
	app.Version = fmt.Sprintf("v%s.rev[%s]", Version, GitCommit)
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "dsn",
			EnvVars: []string{"DSN"},
			Aliases: []string{"d"},
			Value:   "clickhouse://default:@localhost:9000/docker?sslmode=disable&compress=true&debug=false",
			Usage:   "DB connection string",
		},
		&cli.StringFlag{
			Name:    "migrationPath",
			EnvVars: []string{"MIGRATION_PATH"},
			Aliases: []string{"p"},
			Value:   "./migrator/db/clickhouseMigration/test_migrates",
			Usage:   "Directory for migrated files",
		},
		&cli.StringFlag{
			Name:    "migrationTable",
			EnvVars: []string{"MIGRATION_TABLE"},
			Aliases: []string{"t"},
			Value:   "migration",
			Usage:   "Table name for history of migrates",
		},
		&cli.BoolFlag{
			Name:    "compact",
			EnvVars: []string{"COMPACT"},
			Aliases: []string{"c"},
			Usage:   "Indicates whether the console output should be compacted.",
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "interactive",
			EnvVars: []string{"INTERACTIVE"},
			Aliases: []string{"i"},
			Usage:   "Whether to run the command interactively",
			Value:   true,
		},
	}
	app.Commands = []*cli.Command{
		{
			Name: "up",
			Action: func(c *cli.Context) error {
				return controller.Up(c.Args().Get(0))
			},
		},
		{
			Name: "down",
			Action: func(c *cli.Context) error {
				return controller.Down(c.Args().Get(0))
			},
		},
		{
			Name: "redo",
			Action: func(c *cli.Context) error {
				return controller.Redo(c.Args().Get(0))
			},
		},
		{
			Name: "create",
			Action: func(c *cli.Context) error {
				return controller.CreateMigration(c.Args().Get(0))
			},
		},
	}
	app.Before = before
	app.Action = func(c *cli.Context) error {
		return controller.Up(c.Args().Get(0))
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(console.Red(err))
	}
}

func before(c *cli.Context) error {
	fmt.Println(c.Command.Name)

	if c.Bool("debug") {
		go func() {
			fmt.Println(http.ListenAndServe(":6060", nil))
		}()
	}

	var err error
	controller, err = migrator.New(migrator.Options{
		DSN:         c.String("dsn"),
		Directory:   c.String("migrationPath"),
		TableName:   c.String("migrationTable"),
		Compact:     c.Bool("compact"),
		Interactive: c.Bool("interactive"),
	})

	return err
}
