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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
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
	redisCli := func() *redis.Client {
		client := redis.NewClient(&redis.Options{
			Addr:     conf.RedisAddress,
			Password: conf.RedisPass,
			DB:       0,
		})
		if _, err = client.Ping(ctx).Result(); err != nil {
			logger.Fatal("cannot connect redis", zap.Error(err))
		}
		return client
	}()
	r := repository.NewRepository(db, redisCli)
	s := service.NewService(r, logger)
	h := handler.NewHandler(s, logger)

	server := &http.Server{Addr: conf.ServerAddress, Handler: h.Handler()}

	graceShutdown.Add("WEB", server.Shutdown)

	graceShutdown.Add("DB Close", func(ctx context.Context) (err error) {
		if err = db.Close(); err == nil {
			logger.Info("Db Closed")
		}
		return
	})

	graceShutdown.Add("Redis Close", func(ctx context.Context) (err error) {
		if err = redisCli.Close(); err == nil {
			logger.Info("Redis Closed")
		}
		return
	})

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
