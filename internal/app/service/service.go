package service

import (
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/repository"
	"context"
)

type Service struct {
	Chart
	Monitor
	Vessel
}

func NewService(r *repository.Repository) *Service {
	return &Service{
		Chart:   NewChartService(r),
		Monitor: NewMonitorService(r),
		Vessel:  NewVesselService(r),
	}
}

type Chart interface {
	Zones(ctx context.Context, query domain.InputVessels) (zones []domain.ZoneName, err error)
	Vessels(ctx context.Context, query domain.InputZone) (vesselIDs []domain.VesselID, err error)
}

type Vessel interface {
	GetVessel(ctx context.Context, vesselId domain.VesselID) (domain.Vessel, error)
	AddVessel(ctx context.Context, vessel domain.InputVessel) (domain.VesselID, error)
}

type Monitor interface {
	IsMonitored(ctx context.Context, vesselId domain.VesselID) (bool, error)
	SetControl(ctx context.Context, vessel domain.Vessel, status bool) error
	GetState(ctx context.Context, vesselId domain.VesselID) (*domain.VesselState, error)
	UpdateState(ctx context.Context, vesselId domain.VesselID, v domain.VesselState) error
	MonitoredVessels(ctx context.Context) (domain.Vessels, error)
}
