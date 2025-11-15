// internal/models/hotel/room_type.go
package hotel

import (
	"time"

	"gorm.io/gorm"
)

var ValidRoomTypes = map[string]bool{
	"superior":  true,
	"deluxe":    true,
	"executive": true,
}

type RoomType struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Type        string         `gorm:"size:20;uniqueIndex;not null" json:"type"`
	Price       int64          `gorm:"not null;column:price"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Request
type CreateRoomTypeRequest struct {
	Type        string `json:"type" binding:"required,oneof=superior deluxe executive"`
	Price       int64  `json:"price" binding:"required,gte=0"`
	Description string `json:"description" binding:"required"`
}

type UpdateRoomTypeRequest struct {
	Price       *int64  `json:"price" binding:"omitempty,gte=0"`
	Description *string `json:"description" binding:"omitempty"`
}