	package handler

	import (
		"net/http"
		"strconv"

		"backend/internal/models"
		"backend/internal/service"
		"github.com/gin-gonic/gin"
	)

	type RoomHandler struct {
		service service.RoomService
	}

	func NewRoomHandler(s service.RoomService) *RoomHandler {
		return &RoomHandler{service: s}
	}

	func (h *RoomHandler) Create(c *gin.Context) {
		var req models.CreateRoomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
			return
		}
		room, err := h.service.Create(req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": room})
	}

	func (h *RoomHandler) GetByID(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		room, err := h.service.GetByID(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": room})
	}

	func (h *RoomHandler) List(c *gin.Context) {
		t := c.Query("type")   // superior/deluxe/executive (opsional)
		q := c.Query("q")      // search (opsional)
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		rooms, total, err := h.service.List(t, q, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list rooms"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"data":  rooms,
			"total": total,
		})
	}

	func (h *RoomHandler) Update(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var req models.UpdateRoomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
			return
		}
		room, err := h.service.Update(uint(id), req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": room})
	}

	func (h *RoomHandler) Delete(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		if err := h.service.Delete(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete room"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "room deleted"})
	}
