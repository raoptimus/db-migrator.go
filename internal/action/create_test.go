package action

import (
	"context"
	"testing"
	"time"

	"github.com/raoptimus/db-migrator.go/internal/action/mockaction"
	"github.com/raoptimus/db-migrator.go/internal/util/timex"
	"github.com/stretchr/testify/require"
)

func TestCreate_Run_ExpectedArguments_NoError(t *testing.T) {
	tests := []struct {
		name         string
		now          time.Time
		version      string
		fileNameUp   string
		fileNameDown string
		safely       bool
	}{
		{
			name:         "safely is true",
			now:          time230527213123,
			version:      "230527_213123_init",
			fileNameUp:   "230527_213123_init.safe.up.sql",
			fileNameDown: "230527_213123_init.safe.down.sql",
			safely:       true,
		},
		{
			name:         "safely is false",
			now:          time230527213123,
			version:      "230527_213123_init",
			fileNameUp:   "230527_213123_init.safe.up.sql",
			fileNameDown: "230527_213123_init.safe.down.sql",
			safely:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tm := timex.New(func() time.Time {
				return time230527213123
			})

			f := mockaction.NewFile(t)
			f.EXPECT().
				Exists("/tmp").
				Return(true, nil).
				Once()
			f.EXPECT().
				Create(tt.fileNameUp).
				Return(nil).
				Once()
			f.EXPECT().
				Create(tt.fileNameDown).
				Return(nil).
				Once()

			c := mockaction.NewConsole(t)
			expectedQuestion := "Create new migration files: \n" +
				"'" + tt.fileNameUp + "' and \n" +
				"'" + tt.fileNameDown + "'?\n"
			c.EXPECT().
				Confirm(expectedQuestion).
				Return(true)
			c.EXPECT().
				SuccessLn("New migration created successfully.")

			fb := mockaction.NewFileNameBuilder(t)
			fb.EXPECT().
				Up(tt.version, true).
				Return(tt.fileNameUp, tt.safely)
			fb.EXPECT().
				Down(tt.version, true).
				Return(tt.fileNameDown, tt.safely)

			create := NewCreate(tm, f, c, fb, "/tmp")
			err := create.Run(ctx, "init")
			require.NoError(t, err)
		})
	}
}
