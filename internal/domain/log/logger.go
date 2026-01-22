/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package log

// Logger defines the logging interface used by domain services.
// This interface is implemented by infrastructure/log package.
//
//go:generate mockery
type Logger interface {
	Info(a ...any)
	Infof(format string, args ...any)
	Success(a ...any)
	Successf(format string, args ...any)
	Warn(a ...any)
	Warnf(format string, args ...any)
	Error(a ...any)
	Errorf(format string, args ...any)
}
