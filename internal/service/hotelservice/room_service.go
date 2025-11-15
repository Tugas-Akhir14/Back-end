// internal/service/hotelservice/room_service.go
package hotelservice

import (
	"context"
	"errors"
	"strings"

	"backend/internal/models/hotel"
	"backend/internal/repository/repohotel"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type RoomService interface {
	Create(c *gin.Context, req hotel.CreateRoomRequest) (*hotel.Room, error)
	GetByID(id uint) (*hotel.Room, error)
	List(t, query, status string, limit, offset int) ([]hotel.Room, int64, error)
	Update(c *gin.Context, id uint, req hotel.UpdateRoomRequest) (*hotel.Room, error)
	Delete(id uint) error
	ListPublic(ctx context.Context) ([]hotel.Room, error)
}

type roomService struct {
	repo         repohotel.RoomRepository
	roomTypeRepo repohotel.RoomTypeRepository
	db           *gorm.DB
}

func NewRoomService(repo repohotel.RoomRepository, roomTypeRepo repohotel.RoomTypeRepository, db *gorm.DB) RoomService {
	return &roomService{
		repo:         repo,
		roomTypeRepo: roomTypeRepo,
		db:           db,
	}
}

// Validasi RoomTypeID
func (s *roomService) validateRoomTypeID(id uint) error {
	_, err := s.roomTypeRepo.FindByID(id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("invalid room_type_id")
	}
	return nil
}

// Parse status
func parseRoomStatus(str string) (hotel.RoomStatus, error) {
	if str == "" {
		return hotel.RoomStatusAvailable, nil
	}
	str = strings.ToLower(str)
	switch str {
	case "available":
		return hotel.RoomStatusAvailable, nil
	case "booked":
		return hotel.RoomStatusBooked, nil
	case "cleaning":
		return hotel.RoomStatusCleaning, nil
	default:
		return "", errors.New("status tidak valid: available, booked, atau cleaning")
	}
}

// CREATE ROOM
func (s *roomService) Create(c *gin.Context, req hotel.CreateRoomRequest) (*hotel.Room, error) {
	// Cek duplikat nomor kamar
	existing, err := s.repo.FindByNumber(req.Number)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("room number already exists")
	}

	// Validasi room_type_id
	if err := s.validateRoomTypeID(req.RoomTypeID); err != nil {
		return nil, err
	}

	// Parse status
	status, err := parseRoomStatus(req.Status)
	if err != nil {
		return nil, err
	}

	// Ambil image URL dari context (dari handler)
	imageURL := ""
	if url, exists := c.Get("image_url"); exists {
		imageURL = url.(string)
	}

	// Buat room
	room := &hotel.Room{
		Number:      req.Number,
		RoomTypeID:  req.RoomTypeID,
		Capacity:    req.Capacity,
		Description: req.Description,
		Image:       imageURL,
		Status:      status,
	}

	if err := s.repo.Create(room); err != nil {
		return nil, err
	}

	// Reload dengan RoomType
	return s.repo.FindByID(room.ID)
}

// GET BY ID
func (s *roomService) GetByID(id uint) (*hotel.Room, error) {
	return s.repo.FindByID(id)
}

// LIST
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

// UPDATE ROOM
func (s *roomService) Update(c *gin.Context, id uint, req hotel.UpdateRoomRequest) (*hotel.Room, error) {
	room, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Update nomor jika diubah
	if req.Number != nil && *req.Number != room.Number {
		existing, err := s.repo.FindByNumber(*req.Number)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != room.ID {
			return nil, errors.New("room number already exists")
		}
		room.Number = *req.Number
	}

	// Update room_type_id
	if req.RoomTypeID != nil {
		if err := s.validateRoomTypeID(*req.RoomTypeID); err != nil {
			return nil, err
		}
		room.RoomTypeID = *req.RoomTypeID
	}

	// Update kapasitas
	if req.Capacity != nil {
		room.Capacity = *req.Capacity
	}

	// Update deskripsi
	if req.Description != nil {
		room.Description = *req.Description
	}

	// Update gambar (dari context)
	if url, exists := c.Get("image_url"); exists {
		room.Image = url.(string)
	}

	// Update status
	if req.Status != nil {
		status, err := parseRoomStatus(*req.Status)
		if err != nil {
			return nil, err
		}
		room.Status = status
	}

	if err := s.repo.Update(room); err != nil {
		return nil, err
	}

	return s.repo.FindByID(room.ID)
}

// DELETE
func (s *roomService) Delete(id uint) error {
	return s.repo.Delete(id)
}

// LIST PUBLIC
func (s *roomService) ListPublic(ctx context.Context) ([]hotel.Room, error) {
	return s.repo.ListPublic(0)
}