// internal/handler/hotel/room_type_handler.go
package hotel

import (
	"net/http"
	"strconv"

	"backend/internal/models/hotel"
	"backend/internal/service/hotelservice"

	"github.com/gin-gonic/gin"
)

type RoomTypeHandler struct {
	service hotelservice.RoomTypeService
}

func NewRoomTypeHandler(service hotelservice.RoomTypeService) *RoomTypeHandler {
	return &RoomTypeHandler{service: service}
}

func (h *RoomTypeHandler) Create(c *gin.Context) {
	var req hotel.CreateRoomTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rt, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": h.enrichWithCurrentPrice(rt)})
}

func (h *RoomTypeHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	rt, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room type not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": h.enrichWithCurrentPrice(rt),
	})
}

func (h *RoomTypeHandler) List(c *gin.Context) {
	types, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Enrich setiap room type dengan current_price
	var enriched []gin.H
	for _, rt := range types {
		enriched = append(enriched, h.enrichWithCurrentPrice(&rt))
	}

	c.JSON(http.StatusOK, gin.H{"data": enriched})
}

func (h *RoomTypeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req hotel.UpdateRoomTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rt, err := h.service.Update(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": h.enrichWithCurrentPrice(rt)})
}

func (h *RoomTypeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "room type deleted"})
}

// Helper untuk menambahkan current_price tanpa mengubah struct asli
func (h *RoomTypeHandler) enrichWithCurrentPrice(rt *hotel.RoomType) gin.H {
	return gin.H{
		"id":                     rt.ID,
		"type":                   rt.Type,
		"base_price":             rt.BasePrice,
		"discount_percentage":    rt.DiscountPercentage,
		"discount_start":         rt.DiscountStart,
		"discount_end":           rt.DiscountEnd,
		"discount_description":   rt.DiscountDescription,
		"description":            rt.Description,
		"created_at":             rt.CreatedAt,
		"updated_at":             rt.UpdatedAt,
		"current_price":          rt.GetCurrentPrice(), // ← Ini yang penting
	}
}