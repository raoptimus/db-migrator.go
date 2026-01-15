/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package presenter

import (
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewMigrationPresenter(t *testing.T) {
	logger := NewMockLogger(t)
	presenter := NewMigrationPresenter(logger)

	assert.NotNil(t, presenter)
	assert.Equal(t, logger, presenter.logger)
}

func TestMigrationPresenter_ShowUpgradePlan_AllMigrations(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf(mock.AnythingOfType("string"), 2, "migrations").
		Return().
		Once()
	logger.EXPECT().
		Infof("\t%s\n\n", "210328_221600_first").
		Return().
		Once()
	logger.EXPECT().
		Infof("\t%s\n\n", "210328_221700_second").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	migrations := model.Migrations{
		{Version: "210328_221600_first"},
		{Version: "210328_221700_second"},
	}
	presenter.ShowUpgradePlan(migrations, 2)
}

func TestMigrationPresenter_ShowUpgradePlan_PartialMigrations(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf(mock.AnythingOfType("string"), 2, 5, "migrations").
		Return().
		Once()
	logger.EXPECT().
		Infof("\t%s\n\n", mock.AnythingOfType("string")).
		Return().
		Times(2)

	presenter := NewMigrationPresenter(logger)
	migrations := model.Migrations{
		{Version: "210328_221600_first"},
		{Version: "210328_221700_second"},
	}
	presenter.ShowUpgradePlan(migrations, 5)
}

func TestMigrationPresenter_ShowDowngradePlan(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf(mock.AnythingOfType("string"), 2, "migrations").
		Return().
		Once()
	logger.EXPECT().
		Infof("\t%s\n\n", mock.AnythingOfType("string")).
		Return().
		Times(2)

	presenter := NewMigrationPresenter(logger)
	migrations := model.Migrations{
		{Version: "210328_221600_first"},
		{Version: "210328_221700_second"},
	}
	presenter.ShowDowngradePlan(migrations)
}

func TestMigrationPresenter_PrintMigrations_WithoutTime(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Infof("\t%s\n\n", "210328_221600_test").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	migrations := model.Migrations{
		{Version: "210328_221600_test"},
	}
	presenter.PrintMigrations(migrations, false)
}

func TestMigrationPresenter_PrintMigrations_WithTime(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Infof("\t(%s) %s\n\n", mock.AnythingOfType("string"), "210328_221600_test").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	migrations := model.Migrations{
		{Version: "210328_221600_test", ApplyTime: 1616961360},
	}
	presenter.PrintMigrations(migrations, true)
}

func TestMigrationPresenter_PrintMigrations_Empty(t *testing.T) {
	logger := NewMockLogger(t)

	presenter := NewMigrationPresenter(logger)
	migrations := model.Migrations{}
	presenter.PrintMigrations(migrations, false)
}

func TestMigrationPresenter_AskUpgradeConfirmation(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "single migration",
			count: 1,
			want:  "Apply the above migration?",
		},
		{
			name:  "multiple migrations",
			count: 5,
			want:  "Apply the above migrations?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter := NewMigrationPresenter(nil)
			got := presenter.AskUpgradeConfirmation(tt.count)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationPresenter_AskDowngradeConfirmation(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "single migration",
			count: 1,
			want:  "Revert the above 1 migration?",
		},
		{
			name:  "multiple migrations",
			count: 3,
			want:  "Revert the above 3 migrations?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter := NewMigrationPresenter(nil)
			got := presenter.AskDowngradeConfirmation(tt.count)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationPresenter_ShowUpgradeError(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Errorf("%d from %d %s applied.\n", 2, 5, "migrations were").
		Return().
		Once()
	logger.EXPECT().
		Error("The rest of the migrations are canceled.").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowUpgradeError(2, 5)
}

func TestMigrationPresenter_ShowDowngradeError(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Errorf(mock.AnythingOfType("string"), 1, 3, "migration was").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowDowngradeError(1, 3)
}

func TestMigrationPresenter_ShowUpgradeSuccess(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Successf("%d %s applied\n", 3, "migrations were").
		Return().
		Once()
	logger.EXPECT().
		Success("Migrated up successfully").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowUpgradeSuccess(3)
}

func TestMigrationPresenter_ShowDowngradeSuccess(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Successf("%d %s reverted\n", 2, "migrations were").
		Return().
		Once()
	logger.EXPECT().
		Success("Migrated down successfully").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowDowngradeSuccess(2)
}

func TestMigrationPresenter_ShowNoNewMigrations(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Success("No new migrations found. Your system is up-to-date.").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowNoNewMigrations()
}

func TestMigrationPresenter_ShowNoMigrationsToRevert(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Success("No migration has been done before.").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowNoMigrationsToRevert()
}

func TestMigrationPresenter_ShowRedoPlan(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf(mock.AnythingOfType("string"), 2, "migrations").
		Return().
		Once()
	logger.EXPECT().
		Infof("\t%s\n\n", mock.AnythingOfType("string")).
		Return().
		Times(2)

	presenter := NewMigrationPresenter(logger)
	migrations := model.Migrations{
		{Version: "210328_221600_first"},
		{Version: "210328_221700_second"},
	}
	presenter.ShowRedoPlan(migrations)
}

func TestMigrationPresenter_AskRedoConfirmation(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "single migration",
			count: 1,
			want:  "Redo the above 1 migration?\n",
		},
		{
			name:  "multiple migrations",
			count: 4,
			want:  "Redo the above 4 migrations?\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenter := NewMigrationPresenter(nil)
			got := presenter.AskRedoConfirmation(tt.count)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMigrationPresenter_ShowRedoError(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Error("Migration failed. The rest of the migrations are canceled.\n").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowRedoError()
}

func TestMigrationPresenter_ShowRedoSuccess(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf("%d %s redone.\n", 2, "migrations were").
		Return().
		Once()
	logger.EXPECT().
		Success("Migration redone successfully.\n").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowRedoSuccess(2)
}

func TestMigrationPresenter_ShowHistoryHeader(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf(mock.AnythingOfType("string"), 10, "migrations").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowHistoryHeader(10)
}

func TestMigrationPresenter_ShowAllHistoryHeader(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf(mock.AnythingOfType("string"), 5, "migrations have").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowAllHistoryHeader(5)
}

func TestMigrationPresenter_ShowNewMigrationsHeader(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf(mock.AnythingOfType("string"), 3, "migrations").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowNewMigrationsHeader(3)
}

func TestMigrationPresenter_ShowNewMigrationsLimitedHeader(t *testing.T) {
	logger := NewMockLogger(t)
	logger.EXPECT().
		Warnf(mock.AnythingOfType("string"), 5, 10, "migrations").
		Return().
		Once()

	presenter := NewMigrationPresenter(logger)
	presenter.ShowNewMigrationsLimitedHeader(5, 10)
}
