/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package log

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

// Std is the standard logger instance that writes to stdout.
var Std = New(os.Stdout)

// Logger provides color-formatted logging with support for different log levels and TTY detection.
type Logger struct {
	colors map[Level]colorFunc
	writer io.Writer
	isTTY  bool
}

const (
	// ColorBlack is the ANSI escape code for bold black text.
	ColorBlack = "\033[1;30m"
	// ColorRed is the ANSI escape code for bold red text.
	ColorRed = "\033[1;31m"
	// ColorGreen is the ANSI escape code for bold green text.
	ColorGreen = "\033[1;32m"
	// ColorYellow is the ANSI escape code for bold yellow text.
	ColorYellow = "\033[1;33m"
	// ColorBlue is the ANSI escape code for bold blue text.
	ColorBlue = "\033[1;34m"
	// ColorMagenta is the ANSI escape code for bold magenta text.
	ColorMagenta = "\033[1;35m"
	// ColorCyan is the ANSI escape code for bold cyan text.
	ColorCyan = "\033[1;36m"
	// ColorWhite is the ANSI escape code for bold white text.
	ColorWhite = "\033[1;37m"
	// ColorReset is the ANSI escape code to reset text color and formatting.
	ColorReset = "\033[0m"
)

// Level represents a logging level for categorizing log messages.
type Level string

const (
	// Info is the log level for informational messages.
	Info Level = "info"
	// Success is the log level for successful operation messages.
	Success Level = "success"
	// Warn is the log level for warning messages.
	Warn Level = "warn"
	// Error is the log level for error messages.
	Error Level = "error"
)

type colorFunc func(a ...any) string

// New creates and returns a new Logger instance that writes to the provided io.Writer.
func New(w io.Writer) *Logger {
	c := &Logger{
		writer: w,
		isTTY:  term.IsTerminal(int(os.Stdout.Fd())),
		colors: make(map[Level]colorFunc),
	}

	c.colors[Info] = c.colorFunc(ColorWhite)
	c.colors[Success] = c.colorFunc(ColorGreen)
	c.colors[Warn] = c.colorFunc(ColorYellow)
	c.colors[Error] = c.colorFunc(ColorRed)

	return c
}

// Infof logs a formatted informational message.
func (c *Logger) Infof(format string, args ...any) {
	_, _ = fmt.Fprint(c.writer, c.colors[Info](fmt.Sprintf(format, args...)))
}

// Info logs an informational message.
func (c *Logger) Info(a ...any) {
	_, _ = fmt.Fprintln(c.writer, c.colors[Info](a...))
}

// Successf logs a formatted success message.
func (c *Logger) Successf(format string, args ...any) {
	_, _ = fmt.Fprint(c.writer, c.colors[Success](fmt.Sprintf(format, args...)))
}

// Success logs a success message.
func (c *Logger) Success(a ...any) {
	_, _ = fmt.Fprintln(c.writer, c.colors[Success](a...))
}

// Warnf logs a formatted warning message.
func (c *Logger) Warnf(format string, args ...any) {
	_, _ = fmt.Fprint(c.writer, c.colors[Warn](fmt.Sprintf(format, args...)))
}

// Warn logs a warning message.
func (c *Logger) Warn(a ...any) {
	_, _ = fmt.Fprintln(c.writer, c.colors[Warn](a...))
}

// Error logs an error message.
func (c *Logger) Error(a ...any) {
	_, _ = fmt.Fprintln(c.writer, c.colors[Error](a...))
}

// Errorf logs a formatted error message.
func (c *Logger) Errorf(format string, args ...any) {
	_, _ = fmt.Fprint(c.writer, c.colors[Error](fmt.Sprintf(format, args...)))
}

// Fatal logs an error message and exits the program with code 1.
func (c *Logger) Fatal(a ...any) {
	_, _ = fmt.Fprintln(c.writer, c.colors[Error](a...))
	os.Exit(1)
}

// Fatalf logs a formatted error message and exits the program with code 1.
func (c *Logger) Fatalf(format string, args ...any) {
	_, _ = fmt.Fprint(c.writer, c.colors[Error](fmt.Sprintf(format, args...)))
	os.Exit(1)
}

func (c *Logger) colorFunc(code string) colorFunc {
	if !c.isTTY {
		return fmt.Sprint
	}
	return func(a ...any) string {
		return fmt.Sprintf(code+"%s"+ColorReset, fmt.Sprint(a...))
	}
}
