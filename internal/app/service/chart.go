package service

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"charts_analyser/internal/app/repository"
	"context"
	"database/sql"
	"errors"
	"github.com/Goldziher/go-utils/sliceutils"
	"github.com/redis/go-redis/v9"
	"time"
)

func NewChartService(r *repository.Repository) *ChartService {
	return &ChartService{r: r}
}

type ChartService struct {
	r *repository.Repository
}

func (s *ChartService) Zones(ctx context.Context, query domain.InputVesselsInterval) (zones []domain.ZoneName, err error) {
	return s.r.Chart.Zones(ctx, query)
}

func (s *ChartService) Vessels(ctx context.Context, query domain.InputZone) (vesselIDs []domain.VesselID, err error) {
	return s.r.Chart.Vessels(ctx, query)
}

func (s *ChartService) Track(ctx context.Context, vesselID domain.VesselID, loc domain.Point) (err error) {
	var (
		track   = new(domain.Track)
		vessels domain.Vessels
	)
	if loc[0] == 0 || loc[1] == 1 ||
		loc[0] < constant.GeoAllowedRange[0] || loc[1] < constant.GeoAllowedRange[1] ||
		loc[0] > constant.GeoAllowedRange[2] || loc[1] > constant.GeoAllowedRange[3] {
		err = myErr.ErrLocationOutOfRange
		return
	}
	track.Location = loc
	vessels, err = s.r.GetVessels(ctx, vesselID)
	track.Vessel = *vessels[0]
	if track.Timestamp.IsZero() {
		track.Timestamp = time.Now()
	}
	// may be set in bg ?
	//go func(ctx context.Context) {
	if er := s.MaybeUpdateState(ctx, vesselID, track); er != nil {
		err = errors.Join(err, er)
	}
	//}(ctx)

	if er := s.r.Chart.Track(ctx, track); er != nil {
		err = errors.Join(err, er)
	}
	return
}

func (s *ChartService) MaybeUpdateState(ctx context.Context, vesselID domain.VesselID, track *domain.Track) (err error) {
	var control bool
	if control, err = s.r.IsMonitored(ctx, vesselID); control {
		var state *domain.VesselState

		if state, err = s.r.GetState(ctx, vesselID); err != nil && !errors.Is(err, redis.Nil) {
			return err
		}
		if state == nil {
			state = domain.NewVesselState(control)
		}
		state.Location = track.Location
		state.Vessel = track.Vessel
		state.Timestamp = track.Timestamp
		var newZones []domain.ZoneName
		newZones, err = s.r.Chart.ZonesByLocation(ctx, track.Location)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		if len(sliceutils.Difference(state.Zones, newZones)) > 0 {
			state.CurrentZone = domain.CurrentZone{
				Zones:  newZones,
				TimeIn: time.Now(),
			}
		}
		state.ZoneTimeSet()
		if er := s.r.Monitor.UpdateState(ctx, vesselID, state); err != nil {
			err = errors.Join(err, er)
		}
	}
	return
}

func (s *ChartService) GetTrack(ctx context.Context, query domain.InputVesselsInterval) (tracks []domain.Track, err error) {
	return s.r.Chart.GetTrack(ctx, query)
}
