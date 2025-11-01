package repohotel

import (
	"backend/internal/models/hotel"
	"gorm.io/gorm"
)

type NewsRepository interface {
	Create(news *hotel.News) error
	Update(news *hotel.News) error
	Delete(id uint) error
	FindByID(id uint) (*hotel.News, error)
	FindBySlug(slug string) (*hotel.News, error)
	FindAll(page, pageSize int, q string, status string) ([]hotel.News, int64, error)
}

type newsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepository{db}
}

func (r *newsRepository) Create(news *hotel.News) error { return r.db.Create(news).Error }
func (r *newsRepository) Update(news *hotel.News) error { return r.db.Save(news).Error }
func (r *newsRepository) Delete(id uint) error          { return r.db.Delete(&hotel.News{}, id).Error }

func (r *newsRepository) FindByID(id uint) (*hotel.News, error) {
	var n hotel.News
	if err := r.db.First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepository) FindBySlug(slug string) (*hotel.News, error) {
	var n hotel.News
	if err := r.db.Where("slug = ?", slug).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepository) FindAll(page, pageSize int, q string, status string) ([]hotel.News, int64, error) {
	var list []hotel.News
	var total int64

	query := r.db.Model(&hotel.News{})
	if q != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
		// kalau minta published, tampilkan yang sudah terbit saja
		if status == "published" {
			query = query.Where("published_at IS NOT NULL AND published_at <= NOW()")
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("published_at DESC, created_at DESC").Limit(pageSize).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
