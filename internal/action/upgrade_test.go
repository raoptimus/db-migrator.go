package action

import (
	"context"
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/action/mockaction"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
	"github.com/stretchr/testify/require"
)

func TestUpgrade_Run_NoMigrations_NoError(t *testing.T) {
	ctx := context.Background()

	serv := mockaction.NewMigrationService(t)
	serv.EXPECT().
		NewMigrations(ctx).
		Return(entity.Migrations{}, nil)

	c := mockaction.NewConsole(t)
	c.EXPECT().
		SuccessLn("No new migrations found. Your system is up-to-date.")

	fb := mockaction.NewFileNameBuilder(t)

	upgrade := NewUpgrade(c, serv, fb, true)
	err := upgrade.Run(ctx, "2")
	require.NoError(t, err)
}
