package service

import (
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"charts_analyser/internal/app/repository"
	"context"
	"database/sql"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewMonitorService(r *repository.Repository) *MonitorService {
	return &MonitorService{r: r}
}

type MonitorService struct {
	r *repository.Repository
}

func (s *MonitorService) IsMonitored(ctx context.Context, vesselId domain.VesselID) (state bool, err error) {
	return s.r.Monitor.IsMonitored(ctx, vesselId)
}
func (s *MonitorService) SetControl(ctx context.Context, status bool, vesselIDs ...domain.VesselID) (err error) {
	/* todo: create new one automatically ? * /
	if vessel.ID == 0 && len([]rune(vessel.Name)) > 0 {
		vessel.ID, err = s.r.AddVessel(ctx, domain.InputVessel{
			VesselName: vessel.Name,
		})
	}
	/**/
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
	// update monitoring status
	go func(status bool, vesselIDs ...domain.VesselID) {
		var states []*domain.VesselState
		if states, err = s.GetStates(ctx, vesselIDs...); len(states) == 0 && err != nil && !errors.Is(err, redis.Nil) {
			// todo handle error
			return
		}
		for _, st := range states {
			st.Control.State = status
			if !status {
				st.Control.ControlEnd = &[]time.Time{time.Now()}[0]
			} else {
				st.Control.ControlStart = &[]time.Time{time.Now()}[0]
				st.Control.ControlEnd = nil
			}
			err = s.r.UpdateState(ctx, st.Vessel.ID, st)
		}
	}(status, vesselIDs...)
	/* log to postgres */
	go func(ctx context.Context, vessels domain.Vessels, status bool) {
		var cLogs []domain.ControlLog
		for _, v := range vessels {
			cLogs = append(cLogs, domain.ControlLog{
				Vessel:    v,
				Timestamp: time.Now(),
				Control:   status,
				Comment:   nil,
			})
		}
		// todo : handle error
		_ = s.r.ControlLogAdd(ctx, cLogs...)
	}(ctx, vessels, status)

	return
}

func (s *MonitorService) GetStates(ctx context.Context, vesselIds ...domain.VesselID) (states []*domain.VesselState, err error) {
	for _, vesselId := range vesselIds {
		state, er := s.r.Monitor.GetState(ctx, vesselId)
		if er != nil && !errors.Is(er, redis.Nil) {
			err = errors.Join(err, er)
		} else if state != nil {
			states = append(states, state)
		}
	}
	if len(states) == 0 {
		err = myErr.ErrNotExist
	}
	return
}

func (s *MonitorService) MonitoredVessels(ctx context.Context) (vessels domain.Vessels, err error) {
	return s.r.Monitor.MonitoredVessels(ctx)
}
