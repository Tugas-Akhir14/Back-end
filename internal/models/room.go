package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoomTypeSuperior  = "superior"
	RoomTypeDeluxe    = "deluxe"
	RoomTypeExecutive = "executive"
)

var AllowedRoomTypes = []string{RoomTypeSuperior, RoomTypeDeluxe, RoomTypeExecutive}

type Room struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    Number      string         `gorm:"size:32;not null;uniqueIndex:ux_room_number_deleted_at" json:"number"`
    Type        string         `gorm:"size:20;index;not null" json:"type"`
    Price       int64          `gorm:"not null" json:"price"`
    Capacity    int            `gorm:"not null" json:"capacity"`
    Description string         `gorm:"type:text" json:"description"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index;uniqueIndex:ux_room_number_deleted_at" json:"-"`

	Galleries  []Gallery      `gorm:"foreignKey:RoomID" json:"galleries,omitempty"`
}


 

// DTO untuk binding request
type CreateRoomRequest struct {
	Number      string `json:"number" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=superior deluxe executive"`
	Price       int64  `json:"price" binding:"required,gte=0"`
	Capacity    int    `json:"capacity" binding:"required,gte=1"`
	Description string `json:"description"`
}

type UpdateRoomRequest struct {
	Number      *string `json:"number" binding:"omitempty"`
	Type        *string `json:"type" binding:"omitempty,oneof=superior deluxe executive"`
	Price       *int64  `json:"price" binding:"omitempty,gte=0"`
	Capacity    *int    `json:"capacity" binding:"omitempty,gte=1"`
	Description *string `json:"description" binding:"omitempty"`
}
