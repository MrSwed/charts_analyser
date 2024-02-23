package service

import (
	"charts_analyser/internal/app/domain"
	"charts_analyser/internal/app/repository"
	"context"
)

type Service struct {
	Vessel
}

func NewService(r *repository.Repository) *Service {
	return &Service{
		Vessel: NewChartsService(r),
	}
}

type Vessel interface {
	Zones(ctx context.Context, query domain.InputVessels) (zones []string, err error)
	Vessels(ctx context.Context, query domain.InputZone) (vesselIds []uint64, err error)
}

func NewChartsService(r *repository.Repository) *VesselService {
	return &VesselService{r: r}
}

type VesselService struct {
	r *repository.Repository
}

func (s *VesselService) Zones(ctx context.Context, query domain.InputVessels) (zones []string, err error) {
	return s.r.Zones(ctx, query)
}
func (s *VesselService) Vessels(ctx context.Context, query domain.InputZone) (vesselIds []uint64, err error) {
	return s.r.Vessels(ctx, query)
}
