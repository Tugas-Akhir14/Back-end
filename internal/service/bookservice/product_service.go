package bookservice

import (
    "backend/internal/models/book"
    "backend/internal/repository/repobook"
)

type ProductService interface {
    Create(input book.ProductBookCreate) (*book.ProductBook, error)
    GetAll(categoryID *uint) ([]book.ProductBook, error)
    GetByID(id uint) (*book.ProductBook, error)
    Update(id uint, input book.ProductBookUpdate) (*book.ProductBook, error)
    Delete(id uint) error
}

type productService struct {
    repo         repobook.ProductRepository
    categoryRepo repobook.CategoryRepository
}

func NewProductService(repo repobook.ProductRepository, categoryRepo repobook.CategoryRepository) ProductService {
    return &productService{repo, categoryRepo}
}

func (s *productService) Create(input book.ProductBookCreate) (*book.ProductBook, error) {
    // Validasi kategori exists
    if _, err := s.categoryRepo.FindByID(input.CategoryID); err != nil {
        return nil, err
    }

    product := &book.ProductBook{
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

func (s *productService) GetAll(categoryID *uint) ([]book.ProductBook, error) {
    return s.repo.FindAll(categoryID)
}

func (s *productService) GetByID(id uint) (*book.ProductBook, error) {
    return s.repo.FindByID(id)
}

func (s *productService) Update(id uint, input book.ProductBookUpdate) (*book.ProductBook, error) {
    product, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // Update hanya field yang dikirim
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