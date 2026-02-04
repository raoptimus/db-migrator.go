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

func TestNewTo_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	to := NewTo(options, presenterMock, fileNameBuilderMock)

	require.NotNil(t, to)
	require.Equal(t, options, to.options)
	require.Equal(t, presenterMock, to.presenter)
	require.Equal(t, fileNameBuilderMock, to.fileNameBuilder)
}

func TestTo_Handle_NoArgument_Failure(t *testing.T) {
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	to := NewTo(
		&Options{},
		presenter,
		fileNameBuilder,
	)

	cmd := &Command{Args: &argsStub{present: false}}
	err := to.Handle(cmd, nil)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrTargetVersionRequired)
}

func TestTo_Handle_InvalidVersionFormat_Failure(t *testing.T) {
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	to := NewTo(
		&Options{},
		presenter,
		fileNameBuilder,
	)

	cmd := &Command{Args: &argsStub{present: true, first: "invalid_format"}}
	err := to.Handle(cmd, nil)

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid version format")
}

func TestTo_Handle_UpgradeDirection_Success(t *testing.T) {
	// Mock setup
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_185401"
	migrations := model.Migrations{
		{Version: "150101_120000"},
		{Version: "150101_185401"},
	}

	// Expectations
	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{}, nil).
		Once()

	svc.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil).
		Once()

	presenter.EXPECT().
		ShowUpgradePlan(mock.Anything, 2).
		Once()

	presenter.EXPECT().
		AskUpgradeConfirmation(2).
		Return("Apply?").
		Once()

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

	presenter.EXPECT().
		ShowUpgradeSuccess(2).
		Once()

	// Execute
	to := NewTo(&Options{Interactive: false}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.NoError(t, err)
}

func TestTo_Handle_UpgradeDirection_PartialMigrations_Success(t *testing.T) {
	// Mock setup
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_150000" // Target is between two migrations
	allMigrations := model.Migrations{
		{Version: "150101_120000"},
		{Version: "150101_150000"},
		{Version: "150101_185401"}, // This should NOT be applied
	}

	// Expectations
	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{}, nil).
		Once()

	svc.EXPECT().
		NewMigrations(mock.Anything).
		Return(allMigrations, nil).
		Once()

	presenter.EXPECT().
		ShowUpgradePlan(mock.Anything, 2).
		Once()

	presenter.EXPECT().
		AskUpgradeConfirmation(2).
		Return("Apply?").
		Once()

	fileNameBuilder.EXPECT().
		Up(mock.Anything, false).
		Return("/path/file.up.sql", true).
		Times(2)

	svc.EXPECT().
		ApplyFile(mock.Anything, mock.Anything, mock.Anything, true).
		Return(nil).
		Times(2)

	presenter.EXPECT().
		ShowUpgradeSuccess(2).
		Once()

	// Execute
	to := NewTo(&Options{Interactive: false}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.NoError(t, err)
}

func TestTo_Handle_UpgradeDirection_NoNewMigrations_Success(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_185401"

	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{}, nil).
		Once()

	svc.EXPECT().
		NewMigrations(mock.Anything).
		Return(model.Migrations{}, nil).
		Once()

	presenter.EXPECT().
		ShowNoNewMigrations().
		Once()

	to := NewTo(&Options{}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.NoError(t, err)
}

func TestTo_Handle_DowngradeDirection_Success(t *testing.T) {
	// Mock setup
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_120000"
	appliedMigrations := model.Migrations{
		{Version: "150101_185401"},
		{Version: "150101_140000"},
	}

	// Expectations
	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{
			{Version: targetVersion + "_test"},
		}, nil).
		Once()

	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(appliedMigrations, nil).
		Once()

	presenter.EXPECT().
		ShowDowngradePlan(mock.Anything).
		Once()

	presenter.EXPECT().
		AskDowngradeConfirmation(2).
		Return("Revert?").
		Once()

	fileNameBuilder.EXPECT().
		Down("150101_185401", false).
		Return("/path/file1.down.sql", true).
		Once()

	svc.EXPECT().
		RevertFile(mock.Anything, &appliedMigrations[0], "/path/file1.down.sql", true).
		Return(nil).
		Once()

	fileNameBuilder.EXPECT().
		Down("150101_140000", false).
		Return("/path/file2.down.sql", true).
		Once()

	svc.EXPECT().
		RevertFile(mock.Anything, &appliedMigrations[1], "/path/file2.down.sql", true).
		Return(nil).
		Once()

	presenter.EXPECT().
		ShowDowngradeSuccess(2).
		Once()

	// Execute
	to := NewTo(&Options{Interactive: false}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.NoError(t, err)
}

func TestTo_Handle_DowngradeDirection_AlreadyAtTarget_Success(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_185401"

	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{
			{Version: targetVersion + "_test"},
		}, nil).
		Once()

	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{}, nil).
		Once()

	presenter.EXPECT().
		ShowNoMigrationsToRevert().
		Once()

	to := NewTo(&Options{}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.NoError(t, err)
}

func TestTo_Handle_UpgradeDirection_WithError_Failure(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_185401"
	migrations := model.Migrations{
		{Version: "150101_120000"},
		{Version: "150101_185401"},
	}

	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{}, nil).
		Once()

	svc.EXPECT().
		NewMigrations(mock.Anything).
		Return(migrations, nil).
		Once()

	presenter.EXPECT().
		ShowUpgradePlan(mock.Anything, 2).
		Once()

	presenter.EXPECT().
		AskUpgradeConfirmation(2).
		Return("Apply?").
		Once()

	fileNameBuilder.EXPECT().
		Up("150101_120000", false).
		Return("/path/file1.up.sql", true).
		Once()

	applyErr := errors.New("migration failed")
	svc.EXPECT().
		ApplyFile(mock.Anything, &migrations[0], "/path/file1.up.sql", true).
		Return(applyErr).
		Once()

	presenter.EXPECT().
		ShowUpgradeError(0, 2).
		Once()

	to := NewTo(&Options{Interactive: false}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.Error(t, err)
	require.Equal(t, applyErr, err)
}

func TestTo_Handle_DowngradeDirection_WithError_Failure(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_120000"
	appliedMigrations := model.Migrations{
		{Version: "150101_185401"},
		{Version: "150101_140000"},
	}

	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{
			{Version: targetVersion + "_test"},
		}, nil).
		Once()

	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(appliedMigrations, nil).
		Once()

	presenter.EXPECT().
		ShowDowngradePlan(mock.Anything).
		Once()

	presenter.EXPECT().
		AskDowngradeConfirmation(2).
		Return("Revert?").
		Once()

	fileNameBuilder.EXPECT().
		Down("150101_185401", false).
		Return("/path/file1.down.sql", true).
		Once()

	revertErr := errors.New("revert failed")
	svc.EXPECT().
		RevertFile(mock.Anything, &appliedMigrations[0], "/path/file1.down.sql", true).
		Return(revertErr).
		Once()

	presenter.EXPECT().
		ShowDowngradeError(0, 2).
		Once()

	to := NewTo(&Options{Interactive: false}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.Error(t, err)
	require.Equal(t, revertErr, err)
}

func TestTo_Handle_ExistsError_Failure(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_185401"
	migrationsErr := errors.New("database error")

	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(nil, migrationsErr).
		Once()

	to := NewTo(&Options{}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.Error(t, err)
	require.Equal(t, migrationsErr, err)
}

func TestTo_Handle_NewMigrationsError_Failure(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_185401"
	newMigrErr := errors.New("cannot load migrations")

	// First Migrations call to check if target is applied
	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{}, nil).
		Once()

	// Second NewMigrations call in handleUpgrade
	svc.EXPECT().
		NewMigrations(mock.Anything).
		Return(nil, newMigrErr).
		Once()

	to := NewTo(&Options{}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.Error(t, err)
	require.Equal(t, newMigrErr, err)
}

func TestTo_Handle_MigrationsError_Failure(t *testing.T) {
	svc := NewMockMigrationService(t)
	presenter := NewMockPresenter(t)
	fileNameBuilder := NewMockFileNameBuilder(t)

	targetVersion := "150101_120000"
	migrErr := errors.New("cannot load applied migrations")

	// First Migrations call to check if target is applied - return applied migration
	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(model.Migrations{
			{Version: "150101_120000_test"},
		}, nil).
		Once()

	// Second Migrations call in handleDowngrade - return error
	svc.EXPECT().
		Migrations(mock.Anything, 0).
		Return(nil, migrErr).
		Once()

	to := NewTo(&Options{}, presenter, fileNameBuilder)
	cmd := &Command{Args: &argsStub{present: true, first: targetVersion}}

	err := to.Handle(cmd, svc)
	require.Error(t, err)
	require.Equal(t, migrErr, err)
}
