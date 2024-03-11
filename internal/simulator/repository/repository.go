package repository

import (
	appDomain "charts_analyser/internal/app/domain"
	"charts_analyser/internal/simulator/domain"

	"context"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

var sq = sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar)

type Repository struct {
	Chart
	Vessels
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Chart:   NewChartRepository(db),
		Vessels: NewVesselRepository(db),
	}
}

type Chart interface {
	GetTrack(ctx context.Context, query appDomain.InputVesselsInterval) (tracks []domain.Track, err error)
}

type Vessels interface {
	GetRandomVessels(ctx context.Context, count uint) (vessels []*domain.VesselItem, err error)
}
