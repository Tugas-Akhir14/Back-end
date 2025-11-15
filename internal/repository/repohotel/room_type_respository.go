// internal/repository/repohotel/room_type_repository.go
package repohotel

import (
	"backend/internal/models/hotel"
	"gorm.io/gorm"
)

type RoomTypeRepository interface {
	Create(rt *hotel.RoomType) error
	FindByID(id uint) (*hotel.RoomType, error)
	FindByType(t string) (*hotel.RoomType, error)
	List() ([]hotel.RoomType, error)
	Update(rt *hotel.RoomType) error
	Delete(id uint) error
}

type RoomTypeRepositoryImpl struct {
	DB *gorm.DB
}

func NewRoomTypeRepository(db *gorm.DB) RoomTypeRepository {
	return &RoomTypeRepositoryImpl{DB: db}
}

func (r *RoomTypeRepositoryImpl) Create(rt *hotel.RoomType) error {
	return r.DB.Create(rt).Error
}

func (r *RoomTypeRepositoryImpl) FindByID(id uint) (*hotel.RoomType, error) {
	var rt hotel.RoomType
	if err := r.DB.First(&rt, id).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *RoomTypeRepositoryImpl) FindByType(t string) (*hotel.RoomType, error) {
	var rt hotel.RoomType
	err := r.DB.Where("type = ?", t).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *RoomTypeRepositoryImpl) List() ([]hotel.RoomType, error) {
	var types []hotel.RoomType
	if err := r.DB.Find(&types).Error; err != nil {
		return nil, err
	}
	return types, nil
}

func (r *RoomTypeRepositoryImpl) Update(rt *hotel.RoomType) error {
	return r.DB.Save(rt).Error
}

func (r *RoomTypeRepositoryImpl) Delete(id uint) error {
	return r.DB.Delete(&hotel.RoomType{}, id).Error
}