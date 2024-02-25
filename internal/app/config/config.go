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
	RedisAddress  string
	RedisPass     string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress: constant.ServerAddress,
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
	if redAddr, ok := os.LookupEnv(constant.EnvNameRedisAddress); ok && redAddr != "" {
		c.RedisAddress = redAddr
	}
	if redPas, ok := os.LookupEnv(constant.EnvNameRedisPass); ok && redPas != "" {
		c.RedisPass = redPas
	}
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Provide the database dsn connect string")
	flag.StringVar(&c.RedisAddress, "ra", c.RedisAddress, "Provide the redis address")
	flag.StringVar(&c.RedisPass, "rp", c.RedisPass, "Provide the redis pass")
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
