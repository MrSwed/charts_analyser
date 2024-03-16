package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"database/sql"
	"errors"
	sqrl "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type VesselRepo struct {
	db *sqlx.DB
}

func NewVesselRepository(db *sqlx.DB) *VesselRepo {
	return &VesselRepo{db: db}
}

func (r *VesselRepo) GetVessels(ctx context.Context, vesselIDs ...domain.VesselID) (vessels domain.Vessels, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("id as vessel_id", "name as vessel_name").
		From(constant.DBVessels).
		Where("id = any($1) and is_deleted is not true", pq.Array(vesselIDs)).
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vessels, sqlStr, args...)
	if vessels == nil {
		vessels = make(domain.Vessels, 0)
	}
	return
}

func (r *VesselRepo) AddVessel(ctx context.Context, vesselNames ...domain.VesselName) (vessels domain.Vessels, err error) {
	var (
		sqlStr      string
		args        []interface{}
		uniqueCheck = make(map[domain.VesselName]struct{})
	)
	sqBuild := sq.Insert(constant.DBVessels + " as t").Columns("name")
	for _, name := range vesselNames {
		if _, ok := uniqueCheck[name]; !ok {
			uniqueCheck[name] = struct{}{}
			sqBuild = sqBuild.Values(name)
		}
	}
	// Имена уникальные, при совпадении добавляемого имени вернем уже существующий
	sqBuild = sqBuild.Suffix("on CONFLICT (name) DO UPDATE SET name=EXCLUDED.name where t.is_deleted is not true returning id as vessel_id, name as vessel_name")

	if sqlStr, args, err = sqBuild.ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vessels, sqlStr, args...)
	if vessels == nil {
		vessels = make([]*domain.Vessel, 0)
	}
	return
}

func (r *VesselRepo) UpdateVessels(ctx context.Context, vessels ...domain.Vessel) (savedVessels domain.Vessels, err error) {
	var tx *sqlx.Tx
	if tx, err = r.db.Beginx(); err != nil {
		return
	}
	defer func() {
		rErr := tx.Rollback()
		if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			err = errors.Join(err, rErr)
			savedVessels = domain.Vessels{}
		}
	}()

	var (
		stmt   *sqlx.Stmt
		sqlStr = "UPDATE" + " " + constant.DBVessels + " set name = $2 " +
			" where is_deleted is not true and id = $1 and (select count(name) from " + constant.DBVessels + " where id <> $1 and name = $2) = 0 " +
			" returning id as vessel_id, name as vessel_name"
	)
	if stmt, err = tx.PreparexContext(ctx, sqlStr); err != nil {
		return
	}
	for _, vessel := range vessels {
		var v domain.Vessel
		if er := stmt.GetContext(ctx, &v, vessel.ID, vessel.Name); er != nil {
			if errors.Is(er, sql.ErrNoRows) {
				continue
			}
			err = errors.Join(err, er)
			return
		}
		savedVessels = append(savedVessels, &v)
	}
	err = tx.Commit()
	if savedVessels == nil {
		savedVessels = make([]*domain.Vessel, 0)
	}
	return
}

func (r *VesselRepo) SetDeleteVessels(ctx context.Context, delete bool, vesselIDs ...domain.VesselID) (err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Update(constant.DBVessels).
		Set("is_deleted", delete).
		Where(sqrl.Eq{"id": vesselIDs}).
		ToSql(); err != nil {
		return
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return
}
