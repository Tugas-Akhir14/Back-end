// internal/repository/admin/admin_repository.go
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
	Approve(id uint) error
	GetPending() ([]auth.Admin, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db}
}

func (r *adminRepository) Create(admin *auth.Admin) error {
	return r.db.Create(admin).Error
}

func (r *adminRepository) FindByEmail(email string) (*auth.Admin, error) {
	var admin auth.Admin
	err := r.db.Where("email = ?", email).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) FindByID(id uint) (*auth.Admin, error) {
	var admin auth.Admin
	err := r.db.Where("id = ?", id).First(&admin).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) Approve(id uint) error {
	return r.db.Model(&auth.Admin{}).Where("id = ?", id).Update("is_approved", true).Error
}

func (r *adminRepository) GetPending() ([]auth.Admin, error) {
	var admins []auth.Admin
	err := r.db.Where("role LIKE 'admin_%' AND is_approved = ?", false).Find(&admins).Error
	return admins, err
}