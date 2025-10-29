// backend/internal/service/cafeservice/product_service.go
package cafeservice

import (
    "backend/internal/models/cafe"
    "backend/internal/repository/repocafe"
)

type ProductService interface {
    Create(input cafe.ProductCafeCreate) (*cafe.ProductCafe, error)
    GetAll(categoryID *uint) ([]cafe.ProductCafe, error)
    GetByID(id uint) (*cafe.ProductCafe, error)
    Update(id uint, input cafe.ProductCafeUpdate) (*cafe.ProductCafe, error)
    Delete(id uint) error
}

type productService struct {
    repo         repocafe.ProductRepository
    categoryRepo repocafe.CategoryRepository
}

func NewProductService(repo repocafe.ProductRepository, categoryRepo repocafe.CategoryRepository) ProductService {
    return &productService{repo, categoryRepo}
}

func (s *productService) Create(input cafe.ProductCafeCreate) (*cafe.ProductCafe, error) {
    if _, err := s.categoryRepo.FindByID(input.CategoryID); err != nil {
        return nil, err
    }

    product := &cafe.ProductCafe{
        Nama:       input.Nama,
        Deskripsi:  input.Deskripsi,
        Harga:      input.Harga,
        Stok:       input.Stok,
        Gambar:     input.Gambar,
        CategoryID: input.CategoryID,
    }

    if err := s.repo.Create(product); err != nil {
        return nil, err
    }
    return s.repo.FindByID(product.ID)
}

func (s *productService) GetAll(categoryID *uint) ([]cafe.ProductCafe, error) {
    return s.repo.FindAll(categoryID)
}

func (s *productService) GetByID(id uint) (*cafe.ProductCafe, error) {
    return s.repo.FindByID(id)
}

func (s *productService) Update(id uint, input cafe.ProductCafeUpdate) (*cafe.ProductCafe, error) {
    product, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }

    if input.Nama != nil {
        product.Nama = *input.Nama
    }
    if input.Deskripsi != nil {
        product.Deskripsi = *input.Deskripsi
    }
    if input.Harga != nil {
        product.Harga = *input.Harga
    }
    if input.Stok != nil {
        product.Stok = *input.Stok
    }
    if input.Gambar != nil {
        product.Gambar = *input.Gambar
    }
    if input.CategoryID != nil {
        if _, err := s.categoryRepo.FindByID(*input.CategoryID); err != nil {
            return nil, err
        }
        product.CategoryID = *input.CategoryID
    }

    if err := s.repo.Update(product); err != nil {
        return nil, err
    }
    return s.repo.FindByID(id)
}

func (s *productService) Delete(id uint) error {
    return s.repo.Delete(id)
}