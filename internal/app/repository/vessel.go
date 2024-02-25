package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/app/domain"
	"context"
	"github.com/jmoiron/sqlx"
)

type VesselRepo struct {
	db *sqlx.DB
}

func NewVesselRepository(db *sqlx.DB) *VesselRepo {
	return &VesselRepo{db: db}
}

func (r *VesselRepo) GetVessel(ctx context.Context, vesselId domain.VesselID) (vessel domain.Vessel, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("id", "name").
		From(constant.DBVessels).
		Where("id = $1", vesselId).
		ToSql(); err != nil {
		return
	}

	err = r.db.GetContext(ctx, &vessel, sqlStr, args...)
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
