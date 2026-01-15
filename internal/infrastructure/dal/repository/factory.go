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
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
)

const minTableNameParts = 2

// Factory defines the interface for creating repository instances for specific database drivers.
type Factory interface {
	// Create creates a new repository instance for the given connection and options.
	Create(conn Connection, options *Options) (Repository, error)
	// Supports returns true if this factory supports the given database driver.
	Supports(driver connection.Driver) bool
}

// FactoryRegistry manages a collection of repository factories for different database drivers.
type FactoryRegistry struct {
	factories []Factory
}

// NewFactoryRegistry creates a new factory registry with all supported database driver factories.
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

// Create creates a repository instance using the first factory that supports the connection's driver.
// It returns an error if no factory supports the driver.
//
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

// ClickhouseFactory creates repository instances for ClickHouse databases.
type ClickhouseFactory struct{}

// Supports returns true if the driver is ClickHouse.
func (f *ClickhouseFactory) Supports(driver connection.Driver) bool {
	return driver == connection.DriverClickhouse
}

// Create creates a new ClickHouse repository instance with validated identifiers.
// It parses the DSN to extract database name and validates schema, table, and cluster names.
//
//nolint:ireturn,nolintlint // its ok
func (f *ClickhouseFactory) Create(conn Connection, options *Options) (Repository, error) {
	opts, err := clickhouse.ParseDSN(conn.DSN())
	if err != nil {
		return nil, errors.WithMessage(err, "parsing dsn")
	}

	// Validate schema name extracted from DSN
	if err := ValidateIdentifier(opts.Auth.Database); err != nil {
		return nil, errors.Wrap(err, "invalid database name in DSN")
	}

	tableName, err := parseAndValidateTableName(options.TableName)
	if err != nil {
		return nil, err
	}

	// Validate cluster name
	if err := ValidateIdentifier(options.ClusterName); err != nil {
		return nil, errors.Wrap(err, "invalid cluster name")
	}

	return NewClickhouse(conn, &Options{
		SchemaName:  opts.Auth.Database,
		TableName:   tableName,
		ClusterName: options.ClusterName,
		Replicated:  options.Replicated,
	}), nil
}

// MySQLFactory

// MySQLFactory creates repository instances for MySQL databases.
type MySQLFactory struct{}

// Supports returns true if the driver is MySQL.
func (f *MySQLFactory) Supports(driver connection.Driver) bool {
	return driver == connection.DriverMySQL
}

// Create creates a new MySQL repository instance with validated identifiers.
// It parses the DSN to extract database name and validates table and schema names.
//
//nolint:ireturn,nolintlint // its ok
func (f *MySQLFactory) Create(conn Connection, options *Options) (Repository, error) {
	cfg, err := mysql.ParseDSN(conn.DSN())
	if err != nil {
		return nil, errors.WithMessage(err, "parsing dsn")
	}

	// Validate table name
	if err := ValidateIdentifier(options.TableName); err != nil {
		return nil, errors.Wrap(err, "invalid table name")
	}

	// Validate schema name extracted from DSN
	if err := ValidateIdentifier(cfg.DBName); err != nil {
		return nil, errors.Wrap(err, "invalid database name in DSN")
	}

	return NewMySQL(conn, &Options{
		TableName:  options.TableName,
		SchemaName: cfg.DBName,
	}), nil
}

// PostgresFactory

// PostgresFactory creates repository instances for PostgreSQL databases.
type PostgresFactory struct{}

// Supports returns true if the driver is PostgreSQL.
func (f *PostgresFactory) Supports(driver connection.Driver) bool {
	return driver == connection.DriverPostgres
}

// Create creates a new PostgreSQL repository instance with validated identifiers.
// It parses schema and table names from the table name option and validates all identifiers.
//
//nolint:ireturn,nolintlint // its ok
func (f *PostgresFactory) Create(conn Connection, options *Options) (Repository, error) {
	schemaName, tableName, err := parseAndValidateSchemaTableName(options.TableName, postgresDefaultSchema)
	if err != nil {
		return nil, err
	}

	return NewPostgres(conn, &Options{
		TableName:  tableName,
		SchemaName: schemaName,
	}), nil
}

// TarantoolFactory

// TarantoolFactory creates repository instances for Tarantool databases.
type TarantoolFactory struct{}

// Supports returns true if the driver is Tarantool.
func (f *TarantoolFactory) Supports(driver connection.Driver) bool {
	return driver == connection.DriverTarantool
}

// Create creates a new Tarantool repository instance with validated table name.
//
//nolint:ireturn,nolintlint // its ok
func (f *TarantoolFactory) Create(conn Connection, options *Options) (Repository, error) {
	// Validate table name
	if err := ValidateIdentifier(options.TableName); err != nil {
		return nil, errors.Wrap(err, "invalid table name")
	}

	return NewTarantool(conn, &Options{
		TableName:  options.TableName,
		SchemaName: "",
	}), nil
}

// parseAndValidateTableName parses and validates a table name that may contain schema prefix.
// For schema.table format, it validates both parts and returns only the table name.
// For simple table names, it validates and returns the name as-is.
func parseAndValidateTableName(tableName string) (string, error) {
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		if len(parts) < minTableNameParts {
			return "", errors.New("invalid table name format: expected schema.table")
		}
		if err := ValidateIdentifier(parts[0]); err != nil {
			return "", errors.Wrap(err, "invalid schema name in table name")
		}
		if err := ValidateIdentifier(parts[1]); err != nil {
			return "", errors.Wrap(err, "invalid table name")
		}
		return parts[1], nil
	}

	if err := ValidateIdentifier(tableName); err != nil {
		return "", errors.Wrap(err, "invalid table name")
	}
	return tableName, nil
}

// parseAndValidateSchemaTableName parses and validates a table name that may contain schema prefix.
// Returns schema name, table name, and error. If no schema prefix, uses defaultSchema.
func parseAndValidateSchemaTableName(tableName, defaultSchema string) (schema, table string, err error) {
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		if len(parts) < minTableNameParts {
			return "", "", errors.New("invalid table name format: expected schema.table")
		}
		if err := ValidateIdentifier(parts[0]); err != nil {
			return "", "", errors.Wrap(err, "invalid schema name in table name")
		}
		if err := ValidateIdentifier(parts[1]); err != nil {
			return "", "", errors.Wrap(err, "invalid table name")
		}
		return parts[0], parts[1], nil
	}

	if err := ValidateIdentifier(tableName); err != nil {
		return "", "", errors.Wrap(err, "invalid table name")
	}
	return defaultSchema, tableName, nil
}
