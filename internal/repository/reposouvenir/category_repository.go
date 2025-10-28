// backend/internal/repository/reposouvenir/category_repository.go
package reposouvenir

import (
    "backend/internal/models/souvenir"
    "gorm.io/gorm"
)

type CategoryRepository interface {
    Create(cat *souvenir.Category) error
    GetByID(id uint) (*souvenir.Category, error)
    GetAll() ([]souvenir.Category, error)
    Update(cat *souvenir.Category) error
    Delete(id uint) error
}

type categoryRepository struct {
    db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
    return &categoryRepository{db}
}

func (r *categoryRepository) Create(cat *souvenir.Category) error {
    return r.db.Create(cat).Error
}

func (r *categoryRepository) GetByID(id uint) (*souvenir.Category, error) {
    var cat souvenir.Category
    err := r.db.First(&cat, id).Error
    return &cat, err
}

func (r *categoryRepository) GetAll() ([]souvenir.Category, error) {
    var cats []souvenir.Category
    err := r.db.Find(&cats).Error
    return cats, err
}

func (r *categoryRepository) Update(cat *souvenir.Category) error {
    return r.db.Save(cat).Error
}

func (r *categoryRepository) Delete(id uint) error {
    return r.db.Delete(&souvenir.Category{}, id).Error
}