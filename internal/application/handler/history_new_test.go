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

// TestNewHistoryNew_InitializesFieldsCorrectly_Successfully verifies that NewHistoryNew
// properly initializes all struct fields.
func TestNewHistoryNew_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	historyNew := NewHistoryNew(options, presenterMock)

	require.NotNil(t, historyNew)
	require.Equal(t, options, historyNew.options)
	require.NotNil(t, historyNew.presenter)
}

// TestHistoryNew_Handle_InvalidStepArgument_Failure tests that Handle returns an error
// when stepOrDefault fails to parse the step argument from command args.
func TestHistoryNew_Handle_InvalidStepArgument_Failure(t *testing.T) {
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

			historyNew := NewHistoryNew(
				&Options{
					DSN:       "postgres://user:pass@localhost:5432/testdb",
					Directory: "/migrations",
				},
				presenterMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			err := historyNew.Handle(cmd, nil)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

// Sentinel errors for testing
var errHistoryNewMigrationsQueryFailed = errors.New("failed to query new migrations")

// TestHistoryNew_Handle_NewMigrationsReturnsError_Failure tests that Handle returns an error
// when MigrationService.NewMigrations fails.
func TestHistoryNew_Handle_NewMigrationsReturnsError_Failure(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrationServiceMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(nil, errHistoryNewMigrationsQueryFailed)

	historyNew := NewHistoryNew(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false},
	}

	err := historyNew.Handle(cmd, migrationServiceMock)

	require.Error(t, err)
	require.ErrorIs(t, err, errHistoryNewMigrationsQueryFailed)
}

// TestHistoryNew_Handle_NoNewMigrationsFound_Successfully tests that Handle shows
// success message when no new migrations are found.
func TestHistoryNew_Handle_NoNewMigrationsFound_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrationServiceMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(model.Migrations{}, nil)

	presenterMock.EXPECT().
		ShowNoNewMigrations().
		Return()

	historyNew := NewHistoryNew(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false},
	}

	err := historyNew.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistoryNew_Handle_SingleNewMigration_Successfully tests that Handle
// properly displays a single new migration.
func TestHistoryNew_Handle_SingleNewMigration_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200905_192800_create_users_table"},
	}

	migrationServiceMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil)

	presenterMock.EXPECT().
		ShowNewMigrationsHeader(1).
		Return()

	presenterMock.EXPECT().
		PrintMigrations(migrations, false).
		Return()

	historyNew := NewHistoryNew(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false},
	}

	err := historyNew.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistoryNew_Handle_MultipleMigrationsWithinLimit_Successfully tests that Handle
// properly displays multiple new migrations when count is within limit.
func TestHistoryNew_Handle_MultipleMigrationsWithinLimit_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200905_192800_first_migration"},
		{Version: "200922_210000_second_migration"},
	}

	migrationServiceMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil)

	presenterMock.EXPECT().
		ShowNewMigrationsHeader(2).
		Return()

	presenterMock.EXPECT().
		PrintMigrations(migrations, false).
		Return()

	historyNew := NewHistoryNew(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false},
	}

	err := historyNew.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistoryNew_Handle_MigrationsExceedLimit_Successfully tests that Handle
// shows limited migrations when count exceeds the limit.
func TestHistoryNew_Handle_MigrationsExceedLimit_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200905_192800_first_migration"},
		{Version: "200922_210000_second_migration"},
		{Version: "201015_143000_third_migration"},
	}

	migrationServiceMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil)

	// Request limit of 2, but 3 migrations exist
	presenterMock.EXPECT().
		ShowNewMigrationsLimitedHeader(2, 3).
		Return()

	// Only first 2 migrations should be printed
	limitedMigrations := migrations[:2]
	presenterMock.EXPECT().
		PrintMigrations(limitedMigrations, false).
		Return()

	historyNew := NewHistoryNew(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: true, first: "2"},
	}

	err := historyNew.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistoryNew_Handle_AllArgument_Successfully tests that Handle shows
// all new migrations when "all" argument is used.
func TestHistoryNew_Handle_AllArgument_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	migrations := model.Migrations{
		{Version: "200905_192800_first_migration"},
		{Version: "200922_210000_second_migration"},
		{Version: "201015_143000_third_migration"},
	}

	migrationServiceMock.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil)

	// With "all" argument, should show all migrations without limit
	presenterMock.EXPECT().
		ShowNewMigrationsHeader(3).
		Return()

	presenterMock.EXPECT().
		PrintMigrations(migrations, false).
		Return()

	historyNew := NewHistoryNew(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := &Command{
		Args: &argsStub{present: true, first: "all"},
	}

	err := historyNew.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistoryNew_Handle_ContextPropagated_Successfully verifies that the context
// from the command is properly propagated to service calls.
func TestHistoryNew_Handle_ContextPropagated_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	migrationServiceMock := NewMockMigrationService(t)

	ctx := context.WithValue(context.Background(), struct{}{}, "test-value")

	migrations := model.Migrations{
		{Version: "200905_192800_migration"},
	}

	migrationServiceMock.EXPECT().
		NewMigrations(ctx).
		Return(migrations, nil)

	presenterMock.EXPECT().
		ShowNewMigrationsHeader(1).
		Return()

	presenterMock.EXPECT().
		PrintMigrations(migrations, false).
		Return()

	historyNew := NewHistoryNew(
		&Options{Interactive: false},
		presenterMock,
	)

	cmd := (&Command{
		Args: &argsStub{present: false},
	}).WithContext(ctx)

	err := historyNew.Handle(cmd, migrationServiceMock)

	require.NoError(t, err)
}

// TestHistoryNew_Handle_BoundaryLimitValues_Successfully tests boundary values
// for the limit argument.
func TestHistoryNew_Handle_BoundaryLimitValues_Successfully(t *testing.T) {
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
				NewMigrations(mock.Anything).
				Return(model.Migrations{}, nil)

			presenterMock.EXPECT().
				ShowNoNewMigrations().
				Return()

			historyNew := NewHistoryNew(
				&Options{Interactive: false},
				presenterMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			err := historyNew.Handle(cmd, migrationServiceMock)

			require.NoError(t, err)
		})
	}
}
