package service

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"charts_analyser/internal/app/repository"
	"context"
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"time"
)

func NewMonitorService(r *repository.Repository, log *zap.Logger) *MonitorService {
	return &MonitorService{r: r, log: log}
}

type MonitorService struct {
	r   *repository.Repository
	log *zap.Logger
}

func (s *MonitorService) SetControl(ctx context.Context, status bool, vesselIDs ...domain.VesselID) (err error) {
	var vessels domain.Vessels
	if vessels, err = s.r.Vessels.GetVessels(ctx, vesselIDs...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
		return
	}
	if len(vessels) == 0 {
		err = myErr.ErrNotExist
		return
	}
	err = s.r.Monitor.SetControl(ctx, status, []*domain.Vessel(vessels)...)
	if err != nil {
		return
	}

	/* log to postgres */
	go func(vessels domain.Vessels, status bool) {
		ctx, cancel := context.WithTimeout(context.Background(), constant.ServerOperationTimeout)
		defer cancel()
		var cLogs []domain.ControlLog
		for _, v := range vessels {
			cLogs = append(cLogs, domain.ControlLog{
				Vessel:    v,
				Timestamp: time.Now(),
				Control:   status,
				Comment:   nil,
			})
		}
		if err := s.r.ControlLogAdd(ctx, cLogs...); err != nil {
			s.log.Error("Background ControlLogAdd", zap.Error(err))
		}
	}(vessels, status)

	return
}

func (s *MonitorService) GetStates(ctx context.Context, vesselIDs ...domain.VesselID) (states []*domain.VesselState, err error) {
	states, err = s.r.Monitor.GetStates(ctx, vesselIDs...)
	if len(states) == 0 {
		states = []*domain.VesselState{}
		err = myErr.ErrNotExist
	}
	return
}

func (s *MonitorService) MonitoredVessels(ctx context.Context) (vessels domain.Vessels, err error) {
	return s.r.Monitor.MonitoredVessels(ctx)
}
