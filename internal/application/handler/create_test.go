/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"errors"
	"testing"
	"time"

	"github.com/raoptimus/db-migrator.go/internal/helper/timex"
	"github.com/stretchr/testify/require"
)

// TestNewCreate_InitializesFieldsCorrectly_Successfully verifies that NewCreate
// properly initializes all struct fields.
func TestNewCreate_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	tests := []struct {
		name    string
		options *Options
	}{
		{
			name: "basic options non-interactive",
			options: &Options{
				DSN:         "postgres://user:pass@localhost:5432/testdb",
				Directory:   "/migrations",
				Interactive: false,
			},
		},
		{
			name: "basic options interactive",
			options: &Options{
				DSN:         "postgres://user:pass@localhost:5432/testdb",
				Directory:   "/migrations",
				Interactive: true,
			},
		},
		{
			name: "clickhouse options",
			options: &Options{
				DSN:         "clickhouse://user:pass@localhost:9000/testdb",
				Directory:   "/var/migrations",
				ClusterName: "test_cluster",
				Replicated:  true,
				Interactive: false,
			},
		},
		{
			name: "mysql options",
			options: &Options{
				DSN:       "mysql://user:pass@localhost:3306/testdb",
				Directory: "./migrations",
			},
		},
		{
			name: "empty directory",
			options: &Options{
				DSN:       "postgres://user:pass@localhost:5432/testdb",
				Directory: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			create := NewCreate(tt.options, loggerMock, fileMock, fileNameBuilderMock)

			require.NotNil(t, create)
			require.Equal(t, tt.options, create.options)
			require.NotNil(t, create.file)
			require.NotNil(t, create.logger)
			require.NotNil(t, create.fileNameBuilder)
		})
	}
}

// TestCreate_Handle_ArgsNotPresent_Failure tests that Handle returns ErrInvalidFileName
// when no arguments are provided (args not present).
func TestCreate_Handle_ArgsNotPresent_Failure(t *testing.T) {
	fileMock := NewMockFile(t)
	loggerMock := NewMockLogger(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)

	create := NewCreate(
		&Options{
			Directory:   "/migrations",
			Interactive: false,
		},
		loggerMock,
		fileMock,
		fileNameBuilderMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false, first: ""},
	}

	err := create.Handle(cmd)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidFileName)
}

// TestCreate_Handle_InvalidMigrationName_Failure tests that Handle returns ErrInvalidFileName
// when the migration name doesn't match the required regex pattern.
// The regex pattern is: ^[\w\\]+$ (letters, digits, underscore, backslash)
func TestCreate_Handle_InvalidMigrationName_Failure(t *testing.T) {
	tests := []struct {
		name          string
		migrationName string
	}{
		{
			name:          "empty string",
			migrationName: "",
		},
		{
			name:          "contains space",
			migrationName: "create user table",
		},
		{
			name:          "contains hyphen",
			migrationName: "create-user-table",
		},
		{
			name:          "contains dot",
			migrationName: "create.user.table",
		},
		{
			name:          "contains at symbol",
			migrationName: "create@table",
		},
		{
			name:          "contains hash",
			migrationName: "create#table",
		},
		{
			name:          "contains dollar sign",
			migrationName: "create$table",
		},
		{
			name:          "contains percent",
			migrationName: "create%table",
		},
		{
			name:          "contains asterisk",
			migrationName: "create*table",
		},
		{
			name:          "contains plus",
			migrationName: "create+table",
		},
		{
			name:          "contains equals",
			migrationName: "create=table",
		},
		{
			name:          "contains parentheses",
			migrationName: "create(table)",
		},
		{
			name:          "contains brackets",
			migrationName: "create[table]",
		},
		{
			name:          "contains curly braces",
			migrationName: "create{table}",
		},
		{
			name:          "contains pipe",
			migrationName: "create|table",
		},
		{
			name:          "contains semicolon",
			migrationName: "create;table",
		},
		{
			name:          "contains colon",
			migrationName: "create:table",
		},
		{
			name:          "contains quote",
			migrationName: "create'table",
		},
		{
			name:          "contains double quote",
			migrationName: "create\"table",
		},
		{
			name:          "contains less than",
			migrationName: "create<table",
		},
		{
			name:          "contains greater than",
			migrationName: "create>table",
		},
		{
			name:          "contains comma",
			migrationName: "create,table",
		},
		{
			name:          "contains question mark",
			migrationName: "create?table",
		},
		{
			name:          "contains exclamation mark",
			migrationName: "create!table",
		},
		{
			name:          "contains tilde",
			migrationName: "create~table",
		},
		{
			name:          "contains backtick",
			migrationName: "create`table",
		},
		{
			name:          "contains caret",
			migrationName: "create^table",
		},
		{
			name:          "contains ampersand",
			migrationName: "create&table",
		},
		{
			name:          "only spaces",
			migrationName: "   ",
		},
		{
			name:          "leading space",
			migrationName: " create_table",
		},
		{
			name:          "trailing space",
			migrationName: "create_table ",
		},
		{
			name:          "newline character",
			migrationName: "create\ntable",
		},
		{
			name:          "tab character",
			migrationName: "create\ttable",
		},
		{
			name:          "forward slash",
			migrationName: "create/table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			create := NewCreate(
				&Options{
					Directory:   "/migrations",
					Interactive: false,
				},
				loggerMock,
				fileMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.migrationName},
			}

			err := create.Handle(cmd)

			require.Error(t, err)
			require.ErrorIs(t, err, ErrInvalidFileName)
		})
	}
}

// TestCreate_Handle_ValidMigrationName_MatchesRegex_Successfully tests that valid
// migration names that match the regex pattern ^[\w\\]+$ are accepted.
func TestCreate_Handle_ValidMigrationName_MatchesRegex_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		migrationName string
	}{
		{
			name:          "simple lowercase",
			migrationName: "createusertable",
		},
		{
			name:          "with underscore",
			migrationName: "create_user_table",
		},
		{
			name:          "with digits",
			migrationName: "create_user_table_123",
		},
		{
			name:          "only digits",
			migrationName: "123456",
		},
		{
			name:          "uppercase letters",
			migrationName: "CreateUserTable",
		},
		{
			name:          "mixed case with underscore",
			migrationName: "Create_User_Table",
		},
		{
			name:          "single character",
			migrationName: "a",
		},
		{
			name:          "single digit",
			migrationName: "1",
		},
		{
			name:          "single underscore",
			migrationName: "_",
		},
		{
			name:          "leading underscore",
			migrationName: "_create_table",
		},
		{
			name:          "trailing underscore",
			migrationName: "create_table_",
		},
		{
			name:          "multiple underscores",
			migrationName: "create__user__table",
		},
		{
			name:          "backslash character",
			migrationName: "create\\table",
		},
		{
			name:          "multiple backslashes",
			migrationName: "create\\\\table",
		},
		{
			name:          "mixed with backslash",
			migrationName: "create_user\\table_123",
		},
		{
			name:          "long migration name",
			migrationName: "create_very_long_migration_name_with_many_words_and_numbers_123456789",
		},
	}

	// Set a fixed time for predictable test output
	originalStdTime := timex.StdTime
	fixedTime := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
	timex.StdTime = timex.New(func() time.Time { return fixedTime })
	defer func() { timex.StdTime = originalStdTime }()

	expectedPrefix := "240515_103045"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			expectedVersion := expectedPrefix + "_" + tt.migrationName
			expectedFileNameUp := "/migrations/" + expectedVersion + ".safe.up.sql"
			expectedFileNameDown := "/migrations/" + expectedVersion + ".safe.down.sql"

			fileNameBuilderMock.EXPECT().
				Up(expectedVersion, true).
				Return(expectedFileNameUp, true)

			fileNameBuilderMock.EXPECT().
				Down(expectedVersion, true).
				Return(expectedFileNameDown, true)

			fileMock.EXPECT().
				Exists("/migrations").
				Return(true, nil)

			fileMock.EXPECT().
				Create(expectedFileNameUp).
				Return(nil)

			fileMock.EXPECT().
				Create(expectedFileNameDown).
				Return(nil)

			loggerMock.EXPECT().
				Success("New migration created successfully.").
				Return()

			create := NewCreate(
				&Options{
					Directory:   "/migrations",
					Interactive: false,
				},
				loggerMock,
				fileMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.migrationName},
			}

			err := create.Handle(cmd)

			require.NoError(t, err)
		})
	}
}

// TestCreate_Handle_DirectoryExists_Successfully tests that Handle works correctly
// when the migration directory already exists.
func TestCreate_Handle_DirectoryExists_Successfully(t *testing.T) {
	// Set a fixed time for predictable test output
	originalStdTime := timex.StdTime
	fixedTime := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
	timex.StdTime = timex.New(func() time.Time { return fixedTime })
	defer func() { timex.StdTime = originalStdTime }()

	fileMock := NewMockFile(t)
	loggerMock := NewMockLogger(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)

	expectedVersion := "240515_103045_create_users"
	expectedFileNameUp := "/var/migrations/240515_103045_create_users.safe.up.sql"
	expectedFileNameDown := "/var/migrations/240515_103045_create_users.safe.down.sql"

	fileNameBuilderMock.EXPECT().
		Up(expectedVersion, true).
		Return(expectedFileNameUp, true)

	fileNameBuilderMock.EXPECT().
		Down(expectedVersion, true).
		Return(expectedFileNameDown, true)

	// Directory exists
	fileMock.EXPECT().
		Exists("/var/migrations").
		Return(true, nil)

	fileMock.EXPECT().
		Create(expectedFileNameUp).
		Return(nil)

	fileMock.EXPECT().
		Create(expectedFileNameDown).
		Return(nil)

	loggerMock.EXPECT().
		Success("New migration created successfully.").
		Return()

	create := NewCreate(
		&Options{
			Directory:   "/var/migrations",
			Interactive: false,
		},
		loggerMock,
		fileMock,
		fileNameBuilderMock,
	)

	cmd := &Command{
		Args: &argsStub{present: true, first: "create_users"},
	}

	err := create.Handle(cmd)

	require.NoError(t, err)
}

// TestCreate_Handle_DirectoryExistsCheckReturnsError_Failure tests that Handle
// returns an error when checking directory existence fails.
func TestCreate_Handle_DirectoryExistsCheckReturnsError_Failure(t *testing.T) {
	// Set a fixed time for predictable test output
	originalStdTime := timex.StdTime
	fixedTime := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
	timex.StdTime = timex.New(func() time.Time { return fixedTime })
	defer func() { timex.StdTime = originalStdTime }()

	tests := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "permission denied",
			expectedErr: errors.New("permission denied"),
		},
		{
			name:        "filesystem error",
			expectedErr: errors.New("filesystem error"),
		},
		{
			name:        "io error",
			expectedErr: errors.New("io error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			expectedVersion := "240515_103045_create_users"
			expectedFileNameUp := "/migrations/240515_103045_create_users.safe.up.sql"
			expectedFileNameDown := "/migrations/240515_103045_create_users.safe.down.sql"

			fileNameBuilderMock.EXPECT().
				Up(expectedVersion, true).
				Return(expectedFileNameUp, true)

			fileNameBuilderMock.EXPECT().
				Down(expectedVersion, true).
				Return(expectedFileNameDown, true)

			// Directory exists check fails with error
			fileMock.EXPECT().
				Exists("/migrations").
				Return(false, tt.expectedErr)

			create := NewCreate(
				&Options{
					Directory:   "/migrations",
					Interactive: false,
				},
				loggerMock,
				fileMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: "create_users"},
			}

			err := create.Handle(cmd)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

// TestCreate_Handle_CreateUpFileReturnsError_Failure tests that Handle returns an error
// when creating the up migration file fails.
func TestCreate_Handle_CreateUpFileReturnsError_Failure(t *testing.T) {
	// Set a fixed time for predictable test output
	originalStdTime := timex.StdTime
	fixedTime := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
	timex.StdTime = timex.New(func() time.Time { return fixedTime })
	defer func() { timex.StdTime = originalStdTime }()

	tests := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "permission denied",
			expectedErr: errors.New("permission denied"),
		},
		{
			name:        "disk full",
			expectedErr: errors.New("no space left on device"),
		},
		{
			name:        "file already exists",
			expectedErr: errors.New("file already exists"),
		},
		{
			name:        "io error",
			expectedErr: errors.New("io error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			expectedVersion := "240515_103045_create_users"
			expectedFileNameUp := "/migrations/240515_103045_create_users.safe.up.sql"
			expectedFileNameDown := "/migrations/240515_103045_create_users.safe.down.sql"

			fileNameBuilderMock.EXPECT().
				Up(expectedVersion, true).
				Return(expectedFileNameUp, true)

			fileNameBuilderMock.EXPECT().
				Down(expectedVersion, true).
				Return(expectedFileNameDown, true)

			fileMock.EXPECT().
				Exists("/migrations").
				Return(true, nil)

			// Creating up file fails
			fileMock.EXPECT().
				Create(expectedFileNameUp).
				Return(tt.expectedErr)

			create := NewCreate(
				&Options{
					Directory:   "/migrations",
					Interactive: false,
				},
				loggerMock,
				fileMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: "create_users"},
			}

			err := create.Handle(cmd)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

// TestCreate_Handle_CreateDownFileReturnsError_Failure tests that Handle returns an error
// when creating the down migration file fails (after up file was created successfully).
func TestCreate_Handle_CreateDownFileReturnsError_Failure(t *testing.T) {
	// Set a fixed time for predictable test output
	originalStdTime := timex.StdTime
	fixedTime := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
	timex.StdTime = timex.New(func() time.Time { return fixedTime })
	defer func() { timex.StdTime = originalStdTime }()

	tests := []struct {
		name        string
		expectedErr error
	}{
		{
			name:        "permission denied",
			expectedErr: errors.New("permission denied"),
		},
		{
			name:        "disk full",
			expectedErr: errors.New("no space left on device"),
		},
		{
			name:        "file already exists",
			expectedErr: errors.New("file already exists"),
		},
		{
			name:        "io error",
			expectedErr: errors.New("io error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			expectedVersion := "240515_103045_create_users"
			expectedFileNameUp := "/migrations/240515_103045_create_users.safe.up.sql"
			expectedFileNameDown := "/migrations/240515_103045_create_users.safe.down.sql"

			fileNameBuilderMock.EXPECT().
				Up(expectedVersion, true).
				Return(expectedFileNameUp, true)

			fileNameBuilderMock.EXPECT().
				Down(expectedVersion, true).
				Return(expectedFileNameDown, true)

			fileMock.EXPECT().
				Exists("/migrations").
				Return(true, nil)

			// Up file created successfully
			fileMock.EXPECT().
				Create(expectedFileNameUp).
				Return(nil)

			// Creating down file fails
			fileMock.EXPECT().
				Create(expectedFileNameDown).
				Return(tt.expectedErr)

			create := NewCreate(
				&Options{
					Directory:   "/migrations",
					Interactive: false,
				},
				loggerMock,
				fileMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: "create_users"},
			}

			err := create.Handle(cmd)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

// TestCreate_Handle_DifferentDirectories_Successfully tests that Handle works correctly
// with various directory paths.
func TestCreate_Handle_DifferentDirectories_Successfully(t *testing.T) {
	tests := []struct {
		name      string
		directory string
	}{
		{
			name:      "absolute path",
			directory: "/var/lib/migrations",
		},
		{
			name:      "relative path",
			directory: "./migrations",
		},
		{
			name:      "nested path",
			directory: "/app/db/migrations/v1",
		},
		{
			name:      "path with spaces",
			directory: "/var/lib/my migrations",
		},
		{
			name:      "current directory",
			directory: ".",
		},
		{
			name:      "parent directory",
			directory: "..",
		},
		{
			name:      "home directory style",
			directory: "~/migrations",
		},
	}

	// Set a fixed time for predictable test output
	originalStdTime := timex.StdTime
	fixedTime := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
	timex.StdTime = timex.New(func() time.Time { return fixedTime })
	defer func() { timex.StdTime = originalStdTime }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			expectedVersion := "240515_103045_test_migration"
			expectedFileNameUp := tt.directory + "/240515_103045_test_migration.safe.up.sql"
			expectedFileNameDown := tt.directory + "/240515_103045_test_migration.safe.down.sql"

			fileNameBuilderMock.EXPECT().
				Up(expectedVersion, true).
				Return(expectedFileNameUp, true)

			fileNameBuilderMock.EXPECT().
				Down(expectedVersion, true).
				Return(expectedFileNameDown, true)

			fileMock.EXPECT().
				Exists(tt.directory).
				Return(true, nil)

			fileMock.EXPECT().
				Create(expectedFileNameUp).
				Return(nil)

			fileMock.EXPECT().
				Create(expectedFileNameDown).
				Return(nil)

			loggerMock.EXPECT().
				Success("New migration created successfully.").
				Return()

			create := NewCreate(
				&Options{
					Directory:   tt.directory,
					Interactive: false,
				},
				loggerMock,
				fileMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: "test_migration"},
			}

			err := create.Handle(cmd)

			require.NoError(t, err)
		})
	}
}

// TestCreate_Handle_FileNameBuilderReturnsUnsafeFile_Successfully tests that Handle
// works correctly when FileNameBuilder returns files without .safe prefix.
func TestCreate_Handle_FileNameBuilderReturnsUnsafeFile_Successfully(t *testing.T) {
	// Set a fixed time for predictable test output
	originalStdTime := timex.StdTime
	fixedTime := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
	timex.StdTime = timex.New(func() time.Time { return fixedTime })
	defer func() { timex.StdTime = originalStdTime }()

	fileMock := NewMockFile(t)
	loggerMock := NewMockLogger(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)

	expectedVersion := "240515_103045_create_users"
	// Files without .safe prefix
	expectedFileNameUp := "/migrations/240515_103045_create_users.up.sql"
	expectedFileNameDown := "/migrations/240515_103045_create_users.down.sql"

	fileNameBuilderMock.EXPECT().
		Up(expectedVersion, true).
		Return(expectedFileNameUp, false) // safely = false

	fileNameBuilderMock.EXPECT().
		Down(expectedVersion, true).
		Return(expectedFileNameDown, false) // safely = false

	fileMock.EXPECT().
		Exists("/migrations").
		Return(true, nil)

	fileMock.EXPECT().
		Create(expectedFileNameUp).
		Return(nil)

	fileMock.EXPECT().
		Create(expectedFileNameDown).
		Return(nil)

	loggerMock.EXPECT().
		Success("New migration created successfully.").
		Return()

	create := NewCreate(
		&Options{
			Directory:   "/migrations",
			Interactive: false,
		},
		loggerMock,
		fileMock,
		fileNameBuilderMock,
	)

	cmd := &Command{
		Args: &argsStub{present: true, first: "create_users"},
	}

	err := create.Handle(cmd)

	require.NoError(t, err)
}

// TestCreate_Handle_BoundaryMigrationNames_Successfully tests boundary cases for
// migration name validation.
func TestCreate_Handle_BoundaryMigrationNames_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		migrationName string
	}{
		{
			name:          "boundary: single character 'a'",
			migrationName: "a",
		},
		{
			name:          "boundary: single character 'Z'",
			migrationName: "Z",
		},
		{
			name:          "boundary: single digit '0'",
			migrationName: "0",
		},
		{
			name:          "boundary: single digit '9'",
			migrationName: "9",
		},
		{
			name:          "boundary: single underscore",
			migrationName: "_",
		},
		{
			name:          "boundary: two characters",
			migrationName: "ab",
		},
		{
			name:          "boundary: 100 characters",
			migrationName: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		},
	}

	// Set a fixed time for predictable test output
	originalStdTime := timex.StdTime
	fixedTime := time.Date(2024, 5, 15, 10, 30, 45, 0, time.UTC)
	timex.StdTime = timex.New(func() time.Time { return fixedTime })
	defer func() { timex.StdTime = originalStdTime }()

	expectedPrefix := "240515_103045"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			expectedVersion := expectedPrefix + "_" + tt.migrationName
			expectedFileNameUp := "/migrations/" + expectedVersion + ".safe.up.sql"
			expectedFileNameDown := "/migrations/" + expectedVersion + ".safe.down.sql"

			fileNameBuilderMock.EXPECT().
				Up(expectedVersion, true).
				Return(expectedFileNameUp, true)

			fileNameBuilderMock.EXPECT().
				Down(expectedVersion, true).
				Return(expectedFileNameDown, true)

			fileMock.EXPECT().
				Exists("/migrations").
				Return(true, nil)

			fileMock.EXPECT().
				Create(expectedFileNameUp).
				Return(nil)

			fileMock.EXPECT().
				Create(expectedFileNameDown).
				Return(nil)

			loggerMock.EXPECT().
				Success("New migration created successfully.").
				Return()

			create := NewCreate(
				&Options{
					Directory:   "/migrations",
					Interactive: false,
				},
				loggerMock,
				fileMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.migrationName},
			}

			err := create.Handle(cmd)

			require.NoError(t, err)
		})
	}
}

// TestCreate_Handle_VersionFormatting_Successfully tests that the version is formatted
// correctly with the timestamp prefix.
func TestCreate_Handle_VersionFormatting_Successfully(t *testing.T) {
	tests := []struct {
		name            string
		fixedTime       time.Time
		expectedPrefix  string
		migrationName   string
		expectedVersion string
	}{
		{
			name:            "morning time",
			fixedTime:       time.Date(2024, 1, 1, 8, 5, 3, 0, time.UTC),
			expectedPrefix:  "240101_080503",
			migrationName:   "init",
			expectedVersion: "240101_080503_init",
		},
		{
			name:            "afternoon time",
			fixedTime:       time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			expectedPrefix:  "241231_235959",
			migrationName:   "final",
			expectedVersion: "241231_235959_final",
		},
		{
			name:            "midnight",
			fixedTime:       time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			expectedPrefix:  "240615_000000",
			migrationName:   "midnight_migration",
			expectedVersion: "240615_000000_midnight_migration",
		},
		{
			name:            "year boundary 2025",
			fixedTime:       time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			expectedPrefix:  "250101_120000",
			migrationName:   "new_year",
			expectedVersion: "250101_120000_new_year",
		},
		{
			name:            "leap year date",
			fixedTime:       time.Date(2024, 2, 29, 14, 30, 0, 0, time.UTC),
			expectedPrefix:  "240229_143000",
			migrationName:   "leap_day",
			expectedVersion: "240229_143000_leap_day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the fixed time
			originalStdTime := timex.StdTime
			timex.StdTime = timex.New(func() time.Time { return tt.fixedTime })
			defer func() { timex.StdTime = originalStdTime }()

			fileMock := NewMockFile(t)
			loggerMock := NewMockLogger(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			expectedFileNameUp := "/migrations/" + tt.expectedVersion + ".safe.up.sql"
			expectedFileNameDown := "/migrations/" + tt.expectedVersion + ".safe.down.sql"

			fileNameBuilderMock.EXPECT().
				Up(tt.expectedVersion, true).
				Return(expectedFileNameUp, true)

			fileNameBuilderMock.EXPECT().
				Down(tt.expectedVersion, true).
				Return(expectedFileNameDown, true)

			fileMock.EXPECT().
				Exists("/migrations").
				Return(true, nil)

			fileMock.EXPECT().
				Create(expectedFileNameUp).
				Return(nil)

			fileMock.EXPECT().
				Create(expectedFileNameDown).
				Return(nil)

			loggerMock.EXPECT().
				Success("New migration created successfully.").
				Return()

			create := NewCreate(
				&Options{
					Directory:   "/migrations",
					Interactive: false,
				},
				loggerMock,
				fileMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.migrationName},
			}

			err := create.Handle(cmd)

			require.NoError(t, err)
		})
	}
}

// TestRegexpFileName_MatchesValidPatterns_Successfully tests that the regex pattern
// correctly identifies valid migration names.
func TestRegexpFileName_MatchesValidPatterns_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		migrationName string
		shouldMatch   bool
	}{
		// Valid patterns (should match)
		{name: "simple lowercase", migrationName: "abc", shouldMatch: true},
		{name: "simple uppercase", migrationName: "ABC", shouldMatch: true},
		{name: "mixed case", migrationName: "AbCdEf", shouldMatch: true},
		{name: "with digits", migrationName: "abc123", shouldMatch: true},
		{name: "only digits", migrationName: "123", shouldMatch: true},
		{name: "with underscores", migrationName: "a_b_c", shouldMatch: true},
		{name: "only underscores", migrationName: "___", shouldMatch: true},
		{name: "with backslash", migrationName: "a\\b", shouldMatch: true},
		{name: "complex valid", migrationName: "Create_User_Table_123\\test", shouldMatch: true},
		{name: "single char a", migrationName: "a", shouldMatch: true},
		{name: "single char Z", migrationName: "Z", shouldMatch: true},
		{name: "single char 0", migrationName: "0", shouldMatch: true},
		{name: "single char 9", migrationName: "9", shouldMatch: true},
		{name: "single underscore", migrationName: "_", shouldMatch: true},

		// Invalid patterns (should not match)
		{name: "empty string", migrationName: "", shouldMatch: false},
		{name: "with space", migrationName: "a b", shouldMatch: false},
		{name: "with hyphen", migrationName: "a-b", shouldMatch: false},
		{name: "with dot", migrationName: "a.b", shouldMatch: false},
		{name: "with at", migrationName: "a@b", shouldMatch: false},
		{name: "with hash", migrationName: "a#b", shouldMatch: false},
		{name: "with dollar", migrationName: "a$b", shouldMatch: false},
		{name: "with percent", migrationName: "a%b", shouldMatch: false},
		{name: "with asterisk", migrationName: "a*b", shouldMatch: false},
		{name: "with plus", migrationName: "a+b", shouldMatch: false},
		{name: "with forward slash", migrationName: "a/b", shouldMatch: false},
		{name: "leading space", migrationName: " abc", shouldMatch: false},
		{name: "trailing space", migrationName: "abc ", shouldMatch: false},
		{name: "only spaces", migrationName: "   ", shouldMatch: false},
		{name: "newline", migrationName: "a\nb", shouldMatch: false},
		{name: "tab", migrationName: "a\tb", shouldMatch: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := regexpFileName.MatchString(tt.migrationName)
			require.Equal(t, tt.shouldMatch, result, "regex match result for %q", tt.migrationName)
		})
	}
}

// TestErrInvalidFileName_ErrorMessage_Successfully verifies the error message content.
func TestErrInvalidFileName_ErrorMessage_Successfully(t *testing.T) {
	expectedMessage := "the migration name should contain letters, digits, underscore and/or backslash characters only"
	require.Equal(t, expectedMessage, ErrInvalidFileName.Error())
}

// Note: Tests for interactive mode (console.Confirm) are not included because
// console.Confirm is a package-level function that reads from stdin.
// To test interactive mode, the code would need to accept a Confirmer interface
// or use a different approach for user input.
//
// Scenarios that require interactive mode testing:
// - TestCreate_Handle_InteractiveMode_UserConfirms_Successfully
// - TestCreate_Handle_InteractiveMode_UserDeclines_Successfully

// Note: Tests for os.Mkdir in createDirectory are not included because os.Mkdir
// is a standard library function that cannot be easily mocked. The createDirectory
// method is tested indirectly through the directory existence check scenarios.
// To fully unit test directory creation, the code would need to accept a filesystem
// interface or use a different approach for file system operations.
//
// Scenarios that would require filesystem abstraction:
// - TestCreate_Handle_DirectoryNotExists_CreatesDirectory_Successfully
// - TestCreate_Handle_MkdirFails_ReturnsError_Failure
