package constant

const (
	EnvNameServerAddress = "ADDRESS"
	EnvNameDBDSN         = "DATABASE_DSN"
	EnvNameTrackInterval = "TRACK_INTERVAL"
	EnvNameVesselCount   = "VESSEL_COUNT"

	DefaultNumVessels         = 5
	DefaultTrackInterval      = 10
	DefaultTracksSecondsCache = 1 * 60 * 60 * 24
	DefaultTracksItemsCache   = 50

	RouteTrack   = "/api/track"
	RouteMonitor = "/api/monitor"
)
