package hotel

import (
	"fmt"
	"io"
	"mime/multipart"
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

type GalleryHandler struct{ service hotelservice.GalleryService }

func NewGalleryHandler(s hotelservice.GalleryService) *GalleryHandler { return &GalleryHandler{s} }

// POST /api/galleries  (multipart/form-data)
// fields: image (file), title, caption, room_id (optional)
func (h *GalleryHandler) Create(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	var roomIDPtr *uint
	if v := c.PostForm("room_id"); v != "" {
		if rid, err := strconv.Atoi(v); err == nil {
			u := uint(rid)
			roomIDPtr = &u
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room_id"})
			return
		}
	}
	title := c.PostForm("title")
	caption := c.PostForm("caption")

	url, mimeType, size, saveErr := saveImageFile("uploads/gallery", file)
	if saveErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": saveErr.Error()})
		return
	}

	item := &hotel.Gallery{
		RoomID:   roomIDPtr,
		Title:    title,
		Caption:  caption,
		URL:      "/" + filepath.ToSlash(url),
		MimeType: mimeType,
		Size:     size,
	}
	if err := h.service.Create(item); err != nil {
		_ = os.Remove(url)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save gallery item"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

// GET /api/galleries?room_id=&room_type=&include_global=true|false&limit=&offset=
func (h *GalleryHandler) List(c *gin.Context) {
	var roomIDPtr *uint
	if v := c.Query("room_id"); v != "" {
		if rid, err := strconv.Atoi(v); err == nil {
			u := uint(rid)
			roomIDPtr = &u
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room_id"})
			return
		}
	}

	roomType := strings.ToLower(strings.TrimSpace(c.Query("room_type")))
	if roomType != "" && roomType != "superior" && roomType != "deluxe" && roomType != "executive" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_type must be one of: superior|deluxe|executive"})
		return
	}

	includeGlobal, _ := strconv.ParseBool(c.DefaultQuery("include_global", "false"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.service.List(roomIDPtr, roomType, includeGlobal, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list gallery"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
}

// GET /api/galleries/:id
func (h *GalleryHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gallery item not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// PUT /api/galleries/:id  (JSON metadata)
func (h *GalleryHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req hotel.UpdateGalleryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}
	item, err := h.service.Update(uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update gallery"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// PUT /api/galleries/:id/image  (multipart/form-data; field: image)
func (h *GalleryHandler) UpdateImage(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.service.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "gallery item not found"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
	 c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
	 return
	}

	newURL, mimeType, size, saveErr := saveImageFile("uploads/gallery", file)
	if saveErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": saveErr.Error()})
		return
	}

	// hapus file lama (opsional)
	if item.URL != "" {
		oldPath := strings.TrimPrefix(item.URL, "/")
		_ = os.Remove(oldPath)
	}

	item.URL = "/" + filepath.ToSlash(newURL)
	item.MimeType = mimeType
	item.Size = size

	// PENTING: simpan perubahan penuh
	if err := h.service.Save(item); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update image metadata"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}


// DELETE /api/galleries/:id
func (h *GalleryHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	item, _ := h.service.GetByID(uint(id)) // untuk hapus file fisiknya
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete gallery"})
		return
	}
	if item != nil && item.URL != "" {
		_ = os.Remove(strings.TrimPrefix(item.URL, "/"))
	}
	c.JSON(http.StatusOK, gin.H{"message": "gallery deleted"})
}

// ===== utilities =====
func saveImageFile(dir string, fh *multipart.FileHeader) (string, string, int64, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", "", 0, fmt.Errorf("cannot create dir: %w", err)
	}
	src, err := fh.Open()
	if err != nil {
		return "", "", 0, fmt.Errorf("cannot open file: %w", err)
	}
	defer src.Close()

	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	mime := http.DetectContentType(head[:n])
	allowed := map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true}
	if !allowed[mime] {
		return "", "", 0, fmt.Errorf("unsupported image type: %s", mime)
	}
	_, _ = src.Seek(0, 0)

	ext := extFromMime(mime, fh.Filename)
	name := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), slugBase(fh.Filename), ext)
	dst := filepath.Join(dir, name)

	// simpan manual
	out, err := os.Create(dst)
	if err != nil {
		return "", "", 0, err
	}
	defer out.Close()
	if _, err := io.Copy(out, src); err != nil {
		return "", "", 0, err
	}
	return dst, mime, fh.Size, nil
}

func extFromMime(mime, fallback string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return strings.ToLower(filepath.Ext(fallback))
	}
}

func slugBase(name string) string {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	base = strings.ToLower(base)
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return -1
	}, base)
	if base == "" {
		base = "img"
	}
	return base
}
