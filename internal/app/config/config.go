package config

import (
	"charts_analyser/internal/app/constant"
	"flag"
	"os"
	"strings"
)

type Config struct {
	ServerAddress string
	DatabaseDSN   string
	JWT
}

type JWT struct {
	JWTSigningKey string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: constant.ServerAddress,
		JWT: JWT{
			JWTSigningKey: constant.JWTSigningKey,
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
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server "+constant.EnvNameServerAddress)
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Provide the database dsn connect string "+constant.EnvNameDBDSN)
	flag.StringVar(&c.JWTSigningKey, "j", c.JWTSigningKey, "Provide the jwt secret key "+constant.EnvNameJWTSecretKey)
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
