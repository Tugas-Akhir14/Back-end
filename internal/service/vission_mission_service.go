package service

import (
	"context"

	"backend/internal/models"
	"backend/internal/repository"
)

type VisionMissionService interface {
	Get(ctx context.Context) (*models.VisionMission, error)
	Upsert(ctx context.Context, req models.UpsertVisionMissionRequest, userID *uint) (*models.VisionMission, error)
}

type visionMissionService struct {
	repo repository.VisionMissionRepository
}

func NewVisionMissionService(r repository.VisionMissionRepository) VisionMissionService {
	return &visionMissionService{repo: r}
}

func (s *visionMissionService) Get(ctx context.Context) (*models.VisionMission, error) {
	return s.repo.Get(ctx)
}

func (s *visionMissionService) Upsert(ctx context.Context, req models.UpsertVisionMissionRequest, userID *uint) (*models.VisionMission, error) {
	return s.repo.Upsert(ctx, req, userID)
}
