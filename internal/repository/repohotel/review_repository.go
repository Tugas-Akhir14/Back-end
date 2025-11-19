package repohotel

import (
	"backend/internal/models/hotel"
	"gorm.io/gorm"
	
)

type ReviewRepository interface {
	Create(review *hotel.GuestReview) error
	GetAll() ([]hotel.GuestReview, error)
	GetByID(id uint) (*hotel.GuestReview, error)
	GetByUserID(userID uint) ([]hotel.GuestReview, error)
	Update(review *hotel.GuestReview) error
	Delete(id uint) error
	HasUserReviewed(userID uint) (bool, error) // BARU: cek apakah user sudah pernah review
}

type repo struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &repo{db}
}

func (r *repo) HasUserReviewed(userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&hotel.GuestReview{}).
		Where("admin_id = ?", userID).
		Limit(1).
		Count(&count).Error

	return count > 0, err
}

func (r *repo) Create(review *hotel.GuestReview) error {
	return r.db.Create(review).Error
}

func (r *repo) GetAll() ([]hotel.GuestReview, error) {
	var reviews []hotel.GuestReview
	err := r.db.Preload("Admin").Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *repo) GetByID(id uint) (*hotel.GuestReview, error) {
	var review hotel.GuestReview
	err := r.db.Preload("Admin").First(&review, id).Error
	return &review, err
}

func (r *repo) GetByUserID(userID uint) ([]hotel.GuestReview, error) {
	var reviews []hotel.GuestReview
	err := r.db.Preload("Admin").Where("admin_id = ?", userID).Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *repo) Update(review *hotel.GuestReview) error {
	return r.db.Save(review).Error
}

func (r *repo) Delete(id uint) error {
	return r.db.Delete(&hotel.GuestReview{}, id).Error
}