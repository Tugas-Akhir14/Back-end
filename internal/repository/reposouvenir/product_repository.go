// backend/internal/repository/reposouvenir/product_repository.go
package reposouvenir

import (
    "backend/internal/models/souvenir"
    "gorm.io/gorm"
)

type ProductRepository interface {
    Create(product *souvenir.Product) error
    GetByID(id uint) (*souvenir.Product, error)
    GetAll(page, limit int) ([]souvenir.Product, int64, error)
    Update(product *souvenir.Product) error
    Delete(id uint) error
    GetByCategoryID(categoryID uint, page, limit int) ([]souvenir.Product, int64, error)
}

type productRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{db}
}

// CREATE
func (r *productRepository) Create(product *souvenir.Product) error {
    return r.db.Create(product).Error
}

// GET BY ID (dengan preload Category)
func (r *productRepository) GetByID(id uint) (*souvenir.Product, error) {
    var product souvenir.Product
    err := r.db.Preload("Category").First(&product, id).Error
    if err != nil {
        return nil, err
    }
    return &product, nil
}

// GET ALL (dengan pagination + preload)
func (r *productRepository) GetAll(page, limit int) ([]souvenir.Product, int64, error) {
    var products []souvenir.Product
    var total int64

    offset := (page - 1) * limit

    err := r.db.Model(&souvenir.Product{}).
        Count(&total).
        Preload("Category").
        Offset(offset).
        Limit(limit).
        Find(&products).Error

    if err != nil {
        return nil, 0, err
    }
    return products, total, nil
}

// UPDATE
func (r *productRepository) Update(product *souvenir.Product) error {
    return r.db.Save(product).Error
}

// DELETE
func (r *productRepository) Delete(id uint) error {
    return r.db.Delete(&souvenir.Product{}, id).Error
}

// GET BY CATEGORY ID (dengan pagination + preload)
func (r *productRepository) GetByCategoryID(categoryID uint, page, limit int) ([]souvenir.Product, int64, error) {
    var products []souvenir.Product
    var total int64

    offset := (page - 1) * limit

    query := r.db.Model(&souvenir.Product{}).Where("category_id = ?", categoryID)

    err := query.Count(&total).
        Preload("Category").
        Offset(offset).
        Limit(limit).
        Find(&products).Error

    if err != nil {
        return nil, 0, err
    }
    return products, total, nil
}