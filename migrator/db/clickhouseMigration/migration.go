/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package clickhouseMigration

import (
	"database/sql"
	"fmt"
	"github.com/ClickHouse/clickhouse-go"
	"github.com/raoptimus/db-migrator.go/console"
	"github.com/raoptimus/db-migrator.go/migrator/db"
	"log"
	"time"
)

type (
	Migration struct {
		connection  *sql.DB
		tableName   string
		directory   string
		clusterName string
	}
)

func New(connection *sql.DB, tableName, clusterName, directory string) *Migration {
	return &Migration{
		connection:  connection,
		tableName:   tableName,
		clusterName: clusterName,
		directory:   directory,
	}
}

func (s *Migration) ConvertError(err error, query string) error {
	if exception, ok := err.(*clickhouse.Exception); ok {
		return fmt.Errorf("exception: [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
	}

	return fmt.Errorf("exception: %v", err)
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
			WHERE is_deleted = 0 
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
		return nil, err
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
		INSERT INTO %s (version, apply_time, is_deleted) 
		VALUES(?, ?, ?)`,
		s.tableName,
	)
	tx, err := s.connection.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(version, now, 0); err != nil {
		tx.Rollback()

		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return s.optimizeTable()
}

func (s *Migration) RemoveMigrationHistory(version string) error {
	now := uint32(time.Now().Unix())
	q := fmt.Sprintf(`
		INSERT INTO %s (version, apply_time, is_deleted) 
		VALUES(?, ?, ?)`,
		s.tableName,
	)
	tx, err := s.connection.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(q)
	if err != nil {
		return err
	}

	if _, err := stmt.Exec(version, now, 1); err != nil {
		tx.Rollback()

		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return s.optimizeTable()
}

func (s *Migration) optimizeTable() error {
	var sqlQuery string
	if s.clusterName == "" {
		sqlQuery = fmt.Sprintf("OPTIMIZE TABLE %s FINAL", s.tableName)
	} else {
		sqlQuery = fmt.Sprintf("OPTIMIZE TABLE %s ON CLUSTER %s FINAL", s.tableName, s.clusterName)
	}
	_, err := s.connection.Exec(sqlQuery)

	return err
}

func (s *Migration) createMigrationHistoryTable() error {
	log.Printf(console.Yellow("Creating migration history table %s..."), s.tableName)
	var sqlQuery string
	if s.clusterName == "" {
		sqlQuery = fmt.Sprintf(
			`
			CREATE TABLE %s (
				version String, 
				date Date DEFAULT toDate(apply_time),
				apply_time UInt32,
				is_deleted UInt8
			) ENGINE = ReplacingMergeTree(apply_time)
			PRIMARY KEY (version)
			PARTITION BY (toYYYYMM(date))
			ORDER BY (version)
			SETTINGS index_granularity=8192
			`,
			s.tableName,
		)
	} else {
		sqlQuery = fmt.Sprintf(
			`
			CREATE TABLE %s ON CLUSTER %s (
				version String, 
				date Date DEFAULT toDate(apply_time),
				apply_time UInt32,
				is_deleted UInt8
			) ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/%s_%s', '{replica}', apply_time)
			PRIMARY KEY (version)
			PARTITION BY (toYYYYMM(date))
			ORDER BY (version)
			SETTINGS index_granularity=8192
			`,
			s.tableName,
			s.clusterName,
			s.clusterName,
			s.tableName,
		)
	}

	fmt.Println(sqlQuery)
	if _, err := s.connection.Exec(sqlQuery); err != nil {
		return err
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
		SELECT database, table 
		FROM system.columns 
		WHERE table = ? AND database = currentDatabase()
		`
		rows *sql.Rows
	)

	rows, err = s.connection.Query(q, s.tableName)
	if err != nil {
		return false, err
	}

	for rows.Next() {
		var (
			database string
			table    string
		)
		if err := rows.Scan(&database, &table); err != nil {
			return false, err
		}

		//todo scan columns to tableScheme
		if table == s.tableName {
			return true, nil
		}
	}

	return false, nil
}
