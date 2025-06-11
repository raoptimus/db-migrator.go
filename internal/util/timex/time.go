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

type Time interface {
	Now() time.Time
}

var StdTime = New(time.Now)

type stdTime struct {
	nowFunc func() time.Time
}

//nolint:ireturn,nolintlint // its ok
func New(nowFunc func() time.Time) Time {
	return &stdTime{nowFunc: nowFunc}
}

func (s *stdTime) Now() time.Time {
	return s.nowFunc()
}
