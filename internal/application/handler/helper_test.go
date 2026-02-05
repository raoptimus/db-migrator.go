/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// argsStub is a simple stub implementation of cli.Args for testing purposes.
type argsStub struct {
	present bool
	first   string
}

func (a *argsStub) Get(_ int) string { return "" }
func (a *argsStub) First() string    { return a.first }
func (a *argsStub) Tail() []string   { return nil }
func (a *argsStub) Len() int         { return 0 }
func (a *argsStub) Present() bool    { return a.present }
func (a *argsStub) Slice() []string  { return nil }

func TestStepOrDefault_ArgsNotPresent_ReturnsDefault_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		defaults     int
		expectedStep int
	}{
		{
			name:         "default value is 1",
			defaults:     1,
			expectedStep: 1,
		},
		{
			name:         "default value is 10",
			defaults:     10,
			expectedStep: 10,
		},
		{
			name:         "default value is 0",
			defaults:     0,
			expectedStep: 0,
		},
		{
			name:         "default value is 100",
			defaults:     100,
			expectedStep: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: false, first: ""},
			}

			step, err := stepOrDefault(cmd, tt.defaults)

			require.NoError(t, err)
			require.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestStepOrDefault_ArgsIsEmptyString_ReturnsDefault_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		defaults     int
		expectedStep int
	}{
		{
			name:         "default value is 1",
			defaults:     1,
			expectedStep: 1,
		},
		{
			name:         "default value is 5",
			defaults:     5,
			expectedStep: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: ""},
			}

			step, err := stepOrDefault(cmd, tt.defaults)

			require.NoError(t, err)
			require.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestStepOrDefault_ArgsIsAll_ReturnsZero_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		defaults     int
		expectedStep int
	}{
		{
			name:         "default value is 1",
			defaults:     1,
			expectedStep: 0,
		},
		{
			name:         "default value is 100",
			defaults:     100,
			expectedStep: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: "all"},
			}

			step, err := stepOrDefault(cmd, tt.defaults)

			require.NoError(t, err)
			require.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestStepOrDefault_ValidPositiveInteger_ReturnsStep_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		argValue     string
		defaults     int
		expectedStep int
	}{
		{
			name:         "boundary value 1",
			argValue:     "1",
			defaults:     10,
			expectedStep: 1,
		},
		{
			name:         "value 2",
			argValue:     "2",
			defaults:     10,
			expectedStep: 2,
		},
		{
			name:         "value 5",
			argValue:     "5",
			defaults:     1,
			expectedStep: 5,
		},
		{
			name:         "value 10",
			argValue:     "10",
			defaults:     1,
			expectedStep: 10,
		},
		{
			name:         "value 100",
			argValue:     "100",
			defaults:     5,
			expectedStep: 100,
		},
		{
			name:         "large value 999999",
			argValue:     "999999",
			defaults:     1,
			expectedStep: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			step, err := stepOrDefault(cmd, tt.defaults)

			require.NoError(t, err)
			require.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestStepOrDefault_InvalidNonNumericArg_ReturnsError_Failure(t *testing.T) {
	tests := []struct {
		name              string
		argValue          string
		expectedErrSubstr string
	}{
		{
			name:              "alphabetic string",
			argValue:          "abc",
			expectedErrSubstr: "the step argument abc is not valid",
		},
		{
			name:              "special characters",
			argValue:          "@#$",
			expectedErrSubstr: "the step argument @#$ is not valid",
		},
		{
			name:              "mixed alphanumeric",
			argValue:          "12abc",
			expectedErrSubstr: "the step argument 12abc is not valid",
		},
		{
			name:              "float value",
			argValue:          "1.5",
			expectedErrSubstr: "the step argument 1.5 is not valid",
		},
		{
			name:              "space in value",
			argValue:          "1 2",
			expectedErrSubstr: "the step argument 1 2 is not valid",
		},
		{
			name:              "hexadecimal format",
			argValue:          "0x10",
			expectedErrSubstr: "the step argument 0x10 is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			step, err := stepOrDefault(cmd, 1)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
			require.Equal(t, -1, step)
		})
	}
}

func TestStepOrDefault_ZeroOrNegativeArg_ReturnsError_Failure(t *testing.T) {
	tests := []struct {
		name        string
		argValue    string
		expectedErr error
	}{
		{
			name:        "boundary value 0",
			argValue:    "0",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
		{
			name:        "negative value -1",
			argValue:    "-1",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
		{
			name:        "negative value -5",
			argValue:    "-5",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
		{
			name:        "large negative value -100",
			argValue:    "-100",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
		{
			name:        "negative value -999999",
			argValue:    "-999999",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			step, err := stepOrDefault(cmd, 1)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Equal(t, -1, step)
		})
	}
}

func TestParseTargetVersion_Timestamp_Success(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedVersion string
	}{
		{
			name:            "valid timestamp format",
			input:           "150101_185401",
			expectedVersion: "150101_185401",
		},
		{
			name:            "another valid timestamp",
			input:           "200905_192800",
			expectedVersion: "200905_192800",
		},
		{
			name:            "minimum timestamp",
			input:           "000000_000000",
			expectedVersion: "000000_000000",
		},
		{
			name:            "maximum timestamp",
			input:           "991231_235959",
			expectedVersion: "991231_235959",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseTargetVersion(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestParseTargetVersion_FullName_Success(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedVersion string
	}{
		{
			name:            "full migration name",
			input:           "150101_185401_create_news_table",
			expectedVersion: "150101_185401",
		},
		{
			name:            "full name with multiple underscores",
			input:           "200905_192800_add_column_to_users_table",
			expectedVersion: "200905_192800",
		},
		{
			name:            "full name with short description",
			input:           "140220_014658_init",
			expectedVersion: "140220_014658",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseTargetVersion(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestParseTargetVersion_DateTime_Success(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedVersion string
	}{
		{
			name:            "datetime format",
			input:           "2015-01-01 18:54:01",
			expectedVersion: "150101_185401",
		},
		{
			name:            "datetime midnight",
			input:           "2020-09-05 00:00:00",
			expectedVersion: "200905_000000",
		},
		{
			name:            "datetime end of day",
			input:           "2020-09-05 23:59:59",
			expectedVersion: "200905_235959",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseTargetVersion(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestParseTargetVersion_UnixTimestamp_Success(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedVersion string
	}{
		{
			name:            "unix timestamp 10 digits",
			input:           "1392853618",
			expectedVersion: "140219_234658", // UTC time: 2014-02-19 23:46:58
		},
		{
			name:            "unix timestamp year 2015",
			input:           "1420135441",
			expectedVersion: "150101_180401", // UTC time: 2015-01-01 18:04:01
		},
		{
			name:            "unix timestamp 9 digits (1999)",
			input:           "946684800",
			expectedVersion: "000101_000000", // UTC time: 2000-01-01 00:00:00
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseTargetVersion(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expectedVersion, version)
		})
	}
}

func TestParseTargetVersion_EmptyString_Failure(t *testing.T) {
	_, err := parseTargetVersion("")
	require.Error(t, err)
	require.ErrorIs(t, err, ErrTargetVersionRequired)
}

func TestParseTargetVersion_InvalidFormat_Failure(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedErrSubstr string
	}{
		{
			name:              "invalid format",
			input:             "invalid_format",
			expectedErrSubstr: "invalid version format",
		},
		{
			name:              "partial timestamp",
			input:             "150101",
			expectedErrSubstr: "invalid version format",
		},
		{
			name:              "wrong separator",
			input:             "150101-185401",
			expectedErrSubstr: "invalid version format",
		},
		{
			name:              "alphabetic characters",
			input:             "abcdef_ghijkl",
			expectedErrSubstr: "invalid version format",
		},
		{
			name:              "too short for unix timestamp",
			input:             "150101",
			expectedErrSubstr: "invalid version format",
		},
		{
			name:              "too long for unix timestamp",
			input:             "14208538166",
			expectedErrSubstr: "invalid version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTargetVersion(tt.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

func TestApplyMigrations_Success(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	migrations := model.Migrations{
		{Version: "150101_120000"},
		{Version: "150101_185401"},
	}

	fileNameBuilder.EXPECT().
		Up("150101_120000", false).
		Return("/path/file1.up.sql", true).
		Once()

	svc.EXPECT().
		ApplyFile(mock.Anything, &migrations[0], "/path/file1.up.sql", true).
		Return(nil).
		Once()

	fileNameBuilder.EXPECT().
		Up("150101_185401", false).
		Return("/path/file2.up.sql", true).
		Once()

	svc.EXPECT().
		ApplyFile(mock.Anything, &migrations[1], "/path/file2.up.sql", true).
		Return(nil).
		Once()

	cmd := &Command{Args: &argsStub{}}
	applied, err := applyMigrations(cmd, svc, presenter, fileNameBuilder, migrations)

	require.NoError(t, err)
	require.Equal(t, 2, applied)
}

func TestApplyMigrations_PartialFailure(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	migrations := model.Migrations{
		{Version: "150101_120000"},
		{Version: "150101_185401"},
	}

	fileNameBuilder.EXPECT().
		Up("150101_120000", false).
		Return("/path/file1.up.sql", true).
		Once()

	svc.EXPECT().
		ApplyFile(mock.Anything, &migrations[0], "/path/file1.up.sql", true).
		Return(nil).
		Once()

	fileNameBuilder.EXPECT().
		Up("150101_185401", false).
		Return("/path/file2.up.sql", true).
		Once()

	applyErr := errors.New("apply failed")
	svc.EXPECT().
		ApplyFile(mock.Anything, &migrations[1], "/path/file2.up.sql", true).
		Return(applyErr).
		Once()

	presenter.EXPECT().
		ShowUpgradeError(1, 2).
		Once()

	cmd := &Command{Args: &argsStub{}}
	applied, err := applyMigrations(cmd, svc, presenter, fileNameBuilder, migrations)

	require.Error(t, err)
	require.Equal(t, applyErr, err)
	require.Equal(t, 1, applied)
}

func TestApplyMigrations_EmptyList_Success(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	migrations := model.Migrations{}

	cmd := &Command{Args: &argsStub{}}
	applied, err := applyMigrations(cmd, svc, presenter, fileNameBuilder, migrations)

	require.NoError(t, err)
	require.Equal(t, 0, applied)
}

func TestRevertMigrations_Success(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	migrations := model.Migrations{
		{Version: "150101_185401"},
		{Version: "150101_120000"},
	}

	fileNameBuilder.EXPECT().
		Down("150101_185401", false).
		Return("/path/file1.down.sql", true).
		Once()

	svc.EXPECT().
		RevertFile(mock.Anything, &migrations[0], "/path/file1.down.sql", true).
		Return(nil).
		Once()

	fileNameBuilder.EXPECT().
		Down("150101_120000", false).
		Return("/path/file2.down.sql", true).
		Once()

	svc.EXPECT().
		RevertFile(mock.Anything, &migrations[1], "/path/file2.down.sql", true).
		Return(nil).
		Once()

	cmd := &Command{Args: &argsStub{}}
	reverted, err := revertMigrations(cmd, svc, presenter, fileNameBuilder, migrations)

	require.NoError(t, err)
	require.Equal(t, 2, reverted)
}

func TestRevertMigrations_PartialFailure(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	migrations := model.Migrations{
		{Version: "150101_185401"},
		{Version: "150101_120000"},
	}

	fileNameBuilder.EXPECT().
		Down("150101_185401", false).
		Return("/path/file1.down.sql", true).
		Once()

	svc.EXPECT().
		RevertFile(mock.Anything, &migrations[0], "/path/file1.down.sql", true).
		Return(nil).
		Once()

	fileNameBuilder.EXPECT().
		Down("150101_120000", false).
		Return("/path/file2.down.sql", true).
		Once()

	revertErr := errors.New("revert failed")
	svc.EXPECT().
		RevertFile(mock.Anything, &migrations[1], "/path/file2.down.sql", true).
		Return(revertErr).
		Once()

	presenter.EXPECT().
		ShowDowngradeError(1, 2).
		Once()

	cmd := &Command{Args: &argsStub{}}
	reverted, err := revertMigrations(cmd, svc, presenter, fileNameBuilder, migrations)

	require.Error(t, err)
	require.Equal(t, revertErr, err)
	require.Equal(t, 1, reverted)
}

func TestRevertMigrations_EmptyList_Success(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	migrations := model.Migrations{}

	cmd := &Command{Args: &argsStub{}}
	reverted, err := revertMigrations(cmd, svc, presenter, fileNameBuilder, migrations)

	require.NoError(t, err)
	require.Equal(t, 0, reverted)
}
