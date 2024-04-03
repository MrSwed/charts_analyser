package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"time"
)

type ChartRepo struct {
	db *sqlx.DB
}

func NewChartRepository(db *sqlx.DB) *ChartRepo {
	return &ChartRepo{db: db}
}

func (r *ChartRepo) Zones(ctx context.Context, q domain.InputVesselsInterval) (zones []domain.ZoneName, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("name").
		InnerJoin(constant.DBTracks+" t on st_contains(z.geometry, t.location)").
		From(constant.DBZones+" z").
		Where("t.time between $1 and $2 and t.vessel_id = any ($3)", q.StartOrLastPeriod(), q.FinishOrNow(), pq.Array(q.VesselIDs)).
		GroupBy("name").
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &zones, sqlStr, args...)
	if zones == nil {
		zones = make([]domain.ZoneName, 0)
	}
	/*
		 `select name from zones z
			inner join tracks t on st_contains(z.geometry, t.location)
			where t.time between '2017-01-08 00:00:00' and '2020-10-09 00:00:00'
			and t.vessel_id = 9110913
			group by z.name`
	*/
	return
}

func (r *ChartRepo) ZonesByLocation(ctx context.Context, location domain.Point) (zones []domain.ZoneName, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("name").
		From(constant.DBZones+" z").
		Where("st_contains(z.geometry, $1)", location).
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &zones, sqlStr, args...)
	return
}

func (r *ChartRepo) Vessels(ctx context.Context, q domain.InputZones) (vesselIDs []domain.VesselID, err error) {
	var (
		sqlStr string
		args   []interface{}
	)

	if sqlStr, args, err = sq.Select("vessel_id").
		InnerJoin(constant.DBZones+" z on st_contains(z.geometry, t.location)").
		From(constant.DBTracks+" t").
		Where("time between $1 and $2 and z.name = any ($3)", q.StartOrLastPeriod(), q.FinishOrNow(), pq.Array(q.ZoneNames)).
		GroupBy("vessel_id").
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vesselIDs, sqlStr, args...)
	if vesselIDs == nil {
		vesselIDs = make([]domain.VesselID, 0)
	}
	/*		`select vessel_id
	from tracks t
	     inner join zones z on st_contains(z.geometry, t.location)
	where time between '2017-01-08 00:00:00' and '2020-10-09 00:00:00'
	and z.name = 'zone_205'
	group by vessel_id`*/

	return
}

func (r *ChartRepo) Track(ctx context.Context, track *domain.Track) (err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if track.Timestamp.IsZero() {
		track.Timestamp = time.Now()
	}
	if sqlStr, args, err = sq.Insert(constant.DBTracks).
		Columns("vessel_id", "time", "location").
		Values(track.Vessel.ID, track.Timestamp, track.Location).
		ToSql(); err != nil {
		return
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return
}

func (r *ChartRepo) GetTrack(ctx context.Context, q domain.InputVesselsInterval) (tracks []domain.Track, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("time", "ST_AsGeoJSON(location)::json->>'coordinates' as location", "vessel_id", "v.name as vessel_name").
		From(constant.DBTracks+" t").
		LeftJoin(constant.DBVessels+" v on v.id = t.vessel_id ").
		Where("time between $1 and $2 and vessel_id = any ($3)", q.StartOrLastPeriod(), q.FinishOrNow(), pq.Array(q.VesselIDs)).
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &tracks, sqlStr, args...)
	if tracks == nil {
		tracks = make([]domain.Track, 0)
	}

	return
}
