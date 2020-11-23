/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev <resmus@gmail.com>
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */
package migrator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileNameBuilder_BuildUpFileName(t *testing.T) {
	fb := NewFileNameBuilder("/someDir")

	fileName, safely := fb.BuildUpFileName("210328_221600_test", true)
	assert.True(t, safely)
	assert.Equal(t, "/someDir/210328_221600_test.safe.up.sql", fileName)

	fileName, safely = fb.BuildUpFileName("210328_221600_test", false)
	assert.False(t, safely)
	assert.Equal(t, "/someDir/210328_221600_test.up.sql", fileName)
}

func TestFileNameBuilder_BuildDownFileName(t *testing.T) {
	fb := NewFileNameBuilder("/someDir")

	fileName, safely := fb.BuildDownFileName("210328_221600_test", true)
	assert.True(t, safely)
	assert.Equal(t, "/someDir/210328_221600_test.safe.down.sql", fileName)

	fileName, safely = fb.BuildDownFileName("210328_221600_test", false)
	assert.False(t, safely)
	assert.Equal(t, "/someDir/210328_221600_test.down.sql", fileName)
}
