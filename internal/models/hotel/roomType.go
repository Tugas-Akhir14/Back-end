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
	ID                   uint           `gorm:"primaryKey" json:"id"`
	Type                 string         `gorm:"size:20;uniqueIndex;not null" json:"type"`
	BasePrice            int64          `gorm:"not null;column:base_price" json:"base_price"`
	DiscountPercentage   float64        `gorm:"default:0" json:"discount_percentage"`
	DiscountStart        *time.Time     `json:"discount_start,omitempty"`
	DiscountEnd          *time.Time     `json:"discount_end,omitempty"`
	DiscountDescription  string         `gorm:"type:text" json:"discount_description,omitempty"`
	Description          string         `gorm:"type:text" json:"description"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}

// GetCurrentPrice calculates the current price considering the discount if active
func (rt *RoomType) GetCurrentPrice() int64 {
	if rt.DiscountPercentage > 0 && rt.DiscountStart != nil && rt.DiscountEnd != nil {
		now := time.Now()
		if now.After(*rt.DiscountStart) && now.Before(*rt.DiscountEnd) {
			discountAmount := float64(rt.BasePrice) * (rt.DiscountPercentage / 100)
			return rt.BasePrice - int64(discountAmount)
		}
	}
	return rt.BasePrice
}

// Request
type CreateRoomTypeRequest struct {
	Type                 string     `json:"type" binding:"required,oneof=superior deluxe executive"`
	BasePrice            int64      `json:"base_price" binding:"required,gte=0"`
	DiscountPercentage   float64    `json:"discount_percentage" binding:"omitempty,gte=0,lte=100"`
	DiscountStart        *time.Time `json:"discount_start,omitempty" binding:"omitempty"`
	DiscountEnd          *time.Time `json:"discount_end,omitempty" binding:"omitempty"`
	DiscountDescription  string     `json:"discount_description" binding:"omitempty"`
	Description          string     `json:"description" binding:"required"`
}

type UpdateRoomTypeRequest struct {
	BasePrice            *int64      `json:"base_price" binding:"omitempty,gte=0"`
	DiscountPercentage   *float64    `json:"discount_percentage" binding:"omitempty,gte=0,lte=100"`
	DiscountStart        *time.Time  `json:"discount_start,omitempty" binding:"omitempty"`
	DiscountEnd          *time.Time  `json:"discount_end,omitempty" binding:"omitempty"`
	DiscountDescription  *string     `json:"discount_description" binding:"omitempty"`
	Description          *string     `json:"description" binding:"omitempty"`
}