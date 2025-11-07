// backend/internal/models/book/category.go
package book



type CategoryBook struct {
    ID   uint   `json:"id" gorm:"primaryKey"`
    Nama string `json:"nama" gorm:"unique;not null;size:100"` // SESUAI DENGAN SERVICE
    
}