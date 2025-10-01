package repository

import (
	"backend/internal/models"
	"gorm.io/gorm"
	"errors"
)

type AdminRepository interface {
	Create(admin *models.Admin) error
	FindByEmail(email string) (*models.Admin, error)
	FindByID(id uint) (*models.Admin, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db}
}

func (r *adminRepository) Create(admin *models.Admin) error {
    if err := r.db.Create(admin).Error; err != nil {
        return err // Menangani error jika ada masalah saat penyimpanan
    }
    return nil
}


// FindByEmail finds an admin by email
func (r *adminRepository) FindByEmail(email string) (*models.Admin, error) {
	var admin models.Admin
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
func (r *adminRepository) FindByID(id uint) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Where("id = ?", id).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no record is found
		}
		return nil, err // Return the error if something else goes wrong
	}
	return &admin, nil
}
