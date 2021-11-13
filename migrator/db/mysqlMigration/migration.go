/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package mysqlMigration

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/raoptimus/db-migrator.go/console"
	"github.com/raoptimus/db-migrator.go/migrator/db"
	"log"
	"time"
)

type (
	Migration struct {
		connection  *sql.DB
		tableSchema string
		tableName   string
		directory   string
	}
)

func New(connection *sql.DB, tableName, tableSchema, directory string) *Migration {
	return &Migration{
		connection:  connection,
		tableSchema: tableSchema,
		tableName:   tableName,
		directory:   directory,
	}
}

func (s *Migration) internalConvertError(err error, query string) error {
	if ex, ok := err.(*mysql.MySQLError); ok {
		return fmt.Errorf("SQLSTATE[%d]: %s\nThe SQL being executed was: %s\n",
			ex.Number,
			ex.Message,
			query,
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
			LIMIT ?`,
			s.tableName,
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Migration) AddMigrationHistory(version string) error {
	now := uint32(time.Now().Unix())
	q := fmt.Sprintf(`
		INSERT INTO %s (version, apply_time) 
		VALUES (?, ?)`,
		s.tableName,
	)
	_, err := s.connection.Exec(q, version, now)

	return s.internalConvertError(err, q)
}

func (s *Migration) RemoveMigrationHistory(version string) error {
	q := fmt.Sprintf(`DELETE FROM %s WHERE version = ?`, s.tableName)
	_, err := s.connection.Exec(q, version)

	return err
}

func (s *Migration) createMigrationHistoryTable() error {
	log.Printf(console.Yellow("Creating migration history table %s..."), s.tableName)

	q := fmt.Sprintf(
		`
				CREATE TABLE %s (
				  version VARCHAR(180) PRIMARY KEY,
				  apply_time INT
				)
				ENGINE=InnoDB
			`,
		s.tableName,
	)

	if _, err := s.connection.Exec(q); err != nil {
		return s.internalConvertError(err, q)
	}
	if err := s.AddMigrationHistory(db.BaseMigration); err != nil {
		q2 := fmt.Sprintf(`DROP TABLE %s`, s.tableName)
		_, _ = s.connection.Exec(q2)

		return err
	}

	log.Println(console.Green("Done"))

	return nil
}

func (s *Migration) getTableScheme() (exists bool, err error) {
	var (
		q = `
			SELECT EXISTS(
			    SELECT *
				FROM information_schema.tables
				WHERE table_schema = ? AND table_name = ?
			)
		`
		rows *sql.Rows
	)

	rows, err = s.connection.Query(q, s.tableSchema, s.tableName)
	if err != nil {
		return false, s.internalConvertError(err, q)
	}

	for rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, s.internalConvertError(err, q)
		}
	}

	if err = rows.Err(); err != nil {
		return false, err
	}

	return exists, nil
}
