package service

import (
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"charts_analyser/internal/app/repository"
	"context"
	"database/sql"
	"errors"
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
	/* todo: log to postgres * /
	if err != nil {
		go func(ctx context.Context, vessel domain.Vessel, status bool) {
			err := s.r.ControlLogAdd(ctx, vessel, status)

		}(ctx, vessel, status)
	}
	/**/
	return
}
func (s *MonitorService) GetStates(ctx context.Context, vesselIds ...domain.VesselID) (states []*domain.VesselState, err error) {
	for _, vesselId := range vesselIds {
		control, er := s.IsMonitored(ctx, vesselId)
		if er != nil || !control {
			continue
		}
		state, er := s.r.Monitor.GetState(ctx, vesselId)
		if er != nil {
			if errors.Is(err, sql.ErrNoRows) {
				er = myErr.ErrNotExist
			}
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

func (s *MonitorService) UpdateState(ctx context.Context, vesselId domain.VesselID, state domain.VesselState) (err error) {
	var control bool
	if control, err = s.IsMonitored(ctx, vesselId); err != nil {
		return
	}
	if !control {
		return myErr.ErrNotControlled
	}
	err = s.r.Monitor.UpdateState(ctx, vesselId, state)
	/* todo: * /
	if err != nil {
		go func(ctx context.Context, vesselId domain.VesselID, state domain.VesselState) {
			err := s.r.TrackAdd(ctx, vesselId, state)

		}(ctx, vesselId, state)
	}
	/**/
	return
}
func (s *MonitorService) MonitoredVessels(ctx context.Context) (vessels domain.Vessels, err error) {
	return s.r.Monitor.MonitoredVessels(ctx)
}
