// internal/handler/hotel/booking_handler.go
package hotel

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/models/hotel"
	"backend/internal/service/hotelservice"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type response struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type BookingHandler struct {
	service hotelservice.BookingService
}

func NewBookingHandler(service hotelservice.BookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

// ==================================================================
// CREATE BOOKING – BISA 1 SAMPAI N KAMAR SEKALIGUS (UNTUK GUEST, SOURCE WEB)
// ==================================================================
func (h *BookingHandler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.unauthorized(c, "token tidak valid")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		h.unauthorized(c, "user id tidak valid")
		return
	}

	var req hotel.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "data tidak valid: "+err.Error())
		return
	}

	// Validasi nomor WhatsApp harus diawali 62
	phone := strings.ReplaceAll(req.Phone, " ", "")
	if !strings.HasPrefix(phone, "62") {
		h.badRequest(c, "nomor WhatsApp harus diawali 62 (contoh: 6281234567890)")
		return
	}
	req.Phone = phone // normalisasi

	resp, err := h.service.Create(uid, req, hotel.SourceWeb, "", hotel.BookingStatusPending)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.created(c, resp) // → MultiBookingResponse { booking_ids: [], whatsapp_url: "" }
}

// ==================================================================
// CREATE MANUAL BOOKING – UNTUK ADMIN HOTEL (ONSITE/OTA)
// ==================================================================
func (h *BookingHandler) CreateManual(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.unauthorized(c, "token tidak valid")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		h.unauthorized(c, "user id tidak valid")
		return
	}

	var req hotel.CreateManualBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "data tidak valid: "+err.Error())
		return
	}

	// Normalisasi phone jika ada
	if req.Phone != "" {
		phone := strings.ReplaceAll(req.Phone, " ", "")
		if !strings.HasPrefix(phone, "62") {
			h.badRequest(c, "nomor WhatsApp harus diawali 62 (contoh: 6281234567890)")
			return
		}
		req.Phone = phone
	}

	status := hotel.BookingStatusPending
	if req.Status != "" {
		status = hotel.BookingStatus(req.Status)
	}

	resp, err := h.service.Create(uid, hotel.CreateBookingRequest{
		RoomType: req.RoomType,
		Rooms:    req.Rooms,
		Name:     req.Name,
		Phone:    req.Phone,
		Email:    req.Email,
		CheckIn:  req.CheckIn,
		CheckOut: req.CheckOut,
		Guests:   req.Guests,
		Notes:    req.Notes,
	}, hotel.BookingSource(req.Source), req.OtaReference, status)
	if err != nil {
		h.handleError(c, err)	
		return
	}

	h.created(c, resp)
}

// ==================================================================
// CEK KETERSEDIAAN KAMAR
// ==================================================================
func (h *BookingHandler) CheckAvailability(c *gin.Context) {
	var req hotel.AvailabilityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.badRequest(c, "parameter query tidak valid: "+err.Error())
		return
	}

	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		h.badRequest(c, "format check_in salah, gunakan YYYY-MM-DD")
		return
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		h.badRequest(c, "format check_out salah, gunakan YYYY-MM-DD")
		return
	}
	if !checkOut.After(checkIn) {
		h.badRequest(c, "check_out harus setelah check_in")
		return
	}

	res, err := h.service.CheckAvailability(checkIn, checkOut, req.Type)
	if err != nil {
		h.internalError(c, err.Error())
		return
	}

	h.ok(c, res)
}

// ==================================================================
// ADMIN: List, Confirm, Cancel, UpdateStatus
// ==================================================================
func (h *BookingHandler) List(c *gin.Context) {
	status := c.Query("status")
	source := c.Query("source") // Tambah filter source
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	bookings, total, err := h.service.List(status, source, limit, offset)
	if err != nil {
		h.internalError(c, err.Error())
		return
	}

	h.ok(c, gin.H{
		"data":   bookings,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *BookingHandler) Confirm(c *gin.Context) {
	id, err := h.parseID(c, "id")
	if err != nil {
		h.badRequest(c, err.Error())
		return
	}
	if err := h.service.Confirm(id); err != nil {
		h.handleError(c, err)
		return
	}
	h.ok(c, gin.H{"message": "Booking berhasil dikonfirmasi"})
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	id, err := h.parseID(c, "id")
	if err != nil {
		h.badRequest(c, err.Error())
		return
	}
	if err := h.service.Cancel(id); err != nil {
		h.handleError(c, err)
		return
	}
	h.ok(c, gin.H{"message": "Booking berhasil dibatalkan"})
}

func (h *BookingHandler) UpdateStatus(c *gin.Context) {
	id, err := h.parseID(c, "id")
	if err != nil {
		h.badRequest(c, err.Error())
		return
	}

	var req hotel.UpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "request body tidak valid: "+err.Error())
		return
	}

	if err := h.service.UpdateStatus(id, hotel.BookingStatus(req.Status)); err != nil {
		h.handleError(c, err)
		return
	}

	h.ok(c, gin.H{
		"message": "Status booking berhasil diubah menjadi " + req.Status,
	})
}

// ==================================================================
// GUEST: Update & Get My Bookings
// ==================================================================
func (h *BookingHandler) Update(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.unauthorized(c, "token tidak valid")
		return
	}
	uid := userID.(uint)

	id, err := h.parseID(c, "id")
	if err != nil {
		h.badRequest(c, err.Error())
		return
	}

	var req hotel.UpdateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "request body tidak valid: "+err.Error())
		return
	}

	if req.CheckIn == nil && req.CheckOut == nil && req.Guests == nil && req.Notes == nil {
		h.badRequest(c, "tidak ada data yang diubah")
		return
	}

	resp, err := h.service.Update(uid, id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.ok(c, resp)
}

func (h *BookingHandler) GetMyBookings(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.unauthorized(c, "token tidak valid")
		return
	}
	uid := userID.(uint)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	bookings, total, err := h.service.GetMyBookings(uid, limit, offset)
	if err != nil {
		h.internalError(c, "gagal mengambil riwayat booking")
		return
	}

	h.ok(c, gin.H{
		"data":   bookings,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// ==================================================================
// HELPER FUNCTIONS
// ==================================================================
func (h *BookingHandler) parseID(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return 0, errors.New("ID tidak valid")
	}
	return uint(id), nil
}

func (h *BookingHandler) handleError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.notFound(c, "booking tidak ditemukan")
		return
	}
	h.badRequest(c, err.Error())
}

// Response helpers
func (h *BookingHandler) ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, response{Data: data})
}

func (h *BookingHandler) created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, response{Data: data})
}

func (h *BookingHandler) badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusUnprocessableEntity, response{Error: msg})
}

func (h *BookingHandler) unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, response{Error: msg})
}

func (h *BookingHandler) notFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, response{Error: msg})
}

func (h *BookingHandler) internalError(c *gin.Context, msg string) {
	c.JSON(http.StatusInternalServerError, response{Error: "internal server error: " + msg})
}