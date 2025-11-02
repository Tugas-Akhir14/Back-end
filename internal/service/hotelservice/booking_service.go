// backend/internal/service/hotelservice/booking_service.go
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
)

type BookingService interface {
	Create(req hotel.CreateBookingRequest) (*hotel.BookingResponse, error)
	Confirm(id uint) error
	Cancel(id uint) error
	List(status string, limit, offset int) ([]hotel.Booking, int64, error)
}

type bookingService struct {
	bookingRepo repohotel.BookingRepository
	roomRepo    repohotel.RoomRepository
	waNumber    string
}

func NewBookingService(bookingRepo repohotel.BookingRepository, roomRepo repohotel.RoomRepository) BookingService {
    waNumber := os.Getenv("HOTEL_WHATSAPP_NUMBER")
    if waNumber == "" {
        waNumber = "6281396554949" // fallback
    }
    return &bookingService{
        bookingRepo: bookingRepo,
        roomRepo:    roomRepo,
        waNumber:    waNumber,
    }
}

func (s *bookingService) Create(req hotel.CreateBookingRequest) (*hotel.BookingResponse, error) {
	// Parse tanggal
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, errors.New("format check_in tidak valid")
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return nil, errors.New("format check_out tidak valid")
	}
	if checkOut.Before(checkIn) || checkOut.Equal(checkIn) {
		return nil, errors.New("check_out harus setelah check_in")
	}

	// Cek kamar
	room, err := s.roomRepo.FindByID(req.RoomID)
	if err != nil {
		return nil, errors.New("kamar tidak ditemukan")
	}
	if room.Status != hotel.RoomStatusAvailable {
		return nil, errors.New("kamar tidak tersedia")
	}

	// Cek bentrok tanggal
	overlaps, err := s.bookingRepo.FindOverlapping(req.RoomID, checkIn, checkOut)
	if err != nil {
		return nil, err
	}
	if len(overlaps) > 0 {
		return nil, errors.New("kamar sudah dipesan pada tanggal tersebut")
	}

	// Hitung
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	totalPrice := int64(nights) * room.Price

	// Simpan
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

	if err := s.bookingRepo.Create(booking); err != nil {
		return nil, err
	}

	// WhatsApp
	waMsg := s.generateWhatsAppMessage(booking, room)
	waURL := fmt.Sprintf("https://wa.me/%s?text=%s", s.waNumber, url.QueryEscape(waMsg))

	return &hotel.BookingResponse{
		ID:          booking.ID,
		WhatsAppURL: waURL,
	}, nil
}

func (s *bookingService) generateWhatsAppMessage(b *hotel.Booking, r *hotel.Room) string {
	return fmt.Sprintf(`*PESANAN KAMAR - MUTIARA HOTEL*

Nama: %s
No. HP: %s
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
		b.Name, b.Phone,
		strings.Title(r.Type), r.Number,
		strings.Title(r.Type),
		b.CheckIn.Format("02 Jan 2006"),
		b.CheckOut.Format("02 Jan 2006"),
		b.TotalNights, b.Guests,
		formatRupiah(b.TotalPrice),
		b.Notes,
	)
}

func formatRupiah(n int64) string {
	return "Rp " + strconv.FormatInt(n, 10)
}

func (s *bookingService) Confirm(id uint) error {
	b, err := s.bookingRepo.FindByID(id)
	if err != nil {
		return err
	}
	if b.Status != hotel.BookingStatusPending {
		return errors.New("hanya pending yang bisa dikonfirmasi")
	}

	// Update booking
	b.Status = hotel.BookingStatusConfirmed
	if err := s.bookingRepo.Update(b); err != nil {
		return err
	}

	// Update room status
	room, err := s.roomRepo.FindByID(b.RoomID)
	if err != nil {
		return err
	}
	room.Status = hotel.RoomStatusBooked
	return s.roomRepo.Update(room)
}

func (s *bookingService) Cancel(id uint) error {
	b, err := s.bookingRepo.FindByID(id)
	if err != nil {
		return err
	}
	b.Status = hotel.BookingStatusCancelled
	if err := s.bookingRepo.Update(b); err != nil {
		return err
	}

	// Kembalikan kamar ke available
	room, _ := s.roomRepo.FindByID(b.RoomID)
	room.Status = hotel.RoomStatusAvailable
	return s.roomRepo.Update(room)
}

func (s *bookingService) List(status string, limit, offset int) ([]hotel.Booking, int64, error) {
	f := repohotel.BookingFilter{
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	return s.bookingRepo.List(f)
}