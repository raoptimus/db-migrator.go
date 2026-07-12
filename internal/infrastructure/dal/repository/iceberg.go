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
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/iceberg/ddl"
)

// ErrNotImplemented is returned by Iceberg repository methods that are not yet
// implemented. History-methods are implemented in task 02;
// DDL execution (ExecQuery) is implemented in task 04.
var ErrNotImplemented = errors.New("iceberg repository: not implemented")

// icebergHistoryKeyPrefix is the key prefix for migration history entries stored
// as namespace properties. Full key format: "mig.<version>".
const icebergHistoryKeyPrefix = "mig."

// Iceberg implements Repository for the Apache Iceberg REST catalog backend.
// Migration history is stored as namespace properties (ARD Р3).
// DDL execution is implemented in task 04 via the iceberg/ddl parser.
type Iceberg struct {
	cat     IcebergCatalog
	options *Options
}

// NewIceberg creates a new Iceberg repository instance.
func NewIceberg(cat IcebergCatalog, options *Options) *Iceberg {
	return &Iceberg{cat: cat, options: options}
}

// TableNameWithSchema returns the migration history namespace name.
func (i *Iceberg) TableNameWithSchema() string {
	return i.options.TableName
}

// SupportsDDLTransactions returns false: Iceberg REST catalog is per-table atomic
// but does not support multi-table DDL transactions (ARD Р4).
func (i *Iceberg) SupportsDDLTransactions() bool {
	return false
}

// ExecQueryTransaction is a passthrough: Iceberg does not use SQL-level transactions
// for DDL. The outer .safe mechanism becomes best-effort per-table (ARD Р4).
func (i *Iceberg) ExecQueryTransaction(ctx context.Context, fnTx func(ctx context.Context) error) error {
	return fnTx(ctx)
}

// historyNS returns the namespace identifier used for the migration history.
// The history namespace name equals the configured TableName (default "migration").
func (i *Iceberg) historyNS() []string {
	return []string{i.options.TableName}
}

// CreateMigrationHistoryTable creates the history namespace in the catalog.
func (i *Iceberg) CreateMigrationHistoryTable(ctx context.Context) error {
	if err := i.cat.CreateNamespace(ctx, i.historyNS(), nil); err != nil {
		return errors.Wrap(i.dbError(err), "create migration history table")
	}
	return nil
}

// DropMigrationHistoryTable drops the history namespace from the catalog.
func (i *Iceberg) DropMigrationHistoryTable(ctx context.Context) error {
	if err := i.cat.DropNamespace(ctx, i.historyNS()); err != nil {
		return errors.Wrap(i.dbError(err), "drop migration history table")
	}
	return nil
}

// HasMigrationHistoryTable checks whether the history namespace exists in the catalog.
func (i *Iceberg) HasMigrationHistoryTable(ctx context.Context) (bool, error) {
	exists, err := i.cat.NamespaceExists(ctx, i.historyNS())
	if err != nil {
		return false, errors.Wrap(i.dbError(err), "check migration history table")
	}
	return exists, nil
}

// InsertMigration inserts a migration record using the current wall-clock time as apply_time.
func (i *Iceberg) InsertMigration(ctx context.Context, version string) error {
	return i.InsertMigrationWithApplyTime(ctx, version, time.Now().Unix())
}

// InsertMigrationWithApplyTime inserts a migration record with an explicit apply_time.
// The record is stored as namespace property "mig.<version>" = "<apply_time>".
func (i *Iceberg) InsertMigrationWithApplyTime(ctx context.Context, version string, applyTime int64) error {
	updates := map[string]string{
		icebergHistoryKeyPrefix + version: strconv.FormatInt(applyTime, 10),
	}
	if err := i.cat.UpdateNamespaceProperties(ctx, i.historyNS(), nil, updates); err != nil {
		return errors.Wrap(i.dbError(err), "insert migration")
	}
	return nil
}

// RemoveMigration removes a migration record from the history namespace properties.
func (i *Iceberg) RemoveMigration(ctx context.Context, version string) error {
	removals := []string{icebergHistoryKeyPrefix + version}
	if err := i.cat.UpdateNamespaceProperties(ctx, i.historyNS(), removals, nil); err != nil {
		return errors.Wrap(i.dbError(err), "remove migration")
	}
	return nil
}

// Migrations returns applied migrations sorted by version descending.
// If limit <= 0 all migrations are returned.
func (i *Iceberg) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	migrations, err := i.loadMigrations(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get migrations")
	}

	// Sort ascending by version first, then reverse for DESC order.
	migrations.SortByVersion()
	reverseSlice(migrations)

	if limit > 0 && len(migrations) > limit {
		migrations = migrations[:limit]
	}

	return migrations, nil
}

// MigrationsCount returns the total number of applied migration records.
func (i *Iceberg) MigrationsCount(ctx context.Context) (int, error) {
	migrations, err := i.loadMigrations(ctx)
	if err != nil {
		return 0, errors.Wrap(err, "get migrations count")
	}
	return len(migrations), nil
}

// ExistsMigration checks whether a migration record for the given version is present.
func (i *Iceberg) ExistsMigration(ctx context.Context, version string) (bool, error) {
	props, err := i.cat.LoadNamespaceProperties(ctx, i.historyNS())
	if err != nil {
		return false, errors.Wrap(i.dbError(err), "check migration exists")
	}
	_, ok := props[icebergHistoryKeyPrefix+version]
	return ok, nil
}

// MigrationsByMaxApplyTime returns all migration records that share the maximum apply_time value.
// This identifies the latest release batch for rollback.
func (i *Iceberg) MigrationsByMaxApplyTime(ctx context.Context) (entity.Migrations, error) {
	migrations, err := i.loadMigrations(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get migrations by max apply time")
	}

	if len(migrations) == 0 {
		return nil, nil
	}

	var maxApplyTime int64 = math.MinInt64
	for _, m := range migrations {
		if m.ApplyTime > maxApplyTime {
			maxApplyTime = m.ApplyTime
		}
	}

	result := make(entity.Migrations, 0)
	for _, m := range migrations {
		if m.ApplyTime == maxApplyTime {
			result = append(result, m)
		}
	}

	// Sort descending by version for consistent rollback ordering (newest first).
	result.SortByVersion()
	reverseSlice(result)

	return result, nil
}

// ExecQuery parses a Spark-SQL DDL statement and dispatches it to the catalog.
// It implements the full translator: parse → operation kind → catalog method.
// Parse errors (ErrUnsupportedDDL, ErrParse, …) are returned as-is (fail-fast).
// down-migrations use the same path: the .down.sql file contains valid DDL that
// the catalog executes; an irreversible change (e.g. type narrowing) is rejected
// by the catalog itself (allowIncompatibleChanges=false in UpdateSchema) and the
// error is propagated without removing the migration record (Р8).
func (i *Iceberg) ExecQuery(ctx context.Context, query string, _ ...any) error {
	op, err := ddl.Parse(i.cat.Warehouse(), query)
	if err != nil {
		return errors.WithStack(err)
	}

	switch op.Kind {
	case ddl.CreateNamespace:
		return i.cat.CreateNamespace(ctx, op.Table.Namespace, op.Props)
	case ddl.DropNamespace:
		return i.cat.DropNamespace(ctx, op.Table.Namespace)
	case ddl.CreateTable:
		if op.Create == nil {
			return errors.New("iceberg: CreateTable IR has nil Create spec")
		}
		return i.cat.CreateTable(ctx, op.Table, *op.Create)
	case ddl.DropTable:
		return i.cat.DropTable(ctx, op.Table)
	case ddl.RenameTable:
		if op.RenameTo == nil {
			return errors.New("iceberg: RenameTable IR has nil RenameTo")
		}
		return i.cat.RenameTable(ctx, op.Table, *op.RenameTo)
	case ddl.AddColumn, ddl.DropColumn, ddl.RenameColumn, ddl.AlterColumnType:
		return i.cat.ApplySchemaChange(ctx, op)
	case ddl.AddPartitionField, ddl.DropPartitionField:
		return i.cat.ApplySpecChange(ctx, op)
	default:
		return errors.WithStack(ddl.ErrUnsupportedDDL)
	}
}

// QueryScalar is not used for the Iceberg driver: all history aggregation is
// performed in Go after loading namespace properties via LoadNamespaceProperties.
// The service layer never calls QueryScalar for non-SQL drivers.
func (i *Iceberg) QueryScalar(_ context.Context, _ string, _ any) error {
	return errors.WithStack(ErrNotImplemented)
}

// loadMigrations loads all migration properties from the history namespace and
// parses them into entity.Migrations. Keys without the "mig." prefix are ignored.
// A property value that cannot be parsed as an integer results in an error.
func (i *Iceberg) loadMigrations(ctx context.Context) (entity.Migrations, error) {
	props, err := i.cat.LoadNamespaceProperties(ctx, i.historyNS())
	if err != nil {
		return nil, i.dbError(err)
	}

	migrations := make(entity.Migrations, 0, len(props))
	for k, v := range props {
		if !strings.HasPrefix(k, icebergHistoryKeyPrefix) {
			continue
		}
		version := strings.TrimPrefix(k, icebergHistoryKeyPrefix)
		applyTime, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, errors.WithMessagef(err, "invalid apply_time value for migration %q: %q", version, v)
		}
		migrations = append(migrations, entity.Migration{
			Version:   version,
			ApplyTime: applyTime,
		})
	}

	return migrations, nil
}

// dbError wraps a catalog error for consistent error reporting.
// Since the Iceberg catalog client already wraps errors with context messages,
// we use errors.WithMessage to add the "iceberg catalog" prefix for traceability.
func (i *Iceberg) dbError(err error) error {
	if err == nil {
		return nil
	}
	return errors.WithMessage(err, "iceberg catalog")
}

// reverseSlice reverses an entity.Migrations slice in-place.
func reverseSlice(s entity.Migrations) {
	for l, r := 0, len(s)-1; l < r; l, r = l+1, r-1 {
		s[l], s[r] = s[r], s[l]
	}
}
