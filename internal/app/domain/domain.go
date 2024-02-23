package domain

import "time"

type InputDates struct {
	Start time.Time `json:"start_time" form:"start_time"`
	End   time.Time `json:"end_time" form:"end_time"`
}

type InputVessels struct {
	VesselIDs []int64 `json:"vessel_id" form:"vessel_id"`
	InputDates
}

type InputZone struct {
	ZoneName string `json:"zone_name" form:"zone_name"`
	InputDates
}
