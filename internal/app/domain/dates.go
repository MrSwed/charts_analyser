package domain

import (
	"charts_analyser/internal/app/constant"
	"time"
)

type DateInterval struct {
	Start  time.Time `json:"start" form:"start"`
	Finish time.Time `json:"finish" form:"finish"`
}

// FinishOrNow return now if not set
func (i *DateInterval) FinishOrNow() time.Time {
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

func (i *DateInterval) Period() time.Duration {
	return i.Finish.Sub(i.Start)
}

func (i *DateInterval) PeriodAuto() time.Duration {
	return i.FinishOrNow().Sub(i.StartOrLastPeriod())
}
