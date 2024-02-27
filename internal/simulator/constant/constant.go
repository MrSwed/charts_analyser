package constant

const (
	EnvNameServerAddress  = "ADDRESS"
	EnvNameDBDSN          = "DATABASE_DSN"
	EnvNameTrackInterval  = "TRACK_INTERVAL"
	EnvNameVesselCount    = "VESSEL_COUNT"
	EnvNameSleepBeforeRun = "SLEEP_BEFORE_RUN"
	EnvNameJWTSecretKey   = "JWT_SECRET_KEY"

	DefaultNumVessels         = 5
	DefaultTrackInterval      = 10
	DefaultTracksSecondsCache = 1 * 60 * 60 * 24
	DefaultTracksItemsCache   = 50
	DefaultSleepBeforeRun     = 10

	RouteTrack   = "/api/track"
	RouteMonitor = "/api/monitor"

	JWTSigningKey = "JWTSigningKeyDefaultString"

	CtxValueKeyJWTOperator = "jwt_operator"
	CtxValueKeyJWTVessel   = "jwt_vessel"
)
