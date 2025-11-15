// internal/service/hotelservice/booking_service.go
package hotelservice

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"backend/internal/models/hotel"
	"backend/internal/repository/repohotel"
	"gorm.io/gorm"
)

type BookingService interface {
	Create(req hotel.CreateBookingRequest) (*hotel.BookingResponse, error)
	Confirm(id uint) error
	Cancel(id uint) error
	List(status string, limit, offset int) ([]hotel.Booking, int64, error)
	CheckAvailability(checkIn, checkOut time.Time, roomType string) ([]hotel.AvailabilityResponse, error)
	GuestBook(userID uint, req hotel.GuestBookingRequest) (*hotel.GuestBookingResponse, error)
}

type bookingService struct {
	bookingRepo repohotel.BookingRepository
	roomRepo    repohotel.RoomRepository
	waNumber    string
	db          *gorm.DB
}

func NewBookingService(bookingRepo repohotel.BookingRepository, roomRepo repohotel.RoomRepository, db *gorm.DB) BookingService {
	waNumber := os.Getenv("HOTEL_WHATSAPP_NUMBER")
	if waNumber == "" {
		waNumber = "6281396554949"
	}
	if !strings.HasPrefix(waNumber, "62") {
		waNumber = "62" + strings.TrimLeft(waNumber, "0")
	}
	return &bookingService{
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		waNumber:    waNumber,
		db:          db,
	}
}

// Create: Booking oleh guest (pilih kamar spesifik)
func (s *bookingService) Create(req hotel.CreateBookingRequest) (*hotel.BookingResponse, error) {
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, errors.New("format check_in tidak valid, gunakan YYYY-MM-DD")
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return nil, errors.New("format check_out tidak valid, gunakan YYYY-MM-DD")
	}
	if !checkOut.After(checkIn) {
		return nil, errors.New("check_out harus setelah check_in")
	}

	room, err := s.roomRepo.FindByID(req.RoomID)
	if err != nil {
		return nil, errors.New("kamar tidak ditemukan")
	}
	if room.Status != hotel.RoomStatusAvailable {
		return nil, errors.New("kamar tidak tersedia")
	}

	count, err := s.bookingRepo.CountOverlapping(req.RoomID, checkIn, checkOut, nil)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("kamar sudah dipesan pada tanggal tersebut")
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		return nil, errors.New("jumlah malam tidak valid")
	}

	totalPrice := int64(nights) * room.RoomType.Price

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.New("gagal memulai transaksi")
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	booking := &hotel.Booking{
		RoomID:      req.RoomID,
		Name:        req.Name,
		Phone:       req.Phone,
		Email:       req.Email,
		CheckIn:     checkIn,
		CheckOut:    checkOut,
		Guests:      req.Guests,
		TotalNights: nights,
		TotalPrice:  totalPrice,
		Status:      hotel.BookingStatusPending,
		Notes:       req.Notes,
	}

	if err := tx.Create(booking).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	waURL := s.generateWhatsAppURL(booking, room, totalPrice, nights)
	return &hotel.BookingResponse{
		ID:          booking.ID,
		WhatsAppURL: waURL,
	}, nil
}

// GuestBook: Booking banyak kamar (tanpa pilih nomor kamar)
func (s *bookingService) GuestBook(userID uint, req hotel.GuestBookingRequest) (*hotel.GuestBookingResponse, error) {
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, errors.New("format check_in tidak valid")
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return nil, errors.New("format check_out tidak valid")
	}
	if !checkOut.After(checkIn) {
		return nil, errors.New("check_out harus setelah check_in")
	}

	avail, err := s.CheckAvailability(checkIn, checkOut, req.RoomType)
	if err != nil {
		return nil, err
	}
	if len(avail) == 0 {
		return nil, errors.New("tipe kamar tidak ditemukan")
	}
	if avail[0].AvailableRooms < req.TotalRooms {
		return nil, fmt.Errorf("hanya %d kamar tersedia", avail[0].AvailableRooms)
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	pricePerNight := avail[0].PricePerNight
	pricePerRoom := int64(nights) * pricePerNight
	totalPrice := pricePerRoom * int64(req.TotalRooms)

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.New("gagal memulai transaksi")
	}
	defer func() { if r := recover(); r != nil { tx.Rollback() } }()

	var rooms []hotel.Room
	if err := tx.Preload("RoomType").
		Where("room_type_id IN (SELECT id FROM room_types WHERE type = ?) AND status = ? AND deleted_at IS NULL", req.RoomType, string(hotel.RoomStatusAvailable)).
		Limit(req.TotalRooms).
		Find(&rooms).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var bookingIDs []uint
	for _, room := range rooms {
		booking := &hotel.Booking{
			RoomID:      room.ID,
			UserID:      &userID,
			Name:        req.Name,
			Phone:       req.Phone,
			Email:       req.Email,
			CheckIn:     checkIn,
			CheckOut:    checkOut,
			Guests:      req.Guests,
			TotalNights: nights,
			TotalPrice:  pricePerRoom, // per kamar
			Status:      hotel.BookingStatusPending,
			Notes:       req.Notes,
		}

		if err := tx.Create(booking).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		bookingIDs = append(bookingIDs, booking.ID)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	waURL := s.generateWhatsAppURLGuest(req, avail[0], nights, totalPrice, len(bookingIDs), checkIn, checkOut)
	return &hotel.GuestBookingResponse{
		BookingIDs:  bookingIDs,
		WhatsAppURL: waURL,
	}, nil
}

// Confirm: Admin konfirmasi → ubah status kamar jadi Booked
func (s *bookingService) Confirm(id uint) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return errors.New("gagal memulai transaksi")
	}

	b, err := s.bookingRepo.FindByID(id)
	if err != nil {
		tx.Rollback()
		return err
	}
	if b.Status != hotel.BookingStatusPending {
		tx.Rollback()
		return errors.New("hanya booking pending yang bisa dikonfirmasi")
	}

	b.Status = hotel.BookingStatusConfirmed
	if err := tx.Save(b).Error; err != nil {
		tx.Rollback()
		return err
	}

	room, err := s.roomRepo.FindByID(b.RoomID)
	if err == nil && room.Status == hotel.RoomStatusAvailable {
		room.Status = hotel.RoomStatusBooked
		tx.Save(room)
	}

	return tx.Commit().Error
}

// Cancel: Batalkan booking
func (s *bookingService) Cancel(id uint) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return errors.New("gagal memulai transaksi")
	}

	b, err := s.bookingRepo.FindByID(id)
	if err != nil {
		tx.Rollback()
		return err
	}

	b.Status = hotel.BookingStatusCancelled
	if err := tx.Save(b).Error; err != nil {
		tx.Rollback()
		return err
	}

	room, err := s.roomRepo.FindByID(b.RoomID)
	if err == nil && room.Status == hotel.RoomStatusBooked {
		room.Status = hotel.RoomStatusAvailable
		tx.Save(room)
	}

	return tx.Commit().Error
}

// List
func (s *bookingService) List(status string, limit, offset int) ([]hotel.Booking, int64, error) {
	f := repohotel.BookingFilter{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	return s.bookingRepo.List(f)
}

// CheckAvailability
func (s *bookingService) CheckAvailability(checkIn, checkOut time.Time, roomType string) ([]hotel.AvailabilityResponse, error) {
	return s.bookingRepo.CheckAvailability(checkIn, checkOut, roomType)
}

// Helper: WhatsApp URL untuk single booking
func (s *bookingService) generateWhatsAppURL(b *hotel.Booking, r *hotel.Room, totalPrice int64, nights int) string {
	msg := fmt.Sprintf(`*PESANAN KAMAR - MUTIARA HOTEL*

Nama: %s
No. HP: %s
Email: %s
Kamar: %s (No. %s)
Tipe: %s
Check-in: %s
Check-out: %s
Malam: %d
Tamu: %d
Total: %s

Catatan:
%s

Silakan transfer ke:
BCA 1234567890 a.n. Hotel Mutiara
Konfirmasi setelah transfer.`,
		b.Name, b.Phone, b.Email,
		strings.Title(r.RoomType.Type), r.Number,
		strings.Title(r.RoomType.Type),
		b.CheckIn.Format("02 Jan 2006"),
		b.CheckOut.Format("02 Jan 2006"),
		nights, b.Guests,
		formatRupiah(totalPrice),
		b.Notes,
	)
	return fmt.Sprintf("https://wa.me/%s?text=%s", s.waNumber, url.QueryEscape(msg))
}

// Helper: WhatsApp untuk guest booking
func (s *bookingService) generateWhatsAppURLGuest(req hotel.GuestBookingRequest, avail hotel.AvailabilityResponse, nights int, totalPrice int64, totalRooms int, checkIn, checkOut time.Time) string {
	msg := fmt.Sprintf(`*PESANAN KAMAR - MUTIARA HOTEL*

Nama: %s
No. HP: %s
Email: %s
Tipe Kamar: %s
Jumlah: %d kamar
Check-in: %s
Check-out: %s
Malam: %d
Tamu: %d
Harga/malam: %s
Total: %s

Catatan:
%s

Silakan transfer ke:
BCA 1234567890 a.n. Hotel Mutiara
Konfirmasi setelah transfer.`,
		req.Name, req.Phone, req.Email,
		strings.Title(req.RoomType),
		totalRooms,
		checkIn.Format("02 Jan 2006"),
		checkOut.Format("02 Jan 2006"),
		nights,
		req.Guests,
		formatRupiah(avail.PricePerNight),
		formatRupiah(totalPrice),
		req.Notes,
	)
	return fmt.Sprintf("https://wa.me/%s?text=%s", s.waNumber, url.QueryEscape(msg))
}

// formatRupiah: Rp 1.500.000
func formatRupiah(n int64) string {
	if n == 0 {
		return "Rp 0"
	}
	s := strconv.FormatInt(n, 10)
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			result.WriteString(".")
		}
		result.WriteByte(s[i])
	}
	return "Rp " + result.String()
}