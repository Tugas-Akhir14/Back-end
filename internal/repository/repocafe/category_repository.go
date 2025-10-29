// backend/internal/repository/repocafe/category_repository.go
package repocafe

import (
    "backend/internal/models/cafe"
    "gorm.io/gorm"
)

type CategoryRepository interface {
    Create(category *cafe.CategoryCafe) error
    FindAll() ([]cafe.CategoryCafe, error)
    FindByID(id uint) (*cafe.CategoryCafe, error)
    Update(category *cafe.CategoryCafe) error
    Delete(id uint) error
}

type categoryRepository struct {
    db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
    return &categoryRepository{db}
}

func (r *categoryRepository) Create(category *cafe.CategoryCafe) error {
    return r.db.Create(category).Error
}

func (r *categoryRepository) FindAll() ([]cafe.CategoryCafe, error) {
    var categories []cafe.CategoryCafe
    err := r.db.Find(&categories).Error
    return categories, err
}

func (r *categoryRepository) FindByID(id uint) (*cafe.CategoryCafe, error) {
    var category cafe.CategoryCafe
    err := r.db.First(&category, id).Error
    if err != nil {
        return nil, err
    }
    return &category, nil
}

func (r *categoryRepository) Update(category *cafe.CategoryCafe) error {
    return r.db.Save(category).Error
}

func (r *categoryRepository) Delete(id uint) error {
    return r.db.Delete(&cafe.CategoryCafe{}, id).Error
}