/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestNewRollback_InitializesFieldsCorrectly_Successfully verifies that NewRollback
// properly initializes all struct fields.
func TestNewRollback_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	rollback := NewRollback(options, presenterMock, fileNameBuilderMock)

	require.NotNil(t, rollback)
	require.Equal(t, options, rollback.options)
	require.Equal(t, presenterMock, rollback.presenter)
	require.Equal(t, fileNameBuilderMock, rollback.fileNameBuilder)
}

// TestRollback_Handle_NoMigrationsToRollback_Successfully tests that Handle shows
// "no migrations to rollback" message when there are no migrations in the latest batch.
func TestRollback_Handle_NoMigrationsToRollback_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	svcMock.EXPECT().
		LatestReleaseMigrations(mock.Anything).
		Return(nil, nil).
		Once()
	presenterMock.EXPECT().
		ShowNoMigrationsToRevert().
		Once()

	rollback := NewRollback(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := rollback.Handle(cmd, svcMock)

	require.NoError(t, err)
}

// TestRollback_Handle_LatestReleaseMigrationsError_Failure tests that Handle propagates
// errors from LatestReleaseMigrations service call.
func TestRollback_Handle_LatestReleaseMigrationsError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	expectedErr := errors.New("database connection failed")
	svcMock.EXPECT().
		LatestReleaseMigrations(mock.Anything).
		Return(nil, expectedErr).
		Once()

	rollback := NewRollback(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := rollback.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

// TestRollback_Handle_MissingDownFiles_Failure tests that Handle returns ErrMissingDownFiles
// when some migration down files are missing.
func TestRollback_Handle_MissingDownFiles_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 1609502400},
		{Version: "200102_120000", ApplyTime: 1609502400},
	}

	svcMock.EXPECT().
		LatestReleaseMigrations(mock.Anything).
		Return(migrations, nil).
		Once()

	// First migration has a down file, second does not
	fileNameBuilderMock.EXPECT().
		Down("200101_120000", false).
		Return("/migrations/200101_120000_test.down.sql", true).
		Once()
	svcMock.EXPECT().
		FileExists("/migrations/200101_120000_test.down.sql").
		Return(true, nil).
		Once()

	fileNameBuilderMock.EXPECT().
		Down("200102_120000", false).
		Return("/migrations/200102_120000_test.down.sql", false).
		Once()
	svcMock.EXPECT().
		FileExists("/migrations/200102_120000_test.down.sql").
		Return(false, nil).
		Once()

	presenterMock.EXPECT().
		ShowMissingDownFiles([]string{"200102_120000"}).
		Once()

	rollback := NewRollback(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := rollback.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, ErrMissingDownFiles, err)
}

// TestRollback_Handle_RevertSuccessfully_NonInteractive tests that Handle reverts
// all migrations atomically within a transaction.
func TestRollback_Handle_RevertSuccessfully_NonInteractive(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 1609502400},
		{Version: "200102_120000", ApplyTime: 1609502400},
	}

	svcMock.EXPECT().
		LatestReleaseMigrations(mock.Anything).
		Return(migrations, nil).
		Once()

	// Check down files exist
	fileNameBuilderMock.EXPECT().
		Down("200101_120000", false).
		Return("/migrations/200101_120000_test.down.sql", true).
		Times(2) // Once for check, once inside tx
	svcMock.EXPECT().
		FileExists("/migrations/200101_120000_test.down.sql").
		Return(true, nil).
		Once()

	fileNameBuilderMock.EXPECT().
		Down("200102_120000", false).
		Return("/migrations/200102_120000_test.down.sql", false).
		Times(2) // Once for check, once inside tx
	svcMock.EXPECT().
		FileExists("/migrations/200102_120000_test.down.sql").
		Return(true, nil).
		Once()

	presenterMock.EXPECT().
		ShowDowngradePlan(migrations).
		Once()
	presenterMock.EXPECT().
		AskDowngradeConfirmation(2).
		Return("Confirm?").
		Once()

	svcMock.EXPECT().
		RevertFile(mock.Anything, &migrations[0], "/migrations/200101_120000_test.down.sql", false).
		Return(nil).
		Once()
	svcMock.EXPECT().
		RevertFile(mock.Anything, &migrations[1], "/migrations/200102_120000_test.down.sql", false).
		Return(nil).
		Once()

	// ExecInTransaction should call the provided function
	svcMock.EXPECT().
		ExecInTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(ctx context.Context, fn func(context.Context) error) {
			_ = fn(ctx)
		}).
		Return(nil).
		Once()

	presenterMock.EXPECT().
		ShowDowngradeSuccess(2).
		Once()

	rollback := NewRollback(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := rollback.Handle(cmd, svcMock)

	require.NoError(t, err)
}

// TestRollback_Handle_TransactionError_Failure tests that Handle shows rollback error
// when the transaction fails.
func TestRollback_Handle_TransactionError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 1609502400},
	}

	svcMock.EXPECT().
		LatestReleaseMigrations(mock.Anything).
		Return(migrations, nil).
		Once()

	fileNameBuilderMock.EXPECT().
		Down("200101_120000", false).
		Return("/migrations/200101_120000_test.down.sql", true).
		Maybe()
	svcMock.EXPECT().
		FileExists("/migrations/200101_120000_test.down.sql").
		Return(true, nil).
		Once()

	presenterMock.EXPECT().
		ShowDowngradePlan(migrations).
		Once()
	presenterMock.EXPECT().
		AskDowngradeConfirmation(1).
		Return("Confirm?").
		Once()

	txErr := errors.New("transaction commit failed")
	svcMock.EXPECT().
		ExecInTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(txErr).
		Once()

	presenterMock.EXPECT().
		ShowRollbackError().
		Once()

	rollback := NewRollback(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := rollback.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, txErr, err)
}

// TestRollback_Handle_FileExistsError_Failure tests that Handle propagates
// errors from FileExists service call.
func TestRollback_Handle_FileExistsError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 1609502400},
	}

	svcMock.EXPECT().
		LatestReleaseMigrations(mock.Anything).
		Return(migrations, nil).
		Once()

	fileNameBuilderMock.EXPECT().
		Down("200101_120000", false).
		Return("/migrations/200101_120000_test.down.sql", true).
		Once()

	fileExistsErr := errors.New("filesystem error")
	svcMock.EXPECT().
		FileExists("/migrations/200101_120000_test.down.sql").
		Return(false, fileExistsErr).
		Once()

	rollback := NewRollback(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := rollback.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, fileExistsErr, err)
}
