package hotel

import (
	"time"
    "backend/internal/models/auth"
)

type GuestReview struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Rating     int       `gorm:"not null" json:"rating"`
	Comment    string    `gorm:"type:text;not null" json:"comment"`
	GuestName  string    `gorm:"size:100" json:"guest_name,omitempty"`
	IPAddress  string    `gorm:"size:45" json:"ip_address,omitempty"`
	IsApproved bool      `gorm:"default:false" json:"is_approved"`
	AdminID    uint      `gorm:"index" json:"admin_id"` // <-- Pakai Admin
	Admin      auth.Admin     `gorm:"foreignKey:AdminID"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}
