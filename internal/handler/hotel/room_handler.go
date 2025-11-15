// internal/handler/hotel/room_handler.go
package hotel

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"backend/internal/models/hotel"
	"backend/internal/service/hotelservice"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	service hotelservice.RoomService
}

func NewRoomHandler(service hotelservice.RoomService) *RoomHandler {
	return &RoomHandler{service: service}
}

// Buat URL publik untuk file statis
func publicURL(c *gin.Context, rel string) string {
	rel = strings.TrimLeft(rel, "/")
	scheme := "http"
	if p := c.Request.Header.Get("X-Forwarded-Proto"); p == "https" {
		scheme = "https"
	}
	host := c.Request.Host
	return fmt.Sprintf("%s://%s/%s", scheme, host, rel)
}

// Simpan file upload ke disk
func saveUploaded(c *gin.Context, formField string) (string, error) {
	file, err := c.FormFile(formField)
	if err != nil {
		return "", err
	}

	// Buat nama file unik
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	dir := "uploads/rooms"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	dst := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", err
	}

	// Kembalikan URL publik
	return publicURL(c, filepath.ToSlash(dst)), nil
}

// CREATE ROOM
func (h *RoomHandler) Create(c *gin.Context) {
	var req hotel.CreateRoomRequest

	// Cek Content-Type
	ct := c.GetHeader("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		// JSON only (tanpa file)
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid JSON: %v", err)})
			return
		}
	} else {
		// multipart/form-data (dengan file)
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid form-data: %v", err)})
			return
		}
	}

	// Upload gambar jika ada
	if req.Image != nil {
		url, err := saveUploaded(c, "image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upload image: " + err.Error()})
			return
		}
		// Simpan URL ke context untuk diambil di service
		c.Set("image_url", url)
		req.Image = nil // kosongkan file dari request
	}

	// Panggil service (pass context)
	room, err := h.service.Create(c, req)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "already exists") ||
			strings.Contains(msg, "invalid room_type_id") ||
			strings.Contains(msg, "status tidak valid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": room})
}

// UPDATE ROOM
func (h *RoomHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req hotel.UpdateRoomRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid form-data: " + err.Error()})
		return
	}

	// Upload gambar baru jika ada
	if req.Image != nil {
		url, err := saveUploaded(c, "image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to upload image: " + err.Error()})
			return
		}
		c.Set("image_url", url)
		req.Image = nil
	}

	// Panggil service
	room, err := h.service.Update(c, uint(id), req)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "already exists") ||
			strings.Contains(msg, "invalid room_type_id") ||
			strings.Contains(msg, "status tidak valid") {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update room"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": room})
}

// GET BY ID
func (h *RoomHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	room, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": room})
}

// LIST
func (h *RoomHandler) List(c *gin.Context) {
	t := c.Query("type")
	q := c.Query("q")
	status := c.Query("status")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	rooms, total, err := h.service.List(t, q, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rooms, "total": total})
}

// DELETE
func (h *RoomHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "room deleted"})
}

// LIST PUBLIC
func (h *RoomHandler) ListPublic(c *gin.Context) {
	rooms, err := h.service.ListPublic(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rooms})
}