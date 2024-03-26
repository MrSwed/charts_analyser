package service

import (
	"charts_analyser/internal/simulator/domain"
	"charts_analyser/internal/simulator/repository"
	"context"
)

func NewVesselsService(r *repository.Repository) *VesselsService {
	return &VesselsService{r: r}
}

type VesselsService struct {
	r *repository.Repository
}

func (s *VesselsService) GetRandomVessels(ctx context.Context, count uint) (vessels []*domain.VesselItem, err error) {
	return s.r.Vessels.GetRandomVessels(ctx, count)
}
