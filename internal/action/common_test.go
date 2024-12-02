package action

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

var time230527213123 = time.Date(
	2023, 05, 27,
	21, 31, 23,
	0, time.UTC)

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
