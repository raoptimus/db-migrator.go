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

const fileModeExecutable = 0o755

type File struct{}

var StdFile = NewFile()

func NewFile() *File {
	return &File{}
}

func (f *File) Exists(fileName string) (bool, error) {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (f *File) Open(filename string) (io.ReadCloser, error) {
	return os.Open(filename)
}

func (f *File) ReadAll(filename string) ([]byte, error) {
	ff, err := f.Open(filename)
	if err != nil {
		return nil, err
	}
	defer ff.Close()

	return io.ReadAll(ff)
}

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
