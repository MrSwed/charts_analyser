package repository

import (
	"charts_analyser/internal/app/constant"
	"charts_analyser/internal/simulator/domain"
	"context"
	"github.com/jmoiron/sqlx"
)

type VesselRepo struct {
	db *sqlx.DB
}

func NewVesselRepository(db *sqlx.DB) *VesselRepo {
	return &VesselRepo{db: db}
}

func (r *VesselRepo) GetRandomVessels(ctx context.Context, count uint) (vessels []*domain.VesselItem, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Select("id as vessel_id", "name as vessel_name", "created_at").
		From(constant.DBVessels).
		Limit(uint64(count)).
		OrderBy("random()").
		ToSql(); err != nil {
		return
	}

	err = r.db.SelectContext(ctx, &vessels, sqlStr, args...)
	return
}
