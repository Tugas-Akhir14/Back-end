// internal/repository/admin/admin_repository.go
package admin

import (
	"backend/internal/models/auth"
	"errors"
	"time"

	"gorm.io/gorm"
)

type AdminRepository interface {
	Create(admin *auth.Admin) error
	FindByEmail(email string) (*auth.Admin, error)
	FindByID(id uint) (*auth.Admin, error)
	Approve(id uint) error
	GetPending() ([]auth.Admin, error)
	Update(admin *auth.Admin) error
	SaveOTP(email, otp string, expiresAt time.Time) error
    VerifyOTP(email, otp string) (bool, error)
    DeleteOTP(email string) error
	FindHotelAdmin() (*auth.Admin, error)

}

type adminRepository struct {
	db *gorm.DB
}

func (r *adminRepository) FindHotelAdmin() (*auth.Admin, error) {
	var admin auth.Admin

	err := r.db.
		Where("role = ? AND is_approved = ?", auth.RoleAdminHotel, true).
		First(&admin).Error

	if err != nil {
		return nil, err
	}
	return &admin, nil
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
    // Tampilkan SEMUA role admin yang belum approved (kecuali superadmin & guest)
    err := r.db.
        Where("is_approved = ? AND role != ? AND role != ?", false, auth.RoleSuperAdmin, auth.RoleGuest).
        Find(&admins).Error
    return admins, err
}

func (r *adminRepository) Update(admin *auth.Admin) error {
	return r.db.Save(admin).Error
}

func (r *adminRepository) SaveOTP(email, otp string, expiresAt time.Time) error {
	r.db.Where("email = ?", email).Delete(&auth.GuestOTP{})
	return r.db.Create(&auth.GuestOTP{Email: email, OTP: otp, ExpiresAt: expiresAt}).Error
}

func (r *adminRepository) VerifyOTP(email, otp string) (bool, error) {
	var record auth.GuestOTP
	err := r.db.Where("email = ? AND otp = ? AND expires_at > NOW()", email, otp).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *adminRepository) DeleteOTP(email string) error {
	return r.db.Where("email = ?", email).Delete(&auth.GuestOTP{}).Error
}