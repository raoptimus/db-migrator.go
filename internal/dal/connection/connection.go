package connection

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type ContextKey string

const contextKeyTX ContextKey = "tx"

type Connection struct {
	driver Driver
	dsn    string
	db     *sql.DB
	ping   bool
}

func New(dsn string) (*Connection, error) {
	switch {
	case strings.HasPrefix(dsn, "clickhouse://"):
		return clickhouse(dsn)
	case strings.HasPrefix(dsn, "postgres://"):
		return postgres(dsn)
	case strings.HasPrefix(dsn, "mysql://"):
		return mysql(dsn)
	default:
		return nil, fmt.Errorf("driver \"%s\" doesn't support", dsn)
	}
}

// DSN returns the connection string.
func (c *Connection) DSN() string {
	return c.dsn
}

// Driver returns the driver name used to connect to the database.
func (c *Connection) Driver() Driver {
	return c.driver
}

// Ping checks connection
func (c *Connection) Ping() error {
	if c.ping {
		return nil
	}
	if err := c.db.Ping(); err != nil {
		return errors.Wrapf(err, "ping %v connection: %v", c.Driver(), c.dsn)
	}
	c.ping = true
	return nil
}

// QueryContext executes a query that returns rows, typically a SELECT.
// The args are for any placeholder parameters in the query.
func (c *Connection) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if err := c.Ping(); err != nil {
		return nil, err
	}
	return c.db.QueryContext(ctx, query, args...)
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (c *Connection) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if err := c.Ping(); err != nil {
		return nil, err
	}
	v := ctx.Value(contextKeyTX)
	if v != nil {
		if tx, ok := v.(*sql.Tx); ok {
			stmt, err := tx.PrepareContext(ctx, query)
			if err != nil {
				return nil, err
			}
			return stmt.ExecContext(ctx, args...)
		}
	}
	return c.db.ExecContext(ctx, query, args...)
}

// Transaction executes body in func txFn into transaction.
func (c *Connection) Transaction(ctx context.Context, txFn func(ctx context.Context) error) error {
	if err := c.Ping(); err != nil {
		return err
	}
	if v := ctx.Value(contextKeyTX); v != nil {
		return errors.New("active transaction does not close")
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	ctxWithTX := context.WithValue(ctx, contextKeyTX, tx)

	if err := txFn(ctxWithTX); err != nil {
		if err2 := tx.Rollback(); err2 != nil {
			return errors.Wrap(err, err2.Error())
		}
		return err
	}

	return tx.Commit()
}

// clickhouse returns repository with clickhouse configuration.
func clickhouse(dsn string) (*Connection, error) {
	dsn, err := NormalizeClickhouseDSN(dsn)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}

	return &Connection{
		driver: DriverClickhouse,
		dsn:    dsn,
		db:     db,
	}, nil
}

// postgres returns repository with postgres configuration.
func postgres(dsn string) (*Connection, error) {
	db, err := sql.Open(DriverPostgres.String(), dsn)
	if err != nil {
		return nil, errors.Wrap(err, "open postgres connection")
	}

	return &Connection{
		driver: DriverPostgres,
		dsn:    dsn,
		db:     db,
	}, nil
}

// mysql returns repository with mysql configuration.
func mysql(dsn string) (*Connection, error) {
	db, err := sql.Open(DriverMySQL.String(), dsn[8:])
	if err != nil {
		return nil, errors.Wrap(err, "open mysql connection")
	}

	return &Connection{
		driver: DriverMySQL,
		dsn:    dsn,
		db:     db,
	}, nil
}
