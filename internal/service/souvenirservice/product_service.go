// backend/internal/service/souvenirservice/product_service.go
package souvenirservice

import (
    "backend/internal/models/souvenir"
    "backend/internal/repository/reposouvenir"
)

type ProductService interface {
    CreateProduct(req *souvenir.ProductCreate) (*souvenir.Product, error)
    GetProductByID(id uint) (*souvenir.Product, error)
    GetAllProducts(page, limit int) ([]souvenir.Product, int64, error)
    UpdateProduct(id uint, req *souvenir.ProductUpdate) (*souvenir.Product, error)
    DeleteProduct(id uint) error
    GetProductsByCategoryID(categoryID uint, page, limit int) ([]souvenir.Product, int64, error)
}

type productService struct {
    repo reposouvenir.ProductRepository
}

func NewProductService(repo reposouvenir.ProductRepository) ProductService {
    return &productService{repo}
}

func (s *productService) CreateProduct(req *souvenir.ProductCreate) (*souvenir.Product, error) {
    product := &souvenir.Product{
        Nama:       req.Nama,
        Deskripsi:  req.Deskripsi,
        Harga:      req.Harga,
        Stok:       req.Stok,
        Gambar:     req.Gambar,
        CategoryID: req.CategoryID,
    }
    err := s.repo.Create(product)
    return product, err
}

func (s *productService) GetProductByID(id uint) (*souvenir.Product, error) {
    return s.repo.GetByID(id)
}

func (s *productService) GetAllProducts(page, limit int) ([]souvenir.Product, int64, error) {
    return s.repo.GetAll(page, limit)
}

func (s *productService) UpdateProduct(id uint, req *souvenir.ProductUpdate) (*souvenir.Product, error) {
    product, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    if req.Nama != nil {
        product.Nama = *req.Nama
    }
    if req.Deskripsi != nil {
        product.Deskripsi = *req.Deskripsi
    }
    if req.Harga != nil {
        product.Harga = *req.Harga
    }
    if req.Stok != nil {
        product.Stok = *req.Stok
    }
    if req.Gambar != nil {
        product.Gambar = *req.Gambar
    }
    if req.CategoryID != nil {
        product.CategoryID = *req.CategoryID
    }
    err = s.repo.Update(product)
    return product, err
}

func (s *productService) DeleteProduct(id uint) error {
    return s.repo.Delete(id)
}

func (s *productService) GetProductsByCategoryID(categoryID uint, page, limit int) ([]souvenir.Product, int64, error) {
    return s.repo.GetByCategoryID(categoryID, page, limit)
}