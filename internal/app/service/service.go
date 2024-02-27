package service

import (
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/repository"
	"context"
	"go.uber.org/zap"
)

type Service struct {
	Chart
	Monitor
	Vessel
}

func NewService(r *repository.Repository, log *zap.Logger) *Service {
	return &Service{
		Chart:   NewChartService(r),
		Monitor: NewMonitorService(r, log),
		Vessel:  NewVesselService(r),
	}
}

type Chart interface {
	Zones(ctx context.Context, query domain.InputVesselsInterval) (zones []domain.ZoneName, err error)
	Vessels(ctx context.Context, query domain.InputZones) (vesselIDs []domain.VesselID, err error)
	Track(ctx context.Context, vesselID domain.VesselID, query domain.Point) (err error)
	MaybeUpdateState(ctx context.Context, vesselId domain.VesselID, track *domain.Track) error
	GetTrack(ctx context.Context, query domain.InputVesselsInterval) (tracks []domain.Track, err error)
}

type Vessel interface {
	GetVessels(ctx context.Context, vesselIDs ...domain.VesselID) (domain.Vessels, error)
	AddVessel(ctx context.Context, vessel domain.InputVessel) (domain.VesselID, error)
}

type Monitor interface {
	IsMonitored(ctx context.Context, vesselId domain.VesselID) (bool, error)
	SetControl(ctx context.Context, status bool, vessels ...domain.VesselID) error
	GetStates(ctx context.Context, vesselId ...domain.VesselID) ([]*domain.VesselState, error)
	MonitoredVessels(ctx context.Context) (domain.Vessels, error)
}
