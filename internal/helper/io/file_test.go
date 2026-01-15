/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package iohelp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFile(t *testing.T) {
	f := NewFile()
	require.NotNil(t, f)
}

func TestStdFile_IsNotNil(t *testing.T) {
	require.NotNil(t, StdFile)
}

func TestFile_Exists_FileExists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	f := NewFile()
	exists, err := f.Exists(tmpFile.Name())

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFile_Exists_FileDoesNotExist(t *testing.T) {
	f := NewFile()
	exists, err := f.Exists("/nonexistent/path/to/file.txt")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFile_Exists_DirectoryExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	f := NewFile()
	exists, err := f.Exists(tmpDir)

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFile_Open_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString("test content")
	require.NoError(t, err)
	tmpFile.Close()

	f := NewFile()
	reader, err := f.Open(tmpFile.Name())

	require.NoError(t, err)
	require.NotNil(t, reader)
	reader.Close()
}

func TestFile_Open_FileNotFound(t *testing.T) {
	f := NewFile()
	reader, err := f.Open("/nonexistent/file.txt")

	require.Error(t, err)
	assert.Nil(t, reader)
}

func TestFile_ReadAll_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := "Hello, World!"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	f := NewFile()
	data, err := f.ReadAll(tmpFile.Name())

	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestFile_ReadAll_FileNotFound(t *testing.T) {
	f := NewFile()
	data, err := f.ReadAll("/nonexistent/file.txt")

	require.Error(t, err)
	assert.Nil(t, data)
}

func TestFile_ReadAll_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	f := NewFile()
	data, err := f.ReadAll(tmpFile.Name())

	require.NoError(t, err)
	assert.Empty(t, data)
}

func TestFile_Create_Success(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filename := filepath.Join(tmpDir, "newfile.txt")

	f := NewFile()
	err = f.Create(filename)

	require.NoError(t, err)

	exists, err := f.Exists(filename)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFile_Create_InvalidPath(t *testing.T) {
	f := NewFile()
	err := f.Create("/nonexistent/directory/file.txt")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating file")
}

func TestFile_Create_TruncatesExistingFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("original content")
	require.NoError(t, err)
	tmpFile.Close()

	f := NewFile()
	err = f.Create(tmpFile.Name())
	require.NoError(t, err)

	data, err := f.ReadAll(tmpFile.Name())
	require.NoError(t, err)
	assert.Empty(t, data)
}
