// backend/internal/service/cafeservice/category_service.go
package cafeservice

import (
    "backend/internal/models/cafe"
    "backend/internal/repository/repocafe"
)

type CategoryService interface {
    Create(nama string) (*cafe.CategoryCafe, error)
    GetAll() ([]cafe.CategoryCafe, error)
    GetByID(id uint) (*cafe.CategoryCafe, error)
    Update(id uint, nama string) (*cafe.CategoryCafe, error)
    Delete(id uint) error
}

type categoryService struct {
    repo repocafe.CategoryRepository
}

func NewCategoryService(repo repocafe.CategoryRepository) CategoryService {
    return &categoryService{repo}
}

func (s *categoryService) Create(nama string) (*cafe.CategoryCafe, error) {
    category := &cafe.CategoryCafe{Nama: nama}
    if err := s.repo.Create(category); err != nil {
        return nil, err
    }
    return category, nil
}

func (s *categoryService) GetAll() ([]cafe.CategoryCafe, error) {
    return s.repo.FindAll()
}

func (s *categoryService) GetByID(id uint) (*cafe.CategoryCafe, error) {
    return s.repo.FindByID(id)
}

func (s *categoryService) Update(id uint, nama string) (*cafe.CategoryCafe, error) {
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

func (s *categoryService) Delete(id uint) error {
    return s.repo.Delete(id)
}