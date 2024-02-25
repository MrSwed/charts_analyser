package service

import (
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"charts_analyser/internal/app/repository"
	"context"
	"github.com/pkg/errors"
)

func NewVesselService(r *repository.Repository) *VesselRepo {
	return &VesselRepo{r: r}
}

type VesselRepo struct {
	r *repository.Repository
}

func (s *VesselRepo) GetVessel(ctx context.Context, vesselId domain.VesselID) (vessel domain.Vessel, err error) {
	if errors.Is(err, errors.Cause(err)) {
		err = myErr.ErrNotExist
	}
	return s.r.GetVessel(ctx, vesselId)
}

func (s *VesselRepo) AddVessel(ctx context.Context, vessel domain.InputVessel) (vesselId domain.VesselID, err error) {
	return s.r.AddVessel(ctx, vessel)
}
