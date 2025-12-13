// internal/models/auth/guest_otp.go
package auth

import "time"

type GuestOTP struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"unique;not null"`
	OTP       string    `gorm:"size:6;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}