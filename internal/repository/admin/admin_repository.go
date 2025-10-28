package admin

import (
	"backend/internal/models/auth"
	"gorm.io/gorm"
	"errors"
)

type AdminRepository interface {
	Create(admin *auth.Admin) error
	FindByEmail(email string) (*auth.Admin, error)
	FindByID(id uint) (*auth.Admin, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db}
}

func (r *adminRepository) Create(admin *auth.Admin) error {
    if err := r.db.Create(admin).Error; err != nil {
        return err // Menangani error jika ada masalah saat penyimpanan
    }
    return nil
}


// FindByEmail finds an admin by email
func (r *adminRepository) FindByEmail(email string) (*auth.Admin, error) {
	var admin auth.Admin
	err := r.db.Where("email = ?", email).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no record is found
		}
		return nil, err // Return the error if something else goes wrong
	}
	return &admin, nil
}

// FindByID finds an admin by ID
func (r *adminRepository) FindByID(id uint) (*auth.Admin, error) {
	var admin auth.Admin
	err := r.db.Where("id = ?", id).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no record is found
		}
		return nil, err // Return the error if something else goes wrong
	}
	return &admin, nil
}
