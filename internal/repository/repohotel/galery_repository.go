package repohotel

import (
	"backend/internal/models/hotel"
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
	Create(item *hotel.Gallery) error
	FindByID(id uint) (*hotel.Gallery, error)
	List(f GalleryFilter) ([]hotel.Gallery, int64, error)
	Update(item *hotel.Gallery) error
	Delete(id uint) error // soft delete
}

type galleryRepository struct{ db *gorm.DB }

func NewGalleryRepository(db *gorm.DB) GalleryRepository { return &galleryRepository{db} }

func (r *galleryRepository) Create(item *hotel.Gallery) error { return r.db.Create(item).Error }

func (r *galleryRepository) FindByID(id uint) (*hotel.Gallery, error) {
	var g hotel.Gallery
	if err := r.db.First(&g, id).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *galleryRepository) List(f GalleryFilter) ([]hotel.Gallery, int64, error) {
	var (
		items []hotel.Gallery
		count int64
	)

	// Base query
	q := r.db.Model(&hotel.Gallery{}).Where("galleries.deleted_at IS NULL")

	// Filter spesifik
	if f.RoomID != nil {
		q = q.Where("galleries.room_id = ?", *f.RoomID)
	}

	if f.RoomType != "" {
		if f.IncludeGlobal {
			// sertakan yang room_id NULL
			q = q.Joins("LEFT JOIN rooms ON rooms.id = galleries.room_id").
				Where("rooms.type = ? OR galleries.room_id IS NULL", f.RoomType)
		} else {
			q = q.Joins("JOIN rooms ON rooms.id = galleries.room_id").
				Where("rooms.type = ?", f.RoomType).
				Where("galleries.room_id IS NOT NULL")
		}
	}

	// COUNT pada session terpisah agar SELECT/DISTINCT tidak menempel ke query utama
	qCount := q.Session(&gorm.Session{}) // clone
	if err := qCount.Select("DISTINCT galleries.id").Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Paging default
	if f.Limit <= 0 {
		f.Limit = 12
	}
	if f.Offset < 0 {
		f.Offset = 0
	}

	// Fetch data lengkap
	if err := q.Select("galleries.*").
		Order("galleries.id DESC").
		Limit(f.Limit).
		Offset(f.Offset).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, count, nil
}

func (r *galleryRepository) Update(item *hotel.Gallery) error { return r.db.Save(item).Error }
func (r *galleryRepository) Delete(id uint) error              { return r.db.Delete(&hotel.Gallery{}, id).Error }
