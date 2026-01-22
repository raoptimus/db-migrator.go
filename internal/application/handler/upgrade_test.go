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

// TestUpgrade_Handle_InvalidStepArgument_Failure tests that Handle returns an error
// when stepOrDefault fails to parse the step argument from command args.
func TestUpgrade_Handle_InvalidStepArgument_Failure(t *testing.T) {
	tests := []struct {
		name              string
		argValue          string
		expectedErrSubstr string
	}{
		{
			name:              "non-numeric argument",
			argValue:          "abc",
			expectedErrSubstr: "the step argument abc is not valid",
		},
		{
			name:              "float argument",
			argValue:          "1.5",
			expectedErrSubstr: "the step argument 1.5 is not valid",
		},
		{
			name:              "negative argument",
			argValue:          "-5",
			expectedErrSubstr: "the step argument must be greater than 0",
		},
		{
			name:              "zero argument",
			argValue:          "0",
			expectedErrSubstr: "the step argument must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenterMock := NewMockPresenter(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			upgrade := NewUpgrade(
				&Options{
					DSN:       "postgres://user:pass@localhost:5432/testdb",
					Directory: "/migrations",
				},
				presenterMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			err := upgrade.Handle(cmd, nil)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

// TestNewUpgrade_InitializesFieldsCorrectly_Successfully verifies that NewUpgrade
// properly initializes all struct fields.
func TestNewUpgrade_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	upgrade := NewUpgrade(options, presenterMock, fileNameBuilderMock)

	require.NotNil(t, upgrade)
	require.Equal(t, options, upgrade.options)
	require.Equal(t, presenterMock, upgrade.presenter)
	require.Equal(t, fileNameBuilderMock, upgrade.fileNameBuilder)
}

func TestUpgrade_Handle_NoNewMigrations_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	svcMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(nil, nil).
		Once()
	presenterMock.EXPECT().
		ShowNoNewMigrations().
		Once()

	upgrade := NewUpgrade(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := upgrade.Handle(cmd, svcMock)

	require.NoError(t, err)
}

func TestUpgrade_Handle_NewMigrationsServiceError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	expectedErr := errors.New("database connection failed")
	svcMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(nil, expectedErr).
		Once()

	upgrade := NewUpgrade(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := upgrade.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func TestUpgrade_Handle_ApplySuccessfully_NonInteractive(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 0},
	}

	svcMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil).
		Once()
	presenterMock.EXPECT().
		ShowUpgradePlan(migrations, 1).
		Once()
	presenterMock.EXPECT().
		AskUpgradeConfirmation(1).
		Return("Confirm?").
		Once()
	fileNameBuilderMock.EXPECT().
		Up("200101_120000", false).
		Return("/migrations/200101_120000_test.up.sql", true).
		Once()
	svcMock.EXPECT().
		ApplyFile(mock.Anything, &migrations[0], "/migrations/200101_120000_test.up.sql", true).
		Return(nil).
		Once()
	presenterMock.EXPECT().
		ShowUpgradeSuccess(1).
		Once()

	upgrade := NewUpgrade(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := upgrade.Handle(cmd, svcMock)

	require.NoError(t, err)
}

func TestUpgrade_Handle_ApplyFileError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 0},
	}

	svcMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil).
		Once()
	presenterMock.EXPECT().
		ShowUpgradePlan(migrations, 1).
		Once()
	presenterMock.EXPECT().
		AskUpgradeConfirmation(1).
		Return("Confirm?").
		Once()
	fileNameBuilderMock.EXPECT().
		Up("200101_120000", false).
		Return("/migrations/200101_120000_test.up.sql", true).
		Once()

	applyErr := errors.New("failed to execute migration")
	svcMock.EXPECT().
		ApplyFile(mock.Anything, &migrations[0], "/migrations/200101_120000_test.up.sql", true).
		Return(applyErr).
		Once()
	presenterMock.EXPECT().
		ShowUpgradeError(0, 1).
		Once()

	upgrade := NewUpgrade(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := upgrade.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, applyErr, err)
}

func TestUpgrade_Handle_WithLimit_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	allMigrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 0},
		{Version: "200102_120000", ApplyTime: 0},
		{Version: "200103_120000", ApplyTime: 0},
	}

	limitedMigrations := allMigrations[:2]

	svcMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(allMigrations, nil).
		Once()
	presenterMock.EXPECT().
		ShowUpgradePlan(limitedMigrations, 3).
		Once()
	presenterMock.EXPECT().
		AskUpgradeConfirmation(2).
		Return("Confirm?").
		Once()
	fileNameBuilderMock.EXPECT().
		Up("200101_120000", false).
		Return("/migrations/200101_120000_test.up.sql", true).
		Once()
	fileNameBuilderMock.EXPECT().
		Up("200102_120000", false).
		Return("/migrations/200102_120000_test.up.sql", false).
		Once()
	svcMock.EXPECT().
		ApplyFile(mock.Anything, &limitedMigrations[0], "/migrations/200101_120000_test.up.sql", true).
		Return(nil).
		Once()
	svcMock.EXPECT().
		ApplyFile(mock.Anything, &limitedMigrations[1], "/migrations/200102_120000_test.up.sql", false).
		Return(nil).
		Once()
	presenterMock.EXPECT().
		ShowUpgradeSuccess(2).
		Once()

	upgrade := NewUpgrade(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: true, first: "2"}}
	err := upgrade.Handle(cmd, svcMock)

	require.NoError(t, err)
}
