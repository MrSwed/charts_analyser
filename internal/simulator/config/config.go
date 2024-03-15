package config

import (
	appConfig "charts_analyser/internal/app/config"
	appConstant "charts_analyser/internal/app/constant"
	"charts_analyser/internal/simulator/constant"
	"flag"
	"os"
	"strconv"
	"strings"
)

type Out struct {
	ServerAddress string
}

type Config struct {
	Out
	DatabaseDSN    string
	VesselCount    uint
	TrackInterval  uint
	SleepBeforeRun uint
	appConfig.JWT
}

func NewConfig() *Config {
	return &Config{
		VesselCount:    constant.DefaultNumVessels,
		TrackInterval:  constant.DefaultTrackInterval,
		SleepBeforeRun: constant.DefaultSleepBeforeRun,
		JWT: appConfig.JWT{
			JWTSigningKey:       appConstant.JWTSigningKey,
			TokenLifeTime:       appConstant.TokenLifeTime,
			TokenVesselLifeTime: appConstant.TokenVesselLifeTime,
		},
	}
}

func (c *Config) Init() *Config {
	return c.withFlags().WithEnv().CleanParameters()
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the server address for send data, can set with env "+constant.EnvNameServerAddress)
	flag.UintVar(&c.TrackInterval, "i", c.TrackInterval, "Provide the interval between track send, 0 - mean send same time from history source "+constant.EnvNameTrackInterval)
	flag.UintVar(&c.VesselCount, "c", c.VesselCount, "Provide the count of simulated vessels "+constant.EnvNameVesselCount)
	flag.UintVar(&c.SleepBeforeRun, "s", c.SleepBeforeRun, "Provide the time sleep before run "+constant.EnvNameSleepBeforeRun)

	// main app config
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Database dsn connect string, can set with env "+appConstant.EnvNameDBDSN)
	flag.StringVar(&c.JWTSigningKey, "j", c.JWTSigningKey, "Provide the jwt secret key "+appConstant.EnvNameJWTSecretKey)
	flag.Uint64Var(&c.TokenLifeTime, "jlt", c.TokenLifeTime, "Provide the jwt token lifetime, sec "+appConstant.EnvNameJWTLifeTime)
	flag.Uint64Var(&c.TokenVesselLifeTime, "jltv", c.TokenVesselLifeTime, "Provide the vessel jwt token lifetime, sec "+appConstant.EnvNameJWTVesselLifeTime)
	flag.Parse()
	return c
}

func (c *Config) WithEnv() *Config {
	if env, ok := os.LookupEnv(constant.EnvNameServerAddress); ok && env != "" {
		c.ServerAddress = env
	}
	if env, ok := os.LookupEnv(appConstant.EnvNameDBDSN); ok {
		c.DatabaseDSN = env
	}
	if env, ok := os.LookupEnv(constant.EnvNameTrackInterval); ok {
		if env, err := strconv.ParseUint(env, 10, 64); err == nil {
			c.TrackInterval = uint(env)
		}
	}
	if env, ok := os.LookupEnv(constant.EnvNameVesselCount); ok {
		if env, err := strconv.ParseUint(env, 10, 64); err == nil {
			c.VesselCount = uint(env)
		}
	}
	if env, ok := os.LookupEnv(constant.EnvNameSleepBeforeRun); ok {
		if env, err := strconv.ParseUint(env, 10, 64); err == nil {
			c.SleepBeforeRun = uint(env)
		}
	}
	// main app config
	if jwt, ok := os.LookupEnv(appConstant.EnvNameJWTSecretKey); ok && jwt != "" {
		c.JWTSigningKey = jwt
	}
	if jwtLt, ok := os.LookupEnv(appConstant.EnvNameJWTLifeTime); ok && jwtLt != "" {
		if v, err := strconv.ParseUint(jwtLt, 10, 64); err == nil {
			c.TokenLifeTime = v
		}
	}
	if jwtVLt, ok := os.LookupEnv(appConstant.EnvNameJWTVesselLifeTime); ok && jwtVLt != "" {
		if v, err := strconv.ParseUint(jwtVLt, 10, 64); err == nil {
			c.TokenVesselLifeTime = v
		}
	}
	return c
}

func (c *Config) CleanParameters() *Config {
	for _, v := range []string{"http://", "https://"} {
		c.ServerAddress = strings.TrimPrefix(c.ServerAddress, v)
	}
	c.ServerAddress = "http://" + c.ServerAddress
	c.DatabaseDSN = strings.Trim(c.DatabaseDSN, "'")
	return c
}
