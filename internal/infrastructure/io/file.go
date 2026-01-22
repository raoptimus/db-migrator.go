/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package iofile

import (
	"os"

	"github.com/pkg/errors"
)

const fileModeExecutable = 0o755

// Exists checks whether a file or directory exists at the specified path.
// It returns true if the path exists, false otherwise.
func Exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// CreateDirectory creates a new directory at the specified path with executable permissions.
// If the directory already exists, it returns nil without error.
func CreateDirectory(path string) error {
	if Exists(path) {
		return nil
	}

	if err := os.Mkdir(path, fileModeExecutable); err != nil {
		return errors.Wrapf(err, "creating directory %s", path)
	}

	return nil
}

// CreateFile creates a new empty file with the specified filename.
// If the file already exists, it will be truncated.
func CreateFile(filename string) error {
	f, err := os.Create(filename)
	if err == nil {
		err = f.Close()
	}

	if err != nil {
		return errors.Wrapf(err, "creating file %s", filename)
	}

	return nil
}
