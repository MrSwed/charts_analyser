package config

import (
	"charts_analyser/internal/app/constant"
	"flag"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerAddress string
	DatabaseDSN   string
	JWT
}

type JWT struct {
	JWTSigningKey       string
	TokenLifeTime       uint64
	TokenVesselLifeTime uint64
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: constant.ServerAddress,
		JWT: JWT{
			JWTSigningKey:       constant.JWTSigningKey,
			TokenLifeTime:       constant.TokenLifeTime,
			TokenVesselLifeTime: constant.TokenVesselLifeTime,
		},
	}
}

func (c *Config) Init() *Config {
	return c.withFlags().WithEnv().CleanSchemes()
}

func (c *Config) WithEnv() *Config {
	if addr, ok := os.LookupEnv(constant.EnvNameServerAddress); ok && addr != "" {
		c.ServerAddress = addr
	}
	if dbDSN, ok := os.LookupEnv(constant.EnvNameDBDSN); ok {
		c.DatabaseDSN = dbDSN
	}
	if jwt, ok := os.LookupEnv(constant.EnvNameJWTSecretKey); ok && jwt != "" {
		c.JWTSigningKey = jwt
	}
	if jwtLt, ok := os.LookupEnv(constant.EnvNameJWTLifeTime); ok && jwtLt != "" {
		if v, err := strconv.ParseUint(jwtLt, 10, 64); err == nil {
			c.TokenLifeTime = v
		}
	}
	if jwtVLt, ok := os.LookupEnv(constant.EnvNameJWTVesselLifeTime); ok && jwtVLt != "" {
		if v, err := strconv.ParseUint(jwtVLt, 10, 64); err == nil {
			c.TokenVesselLifeTime = v
		}
	}
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server "+constant.EnvNameServerAddress)
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Provide the database dsn connect string "+constant.EnvNameDBDSN)
	flag.StringVar(&c.JWTSigningKey, "j", c.JWTSigningKey, "Provide the jwt secret key "+constant.EnvNameJWTSecretKey)
	flag.Uint64Var(&c.TokenLifeTime, "jlt", c.TokenLifeTime, "Provide the jwt token lifetime, sec "+constant.EnvNameJWTLifeTime)
	flag.Uint64Var(&c.TokenVesselLifeTime, "jltv", c.TokenVesselLifeTime, "Provide the vessel jwt token lifetime, sec "+constant.EnvNameJWTVesselLifeTime)
	flag.Parse()
	return c
}

func (c *Config) CleanSchemes() *Config {
	for _, v := range []string{"http://", "https://"} {
		c.ServerAddress = strings.TrimPrefix(c.ServerAddress, v)
	}
	c.DatabaseDSN = strings.Trim(c.DatabaseDSN, "'")
	return c
}
