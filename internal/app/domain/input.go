package domain

type InputVessels struct {
	VesselIDs []VesselID `json:"vessel_id" form:"vessel_id"`
	DateInterval
}

type InputZone struct {
	ZoneName `json:"zone_name" form:"zone_name"`
	DateInterval
}

type InputVessel struct {
	VesselName `json:"vessel_name,omitempty" form:"vessel_name,omitempty"`
}
