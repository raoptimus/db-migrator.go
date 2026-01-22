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
	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/raoptimus/db-migrator.go/internal/domain/service/mapper"
	"github.com/raoptimus/db-migrator.go/internal/domain/validator"
	"github.com/raoptimus/db-migrator.go/internal/helper/sqlio"
)

const (
	baseMigration            = "000000_000000_base"
	defaultLimit             = 10000
	maxLimit                 = 100000
	regexpFileNameGroupCount = 5
	credentialMask           = "****" // Mask for hiding credentials in output
)

// ErrMigrationVersionReserved occurs when attempting to apply or revert the reserved base migration version.
var ErrMigrationVersionReserved = errors.New("migration version reserved")

var (
	regexpFileName = regexp.MustCompile(`^(\d{6}_?\d{6}[A-Za-z0-9_]+)\.((safe)\.)?(up|down)\.sql$`)
)

// Migration is the service handling migration operations.
// It orchestrates migration file processing, SQL execution, and history tracking.
type Migration struct {
	options *Options
	logger  Logger
	file    File
	repo    Repository
}

// NewMigration creates a new Migration service instance.
// It accepts configuration options, logger, file system handler, and repository for database operations.
func NewMigration(
	options *Options,
	logger Logger,
	file File,
	repo Repository,
) *Migration {
	return &Migration{
		options: options,
		logger:  logger,
		file:    file,
		repo:    repo,
	}
}

// InitializeTableHistory creates the migration history table if it does not exist.
// It inserts a base migration record after creating the table.
func (m *Migration) InitializeTableHistory(ctx context.Context) error {
	exists, err := m.repo.HasMigrationHistoryTable(ctx)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	m.logger.Warnf("Creating migration history table %s...\n", m.repo.TableNameWithSchema())

	if err := m.repo.CreateMigrationHistoryTable(ctx); err != nil {
		return err
	}

	if err := m.repo.InsertMigration(ctx, baseMigration); err != nil {
		if err2 := m.repo.DropMigrationHistoryTable(ctx); err2 != nil {
			return errors.Wrap(err, err2.Error())
		}
		return err
	}

	m.logger.Success("Done")
	return nil
}

// Migrations retrieves the list of applied migrations from the database.
// It excludes the base migration from the returned list and uses a default limit if none is provided.
func (m *Migration) Migrations(ctx context.Context, limit int) (model.Migrations, error) {
	if err := m.InitializeTableHistory(ctx); err != nil {
		return nil, err
	}
	if limit < 1 {
		limit = defaultLimit
	}
	entities, err := m.repo.Migrations(ctx, limit)
	if err != nil {
		return nil, err
	}

	migrations := mapper.EntitiesToDomain(entities)

	for i := range migrations {
		if migrations[i].Version == baseMigration {
			return append(migrations[:i], migrations[i+1:]...), nil
		}
	}

	return migrations, nil
}

// NewMigrations retrieves the list of pending migrations that have not been applied yet.
// It compares migration files in the directory against the database history to identify new migrations.
func (m *Migration) NewMigrations(ctx context.Context) (model.Migrations, error) {
	if err := m.InitializeTableHistory(ctx); err != nil {
		return nil, err
	}
	entities, err := m.repo.Migrations(ctx, maxLimit)
	if err != nil {
		return nil, err
	}

	migrations := mapper.EntitiesToDomain(entities)

	files, err := filepath.Glob(filepath.Join(m.options.Directory, "*.up.sql"))
	if err != nil {
		return nil, err
	}

	newMigrations := make(model.Migrations, 0)
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
				model.Migration{
					Version: groups[1],
				},
			)
		}
	}

	newMigrations.SortByVersion()

	return newMigrations, err
}

// ApplySQL applies a migration by executing the provided SQL statements.
// It tracks execution time, logs progress, and records the migration in the history table.
// The safely parameter determines whether to execute statements within a transaction.
func (m *Migration) ApplySQL(
	ctx context.Context,
	safely bool,
	version,
	upSQL string,
) error {
	if version == baseMigration {
		return ErrMigrationVersionReserved
	}
	m.logger.Warnf("*** applying %s\n", version)
	scanner := sqlio.NewScanner(strings.NewReader(upSQL))

	start := time.Now()
	err := m.apply(ctx, scanner, safely)
	elapsedTime := time.Since(start)
	if err != nil {
		m.logger.Errorf("*** failed to apply %s (time: %.3fs)\n", version, elapsedTime.Seconds())
		return err
	}
	if err := m.repo.InsertMigration(ctx, version); err != nil {
		return err
	}
	// todo: save downSQL
	m.logger.Successf("*** applied %s (time: %.3fs)\n", version, elapsedTime.Seconds())

	return nil
}

// RevertSQL reverts a migration by executing the provided SQL statements.
// It tracks execution time, logs progress, and removes the migration from the history table.
// The safely parameter determines whether to execute statements within a transaction.
func (m *Migration) RevertSQL(
	ctx context.Context,
	safely bool,
	version,
	downSQL string,
) error {
	if version == baseMigration {
		return ErrMigrationVersionReserved
	}
	m.logger.Warnf("*** reverting %s\n", version)
	scanner := sqlio.NewScanner(strings.NewReader(downSQL))
	start := time.Now()
	err := m.apply(ctx, scanner, safely)
	elapsedTime := time.Since(start)
	if err != nil {
		m.logger.Errorf("*** failed to revert %s (time: %.3fs)\n", version, elapsedTime.Seconds())

		return err
	}
	if err := m.repo.RemoveMigration(ctx, version); err != nil {
		return err
	}
	m.logger.Warnf("*** reverted %s (time: %.3fs)\n", version, elapsedTime.Seconds())

	return nil
}

// ApplyFile applies a migration by reading and executing SQL from a file.
// It tracks execution time, logs progress, and records the migration in the history table.
// The safely parameter determines whether to execute statements within a transaction.
func (m *Migration) ApplyFile(ctx context.Context, migration *model.Migration, fileName string, safely bool) error {
	if migration.Version == baseMigration {
		return ErrMigrationVersionReserved
	}
	m.logger.Warnf("*** applying %s\n", migration.Version)
	scanner, err := m.scannerByFile(fileName)
	if err != nil {
		return err
	}
	defer func() {
		// Ensure file is closed
		if closeErr := scanner.Close(); closeErr != nil {
			m.logger.Warnf("failed to close SQL scanner: %v", closeErr)
		}
	}()

	start := time.Now()
	err = m.apply(ctx, scanner, safely)
	elapsedTime := time.Since(start)
	if err != nil {
		m.logger.Errorf("*** failed to apply %s (time: %.3fs)\n", migration.Version, elapsedTime.Seconds())

		return err
	}
	if err := m.repo.InsertMigration(ctx, migration.Version); err != nil {
		return err
	}
	m.logger.Successf("*** applied %s (time: %.3fs)\n", migration.Version, elapsedTime.Seconds())

	return nil
}

// RevertFile reverts a migration by reading and executing SQL from a file.
// It tracks execution time, logs progress, and removes the migration from the history table.
// The safely parameter determines whether to execute statements within a transaction.
func (m *Migration) RevertFile(ctx context.Context, migration *model.Migration, fileName string, safely bool) error {
	if migration.Version == baseMigration {
		return ErrMigrationVersionReserved
	}
	m.logger.Warnf("*** reverting %s\n", migration.Version)
	scanner, err := m.scannerByFile(fileName)
	if err != nil {
		return err
	}
	defer func() {
		// Ensure file is closed
		if closeErr := scanner.Close(); closeErr != nil {
			m.logger.Warnf("failed to close SQL scanner: %v", closeErr)
		}
	}()

	start := time.Now()
	err = m.apply(ctx, scanner, safely)
	elapsedTime := time.Since(start)
	if err != nil {
		m.logger.Errorf("*** failed to revert %s (time: %.3fs)\n",
			migration.Version, elapsedTime.Seconds())
		return err
	}
	if err := m.repo.RemoveMigration(ctx, migration.Version); err != nil {
		return err
	}
	m.logger.Warnf("*** reverted %s (time: %.3fs)\n", migration.Version, elapsedTime.Seconds())

	return nil
}

// BeginCommand logs the start of a SQL command execution and returns the start time.
// It sanitizes credentials in the SQL output before logging.
func (m *Migration) BeginCommand(sqlQuery string) time.Time {
	sqlQueryOutput := m.SQLQueryOutput(sqlQuery)
	if !m.options.Compact {
		m.logger.Infof("    > execute SQL: %s ...\n", sqlQueryOutput)
	}

	return time.Now()
}

// ExecQuery executes a SQL query and logs its execution time.
// It wraps the repository's ExecQuery method with timing and logging.
func (m *Migration) ExecQuery(ctx context.Context, sqlQuery string) error {
	start := m.BeginCommand(sqlQuery)
	if err := m.repo.ExecQuery(ctx, sqlQuery); err != nil {
		return err
	}
	m.EndCommand(start)

	return nil
}

// SQLQueryOutput prepares SQL query text for output by sanitizing credentials and truncating if needed.
// It applies credential masking and respects the maximum SQL output length setting.
func (m *Migration) SQLQueryOutput(sqlQuery string) string {
	// First sanitize credentials
	sqlQueryOutput := m.sanitizeCredentials(sqlQuery)

	// Then apply length limit
	if m.options.MaxSQLOutputLength > 0 && m.options.MaxSQLOutputLength < len(sqlQueryOutput) {
		sqlQueryOutput = sqlQueryOutput[:m.options.MaxSQLOutputLength] + "..."
	}

	return sqlQueryOutput
}

// EndCommand logs the completion time of a SQL command execution.
// It is called after a command completes to log the elapsed time.
func (m *Migration) EndCommand(start time.Time) {
	if m.options.Compact {
		m.logger.Infof(" done (time: '%.3fs)\n", time.Since(start).Seconds())
	}
}

// Exists checks whether a migration with the specified version has been applied.
// It queries the repository to determine if the migration exists in the history table.
func (m *Migration) Exists(ctx context.Context, version string) (bool, error) {
	return m.repo.ExistsMigration(ctx, version)
}

func (m *Migration) apply(ctx context.Context, scanner *sqlio.Scanner, safely bool) error {
	processScanFunc := func(ctx context.Context) error {
		var sql string
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			sql = scanner.SQL()
			if sql == "" {
				continue
			}

			sql = strings.ReplaceAll(sql, "{cluster}", m.options.ClusterName)
			sql = strings.ReplaceAll(sql, "{placeholder_custom}", m.options.PlaceholderCustom)
			sql = strings.ReplaceAll(sql, "{username}", m.options.Username)
			sql = strings.ReplaceAll(sql, "{password}", m.options.Password)

			if err := m.ExecQuery(ctx, sql); err != nil {
				return err
			}
		}

		return scanner.Err()
	}

	var err error
	if safely {
		err = m.repo.ExecQueryTransaction(ctx, processScanFunc)
	} else {
		err = processScanFunc(ctx)
	}

	return err
}

func (m *Migration) scannerByFile(fileName string) (*sqlio.Scanner, error) {
	exists, err := m.file.Exists(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "migration file %s does not exist", fileName)
	}
	if !exists {
		return nil, fmt.Errorf("migration file %s does not exist", fileName)
	}

	f, err := m.file.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "migration file %s does not read", fileName)
	}

	return sqlio.NewScanner(f), nil
}

// sanitizeCredentials replaces credential values with masks in SQL output.
// This prevents passwords and usernames from appearing in logs or console output.
func (m *Migration) sanitizeCredentials(sql string) string {
	if m.options.Username == "" && m.options.Password == "" {
		return sql
	}

	sanitized := sql

	// Replace password if present
	if m.options.Password != "" {
		sanitized = strings.ReplaceAll(sanitized, m.options.Password, credentialMask)
	}

	// Replace username if present
	if m.options.Username != "" {
		sanitized = strings.ReplaceAll(sanitized, m.options.Username, credentialMask)
	}

	return sanitized
}
