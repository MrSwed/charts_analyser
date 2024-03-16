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
	domain.VesselID
	jwtVessel   string
	jwtOperator string
	jwtAdmin    string
}

func newTestsEnvConfig() (c *testConfig) {
	err := godotenv.Load(".env.test")
	if err != nil {
		log.Fatal(err)
	}
	c = &testConfig{}
	c.WithEnv().CleanSchemes()

	c.jwtAdmin, err = domain.NewClaimAdmin(&c.JWT, 1, "Test Admin").Token()
	if err != nil {
		log.Fatal(err)
	}
	c.jwtOperator, err = domain.NewClaimOperator(&c.JWT, 12, "Test Operator").Token()
	if err != nil {
		log.Fatal(err)
	}
	c.VesselID = 9110913
	c.jwtVessel, err = domain.NewClaimVessels(&c.JWT, c.VesselID, "Test Vessel").Token()
	if err != nil {
		log.Fatal(err)
	}
	return
}

var (
	conf = newTestsEnvConfig()
	db   = func() *sqlx.DB {
		db, err := sqlx.Connect("postgres", conf.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		return db
	}()
)
