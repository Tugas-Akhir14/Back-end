package hotel

import "time"

type GuestReview struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    Rating     int       `gorm:"not null" json:"rating"`
    Comment    string    `gorm:"type:text;not null" json:"comment"`
    GuestName  string    `gorm:"size:100" json:"guest_name,omitempty"`
    IPAddress  string    `gorm:"size:45" json:"ip_address,omitempty"`
    IsApproved bool      `gorm:"default:false" json:"is_approved"`
    CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}