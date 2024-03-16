package service

import (
	"charts_analyser/internal/app/config"
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/repository"
	"context"
	"go.uber.org/zap"
)

type Service struct {
	Chart
	Monitor
	Vessel
	User
}

func NewService(r *repository.Repository, conf *config.JWT, log *zap.Logger) *Service {
	return &Service{
		Chart:   NewChartService(r),
		Monitor: NewMonitorService(r, log),
		Vessel:  NewVesselService(r),
		User:    NewUserService(r, conf, log),
	}
}

type Chart interface {
	Zones(ctx context.Context, query domain.InputVesselsInterval) (zones []domain.ZoneName, err error)
	Vessels(ctx context.Context, query domain.InputZones) (vesselIDs []domain.VesselID, err error)
	Track(ctx context.Context, vesselID domain.VesselID, loc domain.InputPoint) (err error)
	MaybeUpdateState(ctx context.Context, vesselID domain.VesselID, track *domain.Track) error
	GetTrack(ctx context.Context, query domain.InputVesselsInterval) (tracks []domain.Track, err error)
}

type Vessel interface {
	GetVessels(ctx context.Context, vesselIDs ...domain.VesselID) (domain.Vessels, error)
	AddVessel(ctx context.Context, vesselNames ...domain.VesselName) (vessels domain.Vessels, err error)
	UpdateVessels(ctx context.Context, vessels ...domain.Vessel) (savedVessels domain.Vessels, err error)
	SetDeleteVessels(ctx context.Context, delete bool, vesselIDS ...domain.VesselID) error
}

type User interface {
	Login(ctx context.Context, user domain.LoginForm) (token string, err error)
	GetUser(ctx context.Context, login domain.UserLogin) (user *domain.UserDB, err error)
	AddUser(ctx context.Context, userAdd domain.UserChange) (id domain.UserID, err error)
	UpdateUser(ctx context.Context, user domain.UserChange) (err error)
	SetDeletedUser(ctx context.Context, delete bool, userIDs ...domain.UserID) (err error)
}

type Monitor interface {
	SetControl(ctx context.Context, status bool, vessels ...domain.VesselID) error
	GetStates(ctx context.Context, vesselIDs ...domain.VesselID) ([]*domain.VesselState, error)
	MonitoredVessels(ctx context.Context) (domain.Vessels, error)
}
