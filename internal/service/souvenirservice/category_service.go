// backend/internal/service/souvenirservice/category_service.go
package souvenirservice

import (
    "backend/internal/models/souvenir"
    "backend/internal/repository/reposouvenir"
)

type CategoryService interface {
    Create(req *souvenir.CategoryCreate) (*souvenir.Category, error)
    GetByID(id uint) (*souvenir.Category, error)
    GetAll() ([]souvenir.Category, error)
    Update(id uint, req *souvenir.CategoryUpdate) (*souvenir.Category, error)
    Delete(id uint) error
}

type categoryService struct {
    repo reposouvenir.CategoryRepository
}

func NewCategoryService(repo reposouvenir.CategoryRepository) CategoryService {
    return &categoryService{repo}
}

func (s *categoryService) Create(req *souvenir.CategoryCreate) (*souvenir.Category, error) {
    cat := &souvenir.Category{
        Nama:      req.Nama,
        Slug:      req.Slug,
        Deskripsi: req.Deskripsi,
    }
    return cat, s.repo.Create(cat)
}

func (s *categoryService) GetByID(id uint) (*souvenir.Category, error) {
    return s.repo.GetByID(id)
}

func (s *categoryService) GetAll() ([]souvenir.Category, error) {
    return s.repo.GetAll()
}

func (s *categoryService) Update(id uint, req *souvenir.CategoryUpdate) (*souvenir.Category, error) {
    cat, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    if req.Nama != nil {
        cat.Nama = *req.Nama
    }
    if req.Slug != nil {
        cat.Slug = *req.Slug
    }
    if req.Deskripsi != nil {
        cat.Deskripsi = *req.Deskripsi
    }
    return cat, s.repo.Update(cat)
}

func (s *categoryService) Delete(id uint) error {
    return s.repo.Delete(id)
}