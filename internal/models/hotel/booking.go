// internal/models/hotel/booking.go
package hotel

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type BookingStatus string

const (
	BookingStatusPending     BookingStatus = "pending"
	BookingStatusConfirmed   BookingStatus = "confirmed"
	BookingStatusCancelled   BookingStatus = "cancelled"
	BookingStatusCheckedIn   BookingStatus = "checked_in"
	BookingStatusCheckedOut  BookingStatus = "checked_out"
)

func (s BookingStatus) String() string {
	return string(s)
}

type Booking struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	RoomID      uint            `gorm:"not null" json:"room_id"`
	UserID      *uint           `json:"user_id,omitempty"`
	Room        Room            `gorm:"foreignKey:RoomID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"room,omitempty"`
	Name        string          `gorm:"size:100;not null" json:"name"`
	Phone       string          `gorm:"size:20;not null" json:"phone"`
	Email       string          `gorm:"size:100" json:"email"`
	CheckIn     time.Time       `gorm:"not null" json:"check_in"`
	CheckOut    time.Time       `gorm:"not null" json:"check_out"`
	Guests      int             `gorm:"not null" json:"guests"`
	TotalNights int             `gorm:"not null" json:"total_nights"`
	TotalPrice  int64           `gorm:"not null" json:"total_price"`
	Status      BookingStatus   `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Notes       string          `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index" json:"-"`
}

// Hook: Validasi sebelum create
func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	if b.CheckIn.After(b.CheckOut) || b.CheckIn.Equal(b.CheckOut) {
		return fmt.Errorf("check_in must be before check_out")
	}
	if b.Guests <= 0 {
		return fmt.Errorf("guests must be greater than 0")
	}
	if b.TotalNights <= 0 {
		return fmt.Errorf("total_nights must be greater than 0")
	}
	if b.TotalPrice <= 0 {
		return fmt.Errorf("total_price must be greater than 0")
	}
	return nil
}

// Request structs
type CreateBookingRequest struct {
	RoomID   uint   `json:"room_id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	CheckIn  string `json:"check_in" binding:"required"`
	CheckOut string `json:"check_out" binding:"required"`
	Guests   int    `json:"guests" binding:"required,gt=0"`
	Notes    string `json:"notes,omitempty"`
}

type GuestBookingRequest struct {
	RoomType   string `json:"room_type" binding:"required,oneof=superior deluxe executive"`
	TotalRooms int    `json:"total_rooms" binding:"required,gt=0"`
	Name       string `json:"name" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	CheckIn    string `json:"check_in" binding:"required"`
	CheckOut   string `json:"check_out" binding:"required"`
	Guests     int    `json:"guests" binding:"required,gt=0"`
	Notes      string `json:"notes,omitempty"`
}

type BookingResponse struct {
	ID          uint   `json:"id"`
	WhatsAppURL string `json:"whatsapp_url"`
}

type GuestBookingResponse struct {
	BookingIDs  []uint `json:"booking_ids"`
	WhatsAppURL string `json:"whatsapp_url"`
}

type AvailabilityRequest struct {
	CheckIn string `form:"check_in" binding:"required"`
	CheckOut string `form:"check_out" binding:"required"`
	Type     string `form:"type,omitempty"`
}

type AvailabilityResponse struct {
	Type           string `json:"type"`
	PricePerNight  int64  `json:"price_per_night"`
	AvailableRooms int    `json:"available_rooms"`
	TotalRooms     int    `json:"total_rooms"`
}




