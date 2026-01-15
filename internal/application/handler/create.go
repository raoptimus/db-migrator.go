/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"fmt"
	"os"
	"regexp"

	"github.com/pkg/errors"
	"github.com/raoptimus/db-migrator.go/internal/helper/console"
	"github.com/raoptimus/db-migrator.go/internal/helper/timex"
)

const fileModeExecutable = 0o755

// ErrInvalidFileName is returned when a migration name contains invalid characters.
var ErrInvalidFileName = errors.New("the migration name should contain letters, digits, underscore and/or backslash characters only")
var regexpFileName = regexp.MustCompile(`^[\w\\]+$`)

// Create handles the creation of new migration files.
type Create struct {
	options         *Options
	file            File
	logger          Logger
	fileNameBuilder FileNameBuilder
}

// NewCreate creates a new Create handler instance.
func NewCreate(
	options *Options,
	logger Logger,
	file File,
	fileNameBuilder FileNameBuilder,
) *Create {
	return &Create{
		options:         options,
		file:            file,
		logger:          logger,
		fileNameBuilder: fileNameBuilder,
	}
}

// Handle processes the create command to generate new migration files.
func (c *Create) Handle(cmd *Command) error {
	if !cmd.Args.Present() {
		return errors.WithStack(ErrInvalidFileName)
	}

	migrationName := cmd.Args.First()
	if !regexpFileName.MatchString(migrationName) {
		return ErrInvalidFileName
	}

	prefix := timex.StdTime.Now().Format("060102_150405")
	version := prefix + "_" + migrationName
	fileNameUp, _ := c.fileNameBuilder.Up(version, true)
	fileNameDown, _ := c.fileNameBuilder.Down(version, true)

	question := fmt.Sprintf(
		"Create new migration files: \n'%s' and \n'%s'?\n",
		fileNameUp,
		fileNameDown,
	)
	if c.options.Interactive && !console.Confirm(question) {
		return nil
	}

	if err := c.createDirectory(c.options.Directory); err != nil {
		return err
	}

	if err := c.file.Create(fileNameUp); err != nil {
		return err
	}

	if err := c.file.Create(fileNameDown); err != nil {
		return err
	}

	c.logger.Success("New migration created successfully.")

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
