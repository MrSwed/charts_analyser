package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
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
	sqBuild := sq.Insert(constant.DBVessels).Columns("name")
	for _, name := range vesselNames {
		if _, ok := uniqueCheck[name]; !ok {
			uniqueCheck[name] = struct{}{}
			sqBuild = sqBuild.Values(name)
		}
	}
	// Имена уникальные, при совпадении добавляемого имени вернем уже существующий
	sqBuild = sqBuild.Suffix("on CONFLICT (name) DO UPDATE SET name=EXCLUDED.name returning id as vessel_id, name as vessel_name")

	if sqlStr, args, err = sqBuild.ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vessels, sqlStr, args...)
	return
}

func (r *VesselRepo) SetDeleted(ctx context.Context, delete bool, vesselIDs ...domain.VesselID) (err error) {
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
