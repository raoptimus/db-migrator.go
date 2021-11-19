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
	"github.com/go-sql-driver/mysql"
	"github.com/raoptimus/db-migrator.go/migrator/db"
	"github.com/raoptimus/db-migrator.go/migrator/db/clickhouseMigration"
	"github.com/raoptimus/db-migrator.go/migrator/db/mysqlMigration"
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
		ClusterName        string
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
	case strings.HasPrefix(s.options.DSN, "mysql://"):
		return s.initMysql()
	default:
		return fmt.Errorf("Driver %s doesn't support", s.options.DSN)
	}
}

func (s *Service) initMysql() error {
	connection, err := sql.Open("mysql", s.options.DSN[8:])
	if err != nil {
		return err
	}
	if err := connection.Ping(); err != nil {
		return err
	}

	var cfg *mysql.Config
	cfg, err = mysql.ParseDSN(s.options.DSN)
	if err != nil {
		return err
	}

	s.db = connection
	s.migration = db.NewMigration(
		mysqlMigration.New(connection, s.options.TableName, cfg.DBName, s.options.Directory),
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

func (s *Service) initPostgres() error {
	connection, err := sql.Open("postgres", s.options.DSN)
	if err != nil {
		return err
	}
	if err := connection.Ping(); err != nil {
		return err
	}

	var tableName, tableSchema string
	if strings.Contains(s.options.TableName, ".") {
		parts := strings.Split(s.options.TableName, ".")
		tableSchema = parts[0]
		tableName = parts[1]
	} else {
		tableSchema = postgresMigration.DefaultSchema
		tableName = s.options.TableName
	}

	s.db = connection
	s.migration = db.NewMigration(
		postgresMigration.New(connection, tableName, tableSchema, s.options.Directory),
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
	dsn, err := clickhouseMigration.NormalizeDSN(s.options.DSN)
	if err != nil {
		return err
	}
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
		clickhouseMigration.New(
			connection,
			s.options.TableName,
			s.options.ClusterName,
			s.options.Directory,
		),
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
