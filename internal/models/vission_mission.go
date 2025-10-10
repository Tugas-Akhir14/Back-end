package models

import (
	"time"

	"gorm.io/datatypes"
)

type VisionMission struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Vision    string         `json:"vision" gorm:"type:text;not null"`
	Missions  datatypes.JSON `json:"missions" gorm:"type:json;not null"` // contoh: ["Meningkatkan...","Mengedepankan..."]
	Active    bool           `json:"active" gorm:"type:bool;default:true"`

	CreatedBy *uint      `json:"created_by"`
	UpdatedBy *uint      `json:"updated_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ====== DTO / Request ======
type UpsertVisionMissionRequest struct {
	Vision   string   `json:"vision" binding:"required,min=3"`
	Missions []string `json:"missions" binding:"required,min=1,dive,min=3"`
	Active   *bool    `json:"active"` // opsional
}
