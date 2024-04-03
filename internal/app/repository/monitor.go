package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"database/sql"
	"errors"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"time"
)

type MonitorDBCache struct {
	db *sqlx.DB
}

func NewMonitorDBRepository(db *sqlx.DB) *MonitorDBCache {
	return &MonitorDBCache{db: db}
}

func (r *MonitorDBCache) SetControl(ctx context.Context, control bool, vesselsItems ...*domain.Vessel) (err error) {
	if len(vesselsItems) == 0 {
		return errors.New("no vessels for control")
	}

	var tx *sqlx.Tx
	if tx, err = r.db.Beginx(); err != nil {
		return
	}
	defer func() {
		rErr := tx.Rollback()
		if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			err = errors.Join(err, rErr)
		}
	}()

	var stmt *sqlx.Stmt
	if control {
		if stmt, err = tx.PreparexContext(ctx, "INSERT INTO"+" "+constant.DBControlDashboard+
			" (vessel_id, state, control_start) "+
			" VALUES($1, $2, $3 ) "+
			" on conflict (vessel_id) do update set state = $2, control_start = $3 "); err != nil {
			return
		}
	} else {
		if stmt, err = tx.PreparexContext(ctx, "INSERT INTO"+" "+constant.DBControlDashboard+
			" (vessel_id, state,  control_end) "+
			" VALUES($1, $2, $3 ) "+
			" on conflict (vessel_id) do update set state = $2, control_end = $3 "); err != nil {
			return
		}

	}
	for _, vesselItem := range vesselsItems {
		if _, err = stmt.ExecContext(ctx, vesselItem.ID, control, time.Now()); err != nil {
			return
		}
	}
	err = tx.Commit()

	return
}

func (r *MonitorDBCache) GetStates(ctx context.Context, vesselIDs ...domain.VesselID) (vStates []*domain.VesselState, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select(
		"v.id as vessel_id", "v.name as vessel_name",
		"state as control",
		"timestamp",
		"control_start",
		"control_end",
		"ST_AsGeoJSON(location)::json->>'coordinates' as location",
		"current_zone",
		"extract(epoch from age(timestamp, (current_zone::jsonb->>'timeIn')::timestamptz))::real as zone_duration",
	).
		From(constant.DBControlDashboard + " d").
		LeftJoin(constant.DBVessels + " v on v.id = d.vessel_id ").
		Where(sqrl.Eq{"v.id": vesselIDs}).
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vStates, sqlStr, args...)

	return
}

func (r *MonitorDBCache) UpdateState(ctx context.Context, vesselID domain.VesselID, v *domain.VesselState) (err error) {
	if v == nil {
		err = errors.New("updateState: input data nil")
		return
	}

	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Insert(constant.DBControlDashboard).
		Columns(
			"vessel_id", "state", "timestamp",
			"control_start", "control_end", "location",
			"current_zone").
		Values(vesselID, v.State, v.Timestamp,
			v.ControlStart, v.ControlEnd, v.Location, v.CurrentZone).
		Suffix("on conflict (vessel_id) do update set state = $2, timestamp = $3, control_start = $4,control_end = $5, location = $6, current_zone = $7").
		ToSql(); err != nil {
		return
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return

}

func (r *MonitorDBCache) MonitoredVessels(ctx context.Context) (vessels domain.Vessels, err error) {
	var (
		sqlStr string
		args   []interface{}
	)

	if sqlStr, args, err = sq.Select("v.id as vessel_id", "v.name as vessel_name").
		From(constant.DBControlDashboard + " d").
		LeftJoin(constant.DBVessels + " v on v.id = d.vessel_id ").
		Where("d.state is true").
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vessels, sqlStr, args...)
	return
}
