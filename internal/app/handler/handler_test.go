package handler_test

// used env with default database connection

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/handler"
	"charts_analyser/internal/app/repository"
	"charts_analyser/internal/app/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
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
	serv *service.Service
	app  *fiber.App
)

func init() {

	repo := repository.NewRepository(func() *sqlx.DB {
		db, err := sqlx.Connect("postgres", conf.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		return db
	}())

	logger, _ := zap.NewDevelopment()

	serv = service.NewService(repo, &conf.JWT, logger)

	app = fiber.New()
	app.Use(recover.New())

	_ = handler.NewHandler(app, serv, &conf.Config, logger).Handler()

}
