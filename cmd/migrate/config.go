package main

import (
	"flag"
	"os"
	"strings"
)

type Config struct {
	DatabaseDSN     string
	MigrateDataPath string
	ZonesFile       string
	ChartsPath      string
}

func NewConfig() *Config {
	return &Config{
		//DatabaseDSN: "",
		MigrateDataPath: "file://migrate",
		ZonesFile:       "data/geo_zones.json",
	}
}

func (c *Config) Init() *Config {
	return c.withFlags().WithEnv().CleanParameters()
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Database dsn connect string, can set with env "+EnvNameDBDSN)
	flag.StringVar(&c.MigrateDataPath, "m", c.MigrateDataPath, "Path to migrate and import files, can set with env "+EnvNameMigratePath)
	flag.StringVar(&c.ZonesFile, "z", c.ZonesFile, "Path to geo zones json file "+EnvNameZonesFile)
	flag.StringVar(&c.ChartsPath, "c", c.ZonesFile, "Path to charts "+EnvNameChartsPath)

	flag.Parse()
	return c
}

func (c *Config) WithEnv() *Config {
	if env, ok := os.LookupEnv(EnvNameDBDSN); ok {
		c.DatabaseDSN = env
	}
	if env, ok := os.LookupEnv(EnvNameMigratePath); ok {
		c.MigrateDataPath = env
	}
	return c
}

func (c *Config) CleanParameters() *Config {
	c.DatabaseDSN = strings.Trim(c.DatabaseDSN, "'")
	if !strings.HasPrefix(c.MigrateDataPath, "file://") {
		c.MigrateDataPath = "file://" + c.MigrateDataPath
	}
	return c
}
