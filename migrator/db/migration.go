/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package db

import (
	"database/sql"
	"fmt"
	"github.com/raoptimus/db-migrator/console"
	"github.com/raoptimus/db-migrator/iofile"
	"github.com/raoptimus/db-migrator/migrator/multistmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"time"
)

const BaseMigration = "000000_000000_base"
const DefaultLimit = 10000

var Regex = regexp.MustCompile(`^(\d{6}_?\d{6}[A-Za-z0-9\_]+)\.((safe)\.)?(up|down)\.sql$`)

type (
	MigrationInterface interface {
		InitializeTableHistory() error
		AddMigrationHistory(version string) error
		RemoveMigrationHistory(version string) error
		GetMigrationHistory(limit int) (HistoryItems, error)
		CatchError(err error) error
	}
	MigrationOptions struct {
		MaxSqlOutputLength int
		Directory          string
		Compact            bool
		MultiSTMT          bool
		ForceSafely        bool
	}
	Migration struct {
		MigrationInterface
		connection *sql.DB
		options    MigrationOptions
	}
)

func NewMigration(m MigrationInterface, conn *sql.DB, options MigrationOptions) *Migration {
	return &Migration{
		MigrationInterface: m,
		connection:         conn,
		options:            options,
	}
}

func (s *Migration) GetNewMigrations(limit int) (HistoryItems, error) {
	hist, err := s.GetMigrationHistory(limit)
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(s.options.Directory, "*.up.sql"))
	if err != nil {
		return nil, err
	}

	result := HistoryItems{}
	var baseFilename string
	for _, file := range files {
		baseFilename = filepath.Base(file)
		groups := Regex.FindStringSubmatch(baseFilename)
		if len(groups) != 5 {
			return nil, fmt.Errorf("File name %s is invalid", baseFilename)
		}
		found := false
		for _, item := range hist {
			if item.Version == groups[1] {
				found = true
				break
			}
		}
		if !found {
			result = append(
				result,
				HistoryItem{
					Version: groups[1],
					Safely:  groups[3] == "safe",
				},
			)
		}
	}

	result.SortByVersion()

	return result, err
}

func (s *Migration) MigrateUp(item HistoryItem) error {
	if item.Version == BaseMigration {
		return nil
	}
	log.Printf(console.Yellow("*** applying %s"), item.Version)

	elapsedTime, err := s.executeFile(item.GetUpFileName(), s.options.ForceSafely || item.Safely)
	if err == nil {
		err = s.AddMigrationHistory(item.Version)
	}
	if err == nil {
		log.Printf(console.Green("*** applied %s (time: %.3fs)"), item.Version, elapsedTime)
		return nil
	}

	return fmt.Errorf("*** failed to apply %s (time: %.3fs)\nException: %v",
		item.Version, elapsedTime, err)
}

func (s *Migration) MigrateDown(item HistoryItem) error {
	if item.Version == BaseMigration {
		return nil
	}
	log.Printf(console.Yellow("*** reverting %s"), item.Version)

	elapsedTime, err := s.executeFile(item.GetDownFileName(), s.options.ForceSafely || item.Safely)
	if err == nil {
		err = s.RemoveMigrationHistory(item.Version)
	}
	if err == nil {
		log.Printf(console.Green("*** reverted %s (time: %.3fs)"), item.Version, elapsedTime)
		return nil
	}

	return fmt.Errorf("*** failed to reverted %s (time: %.3fs)\nException: %v",
		item.Version, elapsedTime, err)
}

func (s *Migration) BeginCommand(sqlQuery string) time.Time {
	sqlQueryOutput := s.GetSQLQueryOutput(sqlQuery)
	if !s.options.Compact {
		log.Printf("    > execute SQL: %s ...", sqlQueryOutput)
	}

	return time.Now()
}

func (s *Migration) ExecuteSafely(sqlQuery string, args ...interface{}) error {
	start := s.BeginCommand(sqlQuery)
	tx, err := s.connection.Begin()
	if err != nil {
		return s.CatchError(err)
	}
	stmt, err := tx.Prepare(sqlQuery)
	if err != nil {
		tx.Rollback()
		return s.CatchError(err)
	}
	if _, err := stmt.Exec(args...); err != nil {
		tx.Rollback()
		return s.CatchError(err)
	}
	if err = tx.Commit(); err != nil {
		return s.CatchError(err)
	}
	s.EndCommand(start)

	return nil
}

func (s *Migration) Execute(sqlQuery string, args ...interface{}) error {
	start := s.BeginCommand(sqlQuery)
	if _, err := s.connection.Exec(sqlQuery, args...); err != nil {
		return s.CatchError(err)
	}
	s.EndCommand(start)

	return nil
}

func (s *Migration) GetSQLQueryOutput(sqlQuery string) string {
	sqlQueryOutput := sqlQuery
	if s.options.MaxSqlOutputLength > 0 && s.options.MaxSqlOutputLength < len(sqlQuery) {
		sqlQueryOutput = sqlQuery[:s.options.MaxSqlOutputLength]
	}

	return sqlQueryOutput
}

func (s *Migration) EndCommand(start time.Time) {
	if s.options.Compact {
		log.Printf(" done (time: '%.3fs)", time.Now().Sub(start).Seconds())
	}
}

func (s *Migration) executeFile(filename string, safely bool) (elapsedSeconds float64, err error) {
	start := time.Now()
	filename = filepath.Join(s.options.Directory, filename)
	if !iofile.Exists(filename) {
		return 0, fmt.Errorf("migration file %s does not exists", filename)
	}

	if !s.options.MultiSTMT {
		err = multistmt.ParseSQLFile(filename, func(sqlQuery string) error {
			return s.executeSql(sqlQuery, safely)
		})
	} else {
		var sqlBytes []byte
		sqlBytes, err = ioutil.ReadFile(filename)
		if err == nil {
			err = s.executeSql(string(sqlBytes), safely)
		}
	}

	if err != nil {
		return 0, err
	}

	return time.Now().Sub(start).Seconds(), nil
}

func (s *Migration) executeSql(sqlQuery string, safely bool) error {
	if safely {
		return s.ExecuteSafely(sqlQuery)
	}

	return s.Execute(sqlQuery)
}
