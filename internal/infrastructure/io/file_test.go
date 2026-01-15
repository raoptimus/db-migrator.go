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
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExists_FileExists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	exists := Exists(tmpFile.Name())

	assert.True(t, exists)
}

func TestExists_FileDoesNotExist(t *testing.T) {
	exists := Exists("/nonexistent/path/to/file.txt")

	assert.False(t, exists)
}

func TestExists_DirectoryExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exists := Exists(tmpDir)

	assert.True(t, exists)
}

func TestCreateDirectory_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	newDir := filepath.Join(tmpDir, "subdir")

	err = CreateDirectory(newDir)

	require.NoError(t, err)
	assert.True(t, Exists(newDir))
}

func TestCreateDirectory_AlreadyExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = CreateDirectory(tmpDir)

	require.NoError(t, err)
}

func TestCreateDirectory_InvalidPath(t *testing.T) {
	err := CreateDirectory("/nonexistent/parent/newdir")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating directory")
}

func TestCreateFile_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "newfile.txt")

	err = CreateFile(filename)

	require.NoError(t, err)
	assert.True(t, Exists(filename))
}

func TestCreateFile_InvalidPath(t *testing.T) {
	err := CreateFile("/nonexistent/directory/file.txt")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating file")
}

func TestCreateFile_TruncatesExistingFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("original content")
	require.NoError(t, err)
	tmpFile.Close()

	err = CreateFile(tmpFile.Name())
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Empty(t, data)
}
