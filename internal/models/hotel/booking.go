// backend/internal/models/hotel/booking.go
package hotel

import (
	"time"

	"gorm.io/gorm"
)

const (
	BookingStatusPending    = "pending"
	BookingStatusConfirmed  = "confirmed"
	BookingStatusCancelled  = "cancelled"
	BookingStatusCheckedIn  = "checked_in"
	BookingStatusCheckedOut = "checked_out"
)

type Booking struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	RoomID      uint           `gorm:"not null" json:"room_id"`
	Room        Room           `gorm:"foreignKey:RoomID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"room"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Phone       string         `gorm:"size:20;not null" json:"phone"`
	Email       string         `gorm:"size:100" json:"email"`
	CheckIn     time.Time      `gorm:"not null" json:"check_in"`
	CheckOut    time.Time      `gorm:"not null" json:"check_out"`
	Guests      int            `gorm:"default:1" json:"guests"`
	TotalNights int            `json:"total_nights"`
	TotalPrice  int64          `json:"total_price"`
	Status      string         `gorm:"size:20;default:'pending';index" json:"status"`
	Notes       string         `gorm:"type:text" json:"notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateBookingRequest struct {
	RoomID    uint   `json:"room_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
	Email     string `json:"email" binding:"omitempty,email"`
	CheckIn   string `json:"check_in" binding:"required"` // format: YYYY-MM-DD
	CheckOut  string `json:"check_out" binding:"required"`
	Guests    int    `json:"guests" binding:"omitempty,gte=1"`
	Notes     string `json:"notes" binding:"omitempty"`
}

type BookingResponse struct {
	ID          uint   `json:"id"`
	WhatsAppURL string `json:"whatsapp_url"`
}