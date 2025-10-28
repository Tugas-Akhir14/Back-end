package repohotel

import (
	"context"
	"encoding/json"
	"errors"

	"backend/internal/models/hotel"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type VisionMissionRepository interface {
	Get(ctx context.Context) (*hotel.VisionMission, error)                               // ambil entri aktif (atau entri pertama jika ada)
	Upsert(ctx context.Context, payload hotel.UpsertVisionMissionRequest, userID *uint) (*hotel.VisionMission, error)
}

type visionMissionRepository struct {
	db *gorm.DB
}

func NewVisionMissionRepository(db *gorm.DB) VisionMissionRepository {
	return &visionMissionRepository{db: db}
}

func (r *visionMissionRepository) Get(ctx context.Context) (*hotel.VisionMission, error) {
	var row hotel.VisionMission
	tx := r.db.WithContext(ctx).
		Order("id ASC").
		Limit(1).
		Find(&row)

	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &row, nil
}

func (r *visionMissionRepository) Upsert(ctx context.Context, payload hotel.UpsertVisionMissionRequest, userID *uint) (*hotel.VisionMission, error) {
	// serialize missions → JSON
	b, err := json.Marshal(payload.Missions)
	if err != nil {
		return nil, err
	}

	var existing hotel.VisionMission
	tx := r.db.WithContext(ctx).Order("id ASC").Limit(1).Find(&existing)
	if tx.Error != nil {
		return nil, tx.Error
	}

	// set active default
	active := true
	if payload.Active != nil {
		active = *payload.Active
	}

	if tx.RowsAffected == 0 {
		// create
		row := hotel.VisionMission{
			Vision:    payload.Vision,
			Missions:  datatypes.JSON(b),
			Active:    active,
			CreatedBy: userID,
			UpdatedBy: userID,
		}
		if err := r.db.WithContext(ctx).Create(&row).Error; err != nil {
			return nil, err
		}
		return &row, nil
	}

	// update first row
	existing.Vision = payload.Vision
	existing.Missions = datatypes.JSON(b)
	existing.Active = active
	existing.UpdatedBy = userID

	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

// Helper (opsional) kalau mau force not found → 404
var ErrNotFound = errors.New("vision_mission not found")
