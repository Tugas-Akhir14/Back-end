// backend/internal/models/cafe/product.go
package cafe

import (

	"time"
)

type ProductCafe struct {
	ID         uint         `json:"id" gorm:"primaryKey"`
	Nama       string       `json:"nama" gorm:"not null;size:255"`
	Deskripsi  string       `json:"deskripsi" gorm:"type:text"`
	Harga      float64      `json:"harga" gorm:"not null"`
	Stok       int          `json:"stok" gorm:"not null;default:0"`
	Gambar     string       `json:"gambar" gorm:"size:500"`
	CategoryID uint         `json:"category_id" gorm:"not null"`
	Category   CategoryCafe `json:"category" gorm:"foreignKey:CategoryID"`
	

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductCafeCreate struct {
	Nama       string  `json:"nama" binding:"required"`
	Deskripsi  string  `json:"deskripsi"`
	Harga      float64 `json:"harga" binding:"required"`
	Stok       int     `json:"stok" binding:"required,min=0"`
	Gambar     string  `json:"gambar"`
	CategoryID uint    `json:"category_id" binding:"required"`
}

type ProductCafeUpdate struct {
	Nama       *string  `json:"nama"`
	Deskripsi  *string  `json:"deskripsi"`
	Harga      *float64 `json:"harga"`
	Stok       *int     `json:"stok"`
	Gambar     *string  `json:"gambar"`
	CategoryID *uint    `json:"category_id"`
}
