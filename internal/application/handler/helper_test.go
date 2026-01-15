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

// argsStub is a simple stub implementation of cli.Args for testing purposes.
type argsStub struct {
	present bool
	first   string
}

func (a *argsStub) Get(_ int) string { return "" }
func (a *argsStub) First() string    { return a.first }
func (a *argsStub) Tail() []string   { return nil }
func (a *argsStub) Len() int         { return 0 }
func (a *argsStub) Present() bool    { return a.present }
func (a *argsStub) Slice() []string  { return nil }

func TestStepOrDefault_ArgsNotPresent_ReturnsDefault_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		defaults     int
		expectedStep int
	}{
		{
			name:         "default value is 1",
			defaults:     1,
			expectedStep: 1,
		},
		{
			name:         "default value is 10",
			defaults:     10,
			expectedStep: 10,
		},
		{
			name:         "default value is 0",
			defaults:     0,
			expectedStep: 0,
		},
		{
			name:         "default value is 100",
			defaults:     100,
			expectedStep: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: false, first: ""},
			}

			step, err := stepOrDefault(cmd, tt.defaults)

			require.NoError(t, err)
			require.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestStepOrDefault_ArgsIsEmptyString_ReturnsDefault_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		defaults     int
		expectedStep int
	}{
		{
			name:         "default value is 1",
			defaults:     1,
			expectedStep: 1,
		},
		{
			name:         "default value is 5",
			defaults:     5,
			expectedStep: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: ""},
			}

			step, err := stepOrDefault(cmd, tt.defaults)

			require.NoError(t, err)
			require.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestStepOrDefault_ArgsIsAll_ReturnsZero_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		defaults     int
		expectedStep int
	}{
		{
			name:         "default value is 1",
			defaults:     1,
			expectedStep: 0,
		},
		{
			name:         "default value is 100",
			defaults:     100,
			expectedStep: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: "all"},
			}

			step, err := stepOrDefault(cmd, tt.defaults)

			require.NoError(t, err)
			require.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestStepOrDefault_ValidPositiveInteger_ReturnsStep_Successfully(t *testing.T) {
	tests := []struct {
		name         string
		argValue     string
		defaults     int
		expectedStep int
	}{
		{
			name:         "boundary value 1",
			argValue:     "1",
			defaults:     10,
			expectedStep: 1,
		},
		{
			name:         "value 2",
			argValue:     "2",
			defaults:     10,
			expectedStep: 2,
		},
		{
			name:         "value 5",
			argValue:     "5",
			defaults:     1,
			expectedStep: 5,
		},
		{
			name:         "value 10",
			argValue:     "10",
			defaults:     1,
			expectedStep: 10,
		},
		{
			name:         "value 100",
			argValue:     "100",
			defaults:     5,
			expectedStep: 100,
		},
		{
			name:         "large value 999999",
			argValue:     "999999",
			defaults:     1,
			expectedStep: 999999,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			step, err := stepOrDefault(cmd, tt.defaults)

			require.NoError(t, err)
			require.Equal(t, tt.expectedStep, step)
		})
	}
}

func TestStepOrDefault_InvalidNonNumericArg_ReturnsError_Failure(t *testing.T) {
	tests := []struct {
		name              string
		argValue          string
		expectedErrSubstr string
	}{
		{
			name:              "alphabetic string",
			argValue:          "abc",
			expectedErrSubstr: "the step argument abc is not valid",
		},
		{
			name:              "special characters",
			argValue:          "@#$",
			expectedErrSubstr: "the step argument @#$ is not valid",
		},
		{
			name:              "mixed alphanumeric",
			argValue:          "12abc",
			expectedErrSubstr: "the step argument 12abc is not valid",
		},
		{
			name:              "float value",
			argValue:          "1.5",
			expectedErrSubstr: "the step argument 1.5 is not valid",
		},
		{
			name:              "space in value",
			argValue:          "1 2",
			expectedErrSubstr: "the step argument 1 2 is not valid",
		},
		{
			name:              "hexadecimal format",
			argValue:          "0x10",
			expectedErrSubstr: "the step argument 0x10 is not valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			step, err := stepOrDefault(cmd, 1)

			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErrSubstr)
			require.Equal(t, -1, step)
		})
	}
}

func TestStepOrDefault_ZeroOrNegativeArg_ReturnsError_Failure(t *testing.T) {
	tests := []struct {
		name        string
		argValue    string
		expectedErr error
	}{
		{
			name:        "boundary value 0",
			argValue:    "0",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
		{
			name:        "negative value -1",
			argValue:    "-1",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
		{
			name:        "negative value -5",
			argValue:    "-5",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
		{
			name:        "large negative value -100",
			argValue:    "-100",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
		{
			name:        "negative value -999999",
			argValue:    "-999999",
			expectedErr: ErrArgumentMustBeGreaterThanZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Args: &argsStub{present: true, first: tt.argValue},
			}

			step, err := stepOrDefault(cmd, 1)

			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr)
			require.Equal(t, -1, step)
		})
	}
}
