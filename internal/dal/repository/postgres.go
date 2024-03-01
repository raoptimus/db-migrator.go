package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
)

const postgresDefaultSchema = "public"

type Postgres struct {
	conn    Connection
	options *Options
}

func NewPostgres(conn Connection, options *Options) *Postgres {
	return &Postgres{
		conn:    conn,
		options: options,
	}
}

// Migrations returns applied migrations history.
func (p *Postgres) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	var (
		q = fmt.Sprintf(
			`
			SELECT version, apply_time 
			FROM %s
			ORDER BY apply_time DESC, version DESC
			LIMIT $1`,
			p.TableNameWithSchema(),
		)
		migrations entity.Migrations
	)

	rows, err := p.conn.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, errors.Wrap(p.dbError(err, q), "get migrations")
	}
	for rows.Next() {
		var (
			version   string
			applyTime int
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

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(p.dbError(err, q), "get migrations")
	}

	return migrations, nil
}

// HasMigrationHistoryTable returns true if migration history table exists.
func (p *Postgres) HasMigrationHistoryTable(ctx context.Context) (exists bool, err error) {
	var (
		q = `
			SELECT
				d.nspname AS table_schema,
				c.relname AS table_name
			FROM pg_class c
			LEFT JOIN pg_namespace d ON d.oid = c.relnamespace
			WHERE (c.relname, d.nspname) = ($1, $2)
		`
		rows *sql.Rows
	)

	rows, err = p.conn.QueryContext(ctx, q, p.options.TableName, p.options.SchemaName)
	if err != nil {
		return false, errors.Wrap(p.dbError(err, q), "get schema table")
	}

	for rows.Next() {
		var (
			tableName string
			schema    string
		)
		if err := rows.Scan(&schema, &tableName); err != nil {
			return false, errors.Wrap(p.dbError(err, q), "get schema table")
		}

		//todo: scan columns to tableScheme
		if tableName == p.options.TableName {
			return true, nil
		}
	}

	if err := rows.Err(); err != nil {
		return false, errors.Wrap(p.dbError(err, q), "get schema table")
	}

	return false, nil
}

// InsertMigration inserts the new migration record.
func (p *Postgres) InsertMigration(ctx context.Context, version string) error {
	q := fmt.Sprintf(`
		INSERT INTO %s (version, apply_time)
		VALUES ($1, $2)`,
		p.TableNameWithSchema(),
	)
	now := uint32(time.Now().Unix())
	if _, err := p.conn.ExecContext(ctx, q, version, now); err != nil {
		return errors.Wrap(p.dbError(err, q), "insert migration")
	}
	return nil
}

// RemoveMigration removes the migration record.
func (p *Postgres) RemoveMigration(ctx context.Context, version string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE (version) = ($1)`, p.TableNameWithSchema())
	if _, err := p.conn.ExecContext(ctx, q, version); err != nil {
		return errors.Wrap(p.dbError(err, q), "remove migration")
	}
	return nil
}

// ExecQuery executes a query without returning any rows.
// The args are for any placeholder parameters in the query.
func (p *Postgres) ExecQuery(ctx context.Context, query string, args ...any) error {
	if _, err := p.conn.ExecContext(ctx, query, args...); err != nil {
		return p.dbError(err, query)
	}
	return nil
}

// ExecQueryTransaction executes a query in transaction without returning any rows.
// The args are for any placeholder parameters in the query.
func (p *Postgres) ExecQueryTransaction(ctx context.Context, txFn func(ctx context.Context) error) error {
	return p.conn.Transaction(ctx, txFn)
}

// CreateMigrationHistoryTable creates a new migration history table.
func (p *Postgres) CreateMigrationHistoryTable(ctx context.Context) error {
	q := fmt.Sprintf(
		`
				CREATE TABLE %s (
				  version varchar(180) PRIMARY KEY,
				  apply_time integer
				)
			`,
		p.TableNameWithSchema(),
	)

	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "create migration history table")
	}
	return nil
}

// DropMigrationHistoryTable drops the migration history table.
func (p *Postgres) DropMigrationHistoryTable(ctx context.Context) error {
	q := fmt.Sprintf(`DROP TABLE %s`, p.TableNameWithSchema())
	if _, err := p.conn.ExecContext(ctx, q); err != nil {
		return errors.Wrap(p.dbError(err, q), "drop migration history table")
	}
	return nil
}

// MigrationsCount returns the number of migrations
func (p *Postgres) MigrationsCount(ctx context.Context) (int, error) {
	q := fmt.Sprintf(`SELECT count(*) FROM %s`, p.TableNameWithSchema())
	rows, err := p.conn.QueryContext(ctx, q)
	if err != nil {
		return 0, err
	}
	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, p.dbError(err, q)
		}
	}
	if err := rows.Err(); err != nil {
		return 0, p.dbError(err, q)
	}
	return count, nil
}

func (p *Postgres) TableNameWithSchema() string {
	return p.options.SchemaName + "." + p.options.TableName
}

func (p *Postgres) ForceSafely() bool {
	return false
}

// dbError returns DBError is err is db error else returns got error.
func (p *Postgres) dbError(err error, q string) error {
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		return err
	}

	if q == "" {
		q = pgErr.InternalQuery
	}

	return &DBError{
		Code:          pgErr.SQLState(),
		Severity:      pgErr.Severity,
		Message:       pgErr.Message,
		Details:       pgErr.Detail,
		InternalQuery: q,
	}
}
