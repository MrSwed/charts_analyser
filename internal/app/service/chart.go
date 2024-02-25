package service

import (
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/repository"
	"context"
)

func NewChartService(r *repository.Repository) *ChartRepo {
	return &ChartRepo{r: r}
}

type ChartRepo struct {
	r *repository.Repository
}

func (s *ChartRepo) Zones(ctx context.Context, query domain.InputVesselsInterval) (zones []domain.ZoneName, err error) {
	return s.r.Chart.Zones(ctx, query)
}
func (s *ChartRepo) Vessels(ctx context.Context, query domain.InputZone) (vesselIDs []domain.VesselID, err error) {
	return s.r.Chart.Vessels(ctx, query)
}
