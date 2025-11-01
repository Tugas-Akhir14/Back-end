package repohotel

import (
	"backend/internal/models/hotel"
	"context"
	"errors"

	"gorm.io/gorm"
)

type RoomFilter struct {
	Type   string
	Query  string
	Limit  int
	Offset int
}

type RoomRepository interface {
	Create(room *hotel.Room) error
	FindByID(id uint) (*hotel.Room, error)
	FindByNumber(number string) (*hotel.Room, error)
	List(f RoomFilter) ([]hotel.Room, int64, error)
	Update(room *hotel.Room) error
	Delete(id uint) error
	GetAll(ctx context.Context) ([]hotel.Room, error)
	ListPublic(limit int) ([]hotel.Room, error)
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) Create(room *hotel.Room) error {
	return r.db.Create(room).Error
}

func (r *roomRepository) FindByID(id uint) (*hotel.Room, error) {
	var room hotel.Room
	if err := r.db.First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) FindByNumber(number string) (*hotel.Room, error) {
	var room hotel.Room
	err := r.db.Where("number = ?", number).First(&room).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) List(f RoomFilter) ([]hotel.Room, int64, error) {
	var (
		rooms []hotel.Room
		count int64
		q     = r.db.Model(&hotel.Room{})
	)
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.Query != "" {
		like := "%" + f.Query + "%"
		q = q.Where("number LIKE ? OR description LIKE ?", like, like)
	}
	q.Count(&count)

	if f.Limit <= 0 {
		f.Limit = 10
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	if err := q.Order("id DESC").Limit(f.Limit).Offset(f.Offset).Find(&rooms).Error; err != nil {
		return nil, 0, err
	}
	return rooms, count, nil
}

func (r *roomRepository) Update(room *hotel.Room) error {
	return r.db.Save(room).Error
}

func (r *roomRepository) Delete(id uint) error {
	return r.db.Delete(&hotel.Room{}, id).Error
}

func (r *roomRepository) GetAll(ctx context.Context) ([]hotel.Room, error) {
	var rooms []hotel.Room
	if err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("price ASC").
		Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}


func (r *roomRepository) ListPublic(limit int) ([]hotel.Room, error) {
	var rooms []hotel.Room

	q := r.db.Model(&hotel.Room{}).
		// kalau kamu punya kolom "active" tinggal tambahkan: Where("active = ?", true).
		Where("deleted_at IS NULL").
		Order("price ASC")

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}