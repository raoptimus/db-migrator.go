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
	"github.com/raoptimus/db-migrator/console"
	"github.com/raoptimus/db-migrator/iofile"
	"log"
	"path"
	"regexp"
	"time"
)

var RegexName = regexp.MustCompile(`^[\w\\\\]+$`)

func (s *MigrateController) CreateMigration(name string) error {
	if !RegexName.MatchString(name) {
		return errors.New("The migration name should contain letters, digits, underscore and/or backslash characters only.")
	}

	prefix := time.Now().Format("060102_150405")
	filenameUp := path.Join(s.options.Directory, prefix+"_"+name+".safe.up.sql")
	filenameDown := path.Join(s.options.Directory, prefix+"_"+name+".safe.down.sql")
	question := fmt.Sprintf("Create new migration files: \n'%s' and \n'%s'?\n", filenameUp, filenameDown)

	if !console.Confirm(question) {
		return nil
	}

	if err := iofile.CreateDirectory(s.options.Directory); err != nil {
		return err
	}

	if err := iofile.CreateFile(filenameUp); err != nil {
		return err
	}

	if err := iofile.CreateFile(filenameDown); err != nil {
		return err
	}

	log.Println(console.Green("New migration created successfully."))

	return nil
}
