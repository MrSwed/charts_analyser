package main

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/handler"
	"charts_analyser/internal/app/repository"
	"charts_analyser/internal/app/service"
	"charts_analyser/internal/common/closer"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLog "github.com/gofiber/fiber/v2/middleware/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	runServer(ctx)
}

func runServer(ctx context.Context) {
	conf := config.NewConfig().Init()
	if len(conf.DatabaseDSN) == 0 {
		if conf.DatabaseDSN == "" {
			println("DatabaseDSN is required")
			os.Exit(1)
		}
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("Start server", zap.Any("Config", conf))

	var (
		db            *sqlx.DB
		graceShutdown = &closer.Closer{}
	)
	if db, err = sqlx.Connect("pgx", conf.DatabaseDSN); err != nil {
		logger.Fatal("cannot connect db", zap.Error(err))
	}

	app := fiber.New()
	app.Use(fiberLog.New(fiberLog.Config{
		Format: constant.LogFormat,
		Output: os.Stdout,
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://" + conf.ServerAddress,
		AllowCredentials: true,
		//MaxAge:           defaultCorsMaxAge,
	}))

	r := repository.NewRepository(db)
	s := service.NewService(r, &conf.JWT, logger)
	handler.NewHandler(app, s, conf, logger).Handler()

	graceShutdown.Add("APP", func(ctx context.Context) (err error) {
		if err = app.Shutdown(); err == nil {
			logger.Info("APP Closed")
		}
		return
	})

	graceShutdown.Add("DB", func(ctx context.Context) (err error) {
		if err = db.Close(); err == nil {
			logger.Info("Db Closed")
		}
		return
	})

	go func() {
		if err = app.Listen(conf.ServerAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Start server", zap.Error(err))
		}
	}()

	logger.Info("Server started")

	<-ctx.Done()

	logger.Info("Shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), constant.ServerShutdownTimeout)
	defer cancel()

	if err = graceShutdown.Close(shutdownCtx); err != nil {
		logger.Error("Shutdown", zap.Error(err), zap.Any("timeout, s: ", constant.ServerShutdownTimeout/time.Second))
	}

	logger.Info("Server stopped")

	_ = logger.Sync()
}
