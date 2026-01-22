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
	"io"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/raoptimus/db-migrator.go/internal/infrastructure/dal/entity"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- InitializeTableHistory Tests ---

func TestMigration_InitializeTableHistory_TableExists_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(true, nil)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.InitializeTableHistory(ctx)

	require.NoError(t, err)
}

func TestMigration_InitializeTableHistory_CreateTable_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	tableName := "public.migration"

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(false, nil)
	repo.EXPECT().TableNameWithSchema().Return(tableName)
	repo.EXPECT().CreateMigrationHistoryTable(ctx).Return(nil)
	repo.EXPECT().InsertMigration(ctx, "000000_000000_base").Return(nil)

	logger.EXPECT().Warnf("Creating migration history table %s...\n", tableName)
	logger.EXPECT().Success("Done")

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.InitializeTableHistory(ctx)

	require.NoError(t, err)
}

func TestMigration_InitializeTableHistory_HasMigrationHistoryTableReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	expectedErr := errors.New("database connection error")

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(false, expectedErr)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.InitializeTableHistory(ctx)

	require.ErrorIs(t, err, expectedErr)
}

func TestMigration_InitializeTableHistory_CreateMigrationHistoryTableReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	expectedErr := errors.New("create table error")
	tableName := "migration"

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(false, nil)
	repo.EXPECT().TableNameWithSchema().Return(tableName)
	repo.EXPECT().CreateMigrationHistoryTable(ctx).Return(expectedErr)

	logger.EXPECT().Warnf("Creating migration history table %s...\n", tableName)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.InitializeTableHistory(ctx)

	require.ErrorIs(t, err, expectedErr)
}

func TestMigration_InitializeTableHistory_InsertBaseMigrationReturnsError_DropsTable_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	insertErr := errors.New("insert base migration error")
	tableName := "migration"

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(false, nil)
	repo.EXPECT().TableNameWithSchema().Return(tableName)
	repo.EXPECT().CreateMigrationHistoryTable(ctx).Return(nil)
	repo.EXPECT().InsertMigration(ctx, "000000_000000_base").Return(insertErr)
	repo.EXPECT().DropMigrationHistoryTable(ctx).Return(nil)

	logger.EXPECT().Warnf("Creating migration history table %s...\n", tableName)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.InitializeTableHistory(ctx)

	require.ErrorIs(t, err, insertErr)
}

func TestMigration_InitializeTableHistory_InsertBaseMigrationReturnsError_DropTableAlsoFails_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	insertErr := errors.New("insert base migration error")
	dropErr := errors.New("drop table error")
	tableName := "migration"

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(false, nil)
	repo.EXPECT().TableNameWithSchema().Return(tableName)
	repo.EXPECT().CreateMigrationHistoryTable(ctx).Return(nil)
	repo.EXPECT().InsertMigration(ctx, "000000_000000_base").Return(insertErr)
	repo.EXPECT().DropMigrationHistoryTable(ctx).Return(dropErr)

	logger.EXPECT().Warnf("Creating migration history table %s...\n", tableName)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.InitializeTableHistory(ctx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "insert base migration error")
	require.Contains(t, err.Error(), "drop table error")
}

// --- Migrations Tests ---

func TestMigration_Migrations_ReturnsAppliedMigrations_Successfully(t *testing.T) {
	tests := []struct {
		name               string
		limit              int
		repoMigrations     entity.Migrations
		expectedMigrations model.Migrations
	}{
		{
			name:  "returns migrations without base migration",
			limit: 10,
			repoMigrations: entity.Migrations{
				{Version: "000000_000000_base"},
				{Version: "200101_120000_create_users"},
				{Version: "200102_130000_add_email"},
			},
			expectedMigrations: model.Migrations{
				{Version: "200101_120000_create_users"},
				{Version: "200102_130000_add_email"},
			},
		},
		{
			name:  "base migration in middle of list",
			limit: 10,
			repoMigrations: entity.Migrations{
				{Version: "200101_120000_create_users"},
				{Version: "000000_000000_base"},
				{Version: "200102_130000_add_email"},
			},
			expectedMigrations: model.Migrations{
				{Version: "200101_120000_create_users"},
				{Version: "200102_130000_add_email"},
			},
		},
		{
			name:  "base migration at end of list",
			limit: 10,
			repoMigrations: entity.Migrations{
				{Version: "200101_120000_create_users"},
				{Version: "200102_130000_add_email"},
				{Version: "000000_000000_base"},
			},
			expectedMigrations: model.Migrations{
				{Version: "200101_120000_create_users"},
				{Version: "200102_130000_add_email"},
			},
		},
		{
			name:               "empty migrations list",
			limit:              10,
			repoMigrations:     entity.Migrations{},
			expectedMigrations: model.Migrations{},
		},
		{
			name:  "no base migration in list",
			limit: 10,
			repoMigrations: entity.Migrations{
				{Version: "200101_120000_create_users"},
			},
			expectedMigrations: model.Migrations{
				{Version: "200101_120000_create_users"},
			},
		},
		{
			name:  "limit less than 1 uses default limit",
			limit: 0,
			repoMigrations: entity.Migrations{
				{Version: "000000_000000_base"},
				{Version: "200101_120000_create_users"},
			},
			expectedMigrations: model.Migrations{
				{Version: "200101_120000_create_users"},
			},
		},
		{
			name:  "negative limit uses default limit",
			limit: -5,
			repoMigrations: entity.Migrations{
				{Version: "000000_000000_base"},
				{Version: "200101_120000_create_users"},
			},
			expectedMigrations: model.Migrations{
				{Version: "200101_120000_create_users"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := NewMockRepository(t)
			file := NewMockFile(t)
			logger := NewMockLogger(t)

			repo.EXPECT().HasMigrationHistoryTable(ctx).Return(true, nil)

			expectedLimit := tt.limit
			if expectedLimit < 1 {
				expectedLimit = 10000 // defaultLimit
			}
			repo.EXPECT().Migrations(ctx, expectedLimit).Return(tt.repoMigrations, nil)

			serv := NewMigration(&Options{}, logger, file, repo)
			migrations, err := serv.Migrations(ctx, tt.limit)

			require.NoError(t, err)
			require.Equal(t, tt.expectedMigrations, migrations)
		})
	}
}

func TestMigration_Migrations_InitializeTableHistoryReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	expectedErr := errors.New("init table error")

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(false, expectedErr)

	serv := NewMigration(&Options{}, logger, file, repo)
	migrations, err := serv.Migrations(ctx, 10)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, migrations)
}

func TestMigration_Migrations_RepositoryMigrationsReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	expectedErr := errors.New("query migrations error")

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(true, nil)
	repo.EXPECT().Migrations(ctx, 10).Return(nil, expectedErr)

	serv := NewMigration(&Options{}, logger, file, repo)
	migrations, err := serv.Migrations(ctx, 10)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, migrations)
}

// --- NewMigrations Tests ---

func TestMigration_NewMigrations_ReturnsNewMigrations_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(true, nil)
	repo.EXPECT().Migrations(ctx, 100000).Return(entity.Migrations{
		{Version: "000000_000000_base"},
		{Version: "200101_120000_create_users"},
	}, nil)

	serv := NewMigration(&Options{Directory: "/nonexistent/path"}, logger, file, repo)
	// Note: filepath.Glob will return empty for non-existent path
	migrations, err := serv.NewMigrations(ctx)

	require.NoError(t, err)
	require.Empty(t, migrations)
}

func TestMigration_NewMigrations_InitializeTableHistoryReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	expectedErr := errors.New("init table error")

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(false, expectedErr)

	serv := NewMigration(&Options{Directory: "/tmp"}, logger, file, repo)
	migrations, err := serv.NewMigrations(ctx)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, migrations)
}

func TestMigration_NewMigrations_RepositoryMigrationsReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	expectedErr := errors.New("query migrations error")

	repo.EXPECT().HasMigrationHistoryTable(ctx).Return(true, nil)
	repo.EXPECT().Migrations(ctx, 100000).Return(nil, expectedErr)

	serv := NewMigration(&Options{Directory: "/tmp"}, logger, file, repo)
	migrations, err := serv.NewMigrations(ctx)

	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, migrations)
}

// --- ApplySQL Tests ---

func TestMigration_ApplySQL_SimpleSQLStatement_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	upSQL := "CREATE TABLE users (id INT);"

	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, false, version, upSQL)

	require.NoError(t, err)
}

func TestMigration_ApplySQL_MultipleStatements_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_tables"
	upSQL := "CREATE TABLE users (id INT); CREATE TABLE posts (id INT);"

	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(nil)
	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE posts (id INT)").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE posts (id INT)")
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, false, version, upSQL)

	require.NoError(t, err)
}

func TestMigration_ApplySQL_WithTransaction_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	upSQL := "CREATE TABLE users (id INT);"

	repo.EXPECT().
		ExecQueryTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, true, version, upSQL)

	require.NoError(t, err)
}

func TestMigration_ApplySQL_ReplacesCredentialPlaceholders_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_user"
	upSQL := "CREATE USER {username} WITH PASSWORD '{password}';"

	repo.EXPECT().ExecQuery(ctx, "CREATE USER admin WITH PASSWORD 'secret123'").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE USER **** WITH PASSWORD '****'")
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{Username: "admin", Password: "secret123"}, logger, file, repo)
	err := serv.ApplySQL(ctx, false, version, upSQL)

	require.NoError(t, err)
}

func TestMigration_ApplySQL_BaseMigrationVersion_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "000000_000000_base"
	upSQL := "SELECT 1;"

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, false, version, upSQL)

	require.ErrorIs(t, err, ErrMigrationVersionReserved)
}

func TestMigration_ApplySQL_ExecQueryReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	upSQL := "CREATE TABLE users (id INT);"
	expectedErr := errors.New("exec query error")

	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(expectedErr)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")
	logger.EXPECT().Errorf("*** failed to apply %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, false, version, upSQL)

	require.ErrorIs(t, err, expectedErr)
}

func TestMigration_ApplySQL_InsertMigrationReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	upSQL := "CREATE TABLE users (id INT);"
	expectedErr := errors.New("insert migration error")

	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(expectedErr)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, false, version, upSQL)

	require.ErrorIs(t, err, expectedErr)
}

func TestMigration_ApplySQL_TransactionReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	upSQL := "CREATE TABLE users (id INT);"
	expectedErr := errors.New("transaction error")

	repo.EXPECT().ExecQueryTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).Return(expectedErr)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Errorf("*** failed to apply %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, true, version, upSQL)

	require.ErrorIs(t, err, expectedErr)
}

// --- RevertSQL Tests ---

func TestMigration_RevertSQL_SimpleSQLStatement_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	downSQL := "DROP TABLE users;"

	repo.EXPECT().ExecQuery(ctx, "DROP TABLE users").Return(nil)
	repo.EXPECT().RemoveMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "DROP TABLE users")
	logger.EXPECT().Warnf("*** reverted %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertSQL(ctx, false, version, downSQL)

	require.NoError(t, err)
}

func TestMigration_RevertSQL_WithTransaction_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	downSQL := "DROP TABLE users;"

	repo.EXPECT().
		ExecQueryTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	repo.EXPECT().ExecQuery(ctx, "DROP TABLE users").Return(nil)
	repo.EXPECT().RemoveMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "DROP TABLE users")
	logger.EXPECT().Warnf("*** reverted %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertSQL(ctx, true, version, downSQL)

	require.NoError(t, err)
}

func TestMigration_RevertSQL_BaseMigrationVersion_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "000000_000000_base"
	downSQL := "SELECT 1;"

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertSQL(ctx, false, version, downSQL)

	require.ErrorIs(t, err, ErrMigrationVersionReserved)
}

func TestMigration_RevertSQL_ExecQueryReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	downSQL := "DROP TABLE users;"
	expectedErr := errors.New("exec query error")

	repo.EXPECT().ExecQuery(ctx, "DROP TABLE users").Return(expectedErr)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "DROP TABLE users")
	logger.EXPECT().Errorf("*** failed to revert %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertSQL(ctx, false, version, downSQL)

	require.ErrorIs(t, err, expectedErr)
}

func TestMigration_RevertSQL_RemoveMigrationReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	downSQL := "DROP TABLE users;"
	expectedErr := errors.New("remove migration error")

	repo.EXPECT().ExecQuery(ctx, "DROP TABLE users").Return(nil)
	repo.EXPECT().RemoveMigration(ctx, version).Return(expectedErr)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "DROP TABLE users")

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertSQL(ctx, false, version, downSQL)

	require.ErrorIs(t, err, expectedErr)
}

// --- ApplyFile Tests ---

func TestMigration_ApplyFile_SimpleSQLStatement_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.up.sql"
	sqlContent := "CREATE TABLE users (id INT);"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.NoError(t, err)
}

func TestMigration_ApplyFile_MultipleStatements_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_tables"
	fileName := "/migrations/200101_120000_create_tables.up.sql"
	sqlContent := "CREATE TABLE users (id INT); CREATE TABLE posts (id INT);"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(nil)
	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE posts (id INT)").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE posts (id INT)")
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.NoError(t, err)
}

func TestMigration_ApplyFile_WithTransaction_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.safe.up.sql"
	sqlContent := "CREATE TABLE users (id INT);"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().
		ExecQueryTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: version}, fileName, true)

	require.NoError(t, err)
}

func TestMigration_ApplyFile_BaseMigrationVersion_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	fileName := "/migrations/000000_000000_base.up.sql"

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: "000000_000000_base"}, fileName, false)

	require.ErrorIs(t, err, ErrMigrationVersionReserved)
}

func TestMigration_ApplyFile_FileDoesNotExist_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.up.sql"

	file.EXPECT().Exists(fileName).Return(false, nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}

func TestMigration_ApplyFile_FileExistsReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.up.sql"
	expectedErr := errors.New("file system error")

	file.EXPECT().Exists(fileName).Return(false, expectedErr)

	logger.EXPECT().Warnf("*** applying %s\n", version)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}

func TestMigration_ApplyFile_FileOpenReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.up.sql"
	expectedErr := errors.New("permission denied")

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(nil, expectedErr)

	logger.EXPECT().Warnf("*** applying %s\n", version)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not read")
}

func TestMigration_ApplyFile_ExecQueryReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.up.sql"
	sqlContent := "CREATE TABLE users (id INT);"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))
	expectedErr := errors.New("exec query error")

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(expectedErr)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")
	logger.EXPECT().Errorf("*** failed to apply %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.ErrorIs(t, err, expectedErr)
}

func TestMigration_ApplyFile_InsertMigrationReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.up.sql"
	sqlContent := "CREATE TABLE users (id INT);"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))
	expectedErr := errors.New("insert migration error")

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().ExecQuery(ctx, "CREATE TABLE users (id INT)").Return(nil)
	repo.EXPECT().InsertMigration(ctx, version).Return(expectedErr)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE TABLE users (id INT)")

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplyFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.ErrorIs(t, err, expectedErr)
}

// --- RevertFile Tests ---

func TestMigration_RevertFile_SimpleSQLStatement_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.down.sql"
	sqlContent := "DROP TABLE users;"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().ExecQuery(ctx, "DROP TABLE users").Return(nil)
	repo.EXPECT().RemoveMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "DROP TABLE users")
	logger.EXPECT().Warnf("*** reverted %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.NoError(t, err)
}

func TestMigration_RevertFile_WithTransaction_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.safe.down.sql"
	sqlContent := "DROP TABLE users;"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().
		ExecQueryTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).
		RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})
	repo.EXPECT().ExecQuery(ctx, "DROP TABLE users").Return(nil)
	repo.EXPECT().RemoveMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "DROP TABLE users")
	logger.EXPECT().Warnf("*** reverted %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertFile(ctx, &model.Migration{Version: version}, fileName, true)

	require.NoError(t, err)
}

func TestMigration_RevertFile_BaseMigrationVersion_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	fileName := "/migrations/000000_000000_base.down.sql"

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertFile(ctx, &model.Migration{Version: "000000_000000_base"}, fileName, false)

	require.ErrorIs(t, err, ErrMigrationVersionReserved)
}

func TestMigration_RevertFile_FileDoesNotExist_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.down.sql"

	file.EXPECT().Exists(fileName).Return(false, nil)

	logger.EXPECT().Warnf("*** reverting %s\n", version)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")
}

func TestMigration_RevertFile_ExecQueryReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.down.sql"
	sqlContent := "DROP TABLE users;"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))
	expectedErr := errors.New("exec query error")

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().ExecQuery(ctx, "DROP TABLE users").Return(expectedErr)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "DROP TABLE users")
	logger.EXPECT().Errorf("*** failed to revert %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.ErrorIs(t, err, expectedErr)
}

func TestMigration_RevertFile_RemoveMigrationReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.down.sql"
	sqlContent := "DROP TABLE users;"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))
	expectedErr := errors.New("remove migration error")

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().ExecQuery(ctx, "DROP TABLE users").Return(nil)
	repo.EXPECT().RemoveMigration(ctx, version).Return(expectedErr)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "DROP TABLE users")

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertFile(ctx, &model.Migration{Version: version}, fileName, false)

	require.ErrorIs(t, err, expectedErr)
}

func TestMigration_RevertFile_TransactionReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	fileName := "/migrations/200101_120000_create_users.safe.down.sql"
	sqlContent := "DROP TABLE users;"
	sqlReader := io.NopCloser(strings.NewReader(sqlContent))
	expectedErr := errors.New("transaction error")

	file.EXPECT().Exists(fileName).Return(true, nil)
	file.EXPECT().Open(fileName).Return(sqlReader, nil)

	repo.EXPECT().ExecQueryTransaction(ctx, mock.AnythingOfType("func(context.Context) error")).Return(expectedErr)

	logger.EXPECT().Warnf("*** reverting %s\n", version)
	logger.EXPECT().Errorf("*** failed to revert %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.RevertFile(ctx, &model.Migration{Version: version}, fileName, true)

	require.ErrorIs(t, err, expectedErr)
}

// --- ExecQuery Tests ---

func TestMigration_ExecQuery_SimpleSQLStatement_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	sqlQuery := "SELECT 1"

	repo.EXPECT().ExecQuery(ctx, sqlQuery).Return(nil)

	logger.EXPECT().Infof("    > execute SQL: %s ...\n", sqlQuery)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ExecQuery(ctx, sqlQuery)

	require.NoError(t, err)
}

func TestMigration_ExecQuery_CompactMode_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	sqlQuery := "SELECT 1"

	repo.EXPECT().ExecQuery(ctx, sqlQuery).Return(nil)

	logger.EXPECT().Infof(" done (time: '%.3fs)\n", mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{Compact: true}, logger, file, repo)
	err := serv.ExecQuery(ctx, sqlQuery)

	require.NoError(t, err)
}

func TestMigration_ExecQuery_ExecQueryReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	sqlQuery := "SELECT 1"
	expectedErr := errors.New("exec query error")

	repo.EXPECT().ExecQuery(ctx, sqlQuery).Return(expectedErr)

	logger.EXPECT().Infof("    > execute SQL: %s ...\n", sqlQuery)

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ExecQuery(ctx, sqlQuery)

	require.ErrorIs(t, err, expectedErr)
}

// --- SQLQueryOutput Tests ---

func TestMigration_SQLQueryOutput_ReturnsFormattedSQL(t *testing.T) {
	tests := []struct {
		name           string
		options        *Options
		sqlQuery       string
		expectedOutput string
	}{
		{
			name:           "no transformation needed",
			options:        &Options{},
			sqlQuery:       "SELECT * FROM users",
			expectedOutput: "SELECT * FROM users",
		},
		{
			name:           "masks password in output",
			options:        &Options{Password: "secret123"},
			sqlQuery:       "CREATE USER admin WITH PASSWORD 'secret123'",
			expectedOutput: "CREATE USER admin WITH PASSWORD '****'",
		},
		{
			name:           "masks username in output",
			options:        &Options{Username: "admin"},
			sqlQuery:       "GRANT ALL TO admin",
			expectedOutput: "GRANT ALL TO ****",
		},
		{
			name:           "masks both username and password",
			options:        &Options{Username: "admin", Password: "secret123"},
			sqlQuery:       "CREATE USER admin WITH PASSWORD 'secret123'",
			expectedOutput: "CREATE USER **** WITH PASSWORD '****'",
		},
		{
			name:           "truncates long SQL at max length",
			options:        &Options{MaxSQLOutputLength: 10},
			sqlQuery:       "SELECT * FROM users WHERE id = 1",
			expectedOutput: "SELECT * F...",
		},
		{
			name:           "truncates long SQL at boundary",
			options:        &Options{MaxSQLOutputLength: 20},
			sqlQuery:       "SELECT * FROM users WHERE id = 1",
			expectedOutput: "SELECT * FROM users ...",
		},
		{
			name:           "no truncation when length equals limit",
			options:        &Options{MaxSQLOutputLength: 33},
			sqlQuery:       "SELECT * FROM users WHERE id = 1",
			expectedOutput: "SELECT * FROM users WHERE id = 1",
		},
		{
			name:           "no truncation when MaxSQLOutputLength is 0",
			options:        &Options{MaxSQLOutputLength: 0},
			sqlQuery:       "SELECT * FROM users WHERE id = 1 AND name = 'John'",
			expectedOutput: "SELECT * FROM users WHERE id = 1 AND name = 'John'",
		},
		{
			name:           "sanitizes credentials before truncating",
			options:        &Options{Password: "secret123", MaxSQLOutputLength: 30},
			sqlQuery:       "CREATE USER admin WITH PASSWORD 'secret123'",
			expectedOutput: "CREATE USER admin WITH PASSWOR...",
		},
		{
			name:           "empty username and password returns original",
			options:        &Options{Username: "", Password: ""},
			sqlQuery:       "SELECT * FROM users",
			expectedOutput: "SELECT * FROM users",
		},
		{
			name:           "multiple occurrences of password masked",
			options:        &Options{Password: "pass"},
			sqlQuery:       "pass and pass and pass",
			expectedOutput: "**** and **** and ****",
		},
		{
			name:           "MaxSQLOutputLength of 1",
			options:        &Options{MaxSQLOutputLength: 1},
			sqlQuery:       "SELECT * FROM users",
			expectedOutput: "S...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepository(t)
			file := NewMockFile(t)
			logger := NewMockLogger(t)

			serv := NewMigration(tt.options, logger, file, repo)
			output := serv.SQLQueryOutput(tt.sqlQuery)

			require.Equal(t, tt.expectedOutput, output)
		})
	}
}

// --- BeginCommand Tests ---

func TestMigration_BeginCommand_NonCompactMode_LogsSQLQuery(t *testing.T) {
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	sqlQuery := "SELECT * FROM users"

	logger.EXPECT().Infof("    > execute SQL: %s ...\n", sqlQuery)

	serv := NewMigration(&Options{Compact: false}, logger, file, repo)
	startTime := serv.BeginCommand(sqlQuery)

	require.False(t, startTime.IsZero())
}

func TestMigration_BeginCommand_CompactMode_DoesNotLogSQLQuery(t *testing.T) {
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	sqlQuery := "SELECT * FROM users"

	// No logger expectations - should not log in compact mode

	serv := NewMigration(&Options{Compact: true}, logger, file, repo)
	startTime := serv.BeginCommand(sqlQuery)

	require.False(t, startTime.IsZero())
}

func TestMigration_BeginCommand_SanitizesCredentialsInLog(t *testing.T) {
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	sqlQuery := "CREATE USER admin WITH PASSWORD 'secret123'"

	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "CREATE USER **** WITH PASSWORD '****'")

	serv := NewMigration(&Options{Username: "admin", Password: "secret123", Compact: false}, logger, file, repo)
	startTime := serv.BeginCommand(sqlQuery)

	require.False(t, startTime.IsZero())
}

// --- EndCommand Tests ---

func TestMigration_EndCommand_CompactMode_LogsTime(t *testing.T) {
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)

	logger.EXPECT().Infof(" done (time: '%.3fs)\n", mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{Compact: true}, logger, file, repo)
	serv.EndCommand(serv.BeginCommand("SELECT 1"))
}

func TestMigration_EndCommand_NonCompactMode_DoesNotLog(t *testing.T) {
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)

	// No logger expectations in non-compact mode for EndCommand

	serv := NewMigration(&Options{Compact: false}, logger, file, repo)
	// Note: BeginCommand will log in non-compact mode, but EndCommand should not
	logger.EXPECT().Infof("    > execute SQL: %s ...\n", "SELECT 1")
	serv.EndCommand(serv.BeginCommand("SELECT 1"))
}

// --- Exists Tests ---

func TestMigration_Exists_MigrationExists_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"

	repo.EXPECT().ExistsMigration(ctx, version).Return(true, nil)

	serv := NewMigration(&Options{}, logger, file, repo)
	exists, err := serv.Exists(ctx, version)

	require.NoError(t, err)
	require.True(t, exists)
}

func TestMigration_Exists_MigrationDoesNotExist_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"

	repo.EXPECT().ExistsMigration(ctx, version).Return(false, nil)

	serv := NewMigration(&Options{}, logger, file, repo)
	exists, err := serv.Exists(ctx, version)

	require.NoError(t, err)
	require.False(t, exists)
}

func TestMigration_Exists_RepositoryReturnsError_Failure(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_create_users"
	expectedErr := errors.New("database error")

	repo.EXPECT().ExistsMigration(ctx, version).Return(false, expectedErr)

	serv := NewMigration(&Options{}, logger, file, repo)
	exists, err := serv.Exists(ctx, version)

	require.ErrorIs(t, err, expectedErr)
	require.False(t, exists)
}

// --- sanitizeCredentials Tests (private method, tested via SQLQueryOutput) ---

func TestMigration_SanitizeCredentials_VariousInputs(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		inputSQL       string
		expectedOutput string
	}{
		{
			name:           "empty credentials returns original SQL",
			username:       "",
			password:       "",
			inputSQL:       "SELECT * FROM users WHERE name = 'test'",
			expectedOutput: "SELECT * FROM users WHERE name = 'test'",
		},
		{
			name:           "only username is masked",
			username:       "admin",
			password:       "",
			inputSQL:       "GRANT SELECT ON users TO admin",
			expectedOutput: "GRANT SELECT ON users TO ****",
		},
		{
			name:           "only password is masked",
			username:       "",
			password:       "secret",
			inputSQL:       "SET PASSWORD = 'secret'",
			expectedOutput: "SET PASSWORD = '****'",
		},
		{
			name:           "both credentials are masked",
			username:       "admin",
			password:       "secret",
			inputSQL:       "CREATE USER admin IDENTIFIED BY 'secret'",
			expectedOutput: "CREATE USER **** IDENTIFIED BY '****'",
		},
		{
			name:           "multiple username occurrences",
			username:       "user1",
			password:       "",
			inputSQL:       "user1 grants to user1 and user1",
			expectedOutput: "**** grants to **** and ****",
		},
		{
			name:           "multiple password occurrences",
			username:       "",
			password:       "pwd",
			inputSQL:       "pwd is pwd but not pwd2",
			expectedOutput: "**** is **** but not ****2",
		},
		{
			name:           "case sensitive matching",
			username:       "Admin",
			password:       "Secret",
			inputSQL:       "User: Admin, admin, ADMIN with password Secret, secret",
			expectedOutput: "User: ****, admin, ADMIN with password ****, secret",
		},
		{
			name:           "overlapping credentials",
			username:       "abc",
			password:       "abcdef",
			inputSQL:       "credentials: abcdef abc",
			expectedOutput: "credentials: **** ****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepository(t)
			file := NewMockFile(t)
			logger := NewMockLogger(t)

			serv := NewMigration(&Options{
				Username: tt.username,
				Password: tt.password,
			}, logger, file, repo)

			// Test via SQLQueryOutput since sanitizeCredentials is private
			output := serv.SQLQueryOutput(tt.inputSQL)

			require.Equal(t, tt.expectedOutput, output)
		})
	}
}

// --- Boundary Value Tests for MaxSQLOutputLength ---

func TestMigration_SQLQueryOutput_BoundaryValues(t *testing.T) {
	tests := []struct {
		name               string
		maxSQLOutputLength int
		sqlQuery           string
		expectedOutput     string
	}{
		{
			name:               "maxLength 0 means no limit",
			maxSQLOutputLength: 0,
			sqlQuery:           "SELECT",
			expectedOutput:     "SELECT",
		},
		{
			name:               "maxLength 1 truncates to 1 char plus ellipsis",
			maxSQLOutputLength: 1,
			sqlQuery:           "SELECT",
			expectedOutput:     "S...",
		},
		{
			name:               "maxLength equals string length - no truncation",
			maxSQLOutputLength: 6,
			sqlQuery:           "SELECT",
			expectedOutput:     "SELECT",
		},
		{
			name:               "maxLength one less than string length - truncates",
			maxSQLOutputLength: 5,
			sqlQuery:           "SELECT",
			expectedOutput:     "SELEC...",
		},
		{
			name:               "maxLength one more than string length - no truncation",
			maxSQLOutputLength: 7,
			sqlQuery:           "SELECT",
			expectedOutput:     "SELECT",
		},
		{
			name:               "maxLength greater than string length - no truncation",
			maxSQLOutputLength: 100,
			sqlQuery:           "SELECT",
			expectedOutput:     "SELECT",
		},
		{
			name:               "empty SQL query",
			maxSQLOutputLength: 10,
			sqlQuery:           "",
			expectedOutput:     "",
		},
		{
			name:               "negative maxLength treated as 0 (no limit)",
			maxSQLOutputLength: -1,
			sqlQuery:           "SELECT * FROM users",
			expectedOutput:     "SELECT * FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewMockRepository(t)
			file := NewMockFile(t)
			logger := NewMockLogger(t)

			serv := NewMigration(&Options{MaxSQLOutputLength: tt.maxSQLOutputLength}, logger, file, repo)
			output := serv.SQLQueryOutput(tt.sqlQuery)

			require.Equal(t, tt.expectedOutput, output)
		})
	}
}

// --- Credential Placeholder Replacement Tests ---

func TestMigration_Apply_ReplacesCredentialPlaceholders(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		inputSQL    string
		expectedSQL string
	}{
		{
			name:        "replaces username placeholder",
			username:    "dbuser",
			password:    "",
			inputSQL:    "CREATE USER {username};",
			expectedSQL: "CREATE USER dbuser",
		},
		{
			name:        "replaces password placeholder",
			username:    "",
			password:    "dbpass",
			inputSQL:    "SET PASSWORD '{password}';",
			expectedSQL: "SET PASSWORD 'dbpass'",
		},
		{
			name:        "replaces both placeholders",
			username:    "admin",
			password:    "secret",
			inputSQL:    "CREATE USER {username} WITH PASSWORD '{password}';",
			expectedSQL: "CREATE USER admin WITH PASSWORD 'secret'",
		},
		{
			name:        "multiple placeholder occurrences",
			username:    "user1",
			password:    "pass1",
			inputSQL:    "{username} and {username} with {password} and {password};",
			expectedSQL: "user1 and user1 with pass1 and pass1",
		},
		{
			name:        "no placeholders unchanged",
			username:    "admin",
			password:    "secret",
			inputSQL:    "SELECT * FROM users;",
			expectedSQL: "SELECT * FROM users",
		},
		{
			name:        "empty credentials replace with empty strings",
			username:    "",
			password:    "",
			inputSQL:    "CREATE USER {username} WITH PASSWORD '{password}';",
			expectedSQL: "CREATE USER  WITH PASSWORD ''",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := NewMockRepository(t)
			file := NewMockFile(t)
			logger := NewMockLogger(t)
			version := "200101_120000_test"

			repo.EXPECT().ExecQuery(ctx, tt.expectedSQL).Return(nil)
			repo.EXPECT().InsertMigration(ctx, version).Return(nil)

			logger.EXPECT().Warnf("*** applying %s\n", version)
			logger.EXPECT().Infof("    > execute SQL: %s ...\n", mock.Anything)
			logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

			serv := NewMigration(&Options{Username: tt.username, Password: tt.password}, logger, file, repo)
			err := serv.ApplySQL(ctx, false, version, tt.inputSQL)

			require.NoError(t, err)
		})
	}
}

// --- NewMigration Constructor Test ---

func TestNewMigration_ReturnsInitializedStruct(t *testing.T) {
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	options := &Options{
		MaxSQLOutputLength: 100,
		Directory:          "/migrations",
		Compact:            true,
		Username:           "user",
		Password:           "pass",
	}

	serv := NewMigration(options, logger, file, repo)

	require.NotNil(t, serv)
	require.Equal(t, options, serv.options)
	require.Equal(t, logger, serv.logger)
	require.Equal(t, file, serv.file)
	require.Equal(t, repo, serv.repo)
}

// --- Empty SQL Content Tests ---

func TestMigration_ApplySQL_EmptySQLContent_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_empty"
	upSQL := ""

	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, false, version, upSQL)

	require.NoError(t, err)
}

func TestMigration_ApplySQL_OnlySemicolonsAndWhitespace_Successfully(t *testing.T) {
	ctx := context.Background()
	repo := NewMockRepository(t)
	file := NewMockFile(t)
	logger := NewMockLogger(t)
	version := "200101_120000_empty"
	upSQL := "  ;  ;  ;"

	repo.EXPECT().InsertMigration(ctx, version).Return(nil)

	logger.EXPECT().Warnf("*** applying %s\n", version)
	logger.EXPECT().Successf("*** applied %s (time: %.3fs)\n", version, mock.AnythingOfType("float64"))

	serv := NewMigration(&Options{}, logger, file, repo)
	err := serv.ApplySQL(ctx, false, version, upSQL)

	require.NoError(t, err)
}
