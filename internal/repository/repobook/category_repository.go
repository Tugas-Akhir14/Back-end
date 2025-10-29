// backend/internal/repository/repobook/category_repository.go
package repobook

import (
	"backend/internal/models/book"
	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(category *book.CategoryBook) error
	FindAll() ([]book.CategoryBook, error)
	FindByID(id uint) (*book.CategoryBook, error)
	Update(category *book.CategoryBook) error
	Delete(id uint) error
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db}
}

func (r *categoryRepository) Create(category *book.CategoryBook) error {
	return r.db.Create(category).Error
}

func (r *categoryRepository) FindAll() ([]book.CategoryBook, error) {
	var categories []book.CategoryBook
	err := r.db.Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) FindByID(id uint) (*book.CategoryBook, error) {
	var category book.CategoryBook
	err := r.db.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Update(category *book.CategoryBook) error {
	return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
	return r.db.Delete(&book.CategoryBook{}, id).Error
}