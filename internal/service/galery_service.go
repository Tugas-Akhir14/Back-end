package service

import (
	"backend/internal/models"
	"backend/internal/repository"
)

type GalleryService interface {
	Create(item *models.Gallery) error
	GetByID(id uint) (*models.Gallery, error)
	List(roomID *uint, roomType string, includeGlobal bool, limit, offset int) ([]models.Gallery, int64, error)
	Update(id uint, req models.UpdateGalleryRequest) (*models.Gallery, error)
	Delete(id uint) error
}

type galleryService struct{ repo repository.GalleryRepository }

func NewGalleryService(repo repository.GalleryRepository) GalleryService { return &galleryService{repo} }

func (s *galleryService) Create(item *models.Gallery) error { return s.repo.Create(item) }

func (s *galleryService) GetByID(id uint) (*models.Gallery, error) { return s.repo.FindByID(id) }

func (s *galleryService) List(roomID *uint, roomType string, includeGlobal bool, limit, offset int) ([]models.Gallery, int64, error) {
	return s.repo.List(repository.GalleryFilter{
		RoomID: roomID, RoomType: roomType, IncludeGlobal: includeGlobal, Limit: limit, Offset: offset,
	})
}

func (s *galleryService) Update(id uint, req models.UpdateGalleryRequest) (*models.Gallery, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Caption != nil {
		item.Caption = *req.Caption
	}
	if req.RoomID != nil {
		item.RoomID = req.RoomID
	}
	if err := s.repo.Update(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *galleryService) Delete(id uint) error { return s.repo.Delete(id) }
