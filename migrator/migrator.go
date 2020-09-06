/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package migrator

import (
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go"
	"github.com/raoptimus/db-migrator/migrator/db"
	"github.com/raoptimus/db-migrator/migrator/db/clickhouseMigration"
	"strings"
)

type (
	MigrateController struct {
		options   Options
		db        *sql.DB
		migration *db.Migration
	}
	Options struct {
		DSN                string
		Directory          string
		TableName          string
		Compact            bool
		Interactive        bool
		MaxSqlOutputLength int
	}
)

func New(options Options) (*MigrateController, error) {
	controller := MigrateController{options: options}
	if err := controller.init(); err != nil {
		return nil, err
	}
	return &controller, nil
}

func (s *MigrateController) CreateMigration(name string) error {
	return nil
}

func (s *MigrateController) init() error {
	switch {
	case strings.HasPrefix(s.options.DSN, "clickhouse://"):
		return s.initClickhouse()
	default:
		return fmt.Errorf("Driver %s doesn't support", s.options.DSN)
	}
}

func (s *MigrateController) initClickhouse() error {
	dsn := "tcp://" + strings.TrimPrefix(s.options.DSN, "clickhouse://")
	connection, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return err
	}

	if err := connection.Ping(); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			return fmt.Errorf("[%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return err
	}
	s.db = connection
	s.migration = db.NewMigration(
		clickhouseMigration.New(connection, s.options.TableName, s.options.Directory),
		connection,
		db.MigrationOptions{
			MaxSqlOutputLength: s.options.MaxSqlOutputLength,
			Directory:          s.options.Directory,
			Compact:            s.options.Compact,
			MultiSTMT:          false,
			ForceSafely:        true,
		},
	)

	return nil
}
