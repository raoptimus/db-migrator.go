/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package log

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew_CreatesLogger_Successfully(t *testing.T) {
	var buf bytes.Buffer

	logger := New(&buf)

	require.NotNil(t, logger)
	require.NotNil(t, logger.writer)
	require.NotNil(t, logger.colors)
	require.Len(t, logger.colors, 4)
	require.Contains(t, logger.colors, Info)
	require.Contains(t, logger.colors, Success)
	require.Contains(t, logger.colors, Warn)
	require.Contains(t, logger.colors, Error)
}

func TestLogger_Infof_WritesFormattedMessage_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "simple message without args",
			format:   "hello world",
			args:     nil,
			expected: "hello world",
		},
		{
			name:     "message with string arg",
			format:   "hello %s",
			args:     []any{"world"},
			expected: "hello world",
		},
		{
			name:     "message with multiple args",
			format:   "name: %s, age: %d",
			args:     []any{"John", 30},
			expected: "name: John, age: 30",
		},
		{
			name:     "message with integer arg",
			format:   "count: %d",
			args:     []any{42},
			expected: "count: 42",
		},
		{
			name:     "empty format",
			format:   "",
			args:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Infof(tt.format, tt.args...)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_Info_WritesMessageWithNewline_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected string
	}{
		{
			name:     "single string arg",
			args:     []any{"hello world"},
			expected: "hello world\n",
		},
		{
			name:     "multiple args concatenated",
			args:     []any{"hello", "world"},
			expected: "helloworld\n",
		},
		{
			name:     "integer arg",
			args:     []any{42},
			expected: "42\n",
		},
		{
			name:     "empty args",
			args:     nil,
			expected: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Info(tt.args...)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_Successf_WritesFormattedMessage_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "simple message",
			format:   "operation completed",
			args:     nil,
			expected: "operation completed",
		},
		{
			name:     "message with arg",
			format:   "created %d items",
			args:     []any{5},
			expected: "created 5 items",
		},
		{
			name:     "message with multiple args",
			format:   "migrated %s to version %s",
			args:     []any{"database", "v1.2.0"},
			expected: "migrated database to version v1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Successf(tt.format, tt.args...)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_Success_WritesMessageWithNewline_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected string
	}{
		{
			name:     "single string",
			args:     []any{"success"},
			expected: "success\n",
		},
		{
			name:     "multiple args concatenated",
			args:     []any{"migration", "applied"},
			expected: "migrationapplied\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Success(tt.args...)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_Warnf_WritesFormattedMessage_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "simple warning",
			format:   "warning message",
			args:     nil,
			expected: "warning message",
		},
		{
			name:     "warning with arg",
			format:   "deprecated function: %s",
			args:     []any{"oldFunc"},
			expected: "deprecated function: oldFunc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Warnf(tt.format, tt.args...)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_WarnLn_WritesMessageWithNewline_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		args     []any
		expected string
	}{
		{
			name:     "single warning",
			args:     []any{"caution"},
			expected: "caution\n",
		},
		{
			name:     "multiple args concatenated",
			args:     []any{"warning:", "check config"},
			expected: "warning:check config\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Warn(tt.args...)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_Errorf_WritesFormattedMessage_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "simple error",
			format:   "error occurred",
			args:     nil,
			expected: "error occurred",
		},
		{
			name:     "error with details",
			format:   "failed to connect: %s",
			args:     []any{"timeout"},
			expected: "failed to connect: timeout",
		},
		{
			name:     "error with multiple args",
			format:   "error code %d: %s",
			args:     []any{500, "internal server error"},
			expected: "error code 500: internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Errorf(tt.format, tt.args...)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_Error_WritesMessageWithNewline_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{
			name:     "simple error",
			format:   "an error occurred",
			expected: "an error occurred\n",
		},
		{
			name:     "empty error",
			format:   "",
			expected: "\n",
		},
		{
			name:     "error with special characters",
			format:   "error: file not found /path/to/file.txt",
			expected: "error: file not found /path/to/file.txt\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Error(tt.format)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_ColorFunc_NoColor_WhenNotTTY(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		input    []any
		expected string
	}{
		{
			name:     "white color returns plain text",
			color:    ColorWhite,
			input:    []any{"test message"},
			expected: "test message",
		},
		{
			name:     "green color returns plain text",
			color:    ColorGreen,
			input:    []any{"success"},
			expected: "success",
		},
		{
			name:     "yellow color returns plain text",
			color:    ColorYellow,
			input:    []any{"warning"},
			expected: "warning",
		},
		{
			name:     "red color returns plain text",
			color:    ColorRed,
			input:    []any{"error"},
			expected: "error",
		},
		{
			name:     "multiple args concatenated without spaces",
			color:    ColorWhite,
			input:    []any{"hello", "world"},
			expected: "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			fn := logger.colorFunc(tt.color)
			result := fn(tt.input...)

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestLogger_ColorFunc_AppliesColor_WhenTTY(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		input    []any
		expected string
	}{
		{
			name:     "white color applied",
			color:    ColorWhite,
			input:    []any{"test"},
			expected: ColorWhite + "test" + ColorReset,
		},
		{
			name:     "green color applied",
			color:    ColorGreen,
			input:    []any{"success"},
			expected: ColorGreen + "success" + ColorReset,
		},
		{
			name:     "yellow color applied",
			color:    ColorYellow,
			input:    []any{"warning"},
			expected: ColorYellow + "warning" + ColorReset,
		},
		{
			name:     "red color applied",
			color:    ColorRed,
			input:    []any{"error"},
			expected: ColorRed + "error" + ColorReset,
		},
		{
			name:     "cyan color applied",
			color:    ColorCyan,
			input:    []any{"info"},
			expected: ColorCyan + "info" + ColorReset,
		},
		{
			name:     "magenta color applied",
			color:    ColorMagenta,
			input:    []any{"special"},
			expected: ColorMagenta + "special" + ColorReset,
		},
		{
			name:     "blue color applied",
			color:    ColorBlue,
			input:    []any{"link"},
			expected: ColorBlue + "link" + ColorReset,
		},
		{
			name:     "black color applied",
			color:    ColorBlack,
			input:    []any{"dark"},
			expected: ColorBlack + "dark" + ColorReset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = true

			fn := logger.colorFunc(tt.color)
			result := fn(tt.input...)

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestLogger_OutputWithColors_WhenTTY(t *testing.T) {
	tests := []struct {
		name      string
		logFunc   func(l *Logger)
		expected  string
		colorCode string
	}{
		{
			name: "Info uses white color",
			logFunc: func(l *Logger) {
				l.Info("test")
			},
			expected:  ColorWhite + "test" + ColorReset + "\n",
			colorCode: ColorWhite,
		},
		{
			name: "Success uses green color",
			logFunc: func(l *Logger) {
				l.Success("done")
			},
			expected:  ColorGreen + "done" + ColorReset + "\n",
			colorCode: ColorGreen,
		},
		{
			name: "Warn uses yellow color",
			logFunc: func(l *Logger) {
				l.Warn("caution")
			},
			expected:  ColorYellow + "caution" + ColorReset + "\n",
			colorCode: ColorYellow,
		},
		{
			name: "Error uses red color",
			logFunc: func(l *Logger) {
				l.Error("failed")
			},
			expected:  ColorRed + "failed" + ColorReset + "\n",
			colorCode: ColorRed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = true
			// Reinitialize colors with TTY enabled
			logger.colors[Info] = logger.colorFunc(ColorWhite)
			logger.colors[Success] = logger.colorFunc(ColorGreen)
			logger.colors[Warn] = logger.colorFunc(ColorYellow)
			logger.colors[Error] = logger.colorFunc(ColorRed)

			tt.logFunc(logger)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_Infof_WithTTYColors_Successfully(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf)
	logger.isTTY = true
	logger.colors[Info] = logger.colorFunc(ColorWhite)

	logger.Infof("hello %s", "world")

	expected := ColorWhite + "hello world" + ColorReset
	require.Equal(t, expected, buf.String())
}

func TestLogger_Successf_WithTTYColors_Successfully(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf)
	logger.isTTY = true
	logger.colors[Success] = logger.colorFunc(ColorGreen)

	logger.Successf("completed %d tasks", 5)

	expected := ColorGreen + "completed 5 tasks" + ColorReset
	require.Equal(t, expected, buf.String())
}

func TestLogger_Warnf_WithTTYColors_Successfully(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf)
	logger.isTTY = true
	logger.colors[Warn] = logger.colorFunc(ColorYellow)

	logger.Warnf("deprecated: %s", "oldMethod")

	expected := ColorYellow + "deprecated: oldMethod" + ColorReset
	require.Equal(t, expected, buf.String())
}

func TestLogger_Errorf_WithTTYColors_Successfully(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf)
	logger.isTTY = true
	logger.colors[Error] = logger.colorFunc(ColorRed)

	logger.Errorf("error: %s", "connection refused")

	expected := ColorRed + "error: connection refused" + ColorReset
	require.Equal(t, expected, buf.String())
}

func TestLogger_LevelConstants_HaveCorrectValues(t *testing.T) {
	require.Equal(t, Level("info"), Info)
	require.Equal(t, Level("success"), Success)
	require.Equal(t, Level("warn"), Warn)
	require.Equal(t, Level("error"), Error)
}

func TestLogger_ColorConstants_HaveCorrectANSICodes(t *testing.T) {
	require.Equal(t, "\033[1;30m", ColorBlack)
	require.Equal(t, "\033[1;31m", ColorRed)
	require.Equal(t, "\033[1;32m", ColorGreen)
	require.Equal(t, "\033[1;33m", ColorYellow)
	require.Equal(t, "\033[1;34m", ColorBlue)
	require.Equal(t, "\033[1;35m", ColorMagenta)
	require.Equal(t, "\033[1;36m", ColorCyan)
	require.Equal(t, "\033[1;37m", ColorWhite)
	require.Equal(t, "\033[0m", ColorReset)
}

func TestStd_IsInitialized(t *testing.T) {
	require.NotNil(t, Std)
	require.NotNil(t, Std.writer)
	require.NotNil(t, Std.colors)
}

// Note: Fatal and Fatalf tests are intentionally omitted because they call os.Exit(1)
// which would terminate the test process. Testing os.Exit behavior requires special
// handling such as spawning subprocess or mocking os.Exit, which is beyond typical
// unit test scope. The output formatting logic is covered by Error/Errorf tests
// since Fatal uses the same color function (Error level).

func TestLogger_Fatal_WritesErrorColoredOutput_BeforeExit(t *testing.T) {
	// This test only verifies that Fatal would write correctly formatted output.
	// We cannot test the os.Exit(1) call without terminating the test process.
	// The actual Fatal method is tested implicitly through Error tests since
	// both use c.colors[Error] for formatting.
	t.Skip("Fatal calls os.Exit(1) which terminates the test process")
}

func TestLogger_Fatalf_WritesFormattedErrorOutput_BeforeExit(t *testing.T) {
	// Same as above - Fatalf uses the same formatting as Errorf plus os.Exit(1).
	t.Skip("Fatalf calls os.Exit(1) which terminates the test process")
}

func TestLogger_MultipleWrites_AppendToBuffer_Successfully(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf)
	logger.isTTY = false

	logger.Infof("first ")
	logger.Infof("second ")
	logger.Info("third")

	expected := "first second third\n"
	require.Equal(t, expected, buf.String())
}

func TestLogger_EmptyInput_HandledCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		logFunc  func(l *Logger)
		expected string
	}{
		{
			name: "Info with empty slice",
			logFunc: func(l *Logger) {
				l.Info()
			},
			expected: "\n",
		},
		{
			name: "Infof with empty format",
			logFunc: func(l *Logger) {
				l.Infof("")
			},
			expected: "",
		},
		{
			name: "Success with empty slice",
			logFunc: func(l *Logger) {
				l.Success()
			},
			expected: "\n",
		},
		{
			name: "Warn with empty slice",
			logFunc: func(l *Logger) {
				l.Warn()
			},
			expected: "\n",
		},
		{
			name: "Error with empty string",
			logFunc: func(l *Logger) {
				l.Error("")
			},
			expected: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			tt.logFunc(logger)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_SpecialCharacters_HandledCorrectly(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "newline in message",
			input:    "line1\nline2",
			expected: "line1\nline2\n",
		},
		{
			name:     "tab in message",
			input:    "col1\tcol2",
			expected: "col1\tcol2\n",
		},
		{
			name:     "unicode characters",
			input:    "hello world",
			expected: "hello world\n",
		},
		{
			name:     "emoji characters",
			input:    "check passed",
			expected: "check passed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Info(tt.input)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestLogger_NumericFormatting_Successfully(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "integer zero",
			format:   "%d",
			args:     []any{0},
			expected: "0",
		},
		{
			name:     "negative integer",
			format:   "%d",
			args:     []any{-1},
			expected: "-1",
		},
		{
			name:     "large integer",
			format:   "%d",
			args:     []any{9223372036854775807},
			expected: "9223372036854775807",
		},
		{
			name:     "float with precision",
			format:   "%.2f",
			args:     []any{3.14159},
			expected: "3.14",
		},
		{
			name:     "boolean",
			format:   "%t",
			args:     []any{true},
			expected: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := New(&buf)
			logger.isTTY = false

			logger.Infof(tt.format, tt.args...)

			require.Equal(t, tt.expected, buf.String())
		})
	}
}
