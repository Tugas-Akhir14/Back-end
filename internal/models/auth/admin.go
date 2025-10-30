// internal/models/auth/admin.go
package auth

import "time"

type Role string

const (
	RoleSuperAdmin    Role = "superadmin"
	RoleAdminHotel    Role = "admin_hotel"
	RoleAdminSouvenir Role = "admin_souvenir"
	RoleAdminBuku     Role = "admin_buku"
	RoleAdminCafe     Role = "admin_cafe"
	RoleGuest         Role = "guest"
)

type Admin struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	FullName        string    `gorm:"not null" form:"full_name" json:"full_name" binding:"required"`
	Email           string    `gorm:"unique;not null" form:"email" json:"email" binding:"required,email"`
	PhoneNumber     string    `gorm:"not null" form:"phone_number" json:"phone_number" binding:"required"`
	Password        string    `gorm:"not null" form:"password" json:"-" binding:"required,min=8"`
	ConfirmPassword string    `form:"confirm_password" json:"-" binding:"required"`
	Role            Role      `gorm:"type:varchar(50);not null;default:'guest'" json:"role"`
	IsApproved      bool      `gorm:"default:false" json:"is_approved"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"-"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"-"`
}

type AdminResponse struct {
	ID          uint   `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Role        Role   `json:"role"`
	IsApproved  bool   `json:"is_approved"`
}