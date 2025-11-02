package hotelservice

import (
	"context"
	"errors"

	"backend/internal/models/hotel"
	"backend/internal/repository/repohotel"
)

type RoomService interface {
	Create(req hotel.CreateRoomRequest) (*hotel.Room, error)
	GetByID(id uint) (*hotel.Room, error)
	List(t, query, status string, limit, offset int) ([]hotel.Room, int64, error)
	Update(id uint, req hotel.UpdateRoomRequest) (*hotel.Room, error)
	Delete(id uint) error
	ListPublic(ctx context.Context) ([]hotel.Room, error)
}

type roomService struct {
	repo repohotel.RoomRepository
}

func NewRoomService(repo repohotel.RoomRepository) RoomService {
	return &roomService{repo: repo}
}

func (s *roomService) Create(req hotel.CreateRoomRequest) (*hotel.Room, error) {
	existing, err := s.repo.FindByNumber(req.Number)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("room number already exists")
	}

	status := hotel.RoomStatusAvailable
	if req.Status != "" {
		status = req.Status
	}

	room := &hotel.Room{
		Number:      req.Number,
		Type:        req.Type,
		Price:       req.Price,
		Capacity:    req.Capacity,
		Description: req.Description,
		Image:       req.Image,
		Status:      status,
	}
	if err := s.repo.Create(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) GetByID(id uint) (*hotel.Room, error) {
	return s.repo.FindByID(id)
}

func (s *roomService) List(t, query, status string, limit, offset int) ([]hotel.Room, int64, error) {
	f := repohotel.RoomFilter{
		Type:   t,
		Query:  query,
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	return s.repo.List(f)
}

func (s *roomService) Update(id uint, req hotel.UpdateRoomRequest) (*hotel.Room, error) {
	room, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Number != nil && *req.Number != room.Number {
		existing, err := s.repo.FindByNumber(*req.Number)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != room.ID {
			return nil, errors.New("room number already exists")
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
	if req.Image != nil {
		room.Image = *req.Image
	}
	if req.Status != nil {
		room.Status = *req.Status
	}

	if err := s.repo.Update(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *roomService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *roomService) ListPublic(ctx context.Context) ([]hotel.Room, error) {
	return s.repo.ListPublic(0) // 0 = tanpa limit
}