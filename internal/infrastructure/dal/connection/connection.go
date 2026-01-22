/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package connection

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/helper/dsn"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex/tarantool"
)

const durationBeforeNextConnAttempt = 1 * time.Second

var (
	ErrTransactionAlreadyOpened = errors.New("transaction already opened")
)

type Connection struct {
	driver Driver
	dsn    string
	db     SQLDB
	ping   bool
}

func New(dsnStr string) (*Connection, error) {
	parsed, err := dsn.Parse(dsnStr)
	if err != nil {
		return nil, errors.WithMessage(err, "parsing DSN")
	}

	switch parsed.Driver {
	case "clickhouse":
		return clickhouse(dsnStr)
	case "postgres":
		return postgres(dsnStr)
	case "mysql":
		return mysql(dsnStr)
	case "tarantool":
		return tarantoolConn(dsnStr)
	default:
		return nil, fmt.Errorf("driver \"%s\" doesn't support", parsed.Driver)
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
//
//nolint:ireturn,nolintlint // its ok
func (c *Connection) QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

// ExecContext executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
//
//nolint:ireturn,nolintlint // its ok
func (c *Connection) ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error) {
	tx, err := TxFromContext(ctx)
	if err != nil {
		return c.db.ExecContext(ctx, query, args...)
	}

	// maybe need to clickhouse
	// stmt, err := tx.PrepareContext(ctx, query)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// return stmt.ExecContext(ctx, args...)

	return tx.ExecContext(ctx, query, args...)
}

// Transaction executes body in func txFn into transaction.
func (c *Connection) Transaction(ctx context.Context, txFn func(ctx context.Context) error) error {
	if _, err := TxFromContext(ctx); err == nil {
		return ErrTransactionAlreadyOpened
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "begin transaction")
	}

	if err := txFn(ContextWithTx(ctx, tx)); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Wrapf(err, "rollback failed: %v", rbErr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return errors.Wrap(err, "commit transaction")
	}

	return nil
}

func (c *Connection) Close() error {
	return c.db.Close()
}

func Try(dsn string, maxAttempts int) (*Connection, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var (
		conn *Connection
		err  error
	)
	for i := 0; i < maxAttempts; i++ {
		conn, err = New(dsn)
		if err == nil {
			if err = conn.Ping(); err == nil {
				return conn, nil
			}
		}

		if i < maxAttempts-1 {
			time.Sleep(durationBeforeNextConnAttempt)
		}
	}

	return nil, err
}

// clickhouse returns connection with clickhouse configuration.
func clickhouse(dsn string) (*Connection, error) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, err
	}

	return &Connection{
		driver: DriverClickhouse,
		dsn:    dsn,
		db:     &sqlex.DB{DB: db},
	}, nil
}

// postgres returns connection with postgres configuration.
func postgres(dsn string) (*Connection, error) {
	db, err := sql.Open(DriverPostgres.String(), dsn)
	if err != nil {
		return nil, errors.Wrap(err, "open postgres connection")
	}

	return &Connection{
		driver: DriverPostgres,
		dsn:    dsn,
		db:     &sqlex.DB{DB: db},
	}, nil
}

// mysql returns connection with mysql configuration.
func mysql(dsn string) (*Connection, error) {
	db, err := sql.Open(DriverMySQL.String(), dsn[8:])
	if err != nil {
		return nil, errors.Wrap(err, "open mysql connection")
	}

	return &Connection{
		driver: DriverMySQL,
		dsn:    dsn,
		db:     &sqlex.DB{DB: db},
	}, nil
}

// tarantool returns connection with tarantool configuration.
func tarantoolConn(dsn string) (*Connection, error) {
	db, err := tarantool.Open(dsn)
	if err != nil {
		return nil, err
	}

	return &Connection{
		driver: DriverTarantool,
		dsn:    dsn,
		db:     db,
	}, nil
}
