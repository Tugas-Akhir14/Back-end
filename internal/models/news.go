package models

import "time"

type News struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Title       string     `gorm:"size:255;not null" json:"title" form:"title" binding:"required"`
	Slug        string     `gorm:"size:255;uniqueIndex" json:"slug"`
	Content     string     `gorm:"type:text;not null" json:"content" form:"content" binding:"required"`
	ImageURL    string     `gorm:"size:255" json:"image_url"`
	Status      string     `gorm:"size:20;default:draft" json:"status" form:"status"` // draft|published
	PublishedAt *time.Time `json:"published_at"`

	CreatedBy *uint     `json:"created_by"`
	UpdatedBy *uint     `json:"updated_by"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
