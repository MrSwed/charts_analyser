package domain

import (
	"charts_analyser/internal/app/constant"
	"time"
)

type DateInterval struct {
	Start  *time.Time `json:"start" form:"start"`
	Finish *time.Time `json:"finish" form:"finish"`
}

// FinishOrNow return now if not set
func (i *DateInterval) FinishOrNow() time.Time {
	if i.Finish == nil || i.Finish.IsZero() {
		return time.Now()
	}
	return *i.Finish
}

// StartOrLastPeriod return zero, if end time is set and return now subtract monitoring interval if not: monitoring mode
func (i *DateInterval) StartOrLastPeriod() time.Time {
	if (i.Start == nil || i.Start.IsZero()) &&
		(i.Finish == nil || i.Finish.IsZero()) {
		return time.Now().Add(-constant.MonitorLastPeriod)
	}
	if i.Start == nil {
		return time.Time{}
	}
	return *i.Start
}
