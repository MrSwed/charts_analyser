package service

import (
	"charts_analyser/internal/app/domain"
	myErr "charts_analyser/internal/app/error"
	"charts_analyser/internal/app/repository"
	"context"
	"database/sql"
	"github.com/pkg/errors"
)

func NewVesselService(r *repository.Repository) *VesselService {
	return &VesselService{r: r}
}

type VesselService struct {
	r *repository.Repository
}

func (s *VesselService) GetVessels(ctx context.Context, vesselIDs ...domain.VesselID) (vessel domain.Vessels, err error) {
	vessel, err = s.r.GetVessels(ctx, vesselIDs...)
	if errors.Is(err, sql.ErrNoRows) {
		err = myErr.ErrNotExist
	}
	return
}

func (s *VesselService) AddVessel(ctx context.Context, vessel domain.InputVessel) (vesselId domain.VesselID, err error) {
	return s.r.AddVessel(ctx, vessel)
}
