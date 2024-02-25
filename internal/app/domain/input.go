package domain

type InputVessels struct {
	VesselIDs VesselIDs `json:"vessel_id" form:"vessel_id"`
}

type InputVesselsInterval struct {
	InputVessels
	DateInterval
}

type InputZone struct {
	ZoneName `json:"zone_name" form:"zone_name"`
	DateInterval
}

type InputVessel struct {
	VesselName `json:"vessel_name,omitempty" form:"vessel_name,omitempty"`
}
