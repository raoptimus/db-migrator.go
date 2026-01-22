/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileNameBuilder_BuildUpFileName(t *testing.T) {
	tests := []struct {
		name             string
		forceSafely      bool
		version          string
		expectedSafely   bool
		existsFileName   string
		expectedFileName string
		fileExists       bool
	}{
		{
			name:             "unsafely file exists expects unsafely filename",
			forceSafely:      false,
			version:          "210328_221600_test",
			expectedSafely:   false,
			existsFileName:   "/someDir/210328_221600_test.up.sql",
			expectedFileName: "/someDir/210328_221600_test.up.sql",
			fileExists:       true,
		},
		{
			name:             "unsafely file not exists expects safely filename",
			forceSafely:      false,
			version:          "210328_221600_test",
			expectedSafely:   true,
			existsFileName:   "/someDir/210328_221600_test.up.sql",
			expectedFileName: "/someDir/210328_221600_test.safe.up.sql",
			fileExists:       false,
		},
		{
			name:             "unsafely file exists force safely expects safely filename",
			forceSafely:      true,
			version:          "210328_221600_test",
			expectedSafely:   true,
			existsFileName:   "/someDir/210328_221600_test.up.sql",
			expectedFileName: "/someDir/210328_221600_test.safe.up.sql",
			fileExists:       true,
		},
		{
			name:             "unsafely file not exists force safely expects safely filename",
			forceSafely:      true,
			version:          "210328_221600_test",
			expectedSafely:   true,
			existsFileName:   "/someDir/210328_221600_test.up.sql",
			expectedFileName: "/someDir/210328_221600_test.safe.up.sql",
			fileExists:       false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := NewMockFile(t)
			file.EXPECT().
				Exists(test.existsFileName).
				Return(test.fileExists, nil)

			fb := NewFileName(file, "/someDir")
			fileName, safely := fb.Up(test.version, test.forceSafely)

			assert.Equal(t, test.expectedSafely, safely)
			assert.Equal(t, test.expectedFileName, fileName)
		})
	}
}

func TestFileNameBuilder_BuildDownFileName(t *testing.T) {
	tests := []struct {
		name             string
		forceSafely      bool
		version          string
		expectedSafely   bool
		existsFileName   string
		expectedFileName string
		fileExists       bool
	}{
		{
			name:             "unsafely file exists expects unsafely filename",
			forceSafely:      false,
			version:          "210328_221600_test",
			expectedSafely:   false,
			existsFileName:   "/someDir/210328_221600_test.down.sql",
			expectedFileName: "/someDir/210328_221600_test.down.sql",
			fileExists:       true,
		},
		{
			name:             "unsafely file not exists expects safely filename",
			forceSafely:      false,
			version:          "210328_221600_test",
			expectedSafely:   true,
			existsFileName:   "/someDir/210328_221600_test.down.sql",
			expectedFileName: "/someDir/210328_221600_test.safe.down.sql",
			fileExists:       false,
		},
		{
			name:             "unsafely file exists force safely expects safely filename",
			forceSafely:      true,
			version:          "210328_221600_test",
			expectedSafely:   true,
			existsFileName:   "/someDir/210328_221600_test.down.sql",
			expectedFileName: "/someDir/210328_221600_test.safe.down.sql",
			fileExists:       true,
		},
		{
			name:             "unsafely file not exists force safely expects safely filename",
			forceSafely:      true,
			version:          "210328_221600_test",
			expectedSafely:   true,
			existsFileName:   "/someDir/210328_221600_test.down.sql",
			expectedFileName: "/someDir/210328_221600_test.safe.down.sql",
			fileExists:       false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file := NewMockFile(t)
			file.EXPECT().
				Exists(test.existsFileName).
				Return(test.fileExists, nil)

			fb := NewFileName(file, "/someDir")
			fileName, safely := fb.Down(test.version, test.forceSafely)

			assert.Equal(t, test.expectedSafely, safely)
			assert.Equal(t, test.expectedFileName, fileName)
		})
	}
}
