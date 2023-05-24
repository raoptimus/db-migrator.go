/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package builder

import (
	"path/filepath"
)

const (
	safelyUpSuffix     = ".safe.up.sql"
	safelyDownSuffix   = ".safe.down.sql"
	unsafelyUpSuffix   = ".up.sql"
	unsafelyDownSuffix = ".down.sql"
)

type FileName struct {
	file                File
	migrationsDirectory string
}

func NewFileName(file File, migrationsDirectory string) *FileName {
	return &FileName{
		file:                file,
		migrationsDirectory: migrationsDirectory,
	}
}

// Up builds a file name for migration update.
func (s *FileName) Up(version string, forceSafely bool) (fname string, safely bool) {
	return s.build(version, safelyUpSuffix, unsafelyUpSuffix, forceSafely)
}

// Down builds a file name for migration downgrade.
func (s *FileName) Down(version string, forceSafely bool) (fname string, safely bool) {
	return s.build(version, safelyDownSuffix, unsafelyDownSuffix, forceSafely)
}

func (s *FileName) build(
	version,
	safelySuffix,
	unsafelySuffix string,
	forceSafely bool,
) (fname string, safely bool) {
	safelyFile := filepath.Join(s.migrationsDirectory, version+safelySuffix)
	unsafelyFile := filepath.Join(s.migrationsDirectory, version+unsafelySuffix)

	if exists, _ := s.file.Exists(unsafelyFile); exists && !forceSafely {
		return unsafelyFile, false
	}

	return safelyFile, true
}
