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

// TestRedo_Handle_InvalidStepArgument_Failure tests that Handle returns an error
// when stepOrDefault fails to parse the step argument from command args.
func TestRedo_Handle_InvalidStepArgument_Failure(t *testing.T) {
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

			redo := NewRedo(
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

			err := redo.Handle(cmd, nil)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

// TestNewRedo_InitializesFieldsCorrectly_Successfully verifies that NewRedo
// properly initializes all struct fields.
func TestNewRedo_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	redo := NewRedo(options, presenterMock, fileNameBuilderMock)

	require.NotNil(t, redo)
	require.Equal(t, options, redo.options)
	require.Equal(t, presenterMock, redo.presenter)
	require.Equal(t, fileNameBuilderMock, redo.fileNameBuilder)
}

func TestRedo_Handle_NoMigrationsToRevert_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	svcMock.EXPECT().
		Migrations(mock.Anything, 1).
		Return(nil, nil).
		Once()
	presenterMock.EXPECT().
		ShowNoMigrationsToRevert().
		Once()

	redo := NewRedo(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := redo.Handle(cmd, svcMock)

	require.NoError(t, err)
}

func TestRedo_Handle_MigrationsServiceError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	expectedErr := errors.New("database connection failed")
	svcMock.EXPECT().
		Migrations(mock.Anything, 1).
		Return(nil, expectedErr).
		Once()

	redo := NewRedo(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := redo.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

func TestRedo_Handle_RedoSuccessfully_NonInteractive(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 1609502400},
	}

	svcMock.EXPECT().
		Migrations(mock.Anything, 1).
		Return(migrations, nil).
		Once()
	presenterMock.EXPECT().
		ShowRedoPlan(migrations).
		Once()
	presenterMock.EXPECT().
		AskRedoConfirmation(1).
		Return("Confirm?").
		Once()

	// Revert phase
	fileNameBuilderMock.EXPECT().
		Down("200101_120000", false).
		Return("/migrations/200101_120000_test.down.sql", true).
		Once()
	svcMock.EXPECT().
		RevertFile(mock.Anything, &migrations[0], "/migrations/200101_120000_test.down.sql", true).
		Return(nil).
		Once()

	// Apply phase
	fileNameBuilderMock.EXPECT().
		Up("200101_120000", false).
		Return("/migrations/200101_120000_test.up.sql", true).
		Once()
	svcMock.EXPECT().
		ApplyFile(mock.Anything, &migrations[0], "/migrations/200101_120000_test.up.sql", true).
		Return(nil).
		Once()

	presenterMock.EXPECT().
		ShowRedoSuccess(1).
		Once()

	redo := NewRedo(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := redo.Handle(cmd, svcMock)

	require.NoError(t, err)
}

func TestRedo_Handle_RevertFileError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 1609502400},
	}

	svcMock.EXPECT().
		Migrations(mock.Anything, 1).
		Return(migrations, nil).
		Once()
	presenterMock.EXPECT().
		ShowRedoPlan(migrations).
		Once()
	presenterMock.EXPECT().
		AskRedoConfirmation(1).
		Return("Confirm?").
		Once()

	fileNameBuilderMock.EXPECT().
		Down("200101_120000", false).
		Return("/migrations/200101_120000_test.down.sql", true).
		Once()

	revertErr := errors.New("failed to execute migration")
	svcMock.EXPECT().
		RevertFile(mock.Anything, &migrations[0], "/migrations/200101_120000_test.down.sql", true).
		Return(revertErr).
		Once()
	presenterMock.EXPECT().
		ShowRedoError().
		Once()

	redo := NewRedo(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := redo.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, revertErr, err)
}

func TestRedo_Handle_ApplyFileError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 1609502400},
	}

	svcMock.EXPECT().
		Migrations(mock.Anything, 1).
		Return(migrations, nil).
		Once()
	presenterMock.EXPECT().
		ShowRedoPlan(migrations).
		Once()
	presenterMock.EXPECT().
		AskRedoConfirmation(1).
		Return("Confirm?").
		Once()

	// Revert phase succeeds
	fileNameBuilderMock.EXPECT().
		Down("200101_120000", false).
		Return("/migrations/200101_120000_test.down.sql", true).
		Once()
	svcMock.EXPECT().
		RevertFile(mock.Anything, &migrations[0], "/migrations/200101_120000_test.down.sql", true).
		Return(nil).
		Once()

	// Apply phase fails
	fileNameBuilderMock.EXPECT().
		Up("200101_120000", false).
		Return("/migrations/200101_120000_test.up.sql", true).
		Once()

	applyErr := errors.New("failed to apply migration")
	svcMock.EXPECT().
		ApplyFile(mock.Anything, &migrations[0], "/migrations/200101_120000_test.up.sql", true).
		Return(applyErr).
		Once()
	presenterMock.EXPECT().
		ShowRedoError().
		Once()

	redo := NewRedo(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := redo.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, applyErr, err)
}
