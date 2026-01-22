/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package handler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTo_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	loggerMock := NewMockLogger(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	to := NewTo(options, loggerMock, fileNameBuilderMock)

	require.NotNil(t, to)
	require.Equal(t, options, to.options)
	require.Equal(t, loggerMock, to.logger)
	require.Equal(t, fileNameBuilderMock, to.fileNameBuilder)
}

func TestTo_Handle_LogsComingSoon_Successfully(t *testing.T) {
	loggerMock := NewMockLogger(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	loggerMock.EXPECT().Info("coming soon").Once()

	to := NewTo(
		&Options{
			DSN:       "postgres://user:pass@localhost:5432/testdb",
			Directory: "/migrations",
		},
		loggerMock,
		fileNameBuilderMock,
	)

	cmd := &Command{
		Args: &argsStub{present: false},
	}

	err := to.Handle(cmd, svcMock)

	require.NoError(t, err)
}

func TestTo_Handle_WithVersion_Successfully(t *testing.T) {
	loggerMock := NewMockLogger(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	svcMock := NewMockMigrationService(t)

	loggerMock.EXPECT().Info("coming soon").Once()

	to := NewTo(
		&Options{
			DSN:       "postgres://user:pass@localhost:5432/testdb",
			Directory: "/migrations",
		},
		loggerMock,
		fileNameBuilderMock,
	)

	cmd := &Command{
		Args: &argsStub{present: true, first: "200101_120000"},
	}

	err := to.Handle(cmd, svcMock)

	require.NoError(t, err)
}
