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
	"github.com/raoptimus/db-migrator.go/migrator/db"
	"github.com/raoptimus/db-migrator.go/migrator/db/clickhouseMigration"
	"github.com/raoptimus/db-migrator.go/migrator/db/postgresMigration"
	"strings"
)

type (
	Service struct {
		options     Options
		db          *sql.DB
		migration   *db.Migration
		fileBuilder FileNameBuilder
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

func New(options Options) (*Service, error) {
	serv := Service{
		options:     options,
		fileBuilder: NewFileNameBuilder(options.Directory),
	}
	if err := serv.init(); err != nil {
		return nil, err
	}
	return &serv, nil
}

func (s *Service) init() error {
	switch {
	case strings.HasPrefix(s.options.DSN, "clickhouse://"):
		return s.initClickHouse()
	case strings.HasPrefix(s.options.DSN, "postgres://"):
		return s.initPostgres()
	default:
		return fmt.Errorf("Driver %s doesn't support", s.options.DSN)
	}
}

func (s *Service) initPostgres() error {
	connection, err := sql.Open("postgres", s.options.DSN)
	if err != nil {
		return err
	}
	if err := connection.Ping(); err != nil {
		return err
	}

	s.db = connection
	s.migration = db.NewMigration(
		postgresMigration.New(connection, s.options.TableName, s.options.Directory),
		connection,
		db.MigrationOptions{
			MaxSqlOutputLength: s.options.MaxSqlOutputLength,
			Directory:          s.options.Directory,
			Compact:            s.options.Compact,
			MultiSTMT:          false,
			ForceSafely:        false,
		},
	)

	return nil
}

func (s *Service) initClickHouse() error {
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
