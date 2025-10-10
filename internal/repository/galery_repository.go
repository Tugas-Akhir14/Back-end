package repository

import (
	"backend/internal/models"
	"gorm.io/gorm"
)

type GalleryFilter struct {
	RoomID        *uint
	RoomType      string // superior | deluxe | executive
	IncludeGlobal bool   // jika true + RoomType, ikutkan galleries.room_id IS NULL
	Limit         int
	Offset        int
}

type GalleryRepository interface {
	Create(item *models.Gallery) error
	FindByID(id uint) (*models.Gallery, error)
	List(f GalleryFilter) ([]models.Gallery, int64, error)
	Update(item *models.Gallery) error
	Delete(id uint) error // soft delete
}

type galleryRepository struct{ db *gorm.DB }

func NewGalleryRepository(db *gorm.DB) GalleryRepository { return &galleryRepository{db} }

func (r *galleryRepository) Create(item *models.Gallery) error { return r.db.Create(item).Error }

func (r *galleryRepository) FindByID(id uint) (*models.Gallery, error) {
	var g models.Gallery
	if err := r.db.First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *galleryRepository) List(f GalleryFilter) ([]models.Gallery, int64, error) {
	var (
		items []models.Gallery
		count int64
		q     = r.db.Model(&models.Gallery{})
	)

	if f.RoomID != nil {
		q = q.Where("galleries.room_id = ?", *f.RoomID)
	}

	if f.RoomType != "" {
		q = q.Joins("JOIN rooms ON rooms.id = galleries.room_id").
			Where("rooms.type = ?", f.RoomType)
		if !f.IncludeGlobal {
			q = q.Where("galleries.room_id IS NOT NULL")
		}
	}

	q.Distinct("galleries.id").Count(&count)

	if f.Limit <= 0 {
		f.Limit = 12
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	if err := q.Order("galleries.id DESC").
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, count, nil
}

func (r *galleryRepository) Update(item *models.Gallery) error { return r.db.Save(item).Error }
func (r *galleryRepository) Delete(id uint) error              { return r.db.Delete(&models.Gallery{}, id).Error }
