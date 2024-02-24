package domain

import (
	"charts_analyser/internal/app/constant"
	"time"
)

type DateInterval struct {
	Start  time.Time `json:"start" form:"start"`
	Finish time.Time `json:"finish" form:"finish"`
}

// EndOrNow return now if not set
func (i *DateInterval) EndOrNow() time.Time {
	if i.Finish.IsZero() {
		return time.Now()
	}
	return i.Finish
}

// StartOrLastPeriod return zero, if end time is set and return now subtract monitoring interval if not: monitoring mode
func (i *DateInterval) StartOrLastPeriod() time.Time {
	if i.Start.IsZero() && i.Finish.IsZero() {
		return time.Now().Add(-constant.MonitorLastPeriod)
	}
	return i.Start
}
