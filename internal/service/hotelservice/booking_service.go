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
	BASE_GUESTS_INCLUDED = 2                    // 2 orang include breakfast
	EXTRA_GUEST_PRICE    = 150000               // Rp150.000/orang/malam
	MAX_GUESTS_PER_ROOM  = 4                    // maksimal 4 orang per kamar
)

type BookingService interface {
	Create(userID uint, req hotel.CreateBookingRequest, source hotel.BookingSource, otaRef string, initialStatus hotel.BookingStatus) (*hotel.MultiBookingResponse, error)
	Confirm(id uint) error
	Cancel(id uint) error
	List(status, source string, limit, offset int) ([]hotel.Booking, int64, error)
	CheckAvailability(checkIn, checkOut time.Time, roomType string) ([]hotel.AvailabilityResponse, error)
	Update(userID, bookingID uint, req hotel.UpdateBookingRequest) (*hotel.BookingResponse, error)
	UpdateStatus(bookingID uint, newStatus hotel.BookingStatus) error
	GetMyBookings(userID uint, limit, offset int) ([]hotel.Booking, int64, error)
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


func (s *bookingService) Create(userID uint, req hotel.CreateBookingRequest, source hotel.BookingSource, otaRef string, initialStatus hotel.BookingStatus) (*hotel.MultiBookingResponse, error) {
	// Validasi jumlah tamu
	maxGuests := req.Rooms * MAX_GUESTS_PER_ROOM
	if req.Guests > maxGuests {
		return nil, fmt.Errorf("maksimal %d tamu untuk %d kamar (4 orang/kamar)", maxGuests, req.Rooms)
	}
	if req.Guests < 1 {
		return nil, errors.New("minimal 1 tamu")
	}

	// Parse tanggal
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

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights < 1 {
		return nil, errors.New("minimal menginap 1 malam")
	}

	// Cek availability
	avail, err := s.bookingRepo.CheckAvailability(checkIn, checkOut, req.RoomType)
	if err != nil || len(avail) == 0 {
		return nil, errors.New("tipe kamar tidak tersedia")
	}
	if avail[0].AvailableRooms < req.Rooms {
		return nil, fmt.Errorf("hanya tersedia %d kamar %s", avail[0].AvailableRooms, strings.Title(req.RoomType))
	}

	// Hitung extra guests secara adil
	totalBaseGuests := req.Rooms * BASE_GUESTS_INCLUDED
	totalExtraGuests := max(0, req.Guests-totalBaseGuests)
	extraCharge := int64(totalExtraGuests) * EXTRA_GUEST_PRICE * int64(nights)
	basePrice := int64(nights) * avail[0].PricePerNight * int64(req.Rooms)
	totalPrice := basePrice + extraCharge

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.New("gagal memulai transaksi")
	}

	// Ambil kamar yang tersedia
	var rooms []hotel.Room
	err = tx.Raw(`
		SELECT r.* FROM rooms r
		JOIN room_types rt ON r.room_type_id = rt.id
		WHERE rt.type = ? AND r.deleted_at IS NULL
		  AND r.id NOT IN (
		    SELECT room_id FROM bookings
		    WHERE status IN ('pending','confirmed','checked_in')
		      AND (
		        (check_in < ? AND check_out > ?) OR
		        (check_in < ? AND check_out > ?) OR
		        (check_in >= ? AND check_out <= ?) OR
		        (check_in <= ? AND check_out >= ?)
		      )
		  )
		LIMIT ?
	`, req.RoomType, checkOut, checkIn, checkIn, checkOut, checkIn, checkOut, checkIn, checkOut, req.Rooms).
		Scan(&rooms).Error

	if err != nil || len(rooms) < req.Rooms {
		tx.Rollback()
		return nil, errors.New("kamar tidak cukup tersedia")
	}

	// Simpan semua booking
	var bookingIDs []uint
	for _, room := range rooms {
		booking := &hotel.Booking{
			RoomID:       room.ID,
			UserID:       &userID,
			Name:         req.Name,
			Phone:        req.Phone,
			Email:        req.Email,
			CheckIn:      checkIn,
			CheckOut:     checkOut,
			Guests:       req.Guests,
			Rooms:        req.Rooms,
			TotalNights:  nights,
			TotalPrice:   totalPrice,
			ExtraGuests:  totalExtraGuests,
			ExtraCharge:  extraCharge,
			Status:       initialStatus,
			Notes:        req.Notes,
			Source:       source,
			OtaReference: otaRef,
		}
		if err := tx.Create(booking).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		bookingIDs = append(bookingIDs, booking.ID)

		// Jika initial status bukan pending, update room status
		if initialStatus != hotel.BookingStatusPending {
			if err := s.updateRoomStatus(tx, booking, room); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	var waURL string
	if source == hotel.SourceWeb {
		waURL = s.generateWhatsAppURLMulti(req, avail[0], nights, totalPrice, int64(totalExtraGuests), checkIn, checkOut)
	} // Untuk OTA/onsite, tidak generate WA URL, atau generate different jika perlu

	return &hotel.MultiBookingResponse{
		BookingIDs:  bookingIDs,
		WhatsAppURL: waURL,
	}, nil
}


func (s *bookingService) generateWhatsAppURLMulti(
	req hotel.CreateBookingRequest,
	avail hotel.AvailabilityResponse,
	nights int,
	totalPrice int64,
	totalExtraGuests int64,
	checkIn, checkOut time.Time,
) string {
	extraCharge := totalExtraGuests * EXTRA_GUEST_PRICE * int64(nights)
	basePrice := totalPrice - extraCharge

	extraText := ""
	if totalExtraGuests > 0 {
		extraText = fmt.Sprintf("\nTamu Ekstra    : %d orang × Rp150.000 × %d malam = *Rp%s*\n(Sudah termasuk breakfast & amenities)",
			totalExtraGuests, nights, formatRupiah(extraCharge))
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

Harga per Malam: Rp%s
Subtotal Kamar : Rp%s × %d kamar × %d malam = *Rp%s*`,
		req.Name,
		req.Phone,
		req.Email,
		strings.Title(req.RoomType),
		req.Rooms,
		checkIn.Format("02 Jan 2006"),
		checkOut.Format("02 Jan 2006"),
		nights,
		req.Guests,
		extraText,
		formatRupiah(avail.PricePerNight),
		formatRupiah(avail.PricePerNight),
		req.Rooms,
		nights,
		formatRupiah(basePrice),
	)

	// TAMBAHKAN EXTRA CHARGE HANYA SEKALI DI SINI!
	if totalExtraGuests > 0 {
		msg += extraText
	}

	msg += fmt.Sprintf(`
──────────────────────────
*TOTAL BAYAR   : %s*

Catatan:
%s

Jika Sudah bayar Silahkan Kirimkan Bukti Pembayaran untuk melakukan Konfirmasi.
Terima kasih`,
		formatRupiah(totalPrice),
		req.Notes,
	)

	return "https://wa.me/" + s.waNumber + "?text=" + url.QueryEscape(msg)
}

// ==================================================================
// UPDATE BOOKING (hanya untuk 1 kamar & pending)
// ==================================================================
func (s *bookingService) Update(userID, bookingID uint, req hotel.UpdateBookingRequest) (*hotel.BookingResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, errors.New("gagal memulai transaksi")
	}

	booking, err := s.bookingRepo.FindByID(bookingID)
	if err != nil || booking.UserID == nil || *booking.UserID != userID || booking.Status != hotel.BookingStatusPending {
		tx.Rollback()
		return nil, errors.New("booking tidak ditemukan atau tidak bisa diubah")
	}
	if booking.Rooms != 1 {
		tx.Rollback()
		return nil, errors.New("update hanya diperbolehkan untuk pemesanan 1 kamar")
	}

	// Tanggal baru
	newCheckIn := booking.CheckIn
	newCheckOut := booking.CheckOut
	if req.CheckIn != nil {
		ci, err := time.Parse("2006-01-02", *req.CheckIn)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("format check_in tidak valid")
		}
		newCheckIn = ci
	}
	if req.CheckOut != nil {
		co, err := time.Parse("2006-01-02", *req.CheckOut)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("format check_out tidak valid")
		}
		newCheckOut = co
	}
	if !newCheckOut.After(newCheckIn) {
		tx.Rollback()
		return nil, errors.New("check_out harus setelah check_in")
	}

	nights := int(newCheckOut.Sub(newCheckIn).Hours() / 24)
	if nights < 1 {
		tx.Rollback()
		return nil, errors.New("minimal menginap 1 malam")
	}

	// Cek overlapping (kecuali booking ini sendiri)
	if count, _ := s.bookingRepo.CountOverlapping(booking.RoomID, newCheckIn, newCheckOut, &bookingID); count > 0 {
		tx.Rollback()
		return nil, errors.New("kamar sudah dipesan pada tanggal tersebut")
	}

	room, err := s.roomRepo.FindByID(booking.RoomID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	guests := booking.Guests
	if req.Guests != nil {
		guests = *req.Guests
		if guests > MAX_GUESTS_PER_ROOM {
			tx.Rollback()
			return nil, errors.New("maksimal 4 tamu per kamar")
		}
	}

	extraGuests := max(0, guests-BASE_GUESTS_INCLUDED)
	extraCharge := int64(extraGuests) * EXTRA_GUEST_PRICE * int64(nights)
	totalPrice := int64(nights)*room.RoomType.BasePrice + extraCharge

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
	tx.Commit()

	waURL := s.generateWhatsAppURLSingle(booking, room, nights)
	return &hotel.BookingResponse{
		ID:          booking.ID,
		WhatsAppURL: waURL,
	}, nil
}

func (s *bookingService) generateWhatsAppURLSingle(b *hotel.Booking, r *hotel.Room, nights int) string {
	extraText := ""
	if b.ExtraGuests > 0 {
		extraText = fmt.Sprintf("\nTamu Ekstra    : %d orang × Rp150.000 × %d malam = *Rp%s*",
			b.ExtraGuests, nights, formatRupiah(b.ExtraCharge))
	}

	msg := fmt.Sprintf(`*UPDATE PESANAN – MUTIARA HOTEL*

Nama           : %s
Kamar          : %s (No. %s)
Check-in       : %s
Check-out      : %s
Lama Menginap  : %d malam
Jumlah Tamu    : %d orang%s
Total Bayar    : %s

Catatan: %s`,
		b.Name,
		strings.Title(r.RoomType.Type),
		r.Number,
		b.CheckIn.Format("02 Jan 2006"),
		b.CheckOut.Format("02 Jan 2006"),
		nights,
		b.Guests,
		extraText,
		formatRupiah(b.TotalPrice),
		b.Notes,
	)

	return "https://wa.me/" + s.waNumber + "?text=" + url.QueryEscape(msg)
}

// ==================================================================
// ADMIN FUNCTIONS (Confirm, Cancel, dll)
// ==================================================================
func (s *bookingService) Confirm(id uint) error {
	tx := s.db.Begin()
	b, err := s.bookingRepo.FindByID(id)
	if err != nil || b.Status != hotel.BookingStatusPending {
		if tx.Error == nil {
			tx.Rollback()
		}
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
		if tx.Error == nil {
			tx.Rollback()
		}
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

func (s *bookingService) List(status, source string, limit, offset int) ([]hotel.Booking, int64, error) {
	f := repohotel.BookingFilter{Status: status, Source: source, Limit: limit, Offset: offset}
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

func (s *bookingService) GetMyBookings(userID uint, limit, offset int) ([]hotel.Booking, int64, error) {
	return s.bookingRepo.GetByUserID(userID, limit, offset)
}

// Helper untuk update room status di create jika initial != pending
func (s *bookingService) updateRoomStatus(tx *gorm.DB, b *hotel.Booking, room hotel.Room) error {
	switch b.Status {
	case hotel.BookingStatusConfirmed, hotel.BookingStatusCheckedIn:
		room.Status = hotel.RoomStatusBooked
		if err := tx.Save(&room).Error; err != nil {
			return err
		}
	}
	return nil
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