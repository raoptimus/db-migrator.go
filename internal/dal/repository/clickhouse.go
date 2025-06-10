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
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

type Clickhouse struct {
	conn    Connection
	options *Options
}

func NewClickhouse(conn Connection, options *Options) *Clickhouse {
	return &Clickhouse{
		conn:    conn,
		options: options,
	}
}

// Migrations returns applied migrations history.
func (c *Clickhouse) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	var (
		q = `
			SELECT version, apply_time 
			FROM ` + c.dTableNameWithSchema() + `
			WHERE is_deleted = 0 
			ORDER BY apply_time DESC, version DESC
			LIMIT ?
		`
		migrations entity.Migrations
	)

	rows, err := c.conn.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, errors.Wrap(c.dbError(err, q), "get migrations")
	}
	for rows.Next() {
		var (
			version   string
			applyTime int
		)

		if err := rows.Scan(&version, &applyTime); err != nil {
			return nil, errors.Wrap(c.dbError(err, q), "get migrations")
		}

		migrations = append(migrations,
			entity.Migration{
				Version:   version,
				ApplyTime: applyTime,
			},
		)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(c.dbError(err, q), "get migrations")
	}

	return migrations, nil
}

// HasMigrationHistoryTable returns true if migration history table exists.
func (c *Clickhouse) HasMigrationHistoryTable(ctx context.Context) (exists bool, err error) {
	var (
		q = `
			SELECT database, table 
			FROM system.columns 
			WHERE table = ? AND database = currentDatabase()
		`
		rows *sql.Rows
	)

	rows, err = c.conn.QueryContext(ctx, q, c.dTableName())
	if err != nil {
		return false, errors.Wrap(c.dbError(err, q), "get table schema")
	}

	for rows.Next() {
		var (
			database string
			table    string
		)
		if err := rows.Scan(&database, &table); err != nil {
			return false, errors.Wrap(c.dbError(err, q), "get table schema")
		}

		//todo: scan columns to tableScheme
		if table == c.options.TableName {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, errors.Wrap(c.dbError(err, q), "get table schema")
	}

	return false, nil
}

// InsertMigration inserts the new migration record.
func (c *Clickhouse) InsertMigration(ctx context.Context, version string) error {
	return c.insertMigration(ctx, version, false)
}

// RemoveMigration removes the migration record.
func (c *Clickhouse) RemoveMigration(ctx context.Context, version string) error {
	return c.insertMigration(ctx, version, true)
}

// ExecQuery executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (c *Clickhouse) ExecQuery(ctx context.Context, query string, args ...any) error {
	_, err := c.conn.ExecContext(ctx, query, args...)
	return err
}

// ExecQueryTransaction executes txFn in transaction.
// todo: называется ExecQuery но query не принимает. подумать
func (c *Clickhouse) ExecQueryTransaction(ctx context.Context, txFn func(ctx context.Context) error) error {
	return c.conn.Transaction(ctx, txFn)
}

// CreateMigrationHistoryTable creates a new migration history table.
func (c *Clickhouse) CreateMigrationHistoryTable(ctx context.Context) error {
	var (
		q         string
		extQ      string
		engine    string
		onCluster string
	)

	switch {
	case c.isUsedCluster():
		onCluster = "ON CLUSTER " + c.options.ClusterName
		engine = "ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/" +
			c.options.ClusterName + "_" + c.options.TableName + "', '{replica}', apply_time)"
		extQ = fmt.Sprintf(`
				CREATE TABLE %[2]s.d_%[3]s ON CLUSTER %[1]s AS %[2]s.%[3]s
				ENGINE = Distributed('%[1]s', '%[2]s', %[3]s, cityHash64(toString(version)))
			`,
			c.options.ClusterName,
			c.options.SchemaName,
			c.options.TableName,
		)
	case c.options.Replicated:
		engine = "ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/" +
			c.options.ClusterName + "_" + c.options.TableName + "', '{replica}', apply_time)"
	default:
		engine = "ReplacingMergeTree(apply_time)"
	}

	q = fmt.Sprintf(
		`
			CREATE TABLE %s %s (
				version String, 
				date Date DEFAULT toDate(apply_time),
				apply_time UInt32,
				is_deleted UInt8
			) ENGINE = %s
			PRIMARY KEY (version)
			PARTITION BY (toYYYYMM(date))
			ORDER BY (version)
			SETTINGS index_granularity=8192
			`,
		c.TableNameWithSchema(),
		onCluster,
		engine,
	)

	if _, err := c.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(c.dbError(err, q), "create migration history table")
	}

	if len(extQ) == 0 {
		return nil
	}

	if _, err := c.conn.ExecContext(ctx, extQ); err != nil {
		return errors.Wrap(c.dbError(err, extQ), "create migration history table")
	}

	return nil
}

// DropMigrationHistoryTable drops the migration history table.
func (c *Clickhouse) DropMigrationHistoryTable(ctx context.Context) error {
	if err := c.dropTable(ctx, c.TableNameWithSchema()); err != nil {
		return err
	}

	if !c.isUsedCluster() {
		return nil
	}

	if err := c.dropTable(ctx, c.dTableNameWithSchema()); err != nil {
		return err
	}

	return nil
}

func (c *Clickhouse) dropTable(ctx context.Context, tableName string) error {
	q := "DROP TABLE " + tableName
	if c.isUsedCluster() {
		q += " ON CLUSTER " + c.options.ClusterName + " NO DELAY"
	}
	if _, err := c.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(c.dbError(err, q), "drop migration history table")
	}

	return nil
}

// MigrationsCount returns the number of migrations
func (c *Clickhouse) MigrationsCount(ctx context.Context) (int, error) {
	q := "SELECT count(*) FROM " + c.dTableNameWithSchema() + " WHERE is_deleted = 0"
	rows, err := c.conn.QueryContext(ctx, q)
	if err != nil {
		return 0, err
	}
	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, c.dbError(err, q)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, c.dbError(err, q)
	}
	return count, nil
}

func (c *Clickhouse) ExistsMigration(ctx context.Context, version string) (bool, error) {
	q := "SELECT 1 FROM " + c.dTableNameWithSchema() + " WHERE version = ? AND is_deleted = 0"
	rows, err := c.conn.QueryContext(ctx, q, version)
	if err != nil {
		return false, err
	}
	var exists int
	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, c.dbError(err, q)
		}
	}
	if err := rows.Err(); err != nil {
		return false, c.dbError(err, q)
	}

	return exists == 1, nil
}

func (c *Clickhouse) TableNameWithSchema() string {
	return c.options.SchemaName + "." + c.options.TableName
}

func (c *Clickhouse) ForceSafely() bool {
	return false
}

func (c *Clickhouse) dTableName() string {
	if c.isUsedCluster() {
		return "_d" + c.options.TableName
	}

	return c.options.TableName
}

func (c *Clickhouse) dTableNameWithSchema() string {
	return c.options.SchemaName + "." + c.dTableName()
}

func (c *Clickhouse) isUsedCluster() bool {
	return !c.options.Replicated && len(c.options.ClusterName) > 0
}

// insertMigration inserts migration record.
func (c *Clickhouse) insertMigration(ctx context.Context, version string, isDeleted bool) error {
	q := `
		INSERT INTO ` + c.dTableNameWithSchema() + ` (version, apply_time, is_deleted) 
		VALUES(?, ?, ?)
	`

	//nolint:gosec // overflow ok
	now := uint32(time.Now().Unix())
	var isDeletedInt int
	if isDeleted {
		isDeletedInt = 1
	}

	if err := c.ExecQueryTransaction(ctx, func(ctx context.Context) error {
		return c.ExecQuery(ctx, q, version, now, isDeletedInt)
	}); err != nil {
		return errors.Wrap(c.dbError(err, q), "insert migration")
	}

	return c.optimizeTable(ctx)
}

// optimizeTable optimizes tables.
func (c *Clickhouse) optimizeTable(ctx context.Context) error {
	var q string
	if c.options.Replicated || c.options.ClusterName == "" {
		q = fmt.Sprintf("OPTIMIZE TABLE %s FINAL", c.options.TableName)
	} else {
		q = fmt.Sprintf("OPTIMIZE TABLE %s ON CLUSTER %s FINAL", c.options.TableName, c.options.ClusterName)
	}
	if _, err := c.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(c.dbError(err, q), "optimize table")
	}

	return nil
}

// dbError returns DBError is err is db error else returns got error.
func (c *Clickhouse) dbError(err error, q string) error {
	var clickEx *clickhouse.Exception
	if !errors.As(err, &clickEx) {
		return err
	}

	return &DBError{
		Code:          string(clickEx.Code),
		Message:       clickEx.Message,
		Details:       clickEx.StackTrace,
		InternalQuery: q,
	}
}
