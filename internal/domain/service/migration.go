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
	return m.applyFileCore(ctx, migration, fileName, safely, func(ctx context.Context, version string) error {
		return m.repo.InsertMigration(ctx, version)
	})
}

// ApplyFileWithApplyTime applies a migration by reading and executing SQL from a file
// with an explicit apply time. This is used by the release command to ensure all migrations
// in a release batch share the same apply_time for later rollback identification.
// The safely parameter is always false because release runs inside an outer transaction.
func (m *Migration) ApplyFileWithApplyTime(
	ctx context.Context,
	migration *model.Migration,
	fileName string,
	applyTime int64,
) error {
	return m.applyFileCore(ctx, migration, fileName, false, func(ctx context.Context, version string) error {
		return m.repo.InsertMigrationWithApplyTime(ctx, version, applyTime)
	})
}

// applyFileCore contains the shared logic for applying a migration file.
// insertFn controls how the migration record is stored (with or without explicit applyTime).
func (m *Migration) applyFileCore(
	ctx context.Context,
	migration *model.Migration,
	fileName string,
	safely bool,
	insertFn func(ctx context.Context, version string) error,
) error {
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
	if err := insertFn(ctx, migration.Version); err != nil {
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

// LatestReleaseMigrations returns migrations from the latest release batch,
// identified by the maximum apply_time value. It filters out the base migration.
func (m *Migration) LatestReleaseMigrations(ctx context.Context) (model.Migrations, error) {
	if err := m.InitializeTableHistory(ctx); err != nil {
		return nil, err
	}

	entities, err := m.repo.MigrationsByMaxApplyTime(ctx)
	if err != nil {
		return nil, err
	}

	migrations := mapper.EntitiesToDomain(entities)

	// Filter out base migration
	result := make(model.Migrations, 0, len(migrations))
	for _, migration := range migrations {
		if migration.Version != baseMigration {
			result = append(result, migration)
		}
	}

	return result, nil
}

// ExecInTransaction executes fn within a transaction only if the driver supports DDL transactions.
// Drivers that do not support transactional DDL (ClickHouse, MySQL, Tarantool) execute fn directly.
func (m *Migration) ExecInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.repo.SupportsDDLTransactions() {
		return m.repo.ExecQueryTransaction(ctx, fn)
	}
	return fn(ctx)
}

// FileExists checks whether a file exists at the specified path.
func (m *Migration) FileExists(fileName string) (bool, error) {
	return m.file.Exists(fileName)
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

// sanitizeCredentials replaces credential values with masks in SQL/DSN output.
// It masks Username, Password, Token, and Credential (for Iceberg REST auth).
// DSN query-parameter forms (token=…, credential=…) are also masked so that
// DSN strings logged verbatim do not leak secrets (ФТ-10).
func (m *Migration) sanitizeCredentials(sql string) string {
	sanitized := sql

	// Replace password if present
	if m.options.Password != "" {
		sanitized = strings.ReplaceAll(sanitized, m.options.Password, credentialMask)
	}

	// Replace username if present
	if m.options.Username != "" {
		sanitized = strings.ReplaceAll(sanitized, m.options.Username, credentialMask)
	}

	// Mask bearer token value
	if m.options.Token != "" {
		sanitized = strings.ReplaceAll(sanitized, m.options.Token, credentialMask)
	}

	// Mask OAuth2 credential value
	if m.options.Credential != "" {
		sanitized = strings.ReplaceAll(sanitized, m.options.Credential, credentialMask)
	}

	// Mask S3/MinIO secret access key value
	if m.options.S3SecretAccessKey != "" {
		sanitized = strings.ReplaceAll(sanitized, m.options.S3SecretAccessKey, credentialMask)
	}

	// Mask S3/MinIO session token value
	if m.options.S3SessionToken != "" {
		sanitized = strings.ReplaceAll(sanitized, m.options.S3SessionToken, credentialMask)
	}

	// Mask DSN query-parameter forms: token=<value>, credential=<value>,
	// s3.secret-access-key=<value>, s3.session-token=<value>.
	// This covers cases where a DSN string is logged directly.
	sanitized = maskDSNParam(sanitized, "token")
	sanitized = maskDSNParam(sanitized, "credential")
	sanitized = maskDSNParam(sanitized, "s3.secret-access-key")
	sanitized = maskDSNParam(sanitized, "s3.session-token")

	return sanitized
}

// maskDSNParam replaces the value of a URL query parameter (key=value) with credentialMask.
// It handles both "key=value&" and "key=value" (end of string / end of query) forms.
// All occurrences of the parameter in the string are masked (left-to-right, no re-scan).
func maskDSNParam(s, key string) string {
	prefix := key + "="
	var b strings.Builder
	rest := s
	for {
		idx := strings.Index(rest, prefix)
		if idx == -1 {
			b.WriteString(rest)
			break
		}

		// Copy everything up to and including the "key=" prefix.
		b.WriteString(rest[:idx+len(prefix)])
		b.WriteString(credentialMask)

		// Skip past the original value (up to a delimiter or end of string).
		after := rest[idx+len(prefix):]
		end := strings.IndexAny(after, "&# ")
		if end == -1 {
			// Value extends to end of string; nothing more to append.
			break
		}
		// Continue scanning from the delimiter onward.
		rest = after[end:]
	}
	return b.String()
}
