package repository

import (
	"backend/internal/models"

	"gorm.io/gorm"
)

type NewsRepository interface {
	Create(news *models.News) error
	Update(news *models.News) error
	Delete(id uint) error
	FindByID(id uint) (*models.News, error)
	FindBySlug(slug string) (*models.News, error)
	FindAll(page, pageSize int, q string, status string) ([]models.News, int64, error)
}

type newsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return &newsRepository{db}
}

func (r *newsRepository) Create(news *models.News) error {
	return r.db.Create(news).Error
}

func (r *newsRepository) Update(news *models.News) error {
	return r.db.Save(news).Error
}

func (r *newsRepository) Delete(id uint) error {
	return r.db.Delete(&models.News{}, id).Error
}

func (r *newsRepository) FindByID(id uint) (*models.News, error) {
	var n models.News
	if err := r.db.First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepository) FindBySlug(slug string) (*models.News, error) {
	var n models.News
	if err := r.db.Where("slug = ?", slug).First(&n).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *newsRepository) FindAll(page, pageSize int, q string, status string) ([]models.News, int64, error) {
	var list []models.News
	var total int64

	query := r.db.Model(&models.News{})
	if q != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
