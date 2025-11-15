// internal/service/hotelservice/room_type_service.go
package hotelservice

import (
	"errors"

	"backend/internal/models/hotel"
	"backend/internal/repository/repohotel"
)

type RoomTypeService interface {
	Create(req hotel.CreateRoomTypeRequest) (*hotel.RoomType, error)
	GetByID(id uint) (*hotel.RoomType, error)
	List() ([]hotel.RoomType, error)
	Update(id uint, req hotel.UpdateRoomTypeRequest) (*hotel.RoomType, error)
	Delete(id uint) error
}

type roomTypeService struct {
	repo repohotel.RoomTypeRepository
}

func NewRoomTypeService(repo repohotel.RoomTypeRepository) RoomTypeService {
	return &roomTypeService{repo: repo}
}

func (s *roomTypeService) Create(req hotel.CreateRoomTypeRequest) (*hotel.RoomType, error) {
	if existing, _ := s.repo.FindByType(req.Type); existing != nil {
		return nil, errors.New("room type already exists")
	}

	rt := &hotel.RoomType{
		Type:        req.Type,
		Price:       req.Price,
		Description: req.Description,
	}

	if err := s.repo.Create(rt); err != nil {
		return nil, err
	}
	return s.repo.FindByID(rt.ID)
}

func (s *roomTypeService) GetByID(id uint) (*hotel.RoomType, error) {
	return s.repo.FindByID(id)
}

func (s *roomTypeService) List() ([]hotel.RoomType, error) {
	return s.repo.List()
}

func (s *roomTypeService) Update(id uint, req hotel.UpdateRoomTypeRequest) (*hotel.RoomType, error) {
	rt, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Price != nil {
		rt.Price = *req.Price
	}
	if req.Description != nil {
		rt.Description = *req.Description
	}

	if err := s.repo.Update(rt); err != nil {
		return nil, err
	}
	return rt, nil
}

// internal/service/hotelservice/room_type_service.go
func (s *roomTypeService) Delete(id uint) error {
	var count int64
	impl := s.repo.(*repohotel.RoomTypeRepositoryImpl)
	if err := impl.DB.Model(&hotel.Room{}).
		Where("room_type_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete: room type is in use")
	}

	return s.repo.Delete(id)
}