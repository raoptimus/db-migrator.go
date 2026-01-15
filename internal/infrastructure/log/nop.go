package log

// NopLogger is a no-op implementation of the logger interface that discards all log messages.
type NopLogger struct{}

// Infof discards the formatted informational message.
func (c *NopLogger) Infof(_ string, _ ...any) {}

// Info discards the informational message.
func (c *NopLogger) Info(_ ...any) {}

// Successf discards the formatted success message.
func (c *NopLogger) Successf(_ string, _ ...any) {}

// Success discards the success message.
func (c *NopLogger) Success(_ ...any) {}

// Warnf discards the formatted warning message.
func (c *NopLogger) Warnf(_ string, _ ...any) {}

// Warn discards the warning message.
func (c *NopLogger) Warn(_ ...any) {}

// Errorf discards the formatted error message.
func (c *NopLogger) Errorf(_ string, _ ...any) {}

// Error discards the error message.
func (c *NopLogger) Error(_ ...any) {}

// Fatalf discards the formatted error message (no-op, does not exit).
func (c *NopLogger) Fatalf(_ string, _ ...any) {}

// Fatal discards the error message (no-op, does not exit).
func (c *NopLogger) Fatal(_ ...any) {}
