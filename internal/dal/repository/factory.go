/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"fmt"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/dal/connection"
)

type Factory interface {
	Create(conn Connection, options *Options) (Repository, error)
	Supports(driver connection.Driver) bool
}

type FactoryRegistry struct {
	factories []Factory
}

func NewFactoryRegistry() *FactoryRegistry {
	return &FactoryRegistry{
		factories: []Factory{
			&TarantoolFactory{},
			&PostgresFactory{},
			&MySQLFactory{},
			&ClickhouseFactory{},
		},
	}
}

//nolint:ireturn,nolintlint // its ok
func (r *FactoryRegistry) Create(conn Connection, options *Options) (Repository, error) {
	for _, factory := range r.factories {
		if factory.Supports(conn.Driver()) {
			return factory.Create(conn, options)
		}
	}

	return nil, fmt.Errorf("driver \"%s\" doesn't support", conn.Driver())
}

// ClickhouseFactory

type ClickhouseFactory struct{}

func (f *ClickhouseFactory) Supports(driver connection.Driver) bool {
	return driver == connection.DriverClickhouse
}

//nolint:ireturn,nolintlint // its ok
func (f *ClickhouseFactory) Create(conn Connection, options *Options) (Repository, error) {
	opts, err := clickhouse.ParseDSN(conn.DSN())
	if err != nil {
		return nil, errors.WithMessage(err, "parsing dsn")
	}
	var tableName string
	if strings.Contains(options.TableName, ".") {
		parts := strings.Split(options.TableName, ".")
		tableName = parts[1]
	} else {
		tableName = options.TableName
	}

	return NewClickhouse(conn, &Options{
		SchemaName:  opts.Auth.Database,
		TableName:   tableName,
		ClusterName: options.ClusterName,
	}), nil
}

// MySQLFactory

type MySQLFactory struct{}

func (f *MySQLFactory) Supports(driver connection.Driver) bool {
	return driver == connection.DriverMySQL
}

//nolint:ireturn,nolintlint // its ok
func (f *MySQLFactory) Create(conn Connection, options *Options) (Repository, error) {
	cfg, err := mysql.ParseDSN(conn.DSN())
	if err != nil {
		return nil, errors.WithMessage(err, "parsing dsn")
	}

	return NewMySQL(conn, &Options{
		TableName:  options.TableName,
		SchemaName: cfg.DBName,
	}), nil
}

// PostgresFactory

type PostgresFactory struct{}

func (f *PostgresFactory) Supports(driver connection.Driver) bool {
	return driver == connection.DriverPostgres
}

//nolint:ireturn,nolintlint // its ok
func (f *PostgresFactory) Create(conn Connection, options *Options) (Repository, error) {
	var tableName, schemaName string
	if strings.Contains(options.TableName, ".") {
		parts := strings.Split(options.TableName, ".")
		schemaName = parts[0]
		tableName = parts[1]
	} else {
		schemaName = postgresDefaultSchema
		tableName = options.TableName
	}

	return NewPostgres(conn, &Options{
		TableName:  tableName,
		SchemaName: schemaName,
	}), nil
}

// TarantoolFactory

type TarantoolFactory struct{}

func (f *TarantoolFactory) Supports(driver connection.Driver) bool {
	return driver == connection.DriverTarantool
}

//nolint:ireturn,nolintlint // its ok
func (f *TarantoolFactory) Create(conn Connection, options *Options) (Repository, error) {
	return NewTarantool(conn, &Options{
		TableName:  options.TableName,
		SchemaName: "",
	}), nil
}
