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
	"errors"
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestHistory_Handle_InvalidStepArgument_Failure tests that Handle returns an error
// when stepOrDefault fails to parse the step argument from command args.
func TestHistory_Handle_InvalidStepArgument_Failure(t *testing.T) {
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

			history := NewHistory(
				&Options{
					DSN:       "postgres://user:pass@localhost:5432/testdb",
					Directory: "/migrations",
				},
				presenterMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			err := history.Handle(cmd, nil)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

// TestNewHistory_InitializesFieldsCorrectly_Successfully verifies that NewHistory
// properly initializes all struct fields.
func TestNewHistory_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	history := NewHistory(options, presenterMock)

	require.NotNil(t, history)
	require.Equal(t, options, history.options)
	require.NotNil(t, history.presenter)
}

// Sentinel errors for testing
var errHistoryMigrationsQueryFailed = errors.New("failed to query history migrations")

// TestHistory_Handle_MigrationsReturnsError_Failure tests that Handle returns an error
// when MigrationService.Migrations fails.
func TestHistory_Handle_MigrationsReturnsError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrationServiceMock.EXPECT().
		Migrations(mock.Anything, defaultGetHistoryLimit).
		Return(nil, errHistoryMigrationsQueryFailed)

	history := NewHistory(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false},
	}

	err := history.Handle(cmd, migrationServiceMock)

	require.Error(t, err)
	require.ErrorIs(t, err, errHistoryMigrationsQueryFailed)
}

// TestHistory_Handle_NoMigrationsFound_Successfully tests that Handle shows
// success message when no migrations are found.
func TestHistory_Handle_NoMigrationsFound_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrationServiceMock.EXPECT().
		Migrations(mock.Anything, defaultGetHistoryLimit).
		Return(model.Migrations{}, nil)

	presenterMock.EXPECT().
		ShowNoMigrationsToRevert().
		Return()

	history := NewHistory(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false},
	}

	err := history.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistory_Handle_SingleMigrationWithLimit_Successfully tests that Handle
// properly displays a single migration when limit > 0.
func TestHistory_Handle_SingleMigrationWithLimit_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200905_192800_create_users_table", ApplyTime: 1599326880},
	}

	migrationServiceMock.EXPECT().
		Migrations(mock.Anything, defaultGetHistoryLimit).
		Return(migrations, nil)

	presenterMock.EXPECT().
		ShowHistoryHeader(1).
		Return()

	presenterMock.EXPECT().
		PrintMigrations(migrations, true).
		Return()

	history := NewHistory(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false},
	}

	err := history.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistory_Handle_MultipleMigrationsWithLimit_Successfully tests that Handle
// properly displays multiple migrations when limit > 0.
func TestHistory_Handle_MultipleMigrationsWithLimit_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200905_192800_first_migration", ApplyTime: 1599326880},
		{Version: "200922_210000_second_migration", ApplyTime: 1600808400},
	}

	migrationServiceMock.EXPECT().
		Migrations(mock.Anything, 2).
		Return(migrations, nil)

	presenterMock.EXPECT().
		ShowHistoryHeader(2).
		Return()

	presenterMock.EXPECT().
		PrintMigrations(migrations, true).
		Return()

	history := NewHistory(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: true, first: "2"},
	}

	err := history.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistory_Handle_AllArgument_Successfully tests that Handle shows
// "Total N migrations have been applied" when "all" argument is used.
func TestHistory_Handle_AllArgument_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200905_192800_first_migration", ApplyTime: 1599326880},
		{Version: "200922_210000_second_migration", ApplyTime: 1600808400},
		{Version: "201015_143000_third_migration", ApplyTime: 1602770000},
	}

	// When "all" is passed, limit should be 0
	migrationServiceMock.EXPECT().
		Migrations(mock.Anything, 0).
		Return(migrations, nil)

	presenterMock.EXPECT().
		ShowAllHistoryHeader(3).
		Return()

	presenterMock.EXPECT().
		PrintMigrations(migrations, true).
		Return()

	history := NewHistory(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: true, first: "all"},
	}

	err := history.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistory_Handle_ContextPropagated_Successfully verifies that the context
// from the command is properly propagated to service calls.
func TestHistory_Handle_ContextPropagated_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	ctx := context.WithValue(context.Background(), struct{}{}, "test-value")

	migrations := model.Migrations{
		{Version: "200905_192800_migration", ApplyTime: 1599326880},
	}

	migrationServiceMock.EXPECT().
		Migrations(ctx, defaultGetHistoryLimit).
		Return(migrations, nil)

	presenterMock.EXPECT().
		ShowHistoryHeader(1).
		Return()

	presenterMock.EXPECT().
		PrintMigrations(migrations, true).
		Return()

	history := NewHistory(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := (&Command{
		Args: &argsStub{present: false},
	}).WithContext(ctx)

	err := history.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistory_Handle_BoundaryLimitValues_Successfully tests boundary values
// for the limit argument.
func TestHistory_Handle_BoundaryLimitValues_Successfully(t *testing.T) {
	tests := []struct {
		name          string
		argValue      string
		expectedLimit int
	}{
		{
			name:          "minimum valid limit (1)",
			argValue:      "1",
			expectedLimit: 1,
		},
		{
			name:          "default limit boundary (10)",
			argValue:      "10",
			expectedLimit: 10,
		},
		{
			name:          "large limit (999)",
			argValue:      "999",
			expectedLimit: 999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenterMock := NewMockPresenter(t)
			migrationServiceMock := NewMockMigrationService(t)

			migrationServiceMock.EXPECT().
				Migrations(mock.Anything, tt.expectedLimit).
				Return(model.Migrations{}, nil)

			presenterMock.EXPECT().
				ShowNoMigrationsToRevert().
				Return()

			history := NewHistory(
				&Options{Interactive: false},
				presenterMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			err := history.Handle(cmd, migrationServiceMock)

			require.NoError(t, err)
		})
	}
}
