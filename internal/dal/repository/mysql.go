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

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
	"github.com/raoptimus/db-migrator.go/internal/sqlex"
)

type MySQL struct {
	conn    Connection
	options *Options
}

func NewMySQL(conn Connection, options *Options) *MySQL {
	return &MySQL{
		conn:    conn,
		options: options,
	}
}

// Migrations returns applied migrations history.
func (m *MySQL) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	var (
		q = fmt.Sprintf(
			`
			SELECT version, apply_time 
			FROM %s
			ORDER BY apply_time DESC, version DESC
			LIMIT ?`,
			m.options.TableName,
		)
		migrations entity.Migrations
	)

	rows, err := m.conn.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, errors.Wrap(m.dbError(err, q), "get migrations")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			version   string
			applyTime int64
		)

		if err := rows.Scan(&version, &applyTime); err != nil {
			return nil, errors.Wrap(m.dbError(err, q), "get migrations")
		}

		migrations = append(migrations,
			entity.Migration{
				Version:   version,
				ApplyTime: applyTime,
			},
		)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(m.dbError(err, q), "get migrations")
	}

	return migrations, nil
}

// HasMigrationHistoryTable returns true if migration history table exists.
func (m *MySQL) HasMigrationHistoryTable(ctx context.Context) (exists bool, err error) {
	var (
		q = `
			SELECT EXISTS(
			    SELECT *
				FROM information_schema.tables
				WHERE table_name = ? AND table_schema = ?
			)
		`
		rows sqlex.Rows
	)

	rows, err = m.conn.QueryContext(ctx, q, m.options.TableName, m.options.SchemaName)
	if err != nil {
		return false, errors.Wrap(m.dbError(err, q), "get schema table")
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, errors.Wrap(m.dbError(err, q), "get schema table")
		}
	}

	if err = rows.Err(); err != nil {
		return false, errors.Wrap(m.dbError(err, q), "get schema table")
	}

	return exists, nil
}

// InsertMigration inserts the new migration record.
func (m *MySQL) InsertMigration(ctx context.Context, version string) error {
	q := fmt.Sprintf(`
		INSERT INTO %s (version, apply_time)
		VALUES (?, ?)`,
		m.options.TableName,
	)
	//nolint:gosec // overflow ok
	now := uint32(time.Now().Unix())
	if _, err := m.conn.ExecContext(ctx, q, version, now); err != nil {
		return errors.Wrap(m.dbError(err, q), "insert migration")
	}
	return nil
}

// RemoveMigration removes the migration record.
func (m *MySQL) RemoveMigration(ctx context.Context, version string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE version = ?`, m.options.TableName)
	if _, err := m.conn.ExecContext(ctx, q, version); err != nil {
		return errors.Wrap(m.dbError(err, q), "remove migration")
	}
	return nil
}

// ExecQuery executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (m *MySQL) ExecQuery(ctx context.Context, query string, args ...any) error {
	_, err := m.conn.ExecContext(ctx, query, args...)
	return err
}

// ExecQueryTransaction executes a query in transaction without returning any rows.
// The args are for any placeholder parameters in the query.
func (m *MySQL) ExecQueryTransaction(ctx context.Context, txFn func(ctx context.Context) error) error {
	return m.conn.Transaction(ctx, txFn)
}

// CreateMigrationHistoryTable creates a new migration history table.
func (m *MySQL) CreateMigrationHistoryTable(ctx context.Context) error {
	q := fmt.Sprintf(
		`
				CREATE TABLE %s (
				  version VARCHAR(180) PRIMARY KEY,
				  apply_time INT
				)
				ENGINE=InnoDB
			`,
		m.options.TableName,
	)

	if _, err := m.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(m.dbError(err, q), "create migration history table")
	}
	return nil
}

// DropMigrationHistoryTable drops the migration history table.
func (m *MySQL) DropMigrationHistoryTable(ctx context.Context) error {
	q := fmt.Sprintf(`DROP TABLE %s`, m.options.TableName)
	if _, err := m.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(m.dbError(err, q), "drop migration history table")
	}
	return nil
}

// MigrationsCount returns the number of migrations
func (m *MySQL) MigrationsCount(ctx context.Context) (int, error) {
	q := fmt.Sprintf(`SELECT count(*) FROM %s`, m.options.TableName)
	var c int

	return c, m.QueryScalar(ctx, q, &c)
}

func (m *MySQL) QueryScalar(ctx context.Context, query string, ptr any) error {
	if err := checkArgIsPtrAndScalar(ptr); err != nil {
		return err
	}
	rows, err := m.conn.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(ptr); err != nil {
			return m.dbError(err, query)
		}
	}
	if err := rows.Err(); err != nil {
		return m.dbError(err, query)
	}

	return nil
}

func (m *MySQL) ExistsMigration(ctx context.Context, version string) (bool, error) {
	q := fmt.Sprintf(`SELECT 1 FROM %s WHERE version = ?`, m.TableNameWithSchema())
	rows, err := m.conn.QueryContext(ctx, q, version)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var exists int
	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, m.dbError(err, q)
		}
	}
	if err := rows.Err(); err != nil {
		return false, m.dbError(err, q)
	}

	return exists == 1, nil
}

func (m *MySQL) TableNameWithSchema() string {
	return m.options.SchemaName + "." + m.options.TableName
}

// dbError returns DBError is err is db error else returns got error.
func (m *MySQL) dbError(err error, q string) error {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		return err
	}

	return &DBError{
		Code:          strconv.Itoa(int(mysqlErr.Number)),
		Message:       mysqlErr.Message,
		InternalQuery: q,
	}
}
