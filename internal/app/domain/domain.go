package domain

import (
	"charts_analyser/internal/app/constant"
	"time"
)

type InputDates struct {
	Start time.Time `json:"start_time" form:"start_time"`
	End   time.Time `json:"end_time" form:"end_time"`
}

// EndOrNow return now if not set
func (i *InputDates) EndOrNow() time.Time {
	if i.End.IsZero() {
		return time.Now()
	}
	return i.End
}

// StartOrNow return zero, if end time is set and return now subtract monitoring interval if not: monitoring mode
func (i *InputDates) StartOrNow() time.Time {
	if i.Start.IsZero() && i.End.IsZero() {
		return time.Now().Add(-constant.MonitorCurrentInterval)
	}
	return i.Start
}

type InputVessels struct {
	VesselIDs []int64 `json:"vessel_id" form:"vessel_id"`
	InputDates
}

type InputZone struct {
	ZoneName string `json:"zone_name" form:"zone_name"`
	InputDates
}
