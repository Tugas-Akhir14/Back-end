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
	return &RoomHandler{service: service} // Perbaikan nama parameter
}

// publicURL membentuk URL absolut dari path relatif "uploads/...".
func publicURL(c *gin.Context, rel string) string {
	rel = strings.TrimLeft(rel, "/")
	scheme := "http"
	if p := c.Request.Header.Get("X-Forwarded-Proto"); p == "https" {
		scheme = "https"
	}
	host := c.Request.Host
	return fmt.Sprintf("%s://%s/%s", scheme, host, rel)
}

// saveUploaded mengunggah file "image" ke uploads/gallery dan
// mengembalikan URL absolut untuk disimpan ke DB.
func saveUploaded(c *gin.Context, formField string) (string, error) {
	file, err := c.FormFile(formField)
	if err != nil {
		return "", err
	}
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB, tangani error
		return "", err
	}

	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	dir := "uploads/gallery"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", err
	}
	rel := filepath.ToSlash(dst)           // "uploads/gallery/xxx.jpg"
	return publicURL(c, rel), nil          // "http://host:port/uploads/gallery/xxx.jpg"
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req hotel.CreateRoomRequest

	ct := c.GetHeader("Content-Type")
	if strings.HasPrefix(ct, "application/json") {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid JSON payload: %v", err)})
			return
		}
	} else {
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid form-data: %v", err)})
			return
		}
		// Jika ada file, simpan dan set URL absolut
		if url, err := saveUploaded(c, "image"); err == nil && url != "" {
			req.Image = url
		}
	}

	room, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to create room: %v", err)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": room})
}

func (h *RoomHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32) // Gunakan ParseUint untuk uint
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

func (h *RoomHandler) List(c *gin.Context) {
	t := c.Query("type")
	q := c.Query("q")
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 { // Batasi maksimum limit
		limit = 100
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	rooms, total, err := h.service.List(t, q, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list rooms: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rooms, "total": total})
}

func (h *RoomHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req hotel.UpdateRoomRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid payload: %v", err)})
		return
	}

	// Jika ada file baru, simpan & set URL absolut
	if url, err := saveUploaded(c, "image"); err == nil && url != "" {
		req.Image = &url // Gunakan pointer jika field Image adalah *string
	}

	room, err := h.service.Update(uint(id), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to update room: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": room})
}

func (h *RoomHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete room: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "room deleted"})
}