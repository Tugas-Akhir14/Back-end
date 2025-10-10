package service

import (
	"backend/internal/models"
	"backend/internal/repository"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type NewsService interface {
	Create(input models.News, img *multipart.FileHeader, creatorID uint) (*models.News, error)
	Update(id uint, input models.News, img *multipart.FileHeader, updaterID uint) (*models.News, error)
	Delete(id uint) error
	GetByID(id uint) (*models.News, error)
	GetBySlug(slug string) (*models.News, error)
	List(page, pageSize int, q, status string) ([]models.News, int64, error)
}

type newsService struct {
	repo repository.NewsRepository
}

func NewNewsService(r repository.NewsRepository) NewsService {
	return &newsService{repo: r}
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, "\\", "-")
	return s
}

func saveImage(img *multipart.FileHeader) (string, error) {
	if img == nil {
		return "", nil
	}
	_ = os.MkdirAll("uploads/news", 0755)

	filename := fmt.Sprintf("%d-%s", time.Now().Unix(), img.Filename)
	dst := filepath.Join("uploads/news", filename)

	srcFile, err := img.Open()
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := out.ReadFrom(srcFile); err != nil {
		return "", err
	}
	return dst, nil
}

func (s *newsService) Create(input models.News, img *multipart.FileHeader, creatorID uint) (*models.News, error) {
	input.Slug = slugify(input.Title)
	input.CreatedBy = &creatorID
	if strings.ToLower(input.Status) == "published" && input.PublishedAt == nil {
		now := time.Now()
		input.PublishedAt = &now
	}
	if img != nil {
		path, err := saveImage(img)
		if err != nil {
			return nil, err
		}
		input.ImageURL = path
	}
	if err := s.repo.Create(&input); err != nil {
		return nil, err
	}
	return &input, nil
}

func (s *newsService) Update(id uint, input models.News, img *multipart.FileHeader, updaterID uint) (*models.News, error) {
	existing, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if input.Title != "" {
		existing.Title = input.Title
		existing.Slug = slugify(input.Title)
	}
	if input.Content != "" {
		existing.Content = input.Content
	}
	if input.Status != "" {
		existing.Status = input.Status
		if strings.ToLower(input.Status) == "published" && existing.PublishedAt == nil {
			now := time.Now()
			existing.PublishedAt = &now
		}
	}
	if img != nil {
		newPath, err := saveImage(img)
		if err != nil {
			return nil, err
		}
		// hapus file lama jika ada
		if existing.ImageURL != "" {
			_ = os.Remove(existing.ImageURL)
		}
		existing.ImageURL = newPath
	}
	existing.UpdatedBy = &updaterID

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (s *newsService) Delete(id uint) error { return s.repo.Delete(id) }
func (s *newsService) GetByID(id uint) (*models.News, error) { return s.repo.FindByID(id) }
func (s *newsService) GetBySlug(slug string) (*models.News, error) { return s.repo.FindBySlug(slug) }
func (s *newsService) List(page, pageSize int, q, status string) ([]models.News, int64, error) {
	return s.repo.FindAll(page, pageSize, q, status)
}
