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

// TestUpgrade_Handle_InvalidStepArgument_Failure tests that Handle returns an error
// when stepOrDefault fails to parse the step argument from command args.
func TestUpgrade_Handle_InvalidStepArgument_Failure(t *testing.T) {
	tests := []struct {
		name              string
		argValue          string
		expectedErrSubstr string
	}{
		{
			name:              "non-numeric argument",
			argValue:          "abc",
			expectedErrSubstr: "the step argument abc is not valid",
		},
		{
			name:              "float argument",
			argValue:          "1.5",
			expectedErrSubstr: "the step argument 1.5 is not valid",
		},
		{
			name:              "negative argument",
			argValue:          "-5",
			expectedErrSubstr: "the step argument must be greater than 0",
		},
		{
			name:              "zero argument",
			argValue:          "0",
			expectedErrSubstr: "the step argument must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			presenterMock := NewMockPresenter(t)
			fileNameBuilderMock := NewMockFileNameBuilder(t)

			upgrade := NewUpgrade(
				&Options{
					DSN:       "postgres://user:pass@localhost:5432/testdb",
					Directory: "/migrations",
				},
				presenterMock,
				fileNameBuilderMock,
			)

			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			err := upgrade.Handle(cmd, nil)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
		})
	}
}

// TestNewUpgrade_InitializesFieldsCorrectly_Successfully verifies that NewUpgrade
// properly initializes all struct fields.
func TestNewUpgrade_InitializesFieldsCorrectly_Successfully(t *testing.T) {
	presenterMock := NewMockPresenter(t)
	fileNameBuilderMock := NewMockFileNameBuilder(t)
	options := &Options{
		DSN:       "postgres://user:pass@localhost:5432/testdb",
		Directory: "/migrations",
	}

	upgrade := NewUpgrade(options, presenterMock, fileNameBuilderMock)

	require.NotNil(t, upgrade)
	require.Equal(t, options, upgrade.options)
	require.Equal(t, presenterMock, upgrade.presenter)
	require.Equal(t, fileNameBuilderMock, upgrade.fileNameBuilder)
}
