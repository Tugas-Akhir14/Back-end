// internal/service/hotelservice/room_type_service.go
package hotelservice

import (
	"errors"
	"time"

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

	if err := validateDiscount(req.DiscountStart, req.DiscountEnd, req.DiscountPercentage); err != nil {
		return nil, err
	}

	rt := &hotel.RoomType{
		Type:                req.Type,
		BasePrice:           req.BasePrice,
		DiscountPercentage:  req.DiscountPercentage,
		DiscountStart:       req.DiscountStart,
		DiscountEnd:         req.DiscountEnd,
		DiscountDescription: req.DiscountDescription,
		Description:         req.Description,
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

	if req.BasePrice != nil {
		rt.BasePrice = *req.BasePrice
	}
	if req.DiscountPercentage != nil {
		rt.DiscountPercentage = *req.DiscountPercentage
	}
	if req.DiscountStart != nil {
		rt.DiscountStart = req.DiscountStart
	}
	if req.DiscountEnd != nil {
		rt.DiscountEnd = req.DiscountEnd
	}
	if req.DiscountDescription != nil {
		rt.DiscountDescription = *req.DiscountDescription
	}
	if req.Description != nil {
		rt.Description = *req.Description
	}

	if err := validateDiscount(rt.DiscountStart, rt.DiscountEnd, rt.DiscountPercentage); err != nil {
		return nil, err
	}

	if err := s.repo.Update(rt); err != nil {
		return nil, err
	}
	return rt, nil
}

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

func validateDiscount(start, end *time.Time, percentage float64) error {
	if percentage > 0 {
		if start == nil || end == nil {
			return errors.New("discount start and end dates are required when percentage is set")
		}
		if start.After(*end) {
			return errors.New("discount start date must be before end date")
		}
	}
	return nil
}