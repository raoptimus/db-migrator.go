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

func New(nowFunc func() time.Time) Time {
	return &stdTime{nowFunc: nowFunc}
}

func (s *stdTime) Now() time.Time {
	return s.nowFunc()
}
