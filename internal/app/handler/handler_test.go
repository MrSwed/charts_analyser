package handler_test

// used  env fo real databases

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/domain"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
)

type testConfig struct {
	config.Config
	jwtOperator string
}

func newEnvConfig() (c *testConfig) {
	err := godotenv.Load(".env.test")
	if err != nil {
		log.Fatal(err)
	}
	c = &testConfig{}
	c.WithEnv().CleanSchemes()
	c.jwtOperator, err = domain.NewClaimOperator(&c.JWT, 12, "User for test").Token()
	if err != nil {
		log.Fatal(err)
	}
	return
}

var (
	conf = newEnvConfig()
	db   = func() *sqlx.DB {
		db, err := sqlx.Connect("postgres", conf.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		return db
	}()
)
