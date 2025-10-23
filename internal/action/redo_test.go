
package action

import (
	"context"
	"errors"
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/action/mockaction"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRedo_Run_NoMigrations_NoError(t *testing.T) {
	ctx := context.Background()

	serv := mockaction.NewMigrationService(t)
	serv.EXPECT().
		Migrations(ctx, 1).
		Return(entity.Migrations{}, nil)

	fb := mockaction.NewFileNameBuilder(t)

	redo := NewRedo(serv, fb, true)
	err := redo.Run(ctx)
	assert.NoError(t, err)
}

func TestRedo_Run_OneMigration_NoError(t *testing.T) {
	ctx := context.Background()
	migration := entity.Migration{Version: "1"}

	serv := mockaction.NewMigrationService(t)
	serv.EXPECT().
		Migrations(ctx, 1).
		Return(entity.Migrations{migration}, nil)
	serv.EXPECT().
		RevertFile(ctx, &migration, "1.down.sql", false).
		Return(nil)
	serv.EXPECT().
		ApplyFile(ctx, &migration, "1.up.sql", false).
		Return(nil)

	fb := mockaction.NewFileNameBuilder(t)
	fb.EXPECT().
		Down("1", false).
		Return("1.down.sql", false)
	fb.EXPECT().
		Up("1", false).
		Return("1.up.sql", false)

	redo := NewRedo(serv, fb, false)
	err := redo.Run(ctx)
	assert.NoError(t, err)
}

func TestRedo_Run_MultipleMigrations_NoError(t *testing.T) {
	ctx := context.Background()
	migrations := entity.Migrations{
		{Version: "1"},
		{Version: "2"},
	}

	serv := mockaction.NewMigrationService(t)
	serv.EXPECT().
		Migrations(ctx, 2).
		Return(migrations, nil)
	serv.EXPECT().
		RevertFile(ctx, &migrations[0], "1.down.sql", false).
		Return(nil)
	serv.EXPECT().
		RevertFile(ctx, &migrations[1], "2.down.sql", false).
		Return(nil)
	serv.EXPECT().
		ApplyFile(ctx, &migrations[1], "2.up.sql", false).
		Return(nil)
	serv.EXPECT().
		ApplyFile(ctx, &migrations[0], "1.up.sql", false).
		Return(nil)

	fb := mockaction.NewFileNameBuilder(t)
	fb.EXPECT().
		Down("1", false).
		Return("1.down.sql", false)
	fb.EXPECT().
		Down("2", false).
		Return("2.down.sql", false)
	fb.EXPECT().
		Up("1", false).
		Return("1.up.sql", false)
	fb.EXPECT().
		Up("2", false).
		Return("2.up.sql", false)

	redo := NewRedo(serv, fb, false)
	err := redo.Run(ctx, "2")
	assert.NoError(t, err)
}

func TestRedo_Run_InteractiveMode_ConfirmFalse_NoError(t *testing.T) {
	ctx := context.Background()
	migrations := entity.Migrations{
		{Version: "1"},
	}

	serv := mockaction.NewMigrationService(t)
	serv.EXPECT().
		Migrations(ctx, 1).
		Return(migrations, nil)

	c := mockaction.NewConsole(t)
	c.EXPECT().
		Confirm(mock.Anything).
		Return(false)

	fb := mockaction.NewFileNameBuilder(t)

	redo := NewRedo(serv, fb, true)
	err := redo.Run(ctx)
	assert.NoError(t, err)
}

func TestRedo_Run_RevertFileError_Error(t *testing.T) {
	ctx := context.Background()
	migration := entity.Migration{Version: "1"}
	expectedErr := errors.New("revert error")

	serv := mockaction.NewMigrationService(t)
	serv.EXPECT().
		Migrations(ctx, 1).
		Return(entity.Migrations{migration}, nil)
	serv.EXPECT().
		RevertFile(ctx, &migration, "1.down.sql", false).
		Return(expectedErr)

	fb := mockaction.NewFileNameBuilder(t)
	fb.EXPECT().
		Down("1", false).
		Return("1.down.sql", false)

	redo := NewRedo(serv, fb, false)
	err := redo.Run(ctx)
	assert.ErrorIs(t, err, expectedErr)
}

func TestRedo_Run_ApplyFileError_Error(t *testing.T) {
	ctx := context.Background()
	migration := entity.Migration{Version: "1"}
	expectedErr := errors.New("apply error")

	serv := mockaction.NewMigrationService(t)
	serv.EXPECT().
		Migrations(ctx, 1).
		Return(entity.Migrations{migration}, nil)
	serv.EXPECT().
		RevertFile(ctx, &migration, "1.down.sql", false).
		Return(nil)
	serv.EXPECT().
		ApplyFile(ctx, &migration, "1.up.sql", false).
		Return(expectedErr)

	fb := mockaction.NewFileNameBuilder(t)
	fb.EXPECT().
		Down("1", false).
		Return("1.down.sql", false)
	fb.EXPECT().
		Up("1", false).
		Return("1.up.sql", false)

	redo := NewRedo(serv, fb, false)
	err := redo.Run(ctx)
	assert.ErrorIs(t, err, expectedErr)
}
