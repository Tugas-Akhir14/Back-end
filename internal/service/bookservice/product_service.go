package bookservice

import (
	"backend/internal/models/book"
	"backend/internal/repository/repobook"
	"fmt"
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
		return nil, fmt.Errorf("produk tidak ditemukan")
	}

	updated := false

	if input.Nama != nil && *input.Nama != product.Nama {
		product.Nama = *input.Nama
		updated = true
	}
	if input.Deskripsi != nil && *input.Deskripsi != product.Deskripsi {
		product.Deskripsi = *input.Deskripsi
		updated = true
	}
	if input.Harga != nil && *input.Harga != product.Harga {
		product.Harga = *input.Harga
		updated = true
	}
	if input.Stok != nil && *input.Stok != product.Stok {
		product.Stok = *input.Stok
		updated = true
	}
	if input.Gambar != nil && *input.Gambar != product.Gambar {
		// Optional: hapus file lama jika perlu
		// os.Remove("." + product.Gambar) // uncomment jika ingin hapus file lama
		product.Gambar = *input.Gambar
		updated = true
	}
	if input.CategoryID != nil && *input.CategoryID != product.CategoryID {
		// Validasi kategori
		if _, err := s.categoryRepo.FindByID(*input.CategoryID); err != nil {
			return nil, fmt.Errorf("kategori tidak ditemukan")
		}
		product.CategoryID = *input.CategoryID
		updated = true
	}

	// Hanya lakukan save jika ada perubahan
	if updated {
		if err := s.repo.Update(product); err != nil {
			return nil, err
		}
	}

	// Reload untuk mendapatkan data terbaru + preload
	return s.repo.FindByID(id)
}

func (s *productService) Delete(id uint) error {
    return s.repo.Delete(id)
}