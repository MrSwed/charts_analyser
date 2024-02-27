package main

import (
	"charts_analyser/internal/simulator/config"
	"charts_analyser/internal/simulator/constant"
	"charts_analyser/internal/simulator/domain"
	"charts_analyser/internal/simulator/repository"
	"charts_analyser/internal/simulator/service"
	"context"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	var (
		wg   sync.WaitGroup
		conf = config.NewConfig().Init()
	)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	db, err := sqlx.Connect("pgx", conf.DatabaseDSN)
	if err != nil {
		logger.Fatal("cannot connect db", zap.Error(err))
	}

	operatorClaims := domain.ClaimsOperator{}

	if operatorToken, err := operatorClaims.Token(conf.JWTSigningKey); err == nil && operatorToken != "" {
		ctx = context.WithValue(ctx, constant.CtxValueKeyJWTOperator, operatorToken)
	} else if err != nil {
		logger.Fatal("Build operator jwt", zap.Error(err))
	}

	r := repository.NewRepository(db, nil)
	s := service.NewService(r, conf, logger)
	vessels, err := s.Vessels.GetRandomVessels(ctx, conf.VesselCount)
	if err != nil {
		logger.Fatal("GetRandomVessels:", zap.Error(err))
	}

	if len(vessels) == 0 {
		logger.Fatal("GetRandomVessels: no results ")
	}
	wg.Add(len(vessels))
	var ids []int64
	for _, v := range vessels {
		ids = append(ids, int64(v.ID))
		// go run service vessel simulation
		go func(vessel *domain.VesselItem) {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
			logger.Info("Start simulation for", zap.Any("vessel", vessel.String()))

			vesselClaims := domain.ClaimsVessel{Vessel: &vessel.Vessel}
			jwtStr, er := vesselClaims.Token(conf.JWTSigningKey)
			if er != nil {
				logger.Fatal("Build vessel jwt", zap.Error(err))
			}
			vesselCtx := context.WithValue(ctx, constant.CtxValueKeyJWTVessel, jwtStr)
			s.SimulateVessel(vesselCtx, vessel)
		}(v)
	}
	logger.Info("Simulation started for IDs: ", zap.Any("ids", ids))

	wg.Wait()
	logger.Info("Simulation stopped")
}
