package repository

import (
	"charts_analyser/internal/app/domain"
	"context"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var sq = sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar)

type Repository struct {
	Charts
}

type Charts interface {
	Zones(ctx context.Context, query domain.InputVessels) (zones []domain.ZoneName, err error)
	Vessels(ctx context.Context, query domain.InputZone) (vesselIDs []domain.VesselID, err error)
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Charts: NewChartsRepository(db),
	}
}
