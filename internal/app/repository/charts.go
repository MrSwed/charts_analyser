package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type ChartsRepo struct {
	db *sqlx.DB
}

func NewChartsRepository(db *sqlx.DB) *ChartsRepo {
	return &ChartsRepo{db: db}
}

func (r *ChartsRepo) Zones(ctx context.Context, q domain.InputVessels) (zones []string, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("name").
		InnerJoin(constant.DBTracks+" t on st_contains(z.geometry, t.coordinate)").
		From(constant.DBZones+" z").
		Where("t.time between $1 and $2 and t.vessel_id = any ($3)", q.Start, q.End, pq.Array(q.VesselIDs)).
		GroupBy("name").
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &zones, sqlStr, args...)
	if zones == nil {
		zones = make([]string, 0)
	}
	/*
		 `select name from zones z
			inner join tracks t on st_contains(z.geometry, t.coordinate)
			where t.time between '2017-01-08 00:00:00' and '2020-10-09 00:00:00'
			and t.vessel_id = 9110913
			group by z.name`
	*/
	return
}

func (r *ChartsRepo) Vessels(ctx context.Context, q domain.InputZone) (vesselIds []uint64, err error) {
	var (
		sqlStr string
		args   []interface{}
	)

	if sqlStr, args, err = sq.Select("vessel_id").
		InnerJoin(constant.DBZones+" z on st_contains(z.geometry, t.coordinate)").
		From(constant.DBTracks+" t").
		Where("time between $1 and $2 and z.name = $3", q.Start, q.End, q.ZoneName).
		GroupBy("vessel_id").
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vesselIds, sqlStr, args...)
	if vesselIds == nil {
		vesselIds = make([]uint64, 0)
	}
	/*		`select vessel_id
	from tracks t
	     inner join zones z on st_contains(z.geometry, t.coordinate)
	where time between '2017-01-08 00:00:00' and '2020-10-09 00:00:00'
	and z.name = 'zone_205'
	group by vessel_id`*/

	return
}
