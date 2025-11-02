// internal/service/hotelservice/vision_mission_service.go
package hotelservice

import (
	"context"

	"backend/internal/models/hotel"
	"backend/internal/repository/repohotel"
)

type VisionMissionService interface {
	Get(ctx context.Context) (*hotel.VisionMission, error)
	GetPublic(ctx context.Context) (*hotel.VisionMission, error)
	Upsert(ctx context.Context, req hotel.UpsertVisionMissionRequest, userID *uint) (*hotel.VisionMission, error)
}

type visionMissionService struct {
	repo repohotel.VisionMissionRepository
}

func NewVisionMissionService(r repohotel.VisionMissionRepository) VisionMissionService {
	return &visionMissionService{repo: r}
}

func (s *visionMissionService) Get(ctx context.Context) (*hotel.VisionMission, error) {
	return s.repo.Get(ctx)
}

func (s *visionMissionService) GetPublic(ctx context.Context) (*hotel.VisionMission, error) {
	return s.repo.GetActive(ctx)
}

func (s *visionMissionService) Upsert(ctx context.Context, req hotel.UpsertVisionMissionRequest, userID *uint) (*hotel.VisionMission, error) {
	return s.repo.Upsert(ctx, req, userID)
}
