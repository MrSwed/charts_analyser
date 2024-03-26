package constant

const (
	EnvNameServerAddress  = "ADDRESS"
	EnvNameTrackInterval  = "TRACK_INTERVAL"
	EnvNameVesselCount    = "VESSEL_COUNT"
	EnvNameSleepBeforeRun = "SLEEP_BEFORE_RUN"

	DefaultNumVessels         = 5
	DefaultTrackInterval      = 10
	DefaultTracksSecondsCache = 1 * 60 * 60 * 24
	DefaultTracksItemsCache   = 50
	DefaultSleepBeforeRun     = 10

	RouteTrack   = "/api/track"
	RouteMonitor = "/api/monitor"

	CtxValueKeyJWTOperator CtxKey = "jwt_operator"
	CtxValueKeyJWTVessel   CtxKey = "jwt_vessel"
)

type CtxKey string

func (c CtxKey) String() string {
	return string(c)
}
