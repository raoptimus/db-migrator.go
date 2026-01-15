/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package iohelp

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

// File provides file system operations including existence checks, reading, and creation.
type File struct{}

// StdFile is the standard File instance for general use.
var StdFile = NewFile()

// NewFile creates a new File instance.
func NewFile() *File {
	return &File{}
}

// Exists checks whether a file exists at the specified path.
// It returns true if the file exists, false if it does not exist, and an error for other failures.
func (f *File) Exists(fileName string) (bool, error) {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Open opens the named file for reading.
// It returns an io.ReadCloser for the file or an error if the operation fails.
func (f *File) Open(filename string) (io.ReadCloser, error) {
	return os.Open(filename)
}

// ReadAll reads the entire contents of the named file.
// It returns the file contents as a byte slice or an error if the operation fails.
func (f *File) ReadAll(filename string) ([]byte, error) {
	ff, err := f.Open(filename)
	if err != nil {
		return nil, err
	}
	defer ff.Close()

	return io.ReadAll(ff)
}

// Create creates a new file with the specified name.
// If the file already exists, it will be truncated.
func (f *File) Create(filename string) error {
	ff, err := os.Create(filename)
	if err == nil {
		err = ff.Close()
	}

	if err != nil {
		return errors.Wrapf(err, "creating file %s", filename)
	}

	return nil
}
