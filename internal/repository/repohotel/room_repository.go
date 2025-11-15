// internal/repository/repohotel/room_repository.go
package repohotel

import (
	"backend/internal/models/hotel"
	"gorm.io/gorm"
)

type RoomRepository interface {
	Create(r *hotel.Room) error
	FindByID(id uint) (*hotel.Room, error)
	FindByNumber(num string) (*hotel.Room, error)
	List(f RoomFilter) ([]hotel.Room, int64, error)
	ListPublic(limit int) ([]hotel.Room, error)
	Update(r *hotel.Room) error
	Delete(id uint) error
}

type RoomFilter struct {
	Type   string
	Query  string
	Status string
	Limit  int
	Offset int
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
	if err := r.db.Preload("RoomType").First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) FindByNumber(num string) (*hotel.Room, error) {
	var room hotel.Room
	err := r.db.Preload("RoomType").Where("number = ?", num).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) List(f RoomFilter) ([]hotel.Room, int64, error) {
	var rooms []hotel.Room
	var total int64

	query := r.db.Model(&hotel.Room{}).Preload("RoomType")
	if f.Type != "" {
		query = query.Joins("JOIN room_types ON room_types.id = rooms.room_type_id").
			Where("room_types.type = ?", f.Type)
	}
	if f.Query != "" {
		query = query.Where("rooms.number LIKE ?", "%"+f.Query+"%")
	}
	if f.Status != "" {
		query = query.Where("rooms.status = ?", f.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if f.Limit > 0 {
		query = query.Limit(f.Limit)
	}
	if f.Offset > 0 {
		query = query.Offset(f.Offset)
	}

	if err := query.Find(&rooms).Error; err != nil {
		return nil, 0, err
	}
	return rooms, total, nil
}

func (r *roomRepository) ListPublic(limit int) ([]hotel.Room, error) {
	var rooms []hotel.Room
	query := r.db.Preload("RoomType").Where("status = ?", hotel.RoomStatusAvailable)
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) Update(room *hotel.Room) error {
	return r.db.Save(room).Error
}

func (r *roomRepository) Delete(id uint) error {
	return r.db.Delete(&hotel.Room{}, id).Error
}