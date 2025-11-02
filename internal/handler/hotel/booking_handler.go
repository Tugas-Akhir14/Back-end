// backend/internal/handler/hotel/booking_handler.go
package hotel

import (
	"net/http"
	"strconv"

	"backend/internal/models/hotel"
	"backend/internal/service/hotelservice"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	service hotelservice.BookingService
}

func NewBookingHandler(service hotelservice.BookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

func (h *BookingHandler) Create(c *gin.Context) {
	var req hotel.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": resp})
}

func (h *BookingHandler) List(c *gin.Context) {
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	bookings, total, err := h.service.List(status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": bookings, "total": total})
}

func (h *BookingHandler) Confirm(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	if err := h.service.Confirm(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Booking dikonfirmasi"})
}

func (h *BookingHandler) Cancel(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseUint(idStr, 10, 32)
	if err := h.service.Cancel(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Booking dibatalkan"})
}