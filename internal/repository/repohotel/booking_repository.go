// internal/repository/repohotel/booking_repository.go
package repohotel

import (
	"time"

	"backend/internal/models/hotel"
	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(booking *hotel.Booking) error
	FindByID(id uint) (*hotel.Booking, error)
	List(filter BookingFilter) ([]hotel.Booking, int64, error)
	Update(booking *hotel.Booking) error
	CountOverlapping(roomID uint, checkIn, checkOut time.Time, excludeID *uint) (int64, error)
	CheckAvailability(checkIn, checkOut time.Time, roomTypeFilter string) ([]hotel.AvailabilityResponse, error)
	FindBookingsByDateRange(checkIn, checkOut time.Time) ([]hotel.Booking, error)
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
	err := r.db.
		Preload("Room").
		Preload("Room.RoomType").
		First(&b, id).Error
	return &b, err
}

func (r *bookingRepository) List(f BookingFilter) ([]hotel.Booking, int64, error) {
	var bookings []hotel.Booking
	var count int64

	query := r.db.Model(&hotel.Booking{}).
		Preload("Room").
		Preload("Room.RoomType")

	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	if err := query.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&bookings).Error; err != nil {
		return nil, 0, err
	}

	return bookings, count, nil
}

func (r *bookingRepository) Update(booking *hotel.Booking) error {
	return r.db.Save(booking).Error
}

func (r *bookingRepository) CountOverlapping(roomID uint, checkIn, checkOut time.Time, excludeID *uint) (int64, error) {
	var count int64

	query := r.db.Model(&hotel.Booking{}).
		Where("room_id = ?", roomID).
		Where("status IN ?", []string{
			string(hotel.BookingStatusConfirmed),
			string(hotel.BookingStatusCheckedIn),
		}).
		Where("check_in < ? AND check_out > ?", checkOut, checkIn)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.Count(&count).Error
	return count, err
}

// internal/repository/repohotel/booking_repository.go

func (r *bookingRepository) CheckAvailability(checkIn, checkOut time.Time, roomTypeFilter string) ([]hotel.AvailabilityResponse, error) {
    var results []hotel.AvailabilityResponse

    query := r.db.Table("room_types rt").
        Joins("JOIN rooms ON rooms.room_type_id = rt.id AND rooms.deleted_at IS NULL AND rooms.status = ?", string(hotel.RoomStatusAvailable)).
        Joins(`LEFT JOIN bookings ON bookings.room_id = rooms.id 
               AND bookings.check_in < ? 
               AND bookings.check_out > ? 
               AND bookings.status IN ?`,
            checkOut, checkIn,
            []string{string(hotel.BookingStatusConfirmed), string(hotel.BookingStatusCheckedIn)},
        ).
        Where("rooms.deleted_at IS NULL")

    if roomTypeFilter != "" {
        query = query.Where("rt.type = ?", roomTypeFilter)
    }

    query = query.
        Select(`
            rt.type,
            rt.price AS price_per_night,  -- INI YANG DIPERBAIKI
            COUNT(DISTINCT rooms.id) AS total_rooms,
            COUNT(DISTINCT bookings.id) AS booked_rooms
        `).
        Group("rt.id, rt.type, rt.price")

    rows, err := query.Rows()
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var res hotel.AvailabilityResponse
        var total, booked int64
        if err := rows.Scan(&res.Type, &res.PricePerNight, &total, &booked); err != nil {
            return nil, err
        }
        res.TotalRooms = int(total)
        res.AvailableRooms = int(total - booked)
        if res.AvailableRooms < 0 {
            res.AvailableRooms = 0
        }
        results = append(results, res)
    }

    return results, nil
}

func (r *bookingRepository) FindBookingsByDateRange(checkIn, checkOut time.Time) ([]hotel.Booking, error) {
	var bookings []hotel.Booking
	err := r.db.
		Where("check_in <= ? AND check_out >= ?", checkOut, checkIn).
		Where("status IN ?", []string{
			string(hotel.BookingStatusPending),
			string(hotel.BookingStatusConfirmed),
			string(hotel.BookingStatusCheckedIn),
		}).
		Preload("Room").
		Preload("Room.RoomType").
		Find(&bookings).Error
	return bookings, err
}