package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
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
	if sqlStr, args, err = sq.Select("id", "name").
		From(constant.DBVessels).
		Where("id = any($1)", pq.Array(vesselIDs)).
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vessels, sqlStr, args...)
	return
}

func (r *VesselRepo) AddVessel(ctx context.Context, vessel domain.InputVessel) (vesselId domain.VesselID, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Insert(constant.DBVessels).
		Columns("name").
		Values(vessel.VesselName).
		Suffix("RETURNING id").
		ToSql(); err != nil {
		return
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
	return
}
