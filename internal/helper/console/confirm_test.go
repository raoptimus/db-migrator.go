/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package console

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfirmf_AffirmativeResponses_ReturnsTrue(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "lowercase y",
			input: "y\n",
		},
		{
			name:  "lowercase yes",
			input: "yes\n",
		},
		{
			name:  "uppercase Y",
			input: "Y\n",
		},
		{
			name:  "uppercase YES",
			input: "YES\n",
		},
		{
			name:  "mixed case Yes",
			input: "Yes\n",
		},
		{
			name:  "mixed case yEs",
			input: "yEs\n",
		},
		{
			name:  "y with leading spaces",
			input: "  y\n",
		},
		{
			name:  "y with trailing spaces",
			input: "y  \n",
		},
		{
			name:  "y with leading and trailing spaces",
			input: "  y  \n",
		},
		{
			name:  "yes with spaces",
			input: "  yes  \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			oldStdin := os.Stdin
			oldStdout := os.Stdout

			stdinReader, stdinWriter, err := os.Pipe()
			require.NoError(t, err)

			stdoutReader, stdoutWriter, err := os.Pipe()
			require.NoError(t, err)

			os.Stdin = stdinReader
			os.Stdout = stdoutWriter

			t.Cleanup(func() {
				os.Stdin = oldStdin
				os.Stdout = oldStdout
				stdinReader.Close()
				stdinWriter.Close()
				stdoutReader.Close()
				stdoutWriter.Close()
			})

			_, err = stdinWriter.WriteString(tt.input)
			require.NoError(t, err)
			stdinWriter.Close()

			// Act
			result := Confirmf("Test prompt")

			// Assert
			require.True(t, result)
		})
	}
}

func TestConfirmf_NegativeResponses_ReturnsFalse(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "lowercase n",
			input: "n\n",
		},
		{
			name:  "lowercase no",
			input: "no\n",
		},
		{
			name:  "uppercase N",
			input: "N\n",
		},
		{
			name:  "uppercase NO",
			input: "NO\n",
		},
		{
			name:  "random text",
			input: "anything\n",
		},
		{
			name:  "empty response",
			input: "\n",
		},
		{
			name:  "only spaces",
			input: "   \n",
		},
		{
			name:  "partial match ya",
			input: "ya\n",
		},
		{
			name:  "partial match yep",
			input: "yep\n",
		},
		{
			name:  "partial match yeah",
			input: "yeah\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			oldStdin := os.Stdin
			oldStdout := os.Stdout

			stdinReader, stdinWriter, err := os.Pipe()
			require.NoError(t, err)

			stdoutReader, stdoutWriter, err := os.Pipe()
			require.NoError(t, err)

			os.Stdin = stdinReader
			os.Stdout = stdoutWriter

			t.Cleanup(func() {
				os.Stdin = oldStdin
				os.Stdout = oldStdout
				stdinReader.Close()
				stdinWriter.Close()
				stdoutReader.Close()
				stdoutWriter.Close()
			})

			_, err = stdinWriter.WriteString(tt.input)
			require.NoError(t, err)
			stdinWriter.Close()

			// Act
			result := Confirmf("Test prompt")

			// Assert
			require.False(t, result)
		})
	}
}

func TestConfirm_DelegatesToConfirmf_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "affirmative response",
			input:    "y\n",
			expected: true,
		},
		{
			name:     "negative response",
			input:    "n\n",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			oldStdin := os.Stdin
			oldStdout := os.Stdout

			stdinReader, stdinWriter, err := os.Pipe()
			require.NoError(t, err)

			stdoutReader, stdoutWriter, err := os.Pipe()
			require.NoError(t, err)

			os.Stdin = stdinReader
			os.Stdout = stdoutWriter

			t.Cleanup(func() {
				os.Stdin = oldStdin
				os.Stdout = oldStdout
				stdinReader.Close()
				stdinWriter.Close()
				stdoutReader.Close()
				stdoutWriter.Close()
			})

			_, err = stdinWriter.WriteString(tt.input)
			require.NoError(t, err)
			stdinWriter.Close()

			// Act
			result := Confirm("Test prompt")

			// Assert
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestConfirmf_WithFormatArgs_Successfully(t *testing.T) {
	// Arrange
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	stdinReader, stdinWriter, err := os.Pipe()
	require.NoError(t, err)

	stdoutReader, stdoutWriter, err := os.Pipe()
	require.NoError(t, err)

	os.Stdin = stdinReader
	os.Stdout = stdoutWriter

	t.Cleanup(func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		stdinReader.Close()
		stdinWriter.Close()
		stdoutReader.Close()
		stdoutWriter.Close()
	})

	_, err = stdinWriter.WriteString("yes\n")
	require.NoError(t, err)
	stdinWriter.Close()

	// Act
	result := Confirmf("Apply migration %s version %d", "test_migration", 42)

	// Assert
	require.True(t, result)
}
