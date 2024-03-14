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
	if errors.Is(err, sql.ErrNoRows) || len(vessel) == 0 {
		err = myErr.ErrNotExist
	}
	return
}

func (s *VesselService) AddVessel(ctx context.Context, vesselNames ...domain.VesselName) (vessels domain.Vessels, err error) {
	return s.r.AddVessel(ctx, vesselNames...)
}

func (s *VesselService) UpdateVessels(ctx context.Context, vessels ...domain.Vessel) (savedVessels domain.Vessels, err error) {
	return s.r.UpdateVessels(ctx, vessels...)
}

func (s *VesselService) SetDeleted(ctx context.Context, delete bool, vesselIDS ...domain.VesselID) error {
	return s.r.SetDeleted(ctx, delete, vesselIDS...)
}
