package constant

import "time"

const (
	ServerShutdownTimeout  = 30 * time.Second
	ServerOperationTimeout = 30 * time.Second

	ServerAddress = "localhost:8080"

	EnvNameServerAddress = "ADDRESS"
	EnvNameDBDSN         = "DATABASE_DSN"
	EnvNameRedisAddress  = "REDIS_ADDRESS"
	EnvNameRedisPass     = "REDIS_PASS"

	DBZones   = "zones"
	DBTracks  = "tracks"
	DBVessels = "vessels"

	RouteID      = "/:id"
	RouteApi     = "/api"
	RouteVessels = "/vessels"
	RouteZones   = "/zones"
	RouteMonitor = "/monitor"
	RouteTrack   = "/track"

	MonitorLastPeriod = 30 * time.Second

	RedisVeselPrefix    = "vessel:"
	RedisControlIds     = "_control:ids"
	RedisControlVessels = "_control:vessels"
)

var GeoAllowedRange = [4]float64{-180, -75, 180, 75}
