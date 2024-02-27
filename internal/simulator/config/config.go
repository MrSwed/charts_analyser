package config

import (
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
	DatabaseDSN   string
	VesselCount   uint
	TrackInterval uint
	JWT
}

type JWT struct {
	JWTSigningKey string
}

func NewConfig() *Config {
	return &Config{
		VesselCount:   constant.DefaultNumVessels,
		TrackInterval: constant.DefaultTrackInterval,
		JWT: JWT{
			JWTSigningKey: constant.JWTSigningKey},
	}
}

func (c *Config) Init() *Config {
	return c.withFlags().WithEnv().CleanParameters()
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server, can set with env "+constant.EnvNameServerAddress)
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Database dsn connect string, can set with env "+constant.EnvNameDBDSN)
	flag.UintVar(&c.TrackInterval, "i", c.TrackInterval, "Provide the interval between track send, 0 - mean send same time from history source "+constant.EnvNameTrackInterval)
	flag.UintVar(&c.VesselCount, "c", c.VesselCount, "Provide the count of simulated vessels "+constant.EnvNameVesselCount)
	flag.StringVar(&c.JWTSigningKey, "j", c.JWTSigningKey, "Provide the jwt secret key "+constant.EnvNameJWTSecretKey)
	flag.Parse()
	return c
}

func (c *Config) WithEnv() *Config {
	if env, ok := os.LookupEnv(constant.EnvNameServerAddress); ok && env != "" {
		c.ServerAddress = env
	}
	if env, ok := os.LookupEnv(constant.EnvNameDBDSN); ok {
		c.DatabaseDSN = env
	}
	if env, ok := os.LookupEnv(constant.EnvNameTrackInterval); ok {
		if env, err := strconv.ParseUint(env, 10, 64); err != nil {
			c.TrackInterval = uint(env)
		}
	}
	if env, ok := os.LookupEnv(constant.EnvNameVesselCount); ok {
		if env, err := strconv.ParseUint(env, 10, 64); err != nil {
			c.VesselCount = uint(env)
		}
	}
	if jwt, ok := os.LookupEnv(constant.EnvNameJWTSecretKey); ok && jwt != "" {
		c.JWTSigningKey = jwt
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
