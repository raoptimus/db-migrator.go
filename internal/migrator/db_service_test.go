package migrator

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestDBService_Create_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{})
	action := dbServ.Create()
	assert.NotNil(t, action)
}

func TestDBService_Upgrade_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	action, err := dbServ.Upgrade()
	assert.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_Downgrade_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	action, err := dbServ.Downgrade()
	assert.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_To_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	action, err := dbServ.To()
	assert.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_History_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	action, err := dbServ.History()
	assert.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_HistoryNew_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	action, err := dbServ.HistoryNew()
	assert.NoError(t, err)
	assert.NotNil(t, action)
}

func TestDBService_Redo_ReturnsAction(t *testing.T) {
	dbServ := New(&Options{
		DSN: "postgres://docker:docker@postgres:5432/docker?sslmode=disable",
	})
	action, err := dbServ.Redo()
	assert.NoError(t, err)
	assert.NotNil(t, action)
}

func flagSet(t *testing.T, argument string) *flag.FlagSet {
	flagSet := flag.NewFlagSet("test", 0)
	err := flagSet.Parse([]string{argument})
	assert.NoError(t, err)

	return flagSet
}

func cliContext(t *testing.T, argument string) *cli.Context {
	flagSet := flagSet(t, argument)

	return cli.NewContext(nil, flagSet, nil)
}
