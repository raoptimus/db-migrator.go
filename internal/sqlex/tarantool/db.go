/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package tarantool

import (
	"context"
	"database/sql"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/sqlex"
	"github.com/tarantool/go-tarantool/v2"
)

const defaultQueryTimeout = 10 * time.Second

//go:generate mockery
type Connection interface {
	Do(req tarantool.Request) *tarantool.Future
	NewStream() (*tarantool.Stream, error)
	Close() error
}

type DB struct {
	conn Connection
}

func Open(dsn string) (*DB, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.WithMessage(err, "parse tarantool DSN")
	}
	pass, _ := dsnURL.User.Password()

	dialer := tarantool.NetDialer{
		Address:  dsnURL.Host,
		User:     dsnURL.User.Username(),
		Password: pass,
	}
	opts := tarantool.Opts{
		Timeout: defaultQueryTimeout, //todo: extract from query param dsn
	}

	conn, err := tarantool.Connect(context.Background(), dialer, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "connect to tarantool at %s", dsnURL.Host)
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Ping() error {
	_, err := db.conn.Do(tarantool.NewPingRequest()).GetResponse()
	if err != nil {
		return err
	}

	return nil
}

//nolint:ireturn,nolintlint // its ok
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (sqlex.Rows, error) {
	req := tarantool.NewEvalRequest(query).Context(ctx)
	if len(args) > 0 {
		req = req.Args(args)
	}
	data, err := db.conn.
		Do(req).
		Get()
	if err != nil {
		return nil, err
	}

	if len(data) == 1 {
		if indata, ok := data[0].([]any); ok {
			data = indata
		}
	}

	return sqlex.NewRowsWithSlice(data), nil
}

//nolint:ireturn,nolintlint // its ok
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sqlex.Result, error) {
	req := tarantool.NewEvalRequest(query).Context(ctx)
	if len(args) > 0 {
		req = req.Args(args)
	}
	_, err := db.conn.Do(req).Get()
	if err != nil {
		return nil, err
	}

	return Done(true), nil
}

//nolint:ireturn,nolintlint // its ok
func (db *DB) BeginTx(ctx context.Context, _ *sql.TxOptions) (sqlex.Tx, error) {
	stream, err := db.conn.NewStream()
	if err != nil {
		return nil, err
	}

	_, err = stream.
		Do(tarantool.NewBeginRequest().Context(ctx).IsSync(true)).
		Get()
	if err != nil {
		return nil, err
	}

	return NewTx(stream), nil
}

func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}

	return nil
}
