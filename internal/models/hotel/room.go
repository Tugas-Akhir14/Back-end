package hotel

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoomTypeSuperior  = "superior"
	RoomTypeDeluxe    = "deluxe"
	RoomTypeExecutive = "executive"

	RoomStatusAvailable   = "available"
	RoomStatusBooked      = "booked"
	RoomStatusMaintenance = "maintenance"
	RoomStatusCleaning    = "cleaning"
)

var AllowedRoomTypes = []string{RoomTypeSuperior, RoomTypeDeluxe, RoomTypeExecutive}
var AllowedRoomStatuses = []string{RoomStatusAvailable, RoomStatusBooked, RoomStatusMaintenance, RoomStatusCleaning}

type Room struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Number      string         `gorm:"size:32;not null;uniqueIndex:ux_room_number_deleted_at" json:"number"`
	Type        string         `gorm:"size:20;index;not null" json:"type"`
	Price       int64          `gorm:"not null" json:"price"`
	Capacity    int            `gorm:"not null" json:"capacity"`
	Description string         `gorm:"type:text" json:"description"`
	Image       string         `json:"image" binding:"omitempty,url"`
	Status      string         `gorm:"size:20;default:'available';index;not null" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index;uniqueIndex:ux_room_number_deleted_at" json:"-"`
}

type CreateRoomRequest struct {
	Number      string `json:"number" form:"number" binding:"required"`
	Type        string `json:"type" form:"type" binding:"required,oneof=superior deluxe executive"`
	Price       int64  `json:"price" form:"price" binding:"required"`
	Capacity    int    `json:"capacity" form:"capacity" binding:"required"`
	Description string `json:"description" form:"description"`
	Image       string `json:"image" form:"-"` // diisi otomatis
	Status      string `json:"status" form:"status" binding:"omitempty,oneof=available booked maintenance cleaning"`
}

type UpdateRoomRequest struct {
	Number      *string `json:"number" binding:"omitempty"`
	Type        *string `json:"type" binding:"omitempty,oneof=superior deluxe executive"`
	Price       *int64  `json:"price" binding:"omitempty,gte=0"`
	Capacity    *int    `json:"capacity" binding:"omitempty,gte=1"`
	Image       *string `json:"image" binding:"omitempty,url"`
	Description *string `json:"description" binding:"omitempty"`
	Status      *string `json:"status" binding:"omitempty,oneof=available booked maintenance cleaning"`
}
