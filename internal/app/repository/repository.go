package repository

import (
	"charts_analyser/internal/app/domain"
	"context"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

var sq = sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar)

type Repository struct {
	Chart
	Monitor
	Vessels
}

func NewRepository(db *sqlx.DB, rds *redis.Client) *Repository {
	return &Repository{
		Chart:   NewChartsRepository(db),
		Monitor: NewMonitorRepository(rds),
		Vessels: NewVesselRepository(db),
	}
}

type Chart interface {
	Zones(ctx context.Context, query domain.InputVesselsInterval) (zones []domain.ZoneName, err error)
	Vessels(ctx context.Context, query domain.InputZone) (vesselIDs []domain.VesselID, err error)
	ZonesByLocation(ctx context.Context, location domain.Point) (zones []domain.ZoneName, err error)
	Track(ctx context.Context, track *domain.Track) (err error)
}

type Vessels interface {
	GetVessels(ctx context.Context, vesselIDs ...domain.VesselID) (domain.Vessels, error)
	AddVessel(ctx context.Context, vessel domain.InputVessel) (domain.VesselID, error)
}

type Monitor interface {
	IsMonitored(ctx context.Context, vesselId domain.VesselID) (bool, error)
	SetControl(ctx context.Context, status bool, vessels ...*domain.Vessel) error
	GetState(ctx context.Context, vesselId domain.VesselID) (*domain.VesselState, error)
	UpdateState(ctx context.Context, vesselId domain.VesselID, v *domain.VesselState) error
	MonitoredVessels(ctx context.Context) (domain.Vessels, error)
}
