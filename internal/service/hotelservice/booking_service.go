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

const (
	BASE_GUESTS_INCLUDED = 2
	EXTRA_GUEST_PRICE    = 150000 // Rp150.000 per orang per malam
	MAX_GUESTS           = 4
)

type BookingService interface {
	CreateWithUser(userID uint, req hotel.CreateBookingRequest) (*hotel.BookingResponse, error)
	Confirm(id uint) error
	Cancel(id uint) error
	List(status string, limit, offset int) ([]hotel.Booking, int64, error)
	CheckAvailability(checkIn, checkOut time.Time, roomType string) ([]hotel.AvailabilityResponse, error)
	GuestBook(userID uint, req hotel.GuestBookingRequest) (*hotel.GuestBookingResponse, error)
	Update(userID, bookingID uint, req hotel.UpdateBookingRequest) (*hotel.BookingResponse, error)
	UpdateStatus(bookingID uint, newStatus hotel.BookingStatus) error
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

// ==================================================================
// CREATE BOOKING (WAJIB LOGIN)
// ==================================================================
func (s *bookingService) CreateWithUser(userID uint, req hotel.CreateBookingRequest) (*hotel.BookingResponse, error) {
	if req.Guests < 1 {
		return nil, errors.New("minimal 1 tamu")
	}
	if req.Guests > MAX_GUESTS {
		return nil, errors.New("maksimal 4 tamu per kamar")
	}

	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		return nil, errors.New("format check_in tidak valid (YYYY-MM-DD)")
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		return nil, errors.New("format check_out tidak valid (YYYY-MM-DD)")
	}
	if !checkOut.After(checkIn) {
		return nil, errors.New("check_out harus setelah check_in")
	}

	room, err := s.roomRepo.FindByID(req.RoomID)
	if err != nil {
		return nil, errors.New("kamar tidak ditemukan")
	}
	if room.Status != hotel.RoomStatusAvailable {
		return nil, errors.New("kamar sedang tidak tersedia")
	}

	count, err := s.bookingRepo.CountOverlapping(req.RoomID, checkIn, checkOut, nil)
	if err != nil || count > 0 {
		return nil, errors.New("kamar sudah dipesan pada tanggal tersebut")
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		return nil, errors.New("minimal menginap 1 malam")
	}

	extraGuests := max(0, req.Guests-BASE_GUESTS_INCLUDED)
	extraCharge := int64(extraGuests) * EXTRA_GUEST_PRICE * int64(nights)
	basePrice := int64(nights) * room.RoomType.Price
	totalPrice := basePrice + extraCharge

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.New("gagal memulai transaksi")
	}
	defer func() { if r := recover(); r != nil { tx.Rollback() } }()

	booking := &hotel.Booking{
		RoomID:      req.RoomID,
		UserID:      &userID,
		Name:        req.Name,
		Phone:       req.Phone,
		Email:       req.Email,
		CheckIn:     checkIn,
		CheckOut:    checkOut,
		Guests:      req.Guests,
		ExtraGuests: extraGuests,
		ExtraCharge: extraCharge,
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

	waURL := s.generateWhatsAppURL(booking, room, nights)
	return &hotel.BookingResponse{
		ID:          booking.ID,
		WhatsAppURL: waURL,
	}, nil
}

// ==================================================================
// UPDATE BOOKING (Guest) – Hitung ulang extra charge
// ==================================================================
func (s *bookingService) Update(userID, bookingID uint, req hotel.UpdateBookingRequest) (*hotel.BookingResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.New("gagal memulai transaksi")
	}
	defer func() { if r := recover(); r != nil { tx.Rollback() } }()

	booking, err := s.bookingRepo.FindByID(bookingID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("booking tidak ditemukan")
	}
	if booking.UserID == nil || *booking.UserID != userID {
		tx.Rollback()
		return nil, errors.New("akses ditolak")
	}
	if booking.Status != hotel.BookingStatusPending {
		tx.Rollback()
		return nil, errors.New("hanya booking pending yang bisa diubah")
	}

	newCheckIn := booking.CheckIn
	newCheckOut := booking.CheckOut
	if req.CheckIn != nil {
		ci, err := time.Parse("2006-01-02", *req.CheckIn)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("format check_in salah")
		}
		newCheckIn = ci
	}
	if req.CheckOut != nil {
		co, err := time.Parse("2006-01-02", *req.CheckOut)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("format check_out salah")
		}
		newCheckOut = co
	}
	if !newCheckOut.After(newCheckIn) {
		tx.Rollback()
		return nil, errors.New("check_out harus setelah check_in")
	}

	count, err := s.bookingRepo.CountOverlapping(booking.RoomID, newCheckIn, newCheckOut, &bookingID)
	if err != nil || count > 0 {
		tx.Rollback()
		return nil, errors.New("kamar sudah dipesan pada tanggal baru")
	}

	nights := int(newCheckOut.Sub(newCheckIn).Hours() / 24)
	if nights <= 0 {
		tx.Rollback()
		return nil, errors.New("minimal 1 malam")
	}

	room, err := s.roomRepo.FindByID(booking.RoomID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	guests := booking.Guests
	if req.Guests != nil {
		guests = *req.Guests
		if guests > MAX_GUESTS {
			tx.Rollback()
			return nil, errors.New("maksimal 4 tamu per kamar")
		}
	}

	extraGuests := max(0, guests-BASE_GUESTS_INCLUDED)
	extraCharge := int64(extraGuests) * EXTRA_GUEST_PRICE * int64(nights)
	basePrice := int64(nights) * room.RoomType.Price
	totalPrice := basePrice + extraCharge

	booking.CheckIn = newCheckIn
	booking.CheckOut = newCheckOut
	booking.TotalNights = nights
	booking.Guests = guests
	booking.ExtraGuests = extraGuests
	booking.ExtraCharge = extraCharge
	booking.TotalPrice = totalPrice
	if req.Notes != nil {
		booking.Notes = *req.Notes
	}

	if err := tx.Save(booking).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	waURL := s.generateWhatsAppURL(booking, room, nights)
	return &hotel.BookingResponse{
		ID:          booking.ID,
		WhatsAppURL: waURL,
	}, nil
}

// ==================================================================
// WHATSAPP MESSAGE – Profesional & Jelas
// ==================================================================
func (s *bookingService) generateWhatsAppURL(b *hotel.Booking, r *hotel.Room, nights int) string {
	basePrice := b.TotalPrice - b.ExtraCharge
	extraText := ""
	if b.ExtraGuests > 0 {
		extraText = fmt.Sprintf("\nTamu Ekstra    : %d orang × Rp150.000 × %d malam = *Rp%s*\n(Sudah termasuk breakfast & amenities)",
			b.ExtraGuests, nights, formatRupiah(b.ExtraCharge))
	}

	msg := fmt.Sprintf(`*PESANAN BARU - MUTIARA HOTEL*

Nama           : %s
No. WhatsApp   : %s
Email          : %s

Kamar          : %s (No. %s)
Check-in       : %s
Check-out      : %s
Lama Menginap  : %d malam
Jumlah Tamu    : %d orang%s

Harga Kamar    : %s%s
──────────────────────────
*TOTAL BAYAR   : %s*

Catatan:
%s

Mohon segera konfirmasi ke tamu.
Terima kasih`,
		b.Name,
		b.Phone,
		b.Email,
		strings.Title(r.RoomType.Type),
		r.Number,
		b.CheckIn.Format("02 Jan 2006"),
		b.CheckOut.Format("02 Jan 2006"),
		nights,
		b.Guests,
		extraText,                    // ← hanya 1x di sini!
		formatRupiah(basePrice),
		extraText,                    // ← hanya 1x lagi di sini (untuk detail harga)
		formatRupiah(b.TotalPrice),
		b.Notes,
	)

	return "https://wa.me/" + s.waNumber + "?text=" + url.QueryEscape(msg)
}
// ==================================================================
// GUEST BOOK (Multiple Rooms) – Belum ada extra guest (opsional nanti)
// ==================================================================
func (s *bookingService) GuestBook(userID uint, req hotel.GuestBookingRequest) (*hotel.GuestBookingResponse, error) {

	// VALIDASI MAKSIMAL 4 TAMU PER KAMAR
	if req.Guests > MAX_GUESTS {
		return nil, errors.New("maksimal 4 tamu per kamar")
	}
	if req.Guests < 1 {
		return nil, errors.New("minimal 1 tamu")
	}

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
	if err != nil || len(avail) == 0 {
		return nil, errors.New("tipe kamar tidak tersedia")
	}
	if avail[0].AvailableRooms < req.TotalRooms {
		return nil, fmt.Errorf("hanya %d kamar tersedia", avail[0].AvailableRooms)
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	pricePerRoom := int64(nights) * avail[0].PricePerNight
	totalPrice := pricePerRoom * int64(req.TotalRooms)

	tx := s.db.Begin()
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
			TotalPrice:  pricePerRoom,
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

func (s *bookingService) generateWhatsAppURLGuest(req hotel.GuestBookingRequest, avail hotel.AvailabilityResponse, nights int, totalPrice int64, totalRooms int, checkIn, checkOut time.Time) string {
	// Hitung extra guest (sama seperti booking biasa)
	extraGuests := max(0, req.Guests-BASE_GUESTS_INCLUDED)
	extraCharge := int64(extraGuests) * EXTRA_GUEST_PRICE * int64(nights) * int64(totalRooms) // per kamar!
	basePrice := int64(nights) * avail.PricePerNight * int64(totalRooms)
	totalWithExtra := basePrice + extraCharge

	extraText := ""
	if extraGuests > 0 {
		extraText = fmt.Sprintf("\nTamu Ekstra    : %d orang × Rp150.000 × %d malam × %d kamar = *Rp%s*\n(Sudah termasuk breakfast & amenities)",
			extraGuests, nights, totalRooms, formatRupiah(extraCharge))
	}

	msg := fmt.Sprintf(`*PESANAN KAMAR – MUTIARA HOTEL*

Nama           : %s
No. WhatsApp   : %s
Email          : %s

Tipe Kamar     : %s
Jumlah Kamar   : %d kamar
Check-in       : %s
Check-out      : %s
Lama Menginap  : %d malam
Jumlah Tamu    : %d orang%s

Harga Kamar    : %s%s
──────────────────────────
*TOTAL BAYAR   : %s*

Catatan:
%s

Mohon segera konfirmasi ke tamu.
Terima kasih`,
		req.Name,
		req.Phone,
		req.Email,
		strings.Title(req.RoomType),
		totalRooms,
		checkIn.Format("02 Jan 2006"),
		checkOut.Format("02 Jan 2006"),
		nights,
		req.Guests,
		extraText,
		formatRupiah(basePrice),
		extraText,
		formatRupiah(totalWithExtra),
		req.Notes,
	)

	return "https://wa.me/" + s.waNumber + "?text=" + url.QueryEscape(msg)
}

// ==================================================================
// ADMIN: Confirm, Cancel, UpdateStatus, List
// ==================================================================
func (s *bookingService) Confirm(id uint) error {
	tx := s.db.Begin()
	b, err := s.bookingRepo.FindByID(id)
	if err != nil || b.Status != hotel.BookingStatusPending {
		if tx.Error == nil { tx.Rollback() }
		return errors.New("booking tidak valid atau bukan pending")
	}
	b.Status = hotel.BookingStatusConfirmed
	if err := tx.Save(b).Error; err != nil {
		tx.Rollback()
		return err
	}
	if room, _ := s.roomRepo.FindByID(b.RoomID); room != nil {
		room.Status = hotel.RoomStatusBooked
		tx.Save(room)
	}
	return tx.Commit().Error
}

func (s *bookingService) Cancel(id uint) error {
	tx := s.db.Begin()
	b, err := s.bookingRepo.FindByID(id)
	if err != nil {
		if tx.Error == nil { tx.Rollback() }
		return err
	}
	b.Status = hotel.BookingStatusCancelled
	if err := tx.Save(b).Error; err != nil {
		tx.Rollback()
		return err
	}
	if room, _ := s.roomRepo.FindByID(b.RoomID); room != nil && room.Status == hotel.RoomStatusBooked {
		room.Status = hotel.RoomStatusAvailable
		tx.Save(room)
	}
	return tx.Commit().Error
}

func (s *bookingService) List(status string, limit, offset int) ([]hotel.Booking, int64, error) {
	f := repohotel.BookingFilter{Status: status, Limit: limit, Offset: offset}
	return s.bookingRepo.List(f)
}

func (s *bookingService) CheckAvailability(checkIn, checkOut time.Time, roomType string) ([]hotel.AvailabilityResponse, error) {
	return s.bookingRepo.CheckAvailability(checkIn, checkOut, roomType)
}

func (s *bookingService) UpdateStatus(bookingID uint, newStatus hotel.BookingStatus) error {
	tx := s.db.Begin()
	b, err := s.bookingRepo.FindByID(bookingID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !s.isValidStatusTransition(b.Status, newStatus) {
		tx.Rollback()
		return fmt.Errorf("transisi status tidak diperbolehkan")
	}
	b.Status = newStatus
	if err := tx.Save(b).Error; err != nil {
		tx.Rollback()
		return err
	}
	if room, _ := s.roomRepo.FindByID(b.RoomID); room != nil {
		switch newStatus {
		case hotel.BookingStatusConfirmed, hotel.BookingStatusCheckedIn:
			room.Status = hotel.RoomStatusBooked
		case hotel.BookingStatusCancelled, hotel.BookingStatusCheckedOut:
			if count, _ := s.bookingRepo.CountOverlapping(b.RoomID, b.CheckIn, b.CheckOut, nil); count == 0 {
				room.Status = hotel.RoomStatusAvailable
			}
		}
		tx.Save(room)
	}
	return tx.Commit().Error
}

func (s *bookingService) isValidStatusTransition(from, to hotel.BookingStatus) bool {
	allowed := map[hotel.BookingStatus][]hotel.BookingStatus{
		hotel.BookingStatusPending:     {hotel.BookingStatusConfirmed, hotel.BookingStatusCancelled},
		hotel.BookingStatusConfirmed:   {hotel.BookingStatusCheckedIn, hotel.BookingStatusCancelled},
		hotel.BookingStatusCheckedIn:   {hotel.BookingStatusCheckedOut},
		hotel.BookingStatusCancelled:   {},
		hotel.BookingStatusCheckedOut:  {},
	}
	for _, v := range allowed[from] {
		if v == to {
			return true
		}
	}
	return false
}

// ==================================================================
// HELPER
// ==================================================================
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

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