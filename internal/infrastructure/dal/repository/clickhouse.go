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
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/sqlex"
)

// Clickhouse implements Repository interface for ClickHouse database.
// It handles migration history tracking and SQL execution for ClickHouse with support for clusters and replication.
type Clickhouse struct {
	conn    Connection
	options *Options
}

// NewClickhouse creates a new Clickhouse repository instance.
// It returns a repository configured with the provided connection and options.
func NewClickhouse(conn Connection, options *Options) *Clickhouse {
	return &Clickhouse{
		conn:    conn,
		options: options,
	}
}

// Migrations returns applied migrations history.
func (ch *Clickhouse) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	var (
		q = `
			SELECT version, apply_time 
			FROM ` + ch.dTableNameWithSchema() + `
			WHERE is_deleted = 0 
			ORDER BY apply_time DESC, version DESC
			LIMIT ?
		`
		migrations entity.Migrations
	)

	rows, err := ch.conn.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, errors.Wrap(ch.dbError(err, q), "get migrations")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			version   string
			applyTime int64
		)

		if err := rows.Scan(&version, &applyTime); err != nil {
			return nil, errors.Wrap(ch.dbError(err, q), "get migrations")
		}

		migrations = append(migrations,
			entity.Migration{
				Version:   version,
				ApplyTime: applyTime,
			},
		)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(ch.dbError(err, q), "get migrations")
	}

	return migrations, nil
}

// HasMigrationHistoryTable returns true if migration history table exists.
func (ch *Clickhouse) HasMigrationHistoryTable(ctx context.Context) (exists bool, err error) {
	var (
		q = `
			SELECT database, table 
			FROM system.columns 
			WHERE table = ? AND database = currentDatabase()
		`
		rows sqlex.Rows
	)

	rows, err = ch.conn.QueryContext(ctx, q, ch.dTableName())
	if err != nil {
		return false, errors.Wrap(ch.dbError(err, q), "get table schema")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			database string
			table    string
		)
		if err := rows.Scan(&database, &table); err != nil {
			return false, errors.Wrap(ch.dbError(err, q), "get table schema")
		}

		//todo: scan columns to tableScheme
		if table == ch.dTableName() {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, errors.Wrap(ch.dbError(err, q), "get table schema")
	}

	return false, nil
}

// InsertMigration inserts the new migration record.
func (ch *Clickhouse) InsertMigration(ctx context.Context, version string) error {
	return ch.insertMigration(ctx, version, false)
}

// RemoveMigration removes the migration record.
func (ch *Clickhouse) RemoveMigration(ctx context.Context, version string) error {
	return ch.insertMigration(ctx, version, true)
}

// ExecQuery executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (ch *Clickhouse) ExecQuery(ctx context.Context, query string, args ...any) error {
	_, err := ch.conn.ExecContext(ctx, query, args...)

	return ch.dbError(err, query)
}

// ExecQueryTransaction executes txFn in transaction.
// todo: называется ExecQuery но query не принимает. подумать
func (ch *Clickhouse) ExecQueryTransaction(ctx context.Context, txFn func(ctx context.Context) error) error {
	return ch.conn.Transaction(ctx, txFn)
}

// CreateMigrationHistoryTable creates a new migration history table.
func (ch *Clickhouse) CreateMigrationHistoryTable(ctx context.Context) error {
	var (
		q         string
		extQ      string
		engine    string
		onCluster string
	)

	switch {
	case ch.isUsedCluster():
		onCluster = "ON CLUSTER " + ch.options.ClusterName
		engine = "ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/" +
			ch.options.ClusterName + "_" + ch.options.TableName + "', '{replica}', apply_time)"
		extQ = fmt.Sprintf(`
				CREATE TABLE %[2]s.d_%[3]s ON CLUSTER %[1]s AS %[2]s.%[3]s
				ENGINE = Distributed('%[1]s', '%[2]s', %[3]s, cityHash64(toString(version)))
			`,
			ch.options.ClusterName,
			ch.options.SchemaName,
			ch.options.TableName,
		)
	case ch.options.Replicated:
		engine = "ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/" +
			ch.options.ClusterName + "_" + ch.options.TableName + "', '{replica}', apply_time)"
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
		ch.TableNameWithSchema(),
		onCluster,
		engine,
	)

	if _, err := ch.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(ch.dbError(err, q), "create migration history table")
	}

	if len(extQ) == 0 {
		return nil
	}

	if _, err := ch.conn.ExecContext(ctx, extQ); err != nil {
		return errors.Wrap(ch.dbError(err, extQ), "create migration history table")
	}

	return nil
}

// DropMigrationHistoryTable drops the migration history table.
func (ch *Clickhouse) DropMigrationHistoryTable(ctx context.Context) error {
	if err := ch.dropTable(ctx, ch.TableNameWithSchema()); err != nil {
		return err
	}

	if !ch.isUsedCluster() {
		return nil
	}

	if err := ch.dropTable(ctx, ch.dTableNameWithSchema()); err != nil {
		return err
	}

	return nil
}

// MigrationsCount returns the number of migrations
func (ch *Clickhouse) MigrationsCount(ctx context.Context) (int, error) {
	q := "SELECT count(*) FROM " + ch.dTableNameWithSchema() + " WHERE is_deleted = 0"
	var c int
	if err := ch.QueryScalar(ctx, q, &c); err != nil {
		return 0, err
	}

	return c, nil
}

// QueryScalar executes a query and scans a single scalar value into ptr.
// The ptr parameter must be a pointer to a scalar type (int, string, bool, etc).
func (ch *Clickhouse) QueryScalar(ctx context.Context, query string, ptr any) error {
	if err := checkArgIsPtrAndScalar(ptr); err != nil {
		return err
	}
	rows, err := ch.conn.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(ptr); err != nil {
			return ch.dbError(err, query)
		}
	}
	if err := rows.Err(); err != nil {
		return ch.dbError(err, query)
	}

	return nil
}

// ExistsMigration checks if a migration with the given version exists in the history table.
// It returns true if the migration record is found and not marked as deleted, false otherwise.
func (ch *Clickhouse) ExistsMigration(ctx context.Context, version string) (bool, error) {
	q := "SELECT 1 FROM " + ch.dTableNameWithSchema() + " WHERE version = ? AND is_deleted = 0"
	rows, err := ch.conn.QueryContext(ctx, q, version)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var exists int
	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, ch.dbError(err, q)
		}
	}
	if err := rows.Err(); err != nil {
		return false, ch.dbError(err, q)
	}

	return exists == 1, nil
}

// TableNameWithSchema returns the fully qualified table name with schema prefix.
// For ClickHouse, it returns database.table_name format.
func (ch *Clickhouse) TableNameWithSchema() string {
	return ch.options.SchemaName + "." + ch.options.TableName
}

// dropTable drops a table by name, using cluster-aware syntax if cluster is configured.
func (ch *Clickhouse) dropTable(ctx context.Context, tableName string) error {
	q := "DROP TABLE " + tableName
	if ch.isUsedCluster() {
		q += " ON CLUSTER " + ch.options.ClusterName + " NO DELAY"
	}
	if _, err := ch.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(ch.dbError(err, q), "drop migration history table")
	}

	return nil
}

// dTableName returns the distributed table name for cluster deployments.
// It adds the "d_" prefix to the table name when a cluster is used, otherwise returns the original table name.
func (ch *Clickhouse) dTableName() string {
	if ch.isUsedCluster() {
		return "d_" + ch.options.TableName
	}

	return ch.options.TableName
}

// dTableNameWithSchema returns the fully qualified distributed table name with schema prefix.
// For ClickHouse clusters, it returns database.d_table_name format.
func (ch *Clickhouse) dTableNameWithSchema() string {
	return ch.options.SchemaName + "." + ch.dTableName()
}

// isUsedCluster checks if ClickHouse cluster mode is enabled.
// Returns true when a cluster name is configured and replication is not explicitly enabled.
func (ch *Clickhouse) isUsedCluster() bool {
	return !ch.options.Replicated && len(ch.options.ClusterName) > 0
}

// insertMigration inserts migration record.
func (ch *Clickhouse) insertMigration(ctx context.Context, version string, isDeleted bool) error {
	q := `
		INSERT INTO ` + ch.dTableNameWithSchema() + ` (version, apply_time, is_deleted) 
		VALUES(?, ?, ?)
	`

	//nolint:gosec // overflow ok
	now := uint32(time.Now().Unix())
	var isDeletedInt int
	if isDeleted {
		isDeletedInt = 1
	}

	if err := ch.ExecQueryTransaction(ctx, func(ctx context.Context) error {
		return ch.ExecQuery(ctx, q, version, now, isDeletedInt)
	}); err != nil {
		return errors.Wrap(ch.dbError(err, q), "insert migration")
	}

	return ch.optimizeTable(ctx)
}

// optimizeTable optimizes tables.
func (ch *Clickhouse) optimizeTable(ctx context.Context) error {
	var q string
	if ch.options.Replicated || ch.options.ClusterName == "" {
		q = fmt.Sprintf("OPTIMIZE TABLE %s FINAL", ch.options.TableName)
	} else {
		q = fmt.Sprintf("OPTIMIZE TABLE %s ON CLUSTER %s FINAL", ch.options.TableName, ch.options.ClusterName)
	}

	if _, err := ch.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(ch.dbError(err, q), "optimize table")
	}

	return nil
}

// dbError returns DBError is err is db error else returns got error.
func (ch *Clickhouse) dbError(err error, q string) error {
	var clickEx *clickhouse.Exception
	if !errors.As(err, &clickEx) {
		return err
	}

	return errors.WithStack(&DBError{
		Code:          string(clickEx.Code),
		Message:       clickEx.Message,
		Details:       clickEx.StackTrace,
		InternalQuery: q,
		Cause:         err,
	})
}
