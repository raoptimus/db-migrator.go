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

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/sqlex"
	"github.com/tarantool/go-tarantool/v2"
)

type DB struct {
	conn *tarantool.Connection
}

func Open(dsn string) (*DB, error) {
	dsnURL, err := url.Parse(dsn)
	if err != nil {
		return nil, errors.WithMessage(err, "parse tarantool DSN")
	}
	pass, hasPassword := dsnURL.User.Password()
	if !hasPassword {
		pass = ""
	}

	dialer := tarantool.NetDialer{
		Address:  dsnURL.Host,
		User:     dsnURL.User.Username(),
		Password: pass,
	}
	opts := tarantool.Opts{
		//Timeout: time.Second,
	}
	ctx := context.Background()
	conn, err := tarantool.Connect(ctx, dialer, opts)
	if err != nil {
		return nil, errors.Wrapf(err, "connect to tarantool at %s", dsnURL.Host)
	}

	_, err = conn.Do(
		tarantool.NewCallRequest("box.cfg").
			Args([]interface{}{map[string]interface{}{
				"txn_isolation": tarantool.ReadCommittedLevel,
			}}),
	).Get()
	if err != nil {
		return nil, errors.Wrapf(err, "configure the tarantool at %s", dsnURL.Host)
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
