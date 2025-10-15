/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package repository

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
	"github.com/tarantool/go-tarantool/v2"
)

const tarantoolIteratorLT = "LT"
const tarantoolIteratorEQ = "EQ"

type Tarantool struct {
	conn    Connection
	options *Options
}

func NewTarantool(conn Connection, options *Options) *Tarantool {
	return &Tarantool{
		conn:    conn,
		options: options,
	}
}

// Migrations returns applied migrations history.
func (p *Tarantool) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	var migrations entity.Migrations

	q := fmt.Sprintf("return box.space.%s:select({}, {iterator='%s', limit = %d})",
		p.TableNameWithSchema(),
		tarantoolIteratorLT,
		limit,
	)
	rows, err := p.conn.QueryContext(ctx, q)
	if err != nil {
		return nil, errors.Wrap(p.dbError(err, q), "get migrations")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			version   string
			applyTime int64
		)

		if err := rows.Scan(&version, &applyTime); err != nil {
			return nil, errors.Wrap(p.dbError(err, q), "get migrations")
		}

		migrations = append(migrations,
			entity.Migration{
				Version:   version,
				ApplyTime: applyTime,
			},
		)
	}

	return migrations, nil
}

// HasMigrationHistoryTable returns true if migration history table exists.
func (p *Tarantool) HasMigrationHistoryTable(ctx context.Context) (exists bool, err error) {
	q := fmt.Sprintf("return box.space.%s ~= nil", p.TableNameWithSchema())

	rows, err := p.conn.QueryContext(ctx, q)
	if err != nil {
		return false, errors.Wrap(p.dbError(err, q), "get schema table")
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, errors.Wrap(p.dbError(err, q), "get schema table")
		}

		if exists {
			return true, nil
		}
	}

	return false, nil
}

// InsertMigration inserts the new migration record.
func (p *Tarantool) InsertMigration(ctx context.Context, version string) error {
	q := fmt.Sprintf("box.space.%s:insert({'%s', %d})", p.TableNameWithSchema(), version, time.Now().Unix())

	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "insert migration")
	}
	return nil
}

// RemoveMigration removes the migration record.
func (p *Tarantool) RemoveMigration(ctx context.Context, version string) error {
	q := fmt.Sprintf("box.space.%s:delete('%s')", p.TableNameWithSchema(), version)

	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "remove migration")
	}

	return nil
}

// ExecQuery executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (p *Tarantool) ExecQuery(ctx context.Context, query string, args ...any) error {
	if _, err := p.conn.ExecContext(ctx, query, args...); err != nil {
		return p.dbError(err, query)
	}

	return nil
}

// ExecQueryTransaction executes a query in transaction without returning any rows.
// The args are for any placeholder parameters in the query.
func (p *Tarantool) ExecQueryTransaction(ctx context.Context, txFn func(ctx context.Context) error) error {
	return p.conn.Transaction(ctx, txFn)
}

// CreateMigrationHistoryTable creates a new migration history table.
func (p *Tarantool) CreateMigrationHistoryTable(ctx context.Context) error {
	// create space
	q := fmt.Sprintf("box.schema.space.create('%s', {if_not_exists = true})", p.TableNameWithSchema())
	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "create migration history table")
	}

	// set space format
	q = fmt.Sprintf("box.space.%s:format", p.TableNameWithSchema())
	q += "({{'version',type = 'string',is_nullable = false},{'apply_time', type = 'unsigned', is_nullable = false}})"
	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "create migration history table")
	}

	// create primary index
	q = fmt.Sprintf("box.space.%s:create_index", p.TableNameWithSchema())
	q += "('primary', {parts = {'version'}, if_not_exists = true})"
	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "create migration history table")
	}

	// create secondary index
	q = fmt.Sprintf("box.space.%s:create_index", p.TableNameWithSchema())
	q += "('secondary', {parts = {{'apply_time'}, {'version'}}, if_not_exists = true})"
	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "create migration history table")
	}

	return nil
}

// DropMigrationHistoryTable drops the migration history table.
func (p *Tarantool) DropMigrationHistoryTable(ctx context.Context) error {
	q := fmt.Sprintf("box.space.%s:drop()", p.TableNameWithSchema())

	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "drop migration history table")
	}

	return nil
}

// MigrationsCount returns the number of migrations
func (p *Tarantool) MigrationsCount(ctx context.Context) (int, error) {
	q := fmt.Sprintf("return box.space.%s:len()", p.TableNameWithSchema())
	var c int
	
	return c, p.QueryScalar(ctx, q, &c)
}

// QueryScalar returns the number of records by query
func (p *Tarantool) QueryScalar(ctx context.Context, query string, ptr any) error {
	if err := checkArgIsPtrAndScalar(ptr); err != nil {
		return err
	}
	rows, err := p.conn.QueryContext(ctx, query)
	if err != nil {
		return p.dbError(err, query)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(ptr); err != nil {
			return p.dbError(err, query)
		}
	}
	if err := rows.Err(); err != nil {
		return p.dbError(err, query)
	}

	return nil
}

func (p *Tarantool) ExistsMigration(ctx context.Context, version string) (bool, error) {
	q := fmt.Sprintf("box.space.%s:count('%s', {iterator='%s'})",
		p.TableNameWithSchema(),
		version,
		tarantoolIteratorEQ,
	)

	rows, err := p.conn.QueryContext(ctx, q)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var exists bool
	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, p.dbError(err, q)
		}
	}
	if err := rows.Err(); err != nil {
		return false, p.dbError(err, q)
	}

	return exists, nil
}

func (p *Tarantool) TableNameWithSchema() string {
	return p.options.TableName
}

// dbError returns DBError is err is db error else returns got error.
func (p *Tarantool) dbError(err error, q string) error {
	var tErr tarantool.Error
	if !errors.As(err, &tErr) {
		return err
	}

	return &DBError{
		Code:          strconv.Itoa(int(tErr.Code)),
		Severity:      "ERR",
		Message:       tErr.Msg,
		Details:       "",
		InternalQuery: q,
	}
}
