package repobook

import (
    "backend/internal/models/book"
    "gorm.io/gorm"
)

type ProductRepository interface {
    Create(product *book.ProductBook) error
    FindAll(categoryID *uint) ([]book.ProductBook, error)
    FindByID(id uint) (*book.ProductBook, error)
    Update(product *book.ProductBook) error
    Delete(id uint) error
}

type productRepository struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepository{db}
}

func (r *productRepository) Create(product *book.ProductBook) error {
    return r.db.Create(product).Error
}

func (r *productRepository) FindAll(categoryID *uint) ([]book.ProductBook, error) {
    var products []book.ProductBook
    query := r.db.Preload("Category")
    if categoryID != nil {
        query = query.Where("category_id = ?", *categoryID)
    }
    return products, query.Find(&products).Error
}

func (r *productRepository) FindByID(id uint) (*book.ProductBook, error) {
    var product book.ProductBook
    err := r.db.Preload("Category").First(&product, id).Error
    return &product, err
}

func (r *productRepository) Update(product *book.ProductBook) error {
    return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint) error {
    return r.db.Delete(&book.ProductBook{}, id).Error
}