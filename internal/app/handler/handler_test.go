package handler_test

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/handler"
	"charts_analyser/internal/app/repository"
	"charts_analyser/internal/app/service"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"

	"go.uber.org/zap"
	"log"
	"path/filepath"
	"time"
)

type testConfig struct {
	*config.Config
	domain.VesselID
	domain.ZoneName
	jwtVessel   string
	jwtOperator string
	jwtAdmin    string
}

func newConfig() (c *testConfig) {
	var err error
	c = &testConfig{}
	c.Config = config.NewConfig()

	c.jwtAdmin, err = domain.NewClaimAdmin(&c.JWT, 1, "Test Admin").Token()
	if err != nil {
		log.Fatal(err)
	}
	c.jwtOperator, err = domain.NewClaimOperator(&c.JWT, 12, "Test Operator").Token()
	if err != nil {
		log.Fatal(err)
	}

	c.VesselID = 9110913
	c.ZoneName = "zone_47"

	c.jwtVessel, err = domain.NewClaimVessels(&c.JWT, c.VesselID, "Test Vessel").Token()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func CreatePostgresContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgis/postgis:16-3.4-alpine"),

		postgres.WithInitScripts(
			filepath.Join("../../../", "testdata", "scheme.sql"),
			filepath.Join("../../../", "testdata", "tracks.sql"),
			filepath.Join("../../../", "testdata", "vessels.sql"),
			filepath.Join("../../../", "testdata", "zones.sql")),

		postgres.WithDatabase("test-db"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	return pgContainer, nil
}

type HandlerTestSuite struct {
	suite.Suite
	ctx    context.Context
	app    *fiber.App
	srv    *service.Service
	cfg    *testConfig
	pgCont *postgres.PostgresContainer
}

func (suite *HandlerTestSuite) SetupSuite() {
	var err error
	suite.cfg = newConfig()
	suite.ctx = context.Background()
	suite.pgCont, err = CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.cfg.DatabaseDSN, err = suite.pgCont.ConnectionString(suite.ctx, "sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(func() *sqlx.DB {
		db, err := sqlx.Connect("pgx", suite.cfg.DatabaseDSN)
		if err != nil {
			log.Fatal(err)
		}
		return db
	}())

	logger, _ := zap.NewDevelopment()

	suite.srv = service.NewService(repo, &suite.cfg.JWT, logger)

	suite.app = fiber.New()
	suite.app.Use(recover.New())

	_ = handler.NewHandler(suite.app, suite.srv, suite.cfg.Config, logger).Handler()
}

func (suite *HandlerTestSuite) TearDownSuite() {
	if err := suite.pgCont.Terminate(suite.ctx); err != nil {
		log.Fatalf("error terminating postgres container: %s", err)
	}
}

func TestHandlers(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
