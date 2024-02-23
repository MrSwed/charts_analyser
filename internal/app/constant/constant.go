package constant

import "time"

const (
	ServerShutdownTimeout  = 30 * time.Second
	ServerOperationTimeout = 30 * time.Second

	ServerAddress = "localhost:8080"

	EnvNameServerAddress = "ADDRESS"
	EnvNameDBDSN         = "DATABASE_DSN"

	DBZones  = "zones"
	DBTracks = "tracks"

	RouteApi     = "/api"
	RouteVessels = "/vessels"
	RouteZones   = "/zones"
)
