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
	"github.com/raoptimus/db-migrator.go/console"
	"github.com/raoptimus/db-migrator.go/iofile"
	"github.com/raoptimus/db-migrator.go/migrator/multistmt"
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
		GetMigrationHistory(limit int) (MigrationEntityList, error)
		ConvertError(err error, query string) error
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
		connection  *sql.DB
		transaction *sql.Tx
		options     MigrationOptions
	}
)

func NewMigration(m MigrationInterface, conn *sql.DB, options MigrationOptions) *Migration {
	return &Migration{
		MigrationInterface: m,
		connection:         conn,
		options:            options,
	}
}

func (s *Migration) GetNewMigrations(limit int) (MigrationEntityList, error) {
	entityList, err := s.GetMigrationHistory(limit)
	if err != nil {
		return nil, err
	}
	var files []string
	files, err = filepath.Glob(filepath.Join(s.options.Directory, "*.up.sql"))
	if err != nil {
		return nil, err
	}

	result := MigrationEntityList{}
	var baseFilename string
	for _, file := range files {
		baseFilename = filepath.Base(file)
		groups := Regex.FindStringSubmatch(baseFilename)
		if len(groups) != 5 {
			return nil, fmt.Errorf("file name %s is invalid", baseFilename)
		}
		found := false
		for _, entity := range entityList {
			if entity.Version == groups[1] {
				found = true
				break
			}
		}
		if !found {
			result = append(
				result,
				MigrationEntity{
					Version: groups[1],
				},
			)
		}
	}

	result.SortByVersion()

	return result, err
}

func (s *Migration) MigrateUp(entity MigrationEntity, fileName string, safely bool) error {
	if entity.Version == BaseMigration {
		return nil
	}
	log.Printf(console.Yellow("*** applying %s"), entity.Version)

	elapsedTime, err := s.executeFile(fileName, s.options.ForceSafely || safely)
	if err == nil {
		err = s.AddMigrationHistory(entity.Version)
	}
	if err == nil {
		log.Printf(console.Green("*** applied %s (time: %.3fs)"), entity.Version, elapsedTime)
		return nil
	}

	return fmt.Errorf("*** failed to apply %s (time: %.3fs)\nException: %v",
		entity.Version, elapsedTime, err)
}

func (s *Migration) MigrateDown(entity MigrationEntity, fileName string, safely bool) error {
	if entity.Version == BaseMigration {
		return nil
	}
	log.Printf(console.Yellow("*** reverting %s"), entity.Version)

	elapsedTime, err := s.executeFile(fileName, s.options.ForceSafely || safely)
	if err == nil {
		err = s.RemoveMigrationHistory(entity.Version)
	}
	if err == nil {
		log.Printf(console.Green("*** reverted %s (time: %.3fs)"), entity.Version, elapsedTime)
		return nil
	}

	return fmt.Errorf("*** failed to reverted %s (time: %.3fs)\nException: %v",
		entity.Version, elapsedTime, err)
}

func (s *Migration) BeginCommand(sqlQuery string) time.Time {
	sqlQueryOutput := s.GetSQLQueryOutput(sqlQuery)
	if !s.options.Compact {
		log.Printf("    > execute SQL: %s ...", sqlQueryOutput)
	}

	return time.Now()
}

func (s *Migration) ExecuteSafely(tx *sql.Tx, sqlQuery string, args ...interface{}) error {
	start := s.BeginCommand(sqlQuery)
	stmt, err := tx.Prepare(sqlQuery)
	if err != nil {
		return s.ConvertError(err, sqlQuery)
	}
	if _, err := stmt.Exec(args...); err != nil {
		return s.ConvertError(err, sqlQuery)
	}
	s.EndCommand(start)

	return nil
}

func (s *Migration) Execute(sqlQuery string, args ...interface{}) error {
	start := s.BeginCommand(sqlQuery)
	if _, err := s.connection.Exec(sqlQuery, args...); err != nil {
		return s.ConvertError(err, sqlQuery)
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

func (s *Migration) executeFile(fileName string, safely bool) (elapsedSeconds float64, err error) {
	start := time.Now()
	if !iofile.Exists(fileName) {
		return 0, fmt.Errorf("migration file %s does not exists", fileName)
	}

	if safely {
		var tx *sql.Tx
		tx, err = s.connection.Begin()
		if err != nil {
			return 0, err
		}

		err = multistmt.ReadOrParseSQLFile(fileName, s.options.MultiSTMT, func(sqlQuery string) error {
			return s.ExecuteSafely(tx, sqlQuery)
		})

		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}

		if err = tx.Commit(); err != nil {
			return 0, err
		}

		return time.Now().Sub(start).Seconds(), nil
	}

	err = multistmt.ReadOrParseSQLFile(fileName, s.options.MultiSTMT, func(sqlQuery string) error {
		return s.Execute(sqlQuery)
	})

	if err != nil {
		return 0, err
	}

	return time.Now().Sub(start).Seconds(), nil
}
