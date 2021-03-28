/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package migrator

import (
	"errors"
	"fmt"
	"github.com/raoptimus/db-migrator.go/console"
	"github.com/raoptimus/db-migrator.go/iofile"
	"log"
	"regexp"
	"time"
)

var RegexName = regexp.MustCompile(`^[\w\\\\]+$`)

func (s *Service) CreateMigration(name string) error {
	if !RegexName.MatchString(name) {
		return errors.New("The migration name should contain letters, digits, underscore and/or backslash characters only.")
	}

	prefix := time.Now().Format("060102_150405")
	version := prefix + "_" + name
	fileNameUp, _ := s.fileBuilder.BuildUpFileName(version, true)
	fileNameDown, _ := s.fileBuilder.BuildDownFileName(version, true)
	question := fmt.Sprintf("Create new migration files: \n'%s' and \n'%s'?\n", fileNameUp, fileNameDown)

	if !console.Confirm(question) {
		return nil
	}

	if err := iofile.CreateDirectory(s.options.Directory); err != nil {
		return err
	}

	if err := iofile.CreateFile(fileNameUp); err != nil {
		return err
	}

	if err := iofile.CreateFile(fileNameDown); err != nil {
		return err
	}

	log.Println(console.Green("New migration created successfully."))

	return nil
}
