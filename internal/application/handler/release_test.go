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

// TestNewRelease_InitializesFieldsCorrectly_Successfully verifies that NewRelease
// properly initializes all struct fields.
func TestNewRelease_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	release := NewRelease(options, presenterMock, fileNameBuilderMock)

	require.NotNil(t, release)
	require.Equal(t, options, release.options)
	require.Equal(t, presenterMock, release.presenter)
	require.Equal(t, fileNameBuilderMock, release.fileNameBuilder)
}

// TestRelease_Handle_NoNewMigrations_Successfully tests that Handle shows
// "no new migrations" message when there are no pending migrations.
func TestRelease_Handle_NoNewMigrations_Successfully(t *testing.T) {
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

	release := NewRelease(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := release.Handle(cmd, svcMock)

	require.NoError(t, err)
}

// TestRelease_Handle_NewMigrationsServiceError_Failure tests that Handle propagates
// errors from NewMigrations service call.
func TestRelease_Handle_NewMigrationsServiceError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	expectedErr := errors.New("database connection failed")
	svcMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(nil, expectedErr).
		Once()

	release := NewRelease(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := release.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, expectedErr, err)
}

// TestRelease_Handle_ApplySuccessfully_NonInteractive tests that Handle applies
// all migrations atomically within a transaction with a shared applyTime.
func TestRelease_Handle_ApplySuccessfully_NonInteractive(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200101_120000", ApplyTime: 0},
		{Version: "200102_120000", ApplyTime: 0},
	}

	svcMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil).
		Once()
	presenterMock.EXPECT().
		ShowReleasePlan(migrations).
		Once()
	presenterMock.EXPECT().
		AskReleaseConfirmation(2).
		Return("Confirm?").
		Once()

	fileNameBuilderMock.EXPECT().
		Up("200101_120000", false).
		Return("/migrations/200101_120000_test.up.sql", true).
		Once()
	fileNameBuilderMock.EXPECT().
		Up("200102_120000", false).
		Return("/migrations/200102_120000_test.up.sql", true).
		Once()

	// Capture applyTime to verify both migrations use the same value
	var capturedApplyTimes []int64
	svcMock.EXPECT().
		ApplyFileWithApplyTime(mock.Anything, &migrations[0], "/migrations/200101_120000_test.up.sql", mock.AnythingOfType("int64")).
		Run(func(_ context.Context, _ *model.Migration, _ string, applyTime int64) {
			capturedApplyTimes = append(capturedApplyTimes, applyTime)
		}).
		Return(nil).
		Once()
	svcMock.EXPECT().
		ApplyFileWithApplyTime(mock.Anything, &migrations[1], "/migrations/200102_120000_test.up.sql", mock.AnythingOfType("int64")).
		Run(func(_ context.Context, _ *model.Migration, _ string, applyTime int64) {
			capturedApplyTimes = append(capturedApplyTimes, applyTime)
		}).
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
		ShowReleaseSuccess(2).
		Once()

	release := NewRelease(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := release.Handle(cmd, svcMock)

	require.NoError(t, err)
	require.Len(t, capturedApplyTimes, 2)
	require.Equal(t, capturedApplyTimes[0], capturedApplyTimes[1], "both migrations must share the same applyTime")
}

// TestRelease_Handle_TransactionError_Failure tests that Handle shows release error
// when the transaction fails.
func TestRelease_Handle_TransactionError_Failure(t *testing.T) {
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
		ShowReleasePlan(migrations).
		Once()
	presenterMock.EXPECT().
		AskReleaseConfirmation(1).
		Return("Confirm?").
		Once()

	txErr := errors.New("transaction commit failed")
	svcMock.EXPECT().
		ExecInTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(txErr).
		Once()

	presenterMock.EXPECT().
		ShowReleaseError().
		Once()

	release := NewRelease(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := release.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, txErr, err)
}

// TestRelease_Handle_ApplyFileErrorInsideTx_Failure tests that Handle propagates
// errors from ApplyFileWithApplyTime when called inside a transaction.
func TestRelease_Handle_ApplyFileErrorInsideTx_Failure(t *testing.T) {
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
		ShowReleasePlan(migrations).
		Once()
	presenterMock.EXPECT().
		AskReleaseConfirmation(1).
		Return("Confirm?").
		Once()

	applyErr := errors.New("failed to execute migration")
	fileNameBuilderMock.EXPECT().
		Up("200101_120000", false).
		Return("/migrations/200101_120000_test.up.sql", true).
		Once()
	svcMock.EXPECT().
		ApplyFileWithApplyTime(mock.Anything, &migrations[0], "/migrations/200101_120000_test.up.sql", mock.AnythingOfType("int64")).
		Return(applyErr).
		Once()

	// ExecInTransaction calls the function and returns its error
	svcMock.EXPECT().
		ExecInTransaction(mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Run(func(ctx context.Context, fn func(context.Context) error) {
			_ = fn(ctx)
		}).
		Return(applyErr).
		Once()

	presenterMock.EXPECT().
		ShowReleaseError().
		Once()

	release := NewRelease(
		&Options{Interactive: false},
		presenterMock,
		fileNameBuilderMock,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := release.Handle(cmd, svcMock)

	require.Error(t, err)
	require.Equal(t, applyErr, err)
}
