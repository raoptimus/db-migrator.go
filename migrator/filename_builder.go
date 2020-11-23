/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package migrator

import (
	"github.com/raoptimus/db-migrator.go/iofile"
	"path/filepath"
)

const (
	safelyUpSuffix     = ".safe.up.sql"
	safelyDownSuffix   = ".safe.down.sql"
	unsafelyUpSuffix   = ".up.sql"
	unsafelyDownSuffix = ".down.sql"
)

type FileNameBuilder struct {
	migrationsDirectory string
}

func NewFileNameBuilder(migrationsDirectory string) FileNameBuilder {
	return FileNameBuilder{migrationsDirectory: migrationsDirectory}
}

func (s FileNameBuilder) BuildUpFileName(version string, forceSafely bool) (fname string, safely bool) {
	return s.buildFileName(version, safelyUpSuffix, unsafelyUpSuffix, forceSafely)
}

func (s FileNameBuilder) BuildDownFileName(version string, forceSafely bool) (fname string, safely bool) {
	return s.buildFileName(version, safelyDownSuffix, unsafelyDownSuffix, forceSafely)
}

func (s FileNameBuilder) buildFileName(version string, safelySuffix, unsafelySuffix string, forceSafely bool) (fname string, safely bool) {
	safelyFile := filepath.Join(s.migrationsDirectory, version+safelySuffix)
	unsafelyFile := filepath.Join(s.migrationsDirectory, version+unsafelySuffix)

	switch {
	case iofile.Exists(safelyFile):
		return safelyFile, true
	case iofile.Exists(unsafelyFile):
		return unsafelyFile, false
	case forceSafely:
		return safelyFile, true
	default:
		return unsafelyFile, false
	}
}
