package constant

import "time"

const (
	ServerShutdownTimeout  = 30 * time.Second
	ServerOperationTimeout = 30 * time.Second

	ServerAddress = "localhost:8080"

	LogFormat = "[${time}] ${status} - ${latency} ${method} ${path}\n"

	MonitorLastPeriod = 30 * time.Second
)

var GeoAllowedRange = [4]float64{-180, -75, 180, 75}
