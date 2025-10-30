// backend/internal/models/souvenir/product.go
package souvenir

import "time"

type Product struct {
    ID         uint      `json:"id" gorm:"primaryKey"`
    Nama       string    `json:"nama" gorm:"not null;size:255"`
    Deskripsi  string    `json:"deskripsi" gorm:"type:text"`
    Harga      float64   `json:"harga" gorm:"not null"`
    Stok       int       `json:"stok" gorm:"not null;default:0"`
    Gambar     string    `json:"gambar" gorm:"size:1000"` // lebih panjang untuk multiple URL
    CategoryID uint      `json:"category_id" gorm:"not null"`
    Category   Category  `json:"category" gorm:"foreignKey:CategoryID"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type ProductCreate struct {
    Nama       string  `form:"nama" json:"nama" binding:"required"`
    Deskripsi  string  `form:"deskripsi" json:"deskripsi"`
    Harga      float64 `form:"harga" json:"harga" binding:"required"`
    Stok       int     `form:"stok" json:"stok" binding:"required,min=0"`
    Gambar     string  `form:"-" json:"gambar"` // diisi manual
    CategoryID uint    `form:"category_id" json:"category_id" binding:"required"`
}

type ProductUpdate struct {
    Nama       *string `form:"nama" json:"nama"`
    Deskripsi  *string `form:"deskripsi" json:"deskripsi"`
    Harga      *float64 `form:"harga" json:"harga"`
    Stok       *int    `form:"stok" json:"stok"`
    Gambar     *string `form:"-" json:"gambar"` // diisi manual
    CategoryID *uint   `form:"category_id" json:"category_id"`
}