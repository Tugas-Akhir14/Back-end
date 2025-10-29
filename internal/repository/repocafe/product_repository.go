// backend/internal/repository/repocafe/product_repository.go
package repocafe

import (
    "backend/internal/models/cafe"
    "gorm.io/gorm"
)

type ProductRepository interface {
    Create(product *cafe.ProductCafe) error
    FindAll(categoryID *uint) ([]cafe.ProductCafe, error)
    FindByID(id uint) (*cafe.ProductCafe, error)
    Update(product *cafe.ProductCafe) error
    Delete(id uint) error
}

type productRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{db}
}

func (r *productRepository) Create(product *cafe.ProductCafe) error {
    return r.db.Create(product).Error
}

func (r *productRepository) FindAll(categoryID *uint) ([]cafe.ProductCafe, error) {
    var products []cafe.ProductCafe
    query := r.db.Preload("Category")
    if categoryID != nil {
        query = query.Where("category_id = ?", *categoryID)
    }
    return products, query.Find(&products).Error
}

func (r *productRepository) FindByID(id uint) (*cafe.ProductCafe, error) {
    var product cafe.ProductCafe
    err := r.db.Preload("Category").First(&product, id).Error
    return &product, err
}

func (r *productRepository) Update(product *cafe.ProductCafe) error {
    return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint) error {
    return r.db.Delete(&cafe.ProductCafe{}, id).Error
}