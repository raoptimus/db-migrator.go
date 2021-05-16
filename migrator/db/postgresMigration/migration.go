/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package postgresMigration

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/raoptimus/db-migrator.go/console"
	"github.com/raoptimus/db-migrator.go/migrator/db"
	"log"
	"time"
)

const DefaultSchema = "public"

type (
	Migration struct {
		connection  *sql.DB
		tableName   string
		tableSchema string
		directory   string
	}
)

func New(connection *sql.DB, tableName, tableSchema, directory string) *Migration {
	return &Migration{
		connection:  connection,
		tableName:   tableName,
		tableSchema: tableSchema,
		directory:   directory,
	}
}

func (s *Migration) internalConvertError(err error, query string) error {
	if ex, ok := err.(*pq.Error); ok {
		q := ex.InternalQuery
		if q == "" {
			q = query
		}
		return fmt.Errorf("SQLSTATE[%s]: %s: %s\nDETAILS:%s\nThe SQL being executed was: %s\n",
			ex.Code,
			ex.Severity,
			ex.Message,
			ex.Detail,
			q,
		)
	}

	return err
}

func (s *Migration) ConvertError(err error, query string) error {
	return fmt.Errorf("exception: %v", s.internalConvertError(err, query))
}

func (s *Migration) InitializeTableHistory() error {
	exists, err := s.getTableScheme()
	if err != nil {
		return err
	}

	if !exists {
		if err := s.createMigrationHistoryTable(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Migration) GetMigrationHistory(limit int) (db.MigrationEntityList, error) {
	if err := s.InitializeTableHistory(); err != nil {
		return nil, err
	}

	var (
		q = fmt.Sprintf(
			`
			SELECT version, apply_time 
			FROM %s
			ORDER BY apply_time DESC, version DESC
			LIMIT $1`,
			s.getTableNameWithSchema(),
		)
		result db.MigrationEntityList
	)
	if limit < 1 {
		limit = db.DefaultLimit
	}

	rows, err := s.connection.Query(q, limit)
	if err != nil {
		return nil, s.internalConvertError(err, q)
	}
	for rows.Next() {
		var (
			version   string
			applyTime int
		)

		if err := rows.Scan(&version, &applyTime); err != nil {
			return nil, err
		}

		if version == db.BaseMigration {
			continue
		}
		result = append(result,
			db.MigrationEntity{
				Version:   version,
				ApplyTime: applyTime,
			},
		)
	}

	return result, nil
}

func (s *Migration) AddMigrationHistory(version string) error {
	now := uint32(time.Now().Unix())
	q := fmt.Sprintf(`
		INSERT INTO %s (version, apply_time) 
		VALUES ($1, $2)`,
		s.getTableNameWithSchema(),
	)
	_, err := s.connection.Exec(q, version, now)

	return s.internalConvertError(err, q)
}

func (s *Migration) RemoveMigrationHistory(version string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE (version) = ($1)`, s.getTableNameWithSchema())
	_, err := s.connection.Exec(q, version)

	return err
}

func (s *Migration) createMigrationHistoryTable() error {
	log.Printf(console.Yellow("Creating migration history table %s..."), s.getTableNameWithSchema())

	q := fmt.Sprintf(
		`
				CREATE TABLE %s (
				  version varchar(180) PRIMARY KEY,
				  apply_time integer
				)
			`,
		s.getTableNameWithSchema(),
	)

	if _, err := s.connection.Exec(q); err != nil {
		return s.internalConvertError(err, q)
	}
	if err := s.AddMigrationHistory(db.BaseMigration); err != nil {
		q2 := fmt.Sprintf(`DROP TABLE %s`, s.getTableNameWithSchema())
		_, _ = s.connection.Exec(q2)

		return err
	}

	log.Println(console.Green("Done"))

	return nil
}

func (s *Migration) getTableScheme() (exists bool, err error) {
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

	rows, err = s.connection.Query(q, s.tableName, s.tableSchema)
	if err != nil {
		return false, s.internalConvertError(err, q)
	}

	for rows.Next() {
		var (
			tableName string
			schema    string
		)
		if err := rows.Scan(&schema, &tableName); err != nil {
			return false, s.internalConvertError(err, q)
		}

		//todo scan columns to tableScheme
		if tableName == s.tableName {
			return true, nil
		}
	}

	return false, nil
}

func (s *Migration) getTableNameWithSchema() string {
	return s.tableSchema + "." + s.tableName
}
