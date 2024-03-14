package repository

import (
	"charts_analyser/internal/app/domain"
	"context"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var sq = sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar)

type Repository struct {
	Chart
	Monitor
	Vessels
	Log
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Chart:   NewChartRepository(db),
		Monitor: NewMonitorDBRepository(db),
		Vessels: NewVesselRepository(db),
		Log:     NewLogRepository(db),
	}
}

type Chart interface {
	Zones(ctx context.Context, query domain.InputVesselsInterval) (zones []domain.ZoneName, err error)
	Vessels(ctx context.Context, query domain.InputZones) (vesselIDs []domain.VesselID, err error)
	ZonesByLocation(ctx context.Context, location domain.Point) (zones []domain.ZoneName, err error)
	Track(ctx context.Context, track *domain.Track) (err error)
	GetTrack(ctx context.Context, query domain.InputVesselsInterval) (tracks []domain.Track, err error)
}

type Vessels interface {
	GetVessels(ctx context.Context, vesselIDs ...domain.VesselID) (domain.Vessels, error)
	AddVessel(ctx context.Context, vesselNames ...domain.VesselName) (vessels domain.Vessels, err error)
	SetDeleted(ctx context.Context, delete bool, vesselIDS ...domain.VesselID) error
	UpdateVessels(ctx context.Context, vessels ...domain.Vessel) (savedVessels domain.Vessels, err error)
}

type Monitor interface {
	SetControl(ctx context.Context, status bool, vessels ...*domain.Vessel) error
	GetStates(ctx context.Context, vesselID ...domain.VesselID) ([]*domain.VesselState, error)
	UpdateState(ctx context.Context, vesselID domain.VesselID, v *domain.VesselState) error
	MonitoredVessels(ctx context.Context) (domain.Vessels, error)
}

type Log interface {
	ControlLogAdd(ctx context.Context, log ...domain.ControlLog) error
}
