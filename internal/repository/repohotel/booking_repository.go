// internal/repository/repohotel/booking_repository.go
package repohotel

import (
	"database/sql"
	"fmt"
	"strings"
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
	GetByUserID(userID uint, limit, offset int) ([]hotel.Booking, int64, error)
}

type BookingFilter struct {
	Status string
	Source string // Tambah source filter
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
	if f.Source != "" {
		query = query.Where("source = ?", f.Source)
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
			string(hotel.BookingStatusPending),
			string(hotel.BookingStatusConfirmed),
			string(hotel.BookingStatusCheckedIn),
		}).
		Where(`(
			(check_in < ?  AND check_out > ?)  OR  -- booking lama nge-cover baru
			(check_in < ?  AND check_out > ?)  OR  -- baru nge-cover booking lama
			(check_in >= ? AND check_out <= ?) OR  -- baru di dalam booking lama
			(check_in <= ? AND check_out >= ?)     -- booking lama di dalam baru ← FIXED!
		)`, 
			checkOut, checkIn,   
			checkIn,  checkOut,  
			checkIn,  checkOut,  
			checkIn,  checkOut,  
		)

	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.Count(&count).Error
	return count, err
}


func (r *bookingRepository) CheckAvailability(checkIn, checkOut time.Time, roomType string) ([]hotel.AvailabilityResponse, error) {
    var results []hotel.AvailabilityResponse

    // PAKAI KOLOM "price" SAJA → INI YANG PASTI ADA DI DATABASE KAMU
    query := `
        SELECT 
            rt.type AS room_type,
            rt.price AS price_per_night,
            COUNT(r.id) AS total_rooms,
            COUNT(r.id) - COUNT(b.id) AS available_rooms
        FROM room_types rt
        LEFT JOIN rooms r ON r.room_type_id = rt.id AND r.deleted_at IS NULL
        LEFT JOIN bookings b ON b.room_id = r.id 
            AND b.status IN ('pending', 'confirmed', 'checked_in')
            AND b.deleted_at IS NULL
            AND (
                (b.check_in < ? AND b.check_out > ?) OR
                (b.check_in < ? AND b.check_out > ?) OR
                (b.check_in >= ? AND b.check_out <= ?)
            )
        WHERE rt.deleted_at IS NULL
    `

    params := []interface{}{
        checkOut, checkIn,
        checkIn, checkOut,
        checkIn, checkOut,
    }

    if roomType != "" {
        query += " AND LOWER(rt.type) = LOWER(?)"
        params = append(params, roomType)
    }

    query += " GROUP BY rt.id, rt.type, rt.price"

    rows, err := r.db.Raw(query, params...).Rows()
    if err != nil {
        return nil, fmt.Errorf("query error: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var res hotel.AvailabilityResponse
        var roomTypeStr string
        var price int64
        var totalRooms, availableRooms sql.NullInt64

        if err := rows.Scan(&roomTypeStr, &price, &totalRooms, &availableRooms); err != nil {
            return nil, fmt.Errorf("scan error: %w", err)
        }

        res.RoomType = strings.Title(strings.ToLower(roomTypeStr))
        res.PricePerNight = price
        res.TotalRooms = int(totalRooms.Int64)
        res.AvailableRooms = int(availableRooms.Int64)

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

func (r *bookingRepository) GetByUserID(userID uint, limit, offset int) ([]hotel.Booking, int64, error) {
	var bookings []hotel.Booking
	var total int64

	query := r.db.Model(&hotel.Booking{}).
		Where("user_id = ?", userID).
		Preload("Room").
		Preload("Room.RoomType")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&bookings).Error

	return bookings, total, err
}