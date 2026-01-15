/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package log

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNopLogger_Infof_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
	}{
		{
			name:   "empty format without args",
			format: "",
			args:   nil,
		},
		{
			name:   "simple message without args",
			format: "hello world",
			args:   nil,
		},
		{
			name:   "message with string arg",
			format: "hello %s",
			args:   []any{"world"},
		},
		{
			name:   "message with multiple args",
			format: "name: %s, age: %d",
			args:   []any{"John", 30},
		},
		{
			name:   "message with integer arg",
			format: "count: %d",
			args:   []any{42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Infof(tt.format, tt.args...)
			})
		})
	}
}

func TestNopLogger_Info_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "empty args",
			args: nil,
		},
		{
			name: "single string arg",
			args: []any{"hello world"},
		},
		{
			name: "multiple string args",
			args: []any{"hello", "world"},
		},
		{
			name: "integer arg",
			args: []any{42},
		},
		{
			name: "mixed type args",
			args: []any{"message", 123, true, 3.14},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Info(tt.args...)
			})
		})
	}
}

func TestNopLogger_Successf_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
	}{
		{
			name:   "empty format without args",
			format: "",
			args:   nil,
		},
		{
			name:   "simple message",
			format: "operation completed",
			args:   nil,
		},
		{
			name:   "message with arg",
			format: "created %d items",
			args:   []any{5},
		},
		{
			name:   "message with multiple args",
			format: "migrated %s to version %s",
			args:   []any{"database", "v1.2.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Successf(tt.format, tt.args...)
			})
		})
	}
}

func TestNopLogger_Success_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "empty args",
			args: nil,
		},
		{
			name: "single string",
			args: []any{"success"},
		},
		{
			name: "multiple args",
			args: []any{"migration", "applied"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Success(tt.args...)
			})
		})
	}
}

func TestNopLogger_Warnf_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
	}{
		{
			name:   "empty format without args",
			format: "",
			args:   nil,
		},
		{
			name:   "simple warning",
			format: "warning message",
			args:   nil,
		},
		{
			name:   "warning with arg",
			format: "deprecated function: %s",
			args:   []any{"oldFunc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Warnf(tt.format, tt.args...)
			})
		})
	}
}

func TestNopLogger_Warn_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "empty args",
			args: nil,
		},
		{
			name: "single warning",
			args: []any{"caution"},
		},
		{
			name: "multiple args",
			args: []any{"warning:", "check config"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Warn(tt.args...)
			})
		})
	}
}

func TestNopLogger_Errorf_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
	}{
		{
			name:   "empty format without args",
			format: "",
			args:   nil,
		},
		{
			name:   "simple error",
			format: "error occurred",
			args:   nil,
		},
		{
			name:   "error with details",
			format: "failed to connect: %s",
			args:   []any{"timeout"},
		},
		{
			name:   "error with multiple args",
			format: "error code %d: %s",
			args:   []any{500, "internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Errorf(tt.format, tt.args...)
			})
		})
	}
}

func TestNopLogger_Error_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "empty args",
			args: nil,
		},
		{
			name: "single error message",
			args: []any{"an error occurred"},
		},
		{
			name: "error with special characters",
			args: []any{"error: file not found /path/to/file.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Error(tt.args...)
			})
		})
	}
}

func TestNopLogger_Fatalf_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
	}{
		{
			name:   "empty format without args",
			format: "",
			args:   nil,
		},
		{
			name:   "simple fatal",
			format: "fatal error",
			args:   nil,
		},
		{
			name:   "fatal with arg",
			format: "fatal: %s",
			args:   []any{"shutdown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Fatalf(tt.format, tt.args...)
			})
		})
	}
}

func TestNopLogger_Fatal_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name string
		args []any
	}{
		{
			name: "empty args",
			args: nil,
		},
		{
			name: "single fatal message",
			args: []any{"fatal error occurred"},
		},
		{
			name: "multiple args",
			args: []any{"fatal:", "system shutdown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				logger.Fatal(tt.args...)
			})
		})
	}
}

func TestNopLogger_AllMethods_WithNilArgs_DoesNotPanic_Successfully(t *testing.T) {
	tests := []struct {
		name    string
		logFunc func(l *NopLogger)
	}{
		{
			name: "Infof with nil args",
			logFunc: func(l *NopLogger) {
				l.Infof("test", nil)
			},
		},
		{
			name: "Info with nil",
			logFunc: func(l *NopLogger) {
				l.Info(nil)
			},
		},
		{
			name: "Successf with nil args",
			logFunc: func(l *NopLogger) {
				l.Successf("test", nil)
			},
		},
		{
			name: "Success with nil",
			logFunc: func(l *NopLogger) {
				l.Success(nil)
			},
		},
		{
			name: "Warnf with nil args",
			logFunc: func(l *NopLogger) {
				l.Warnf("test", nil)
			},
		},
		{
			name: "Warn with nil",
			logFunc: func(l *NopLogger) {
				l.Warn(nil)
			},
		},
		{
			name: "Errorf with nil args",
			logFunc: func(l *NopLogger) {
				l.Errorf("test", nil)
			},
		},
		{
			name: "Error with nil",
			logFunc: func(l *NopLogger) {
				l.Error(nil)
			},
		},
		{
			name: "Fatalf with nil args",
			logFunc: func(l *NopLogger) {
				l.Fatalf("test", nil)
			},
		},
		{
			name: "Fatal with nil",
			logFunc: func(l *NopLogger) {
				l.Fatal(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &NopLogger{}

			require.NotPanics(t, func() {
				tt.logFunc(logger)
			})
		})
	}
}

func TestNopLogger_CanBeUsedAsInterface_Successfully(t *testing.T) {
	// This test verifies that NopLogger can be used in place of any logger interface
	// that expects the same method signatures.
	type LoggerInterface interface {
		Infof(format string, args ...any)
		Info(args ...any)
		Successf(format string, args ...any)
		Success(args ...any)
		Warnf(format string, args ...any)
		Warn(args ...any)
		Errorf(format string, args ...any)
		Error(args ...any)
		Fatalf(format string, args ...any)
		Fatal(args ...any)
	}

	var logger LoggerInterface = &NopLogger{}

	require.NotPanics(t, func() {
		logger.Info("test message")
		logger.Infof("test %s", "format")
		logger.Success("success message")
		logger.Successf("success %s", "format")
		logger.Warn("warn message")
		logger.Warnf("warn %s", "format")
		logger.Error("error message")
		logger.Errorf("error %s", "format")
		logger.Fatal("fatal message")
		logger.Fatalf("fatal %s", "format")
	})
}

func TestNopLogger_MultipleCalls_DoNotPanic_Successfully(t *testing.T) {
	logger := &NopLogger{}

	require.NotPanics(t, func() {
		for i := 0; i < 100; i++ {
			logger.Info("message", i)
			logger.Infof("iteration %d", i)
			logger.Success("done")
			logger.Successf("completed %d", i)
			logger.Warn("warning")
			logger.Warnf("warning %d", i)
			logger.Error("error")
			logger.Errorf("error %d", i)
		}
	})
}

func TestNopLogger_ZeroValue_DoesNotPanic_Successfully(t *testing.T) {
	// Test that zero-value NopLogger works correctly
	var logger NopLogger

	require.NotPanics(t, func() {
		logger.Info("test")
		logger.Infof("test %s", "arg")
		logger.Success("test")
		logger.Successf("test %s", "arg")
		logger.Warn("test")
		logger.Warnf("test %s", "arg")
		logger.Error("test")
		logger.Errorf("test %s", "arg")
		logger.Fatal("test")
		logger.Fatalf("test %s", "arg")
	})
}
