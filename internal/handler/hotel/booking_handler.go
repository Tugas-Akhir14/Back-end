// internal/handler/hotel/booking_handler.go
package hotel

import (
	"errors"
	"net/http"
	"strconv"
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

func (h *BookingHandler) Create(c *gin.Context) {
	var req hotel.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "invalid request body: "+err.Error())
		return
	}

	resp, err := h.service.Create(req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.created(c, resp)
}

func (h *BookingHandler) GuestBook(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.unauthorized(c, "user not authenticated")
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		h.unauthorized(c, "invalid user id")
		return
	}

	var req hotel.GuestBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "invalid request body: "+err.Error())
		return
	}

	resp, err := h.service.GuestBook(uid, req)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.created(c, resp)
}

func (h *BookingHandler) CheckAvailability(c *gin.Context) {
	var req hotel.AvailabilityRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.badRequest(c, "invalid query parameters: "+err.Error())
		return
	}

	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		h.badRequest(c, "invalid check_in format, use YYYY-MM-DD")
		return
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		h.badRequest(c, "invalid check_out format, use YYYY-MM-DD")
		return
	}
	if !checkOut.After(checkIn) {
		h.badRequest(c, "check_out must be after check_in") 
		return
	}

	res, err := h.service.CheckAvailability(checkIn, checkOut, req.Type)
	if err != nil {
		h.internalError(c, err.Error())
		return
	}

	h.ok(c, res)
}

func (h *BookingHandler) List(c *gin.Context) {
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 {
		limit = 10
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	bookings, total, err := h.service.List(status, limit, offset)
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

	h.ok(c, gin.H{"message": "Booking dikonfirmasi"})
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

	h.ok(c, gin.H{"message": "Booking dibatalkan"})
}

// === Helper Methods ===

func (h *BookingHandler) parseID(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return 0, errors.New("invalid ID")
	}
	return uint(id), nil
}

func (h *BookingHandler) handleError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		h.notFound(c, "booking not found")
		return
	}
	h.badRequest(c, err.Error())
}

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