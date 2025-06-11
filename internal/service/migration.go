/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package service

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	_ "github.com/raoptimus/db-migrator.go/internal/console"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
	"github.com/raoptimus/db-migrator.go/internal/util/sqlio"
	"github.com/raoptimus/db-migrator.go/internal/validator"
)

const (
	baseMigration            = "000000_000000_base"
	baseMigrationsCount      = 1
	defaultLimit             = 10000
	maxLimit                 = 100000
	regexpFileNameGroupCount = 5
)

var (
	regexpFileName = regexp.MustCompile(`^(\d{6}_?\d{6}[A-Za-z0-9_]+)\.((safe)\.)?(up|down)\.sql$`)
)

type Migration struct {
	options *Options
	console Console
	file    File
	repo    Repository
}

func NewMigration(
	options *Options,
	console Console,
	file File,
	repo Repository,
) *Migration {
	return &Migration{
		options: options,
		console: console,
		file:    file,
		repo:    repo,
	}
}

func (m *Migration) InitializeTableHistory(ctx context.Context) error {
	exists, err := m.repo.HasMigrationHistoryTable(ctx)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	m.console.Warnf("Creating migration history table %s...", m.repo.TableNameWithSchema())

	if err := m.repo.CreateMigrationHistoryTable(ctx); err != nil {
		return err
	}

	if err := m.repo.InsertMigration(ctx, baseMigration); err != nil {
		if err2 := m.repo.DropMigrationHistoryTable(ctx); err2 != nil {
			return errors.Wrap(err, err2.Error())
		}
		return err
	}

	m.console.SuccessLn("Done")
	return nil
}

func (m *Migration) Migrations(ctx context.Context, limit int) (entity.Migrations, error) {
	if err := m.InitializeTableHistory(ctx); err != nil {
		return nil, err
	}
	if limit < 1 {
		limit = defaultLimit
	}
	migrations, err := m.repo.Migrations(ctx, limit)
	if err != nil {
		return nil, err
	}

	for i := range migrations {
		if migrations[i].Version == baseMigration {
			return append(migrations[:i], migrations[i+1:]...), nil
		}
	}

	return migrations, nil
}

func (m *Migration) NewMigrations(ctx context.Context) (entity.Migrations, error) {
	if err := m.InitializeTableHistory(ctx); err != nil {
		return nil, err
	}
	migrations, err := m.repo.Migrations(ctx, maxLimit)
	if err != nil {
		return nil, err
	}

	files, err := filepath.Glob(filepath.Join(m.options.Directory, "*.up.sql"))
	if err != nil {
		return nil, err
	}

	newMigrations := make(entity.Migrations, 0)
	var baseFilename string

	for _, file := range files {
		baseFilename = filepath.Base(file)
		if err := validator.ValidateFileName(baseFilename); err != nil {
			return nil, errors.Wrap(err, baseFilename)
		}
		groups := regexpFileName.FindStringSubmatch(baseFilename)
		if len(groups) != regexpFileNameGroupCount {
			return nil, fmt.Errorf("file name %s is invalid", baseFilename)
		}
		found := false
		for _, migration := range migrations {
			if migration.Version == baseMigration {
				continue
			}
			if migration.Version == groups[1] {
				found = true
				break
			}
		}
		if !found {
			newMigrations = append(
				newMigrations,
				entity.Migration{
					Version: groups[1],
				},
			)
		}
	}

	newMigrations.SortByVersion()

	return newMigrations, err
}

func (m *Migration) ApplySQL(
	ctx context.Context,
	safely bool,
	version,
	upSQL string,
) error {
	if version == baseMigration {
		return ErrMigrationVersionReserved
	}
	m.console.Warnf("*** applying %s\n", version)
	scanner := sqlio.NewScanner(strings.NewReader(upSQL))

	start := time.Now()
	err := m.apply(ctx, scanner, safely)
	elapsedTime := time.Since(start)
	if err != nil {
		m.console.Errorf("*** failed to apply %s (time: %.3fs)\n", version, elapsedTime.Seconds())
		return err
	}
	if err := m.repo.InsertMigration(ctx, version); err != nil {
		return err
	}
	// todo: save downSQL
	m.console.Successf("*** applied %s (time: %.3fs)\n", version, elapsedTime.Seconds())

	return nil
}

func (m *Migration) RevertSQL(
	ctx context.Context,
	safely bool,
	version,
	downSQL string,
) error {
	if version == baseMigration {
		return ErrMigrationVersionReserved
	}
	m.console.Warnf("*** reverting %s\n", version)
	scanner := sqlio.NewScanner(strings.NewReader(downSQL))
	start := time.Now()
	err := m.apply(ctx, scanner, safely)
	elapsedTime := time.Since(start)
	if err != nil {
		m.console.Errorf("*** failed to reverted %s (time: %.3fs)\n", version, elapsedTime.Seconds())

		return err
	}
	if err := m.repo.RemoveMigration(ctx, version); err != nil {
		return err
	}
	m.console.Warnf("*** reverted %s (time: %.3fs)\n", version, elapsedTime.Seconds())

	return nil
}

func (m *Migration) ApplyFile(ctx context.Context, entity *entity.Migration, fileName string, safely bool) error {
	if entity.Version == baseMigration {
		return ErrMigrationVersionReserved
	}
	m.console.Warnf("*** applying %s\n", entity.Version)
	scanner, err := m.scannerByFile(fileName)
	if err != nil {
		return err
	}

	start := time.Now()
	err = m.apply(ctx, scanner, safely)
	elapsedTime := time.Since(start)
	if err != nil {
		m.console.Errorf("*** failed to apply %s (time: %.3fs)\n", entity.Version, elapsedTime.Seconds())

		return err
	}
	if err := m.repo.InsertMigration(ctx, entity.Version); err != nil {
		return err
	}
	m.console.Successf("*** applied %s (time: %.3fs)\n", entity.Version, elapsedTime.Seconds())

	return nil
}

func (m *Migration) RevertFile(ctx context.Context, entity *entity.Migration, fileName string, safely bool) error {
	if entity.Version == baseMigration {
		return ErrMigrationVersionReserved
	}
	m.console.Warnf("*** reverting %s\n", entity.Version)
	scanner, err := m.scannerByFile(fileName)
	if err != nil {
		return err
	}

	start := time.Now()
	err = m.apply(ctx, scanner, safely)
	elapsedTime := time.Since(start)
	if err != nil {
		m.console.Errorf("*** failed to reverted %s (time: %.3fs)\n",
			entity.Version, elapsedTime.Seconds())
		return err
	}
	if err := m.repo.RemoveMigration(ctx, entity.Version); err != nil {
		return err
	}
	m.console.Warnf("*** reverted %s (time: %.3fs)\n", entity.Version, elapsedTime.Seconds())

	return nil
}

func (m *Migration) BeginCommand(sqlQuery string) time.Time {
	sqlQueryOutput := m.SQLQueryOutput(sqlQuery)
	if !m.options.Compact {
		m.console.Infof("    > execute SQL: %s ...\n", sqlQueryOutput)
	}

	return time.Now()
}

func (m *Migration) ExecQuery(ctx context.Context, sqlQuery string) error {
	start := m.BeginCommand(sqlQuery)
	if err := m.repo.ExecQuery(ctx, sqlQuery); err != nil {
		return err
	}
	m.EndCommand(start)

	return nil
}

func (m *Migration) SQLQueryOutput(sqlQuery string) string {
	sqlQueryOutput := sqlQuery
	if m.options.MaxSQLOutputLength > 0 && m.options.MaxSQLOutputLength < len(sqlQuery) {
		sqlQueryOutput = sqlQuery[:m.options.MaxSQLOutputLength]
	}

	return sqlQueryOutput
}

func (m *Migration) EndCommand(start time.Time) {
	if m.options.Compact {
		m.console.Infof(" done (time: '%.3fs)\n", time.Since(start).Seconds())
	}
}

func (m *Migration) Exists(ctx context.Context, version string) (bool, error) {
	return m.repo.ExistsMigration(ctx, version)
}

func (m *Migration) apply(ctx context.Context, scanner *sqlio.Scanner, safely bool) error {
	processScanFunc := func(ctx context.Context) error {
		var sql string
		for scanner.Scan() {
			sql = scanner.SQL()
			if sql == "" {
				continue
			}

			sql = strings.ReplaceAll(sql, "{username}", m.options.Username)
			sql = strings.ReplaceAll(sql, "{password}", m.options.Password)

			if err := m.ExecQuery(ctx, sql); err != nil {
				return err
			}
		}
		return scanner.Err()
	}

	var err error
	if m.repo.ForceSafely() || safely {
		err = m.repo.ExecQueryTransaction(ctx, processScanFunc)
	} else {
		err = processScanFunc(ctx)
	}

	return err
}

func (m *Migration) scannerByFile(fileName string) (*sqlio.Scanner, error) {
	exists, err := m.file.Exists(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "migration file %s does not exists", fileName)
	}
	if !exists {
		return nil, fmt.Errorf("migration file %s does not exists", fileName)
	}

	f, err := m.file.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "migration file %s does not read", fileName)
	}

	return sqlio.NewScanner(f), nil
}
