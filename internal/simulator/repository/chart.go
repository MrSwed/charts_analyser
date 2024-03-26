package repository

import (
	appConst "charts_analyser/internal/app/constant"
	appDomain "charts_analyser/internal/app/domain"
	"charts_analyser/internal/simulator/constant"
	"charts_analyser/internal/simulator/domain"

	"context"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ChartRepo struct {
	db *sqlx.DB
}

func NewChartRepository(db *sqlx.DB) *ChartRepo {
	return &ChartRepo{db: db}
}

func (r *ChartRepo) GetTrack(ctx context.Context, q appDomain.InputVesselsInterval) (tracks []domain.Track, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("time", "ST_AsGeoJSON(location)::json->>'coordinates' as location").
		From(appConst.DBTracks).
		Where("time > $1 and vessel_id = any ($2)", q.StartOrLastPeriod(), pq.Array(q.VesselIDs)).
		OrderBy("time asc").
		Limit(constant.DefaultTracksItemsCache).
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &tracks, sqlStr, args...)
	if tracks == nil {
		tracks = make([]domain.Track, 0)
	}

	return
}
