// backend/internal/models/cafe/category.go
package cafe



type CategoryCafe struct {
    ID   uint   `json:"id" gorm:"primaryKey"`
    Nama string `json:"nama" gorm:"unique;not null;size:100"`
    
}