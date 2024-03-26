package service

import (
	appDomain "charts_analyser/internal/app/domain"
	"charts_analyser/internal/simulator/domain"
	"charts_analyser/internal/simulator/repository"
	"context"
)

func NewChartService(r *repository.Repository) *ChartService {
	return &ChartService{r: r}
}

type ChartService struct {
	r *repository.Repository
}

func (s *ChartService) GetTrack(ctx context.Context, query appDomain.InputVesselsInterval) (tracks []domain.Track, err error) {
	return s.r.Chart.GetTrack(ctx, query)
}
