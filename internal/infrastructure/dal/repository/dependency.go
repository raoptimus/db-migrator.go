package repository

import (
	"context"
	"io"

	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/connection"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/iceberg/ddl"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

//go:generate mockery

// IcebergCatalog is the full catalog contract needed by the Iceberg repository.
// It is implemented by catalog.Client, which wraps the Apache Iceberg REST client.
type IcebergCatalog interface {
	// Ping verifies connectivity to the REST catalog.
	Ping(ctx context.Context) error
	// Close releases any resources held by the catalog client.
	Close() error
	// Warehouse returns the warehouse name from the DSN path.
	Warehouse() string
	// CreateNamespace creates a namespace with the given properties.
	CreateNamespace(ctx context.Context, ns []string, props map[string]string) error
	// DropNamespace drops the given namespace.
	DropNamespace(ctx context.Context, ns []string) error
	// NamespaceExists checks whether the given namespace exists.
	NamespaceExists(ctx context.Context, ns []string) (bool, error)
	// LoadNamespaceProperties returns the properties of the given namespace.
	LoadNamespaceProperties(ctx context.Context, ns []string) (map[string]string, error)
	// UpdateNamespaceProperties updates namespace properties by removing and setting keys.
	UpdateNamespaceProperties(ctx context.Context, ns []string, removals []string, updates map[string]string) error

	// CreateTable creates an Iceberg table from the given IR specification.
	CreateTable(ctx context.Context, ident ddl.Ident, spec ddl.CreateTableSpec) error
	// TableExists checks whether the given table exists in the catalog.
	TableExists(ctx context.Context, ident ddl.Ident) (bool, error)
	// DropTable drops an Iceberg table identified by ident.
	DropTable(ctx context.Context, ident ddl.Ident) error
	// RenameTable renames an Iceberg table from from to to.
	RenameTable(ctx context.Context, from, to ddl.Ident) error
	// ApplySchemaChange applies a schema-level DDL operation (AddColumn, DropColumn,
	// RenameColumn, AlterColumnType) via an Iceberg schema update transaction.
	ApplySchemaChange(ctx context.Context, op ddl.Operation) error
	// ApplySpecChange applies a partition-spec DDL operation (AddPartitionField,
	// DropPartitionField) via an Iceberg spec update transaction.
	ApplySpecChange(ctx context.Context, op ddl.Operation) error
	// ApplySortOrderChange sets or clears the table write sort order (WRITE ORDERED BY /
	// WRITE UNORDERED) via a catalog CommitTable call.
	ApplySortOrderChange(ctx context.Context, op ddl.Operation) error
}

//go:generate mockery

// Connection defines the interface for database connection operations used by repositories.
// It provides methods for executing queries, transactions, and managing the connection lifecycle.
type Connection interface {
	io.Closer

	// DSN returns the Data Source Name string for the connection.
	DSN() string
	// Driver returns the database driver type for this connection.
	Driver() connection.Driver
	// Ping verifies the connection to the database is alive.
	Ping() error
	// QueryContext executes a query that returns rows with the provided context and arguments.
	QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error)
	// ExecContext executes a query that doesn't return rows with the provided context and arguments.
	ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error)
	// Transaction executes a function within a database transaction.
	Transaction(ctx context.Context, txFn func(ctx context.Context) error) error
}
