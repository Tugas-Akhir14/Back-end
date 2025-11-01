package repohotel

import (
	"backend/internal/models/hotel"
	"gorm.io/gorm"
	"time"
)

type ReviewRepository interface {
	Create(review *hotel.GuestReview) error
	GetApproved() ([]hotel.GuestReview, error)
	GetPending() ([]hotel.GuestReview, error)
	Approve(id uint) error
	Delete(id uint) error
	CheckRateLimit(ip string, minutes int) (bool, error) // TAMBAH INI
}

type repo struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &repo{db}
}

// TAMBAH METHOD INI
func (r *repo) CheckRateLimit(ip string, minutes int) (bool, error) {
	var count int64
	cutoff := time.Now().Add(-1 * time.Duration(minutes) * time.Minute)
	err := r.db.Model(&hotel.GuestReview{}).
		Where("ip_address = ? AND created_at > ?", ip, cutoff).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repo) Create(review *hotel.GuestReview) error {
	return r.db.Create(review).Error
}

func (r *repo) GetApproved() ([]hotel.GuestReview, error) {
	var reviews []hotel.GuestReview
	err := r.db.Where("is_approved = ?", true).Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *repo) GetPending() ([]hotel.GuestReview, error) {
	var reviews []hotel.GuestReview
	err := r.db.Where("is_approved = ?", false).Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *repo) Approve(id uint) error {
	return r.db.Model(&hotel.GuestReview{}).Where("id = ?", id).Update("is_approved", true).Error
}

func (r *repo) Delete(id uint) error {
	return r.db.Delete(&hotel.GuestReview{}, id).Error
}