/**
 * This file is part of the raoptimus/db-migrator.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/db-migrator.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/db-migrator.go
 */

package timex

import (
	"time"
)

// Time defines the interface for obtaining the current time.
// This abstraction allows for easier testing by enabling time mocking.
type Time interface {
	Now() time.Time
}

// StdTime is the standard Time instance that uses time.Now.
var StdTime = New(time.Now)

// stdTime implements the Time interface using a custom time function.
type stdTime struct {
	nowFunc func() time.Time
}

// New creates a new Time instance that uses the provided function to get the current time.
// This allows for dependency injection of time for testing purposes.
//
//nolint:ireturn,nolintlint // its ok
func New(nowFunc func() time.Time) Time {
	return &stdTime{nowFunc: nowFunc}
}

// Now returns the current time by calling the configured time function.
func (s *stdTime) Now() time.Time {
	return s.nowFunc()
}
