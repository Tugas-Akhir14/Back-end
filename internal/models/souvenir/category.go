// backend/internal/models/souvenir/category.go
package souvenir

import (
	"time"
)

type Category struct {
	ID        uint       `json:"id" gorm:"primaryKey"`
	Nama      string     `json:"nama" gorm:"not null;size:100"`
	Slug      string     `json:"slug" gorm:"unique;not null;size:100"`
	Deskripsi string     `json:"deskripsi" gorm:"type:text"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Untuk Create
type CategoryCreate struct {
	Nama      string `json:"nama" binding:"required"`
	Slug      string `json:"slug" binding:"required"`
	Deskripsi string `json:"deskripsi"`
}

// Untuk Update (partial)
type CategoryUpdate struct {
	Nama      *string `json:"nama"`
	Slug      *string `json:"slug"`
	Deskripsi *string `json:"deskripsi"`
}
