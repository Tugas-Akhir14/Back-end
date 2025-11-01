package hotelservice

import (
	"backend/internal/models/hotel"
	"backend/internal/repository/repohotel"
)

type GalleryService interface {
	Create(item *hotel.Gallery) error
	GetByID(id uint) (*hotel.Gallery, error)
	List(roomID *uint, roomType string, includeGlobal bool, limit, offset int) ([]hotel.Gallery, int64, error)
	Update(id uint, req hotel.UpdateGalleryRequest) (*hotel.Gallery, error)
	Save(item *hotel.Gallery) error // simpan semua field (untuk update image)
	Delete(id uint) error
}

type galleryService struct{ repo repohotel.GalleryRepository }

func NewGalleryService(repo repohotel.GalleryRepository) GalleryService { return &galleryService{repo} }

func (s *galleryService) Create(item *hotel.Gallery) error                 { return s.repo.Create(item) }
func (s *galleryService) GetByID(id uint) (*hotel.Gallery, error)          { return s.repo.FindByID(id) }
func (s *galleryService) Delete(id uint) error                             { return s.repo.Delete(id) }
func (s *galleryService) Save(item *hotel.Gallery) error                   { return s.repo.Update(item) }
func (s *galleryService) List(roomID *uint, roomType string, includeGlobal bool, limit, offset int) ([]hotel.Gallery, int64, error) {
	return s.repo.List(repohotel.GalleryFilter{
		RoomID: roomID, RoomType: roomType, IncludeGlobal: includeGlobal, Limit: limit, Offset: offset,
	})
}

func (s *galleryService) Update(id uint, req hotel.UpdateGalleryRequest) (*hotel.Gallery, error) {
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
