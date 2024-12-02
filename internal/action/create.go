/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package action

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/pkg/timex"
	"github.com/urfave/cli/v3"
)

const fileModeExecutable = 0o755

var regexpFileName = regexp.MustCompile(`^[\w\\]+$`)

type Create struct {
	time            timex.Time
	file            File
	console         Console
	fileNameBuilder FileNameBuilder
	migrationDir    string
}

func NewCreate(
	time timex.Time,
	file File,
	console Console,
	fileNameBuilder FileNameBuilder,
	migrationDir string,
) *Create {
	return &Create{
		time:            time,
		file:            file,
		console:         console,
		fileNameBuilder: fileNameBuilder,
		migrationDir:    migrationDir,
	}
}

func (c *Create) Run(_ context.Context, cmdArgs cli.Args) error {
	migrationName := cmdArgs.Get(0)
	if !regexpFileName.MatchString(migrationName) {
		return ErrInvalidFileName
	}

	prefix := c.time.Now().Format("060102_150405")
	version := prefix + "_" + migrationName
	fileNameUp, _ := c.fileNameBuilder.Up(version, true)
	fileNameDown, _ := c.fileNameBuilder.Down(version, true)

	question := fmt.Sprintf(
		"Create new migration files: \n'%s' and \n'%s'?\n",
		fileNameUp,
		fileNameDown,
	)
	if !c.console.Confirm(question) {
		return nil
	}

	if err := c.createDirectory(c.migrationDir); err != nil {
		return err
	}

	if err := c.file.Create(fileNameUp); err != nil {
		return err
	}

	if err := c.file.Create(fileNameDown); err != nil {
		return err
	}

	c.console.SuccessLn("New migration created successfully.")

	return nil
}

func (c *Create) createDirectory(path string) error {
	if ok, err := c.file.Exists(path); err != nil || ok {
		return err
	}

	if err := os.Mkdir(path, fileModeExecutable); err != nil {
		return errors.Wrapf(err, "creating directory %s", path)
	}

	return nil
}
