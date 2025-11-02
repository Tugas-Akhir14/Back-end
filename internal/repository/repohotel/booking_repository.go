// backend/internal/repository/repohotel/booking_repository.go
package repohotel

import (
	"backend/internal/models/hotel"
	"time"

	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(booking *hotel.Booking) error
	FindByID(id uint) (*hotel.Booking, error)
	List(filter BookingFilter) ([]hotel.Booking, int64, error)
	Update(booking *hotel.Booking) error
	FindOverlapping(roomID uint, checkIn, checkOut time.Time) ([]hotel.Booking, error)
}

type BookingFilter struct {
	Status string
	Limit  int
	Offset int
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(booking *hotel.Booking) error {
	return r.db.Create(booking).Error
}

func (r *bookingRepository) FindByID(id uint) (*hotel.Booking, error) {
	var b hotel.Booking
	if err := r.db.Preload("Room").First(&b, id).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepository) List(f BookingFilter) ([]hotel.Booking, int64, error) {
	var bookings []hotel.Booking
	var count int64

	q := r.db.Model(&hotel.Booking{}).Preload("Room")
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	q.Count(&count)

	if f.Limit <= 0 {
		f.Limit = 10
	}
	if err := q.Order("id DESC").Limit(f.Limit).Offset(f.Offset).Find(&bookings).Error; err != nil {
		return nil, 0, err
	}
	return bookings, count, nil
}

func (r *bookingRepository) Update(booking *hotel.Booking) error {
	return r.db.Save(booking).Error
}

// Cegah double booking
func (r *bookingRepository) FindOverlapping(roomID uint, checkIn, checkOut time.Time) ([]hotel.Booking, error) {
	var bookings []hotel.Booking
	err := r.db.Where("room_id = ? AND status IN ? AND check_out > ? AND check_in < ?",
		roomID,
		[]string{hotel.BookingStatusConfirmed, hotel.BookingStatusCheckedIn},
		checkIn, checkOut,
	).Find(&bookings).Error
	return bookings, err
}