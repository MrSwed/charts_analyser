package service

import (
	appDomain "charts_analyser/internal/app/domain"
	"charts_analyser/internal/simulator/config"
	"charts_analyser/internal/simulator/domain"
	"charts_analyser/internal/simulator/repository"
	"context"
	"go.uber.org/zap"
	"time"
)

type Service struct {
	Chart
	Request
	Vessels
	c *config.Config
	l *zap.Logger
}

func NewService(r *repository.Repository, c *config.Config, l *zap.Logger) *Service {
	return &Service{
		Chart:   NewChartService(r),
		Request: NewRequest(c.Out, l),
		Vessels: NewVesselsService(r),
		c:       c,
		l:       l,
	}
}

type Request interface {
	SendTrack(ctx context.Context, location appDomain.Point)
	SetControl(ctx context.Context, vesselID appDomain.VesselID)
}

type Chart interface {
	GetTrack(ctx context.Context, query appDomain.InputVesselsInterval) (tracks []domain.Track, err error)
}

type Vessels interface {
	GetRandomVessels(ctx context.Context, count uint) (vessels []*domain.VesselItem, err error)
}

func (s *Service) SimulateVessel(ctx context.Context, vessel *domain.VesselItem) {
	var q appDomain.InputVesselsInterval
	q.VesselIDs = appDomain.VesselIDs{vessel.ID}
	n := time.Now()
	n = time.Date(vessel.CreatedAt.Year(), vessel.CreatedAt.Month(), vessel.CreatedAt.Day(),
		n.Hour(), n.Minute(), n.Second(), 0, n.Location())
	q.Start = &n
	tracks, err := s.GetTrack(ctx, q)
	if err != nil {
		s.l.Error("No start tracks", zap.Error(err))
		return
	} else if len(tracks) == 0 {
		s.l.Error("No start tracks", zap.Any("q", q))
		return
	}
	s.l.Info("set control " + vessel.String())
	s.SetControl(ctx, vessel.ID)

	vessel.TrackAdd(tracks...)
	var simulateSt domain.HistoryDate

	var nextTime = time.Second * time.Duration(s.c.TrackInterval)
	if nextTime == 0 {
		nextTime = time.Until(time.Now().Add(time.Duration(s.c.TrackInterval) * time.Second))
	}
	for {
		select {
		case <-ctx.Done():
			s.l.Info("CTX: Stop simulation for", zap.Any("vessel", vessel))
			return
		case <-time.After(nextTime):
			track := vessel.TrackShift()

			s.SendTrack(ctx, track.Location)

			if len(vessel.Tracks()) < 1 {
				tracks, err := s.GetTrack(ctx, q)
				if err != nil {
					s.l.Error("Get tracks err", zap.Error(err), zap.Any("q", q))
				} else if len(tracks) == 0 {
					s.l.Info("No tracks more", zap.Any("q", q))
					return
				}
				vessel.TrackAdd(tracks...)
			}
			simulateSt = domain.HistoryDate(tracks[0].Timestamp)
			if s.c.TrackInterval == 0 {
				nextTime = time.Until(simulateSt.Now())
			}
			q.Start = &tracks[0].Timestamp
		}
	}
}
