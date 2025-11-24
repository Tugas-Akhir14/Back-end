// internal/models/hotel/booking.go
package hotel

import (
    "time"
    "gorm.io/gorm"
    
)

type BookingStatus string

const (
    BookingStatusPending    BookingStatus = "pending"
    BookingStatusConfirmed  BookingStatus = "confirmed"
    BookingStatusCancelled  BookingStatus = "cancelled"
    BookingStatusCheckedIn  BookingStatus = "checked_in"
    BookingStatusCheckedOut BookingStatus = "checked_out"
)

type Booking struct {
    ID          uint            `gorm:"primaryKey" json:"id"`
    RoomID      uint            `gorm:"not null;index" json:"room_id"`
    UserID      *uint           `json:"user_id,omitempty"`
    Room        Room            `gorm:"foreignKey:RoomID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"room,omitempty"`
    Name        string          `gorm:"size:100;not null" json:"name"`
    Phone       string          `gorm:"size:20;not null" json:"phone"`
    Email       string          `gorm:"size:100" json:"email"`
    CheckIn     time.Time       `gorm:"not null" json:"check_in"`
    CheckOut    time.Time       `gorm:"not null" json:"check_out"`
    Guests      int             `gorm:"not null" json:"guests"`           // TOTAL TAMU (semua kamar)
    Rooms       int             `gorm:"not null;default:1" json:"rooms"`  // JUMLAH KAMAR YANG DIPESAN
    TotalNights int             `gorm:"not null" json:"total_nights"`
    TotalPrice  int64           `gorm:"not null" json:"total_price"`
    ExtraGuests int             `gorm:"default:0" json:"extra_guests"`
    ExtraCharge int64           `gorm:"default:0" json:"extra_charge"`
    Status      BookingStatus   `gorm:"type:varchar(20);default:'pending';index" json:"status"`
    Notes       string          `gorm:"type:text" json:"notes,omitempty"`

    // HAPUS KOLOM INI! GAK PERLU LAGI!
    // ActiveRoomKey string `gorm:"column:active_room_key;type:varchar(50);uniqueIndex:idx_active_room_key" json:"-"`

    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
    DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

    // HOOKS: HAPUS LOGIC active_room_key (gak perlu lagi!)
    // func (b *Booking) BeforeCreate(tx *gorm.DB) error { ... }
    // func (b *Booking) BeforeUpdate(tx *gorm.DB) error { ... }
    // func (b *Booking) AfterFind(tx *gorm.DB) error { ... }
}

// REQUEST STRUCTS (SESUAI BACKEND BARU)
type CreateBookingRequest struct {
    RoomType string `json:"room_type" binding:"required,oneof=superior deluxe executive"`
    Rooms    int    `json:"rooms" binding:"required,gt=0"`
    Name     string `json:"name" binding:"required"`
    Phone    string `json:"phone" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    CheckIn  string `json:"check_in" binding:"required"`
    CheckOut string `json:"check_out" binding:"required"`
    Guests   int    `json:"guests" binding:"required,gt=0"`
    Notes    string `json:"notes,omitempty"`
}

type BookingResponse struct {
    ID          uint   `json:"id"`
    WhatsAppURL string `json:"whatsapp_url"`
}

type MultiBookingResponse struct {
    BookingIDs  []uint `json:"booking_ids"`
    WhatsAppURL string `json:"whatsapp_url"`
}

type AvailabilityRequest struct {
    CheckIn string `form:"check_in" binding:"required"`
    CheckOut string `form:"check_out" binding:"required"`
    Type     string `form:"type,omitempty"`
}

type AvailabilityResponse struct {
    RoomType       string `json:"room_type"`
    PricePerNight  int64  `json:"price_per_night"`
    AvailableRooms int    `json:"available_rooms"`
    TotalRooms     int    `json:"total_rooms"`
}

type UpdateBookingRequest struct {
    CheckIn *string `json:"check_in,omitempty"`
    CheckOut *string `json:"check_out,omitempty"`
    Guests  *int    `json:"guests,omitempty"`
    Notes   *string `json:"notes,omitempty"`
}

type UpdateBookingStatusRequest struct {
    Status string `json:"status" binding:"required,oneof=pending confirmed cancelled checked_in checked_out"`
}