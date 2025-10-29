// backend/internal/service/bookservice/category_service.go
package bookservice

import (
	"backend/internal/models/book"
	"backend/internal/repository/repobook"
)

type CategoryService interface {
	Create(nama string) (*book.CategoryBook, error)
	GetAll() ([]book.CategoryBook, error)
	GetByID(id uint) (*book.CategoryBook, error)
	Update(id uint, nama string) (*book.CategoryBook, error)
	Delete(id uint) error
}

type categoryService struct {
	repo repobook.CategoryRepository
}

func NewCategoryService(repo repobook.CategoryRepository) CategoryService {
	return &categoryService{repo}
}

// Create
func (s *categoryService) Create(nama string) (*book.CategoryBook, error) {
	category := &book.CategoryBook{Nama: nama}
	if err := s.repo.Create(category); err != nil {
		return nil, err
	}
	return category, nil
}

// GetAll
func (s *categoryService) GetAll() ([]book.CategoryBook, error) {
	return s.repo.FindAll()
}

// GetByID
func (s *categoryService) GetByID(id uint) (*book.CategoryBook, error) {
	return s.repo.FindByID(id)
}

// Update
func (s *categoryService) Update(id uint, nama string) (*book.CategoryBook, error) {
	category, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	category.Nama = nama
	if err := s.repo.Update(category); err != nil {
		return nil, err
	}
	return category, nil
}

// Delete
func (s *categoryService) Delete(id uint) error {
	return s.repo.Delete(id)
}