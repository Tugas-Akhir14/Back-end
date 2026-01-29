package hotelservice

import (
	"errors"
	"fmt"
	"net/url"
	
	"strconv"
	"strings"
	"time"
	"regexp"
	"backend/internal/models/hotel"
	"backend/internal/repository/admin"
	"backend/internal/repository/repohotel"

	"gorm.io/gorm"
)

const (
	BASE_GUESTS_INCLUDED = 2                    // 2 orang include breakfast
	EXTRA_GUEST_PRICE    = 150000               // Rp150.000/orang/malam
	MAX_GUESTS_PER_ROOM  = 4                    // maksimal 4 orang per kamar
	CHECKIN_HOUR_LIMIT   = 14                   // contoh: check-in setelah jam 14:00 di hari yang sama
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
	adminRepo   admin.AdminRepository  
	waNumber    string
	db          *gorm.DB
}

func NewBookingService(
	bookingRepo repohotel.BookingRepository,
	roomRepo repohotel.RoomRepository,
	adminRepo admin.AdminRepository,
	db *gorm.DB,
) BookingService {

	return &bookingService{
		bookingRepo: bookingRepo,
		roomRepo:    roomRepo,
		adminRepo:   adminRepo,
		db:          db,
	}
}

func (s *bookingService) getAdminWANumber() string {
	admin, err := s.adminRepo.FindHotelAdmin()
	if err != nil || admin.PhoneNumber == "" {
		return "6281396554949" // fallback
	}

	wa := admin.PhoneNumber

	// 🔹 hapus semua karakter non-angka
	re := regexp.MustCompile(`[^0-9]`)
	wa = re.ReplaceAllString(wa, "")

	// 🔹 jika diawali 0 → ubah ke 62
	if strings.HasPrefix(wa, "0") {
		wa = "62" + wa[1:]
	}

	// 🔹 jika belum diawali 62
	if !strings.HasPrefix(wa, "62") {
		wa = "62" + wa
	}
	fmt.Println("WA ADMIN:", wa)


	return wa

	
}



func (s *bookingService) Create(userID uint, req hotel.CreateBookingRequest, source hotel.BookingSource, otaRef string, initialStatus hotel.BookingStatus) (*hotel.MultiBookingResponse, error) {
	// Validasi jumlah tamu
	maxGuests := req.Rooms * MAX_GUESTS_PER_ROOM
	if req.Guests > maxGuests {
		return nil, fmt.Errorf("maksimal %d tamu untuk %d kamar (maks 4 orang/kamar)", maxGuests, req.Rooms)
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

	// Cegah booking di masa lalu
	todayStart := time.Now().Truncate(24 * time.Hour)
	if checkIn.Before(todayStart) {
		return nil, errors.New("check-in tidak boleh di masa lalu")
	}
	// Opsional: jika check-in hari ini, cek jam
	if checkIn.Equal(todayStart) && time.Now().Hour() >= CHECKIN_HOUR_LIMIT {
		// Bisa dihapus atau disesuaikan dengan kebijakan hotel
	}

	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights < 1 {
		return nil, errors.New("minimal menginap 1 malam")
	}

	// Cek ketersediaan
	availList, err := s.bookingRepo.CheckAvailability(checkIn, checkOut, req.RoomType)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa ketersediaan: %w", err)
	}
	if len(availList) == 0 {
		return nil, errors.New("tipe kamar tidak ditemukan")
	}

	avail := availList[0]
	if avail.AvailableRooms < req.Rooms {
		return nil, fmt.Errorf("hanya tersedia %d kamar %s untuk periode tersebut", avail.AvailableRooms, strings.Title(req.RoomType))
	}

	// Gunakan CurrentPrice (sudah termasuk diskon jika aktif)
	roomPricePerNight := avail.CurrentPrice
	if roomPricePerNight == 0 {
		roomPricePerNight = avail.BasePrice // fallback jika diskon tidak aktif
	}

	// Hitung biaya
	totalBaseGuests := req.Rooms * BASE_GUESTS_INCLUDED
	totalExtraGuests := max(0, req.Guests-totalBaseGuests)
	extraCharge := int64(totalExtraGuests) * EXTRA_GUEST_PRICE * int64(nights)

	basePrice := int64(nights) * roomPricePerNight * int64(req.Rooms)
	totalPrice := basePrice + extraCharge

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("gagal memulai transaksi database")
	}

	// Ambil daftar kamar yang benar-benar tersedia
	var availableRoomIDs []uint
	err = tx.Raw(`
		SELECT r.id
		FROM rooms r
		JOIN room_types rt ON rt.id = r.room_type_id
		WHERE LOWER(rt.type) = LOWER(?)
		  AND r.deleted_at IS NULL
		  AND NOT EXISTS (
		      SELECT 1 FROM bookings b
		      WHERE b.room_id = r.id
		        AND b.deleted_at IS NULL
		        AND b.status IN ('pending', 'confirmed', 'checked_in')
		        AND (
		            (b.check_in < ? AND b.check_out > ?) OR
		            (b.check_in < ? AND b.check_out > ?) OR
		            (b.check_in >= ? AND b.check_out <= ?) OR
		            (b.check_in <= ? AND b.check_out >= ?)
		        )
		  )
		ORDER BY r.id ASC
		LIMIT ?
	`, req.RoomType, checkOut, checkIn, checkIn, checkOut, checkIn, checkOut, checkIn, checkOut, req.Rooms).
		Scan(&availableRoomIDs).Error

	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("gagal mendapatkan daftar kamar tersedia: %w", err)
	}

	if len(availableRoomIDs) < req.Rooms {
		tx.Rollback()
		return nil, fmt.Errorf("ketersediaan kamar berubah, hanya tersedia %d kamar %s", len(availableRoomIDs), req.RoomType)
	}

	// Simpan semua booking
	var bookingIDs []uint
	for _, roomID := range availableRoomIDs {
		booking := &hotel.Booking{
			RoomID:       roomID,
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
			return nil, fmt.Errorf("gagal menyimpan booking: %w", err)
		}

		bookingIDs = append(bookingIDs, booking.ID)

		// Update status kamar jika booking langsung confirmed/checked_in
		if initialStatus != hotel.BookingStatusPending {
			if err := tx.Model(&hotel.Room{}).
				Where("id = ?", roomID).
				Update("status", hotel.RoomStatusBooked).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("gagal commit transaksi: %w", err)
	}

	var waURL string
	if source == hotel.SourceWeb {
		waURL = s.generateWhatsAppURLMulti(req, avail, nights, totalPrice, int64(totalExtraGuests), checkIn, checkOut, roomPricePerNight)
	}

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
	pricePerNight int64,
) string {
	extraCharge := totalExtraGuests * EXTRA_GUEST_PRICE * int64(nights)
	basePrice := totalPrice - extraCharge

	extraText := ""
	if totalExtraGuests > 0 {
		extraText = fmt.Sprintf("\nTamu Ekstra    : %d orang × Rp150.000 × %d malam = *%s*\n(Sudah termasuk sarapan & amenities dasar)",
			totalExtraGuests, nights, formatRupiah(extraCharge))
	}

	discountInfo := ""
	if avail.DiscountPercent > 0 {
		discountInfo = fmt.Sprintf("\nDiskon aktif   : %.0f%% (harga normal %s → %s)",
			avail.DiscountPercent,
			formatRupiah(avail.BasePrice),
			formatRupiah(pricePerNight))
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
%s

Harga per Malam: %s
Subtotal Kamar : %s × %d kamar × %d malam = %s
%s
──────────────────────────
*TOTAL PEMBAYARAN : %s*

Catatan tambahan:
%s

Silakan kirim bukti pembayaran untuk konfirmasi lebih cepat.
Terima kasih atas kepercayaannya! `,
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
		discountInfo,
		formatRupiah(pricePerNight),
		formatRupiah(pricePerNight),
		req.Rooms,
		nights,
		formatRupiah(basePrice),
		extraText,
		formatRupiah(totalPrice),
		req.Notes,
	)

	wa := s.getAdminWANumber()
return "https://wa.me/" + wa + "?text=" + url.QueryEscape(msg)

}

func (s *bookingService) Update(userID, bookingID uint, req hotel.UpdateBookingRequest) (*hotel.BookingResponse, error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return nil, errors.New("gagal memulai transaksi")
	}

	booking, err := s.bookingRepo.FindByID(bookingID)
	if err != nil || booking.UserID == nil || *booking.UserID != userID || booking.Status != hotel.BookingStatusPending {
		tx.Rollback()
		return nil, errors.New("booking tidak ditemukan, bukan milik Anda, atau bukan status pending")
	}

	if booking.Rooms != 1 {
		tx.Rollback()
		return nil, errors.New("update hanya diperbolehkan untuk pemesanan 1 kamar saja")
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

	// Gunakan BasePrice (bisa diganti ke CurrentPrice jika ingin diskon berlaku saat update)
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

	wa := s.getAdminWANumber()
return "https://wa.me/" + wa + "?text=" + url.QueryEscape(msg)

}

func (s *bookingService) Confirm(id uint) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	b, err := s.bookingRepo.FindByID(id)
	if err != nil || b.Status != hotel.BookingStatusPending {
		tx.Rollback()
		return errors.New("booking tidak valid atau bukan status pending")
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
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
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
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	b, err := s.bookingRepo.FindByID(bookingID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !s.isValidStatusTransition(b.Status, newStatus) {
		tx.Rollback()
		return fmt.Errorf("transisi status tidak diperbolehkan dari %s ke %s", b.Status, newStatus)
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
		hotel.BookingStatusPending:    {hotel.BookingStatusConfirmed, hotel.BookingStatusCancelled},
		hotel.BookingStatusConfirmed:  {hotel.BookingStatusCheckedIn, hotel.BookingStatusCancelled},
		hotel.BookingStatusCheckedIn:  {hotel.BookingStatusCheckedOut},
		hotel.BookingStatusCancelled:  {},
		hotel.BookingStatusCheckedOut: {},
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