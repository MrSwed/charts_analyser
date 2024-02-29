package domain

type InputVessels struct {
	VesselIDs VesselIDs `json:"vesselIDs"`
}

type InputVesselsInterval struct {
	InputVessels
	DateInterval
}

type InputZones struct {
	ZoneNames []ZoneName `json:"zoneNames"`
	DateInterval
}

type InputVessel struct {
	VesselName `json:"vesselName,omitempty"`
}

type InputPoint []float64
