package action

import (
	"testing"

	"github.com/raoptimus/db-migrator.go/internal/action/mockaction"
	"github.com/raoptimus/db-migrator.go/internal/dal/entity"
	"github.com/stretchr/testify/assert"
)

func TestUpgrade_Run_NoMigrations_NoError(t *testing.T) {
	ctx := cliContext(t, "2")

	serv := mockaction.NewMigrationService(t)
	serv.EXPECT().
		NewMigrations(ctx.Context).
		Return(entity.Migrations{}, nil)

	c := mockaction.NewConsole(t)
	c.EXPECT().
		SuccessLn("No new migrations found. Your system is up-to-date.")

	fb := mockaction.NewFileNameBuilder(t)

	upgrade := NewUpgrade(c, serv, fb, true)
	err := upgrade.Run(ctx)
	assert.NoError(t, err)
}
