package service

import (
	"errors"

	"backend/internal/models"
	"backend/internal/repository"
	"gorm.io/gorm"
)

type RoomService interface {
	Create(req models.CreateRoomRequest) (*models.Room, error)
	GetByID(id uint) (*models.Room, error)
	List(t, query string, limit, offset int) ([]models.Room, int64, error)
	Update(id uint, req models.UpdateRoomRequest) (*models.Room, error)
	Delete(id uint) error
}

type roomService struct {
	repo repository.RoomRepository
}

func NewRoomService(repo repository.RoomRepository) RoomService {
	return &roomService{repo: repo}
}

func (s *roomService) Create(req models.CreateRoomRequest) (*models.Room, error) {
	// Cek duplikasi nomor kamar
	if _, err := s.repo.FindByNumber(req.Number); err == nil {
		return nil, errors.New("room number already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	room := &models.Room{
		Number:      req.Number,
		Type:        req.Type,
		Price:       req.Price,
		Capacity:    req.Capacity,
		Description: req.Description,
	}
	if err := s.repo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) GetByID(id uint) (*models.Room, error) {
	return s.repo.FindByID(id)
}

func (s *roomService) List(t, query string, limit, offset int) ([]models.Room, int64, error) {
	f := repository.RoomFilter{
		Type:   t,
		Query:  query,
		Limit:  limit,
		Offset: offset,
	}
	return s.repo.List(f)
}

func (s *roomService) Update(id uint, req models.UpdateRoomRequest) (*models.Room, error) {
	room, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if req.Number != nil && *req.Number != room.Number {
		// cek unik
		if _, err := s.repo.FindByNumber(*req.Number); err == nil {
			return nil, errors.New("room number already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		room.Number = *req.Number
	}
	if req.Type != nil {
		room.Type = *req.Type
	}
	if req.Price != nil {
		room.Price = *req.Price
	}
	if req.Capacity != nil {
		room.Capacity = *req.Capacity
	}
	if req.Description != nil {
		room.Description = *req.Description
	}
	if err := s.repo.Update(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) Delete(id uint) error {
	return s.repo.Delete(id)
}
