package domain

type InputVessels struct {
	VesselIDs VesselIDs `json:"vesselIDs"`
}

type InputVesselsInterval struct {
	InputVessels
	DateInterval
}

type InputZone struct {
	ZoneName ZoneName `json:"zoneName"`
	DateInterval
}

type InputVessel struct {
	VesselName `json:"vesselName,omitempty"`
}
