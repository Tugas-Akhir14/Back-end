package hotel

import (
	"backend/internal/models/hotel"
	"backend/internal/service/hotelservice"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type NewsHandler struct {
	svc hotelservice.NewsService
}

func NewNewsHandler(s hotelservice.NewsService) *NewsHandler {
	return &NewsHandler{svc: s}
}

// helper: bikin URL publik utk image berdasarkan host/scheme request
func publicImageURL(c *gin.Context, path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	if strings.HasPrefix(path, "/") {
		return scheme + "://" + host + path
	}
	return scheme + "://" + host + "/" + path
}

func (h *NewsHandler) normalizeImageURL(c *gin.Context, n *hotel.News) {
	if n == nil {
		return
	}
	n.ImageURL = publicImageURL(c, n.ImageURL)
}

// ================== PUBLIC ==================

// GET /public/news?page=&page_size=&q=
func (h *NewsHandler) ListPublic(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	q := c.Query("q")

	// paksa published only
	data, total, err := h.svc.List(page, ps, q, "published")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range data {
		h.normalizeImageURL(c, &data[i])
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     total,
		"page":      page,
		"page_size": ps,
	})
}

// GET /public/news/:id
func (h *NewsHandler) GetPublicByID(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := h.svc.GetByID(uint(id64))
	if err != nil || strings.ToLower(item.Status) != "published" {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}
	h.normalizeImageURL(c, item)
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// GET /public/news/slug/:slug
func (h *NewsHandler) GetPublicBySlug(c *gin.Context) {
	slug := c.Param("slug")
	item, err := h.svc.GetBySlug(slug)
	if err != nil || strings.ToLower(item.Status) != "published" {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}
	h.normalizeImageURL(c, item)
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// ================== ADMIN/PROTECTED ==================

// Public: GET /api/news?page=1&page_size=10&q=...&status=published
func (h *NewsHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	ps, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	q := c.Query("q")
	status := c.Query("status")

	data, total, err := h.svc.List(page, ps, q, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range data {
		h.normalizeImageURL(c, &data[i])
	}
	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"total":     total,
		"page":      page,
		"page_size": ps,
	})
}

func (h *NewsHandler) GetByID(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := h.svc.GetByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}
	h.normalizeImageURL(c, item)
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *NewsHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	item, err := h.svc.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}
	h.normalizeImageURL(c, item)
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *NewsHandler) Create(c *gin.Context) {
	var input hotel.News
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	img, _ := c.FormFile("image")

	var adminID uint = 0
	if val, ok := c.Get("admin_id"); ok {
		if v, ok2 := val.(uint); ok2 {
			adminID = v
		}
	}

	item, err := h.svc.Create(input, img, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.normalizeImageURL(c, item)
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *NewsHandler) Update(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var input hotel.News
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	img, _ := c.FormFile("image")

	var adminID uint = 0
	if val, ok := c.Get("admin_id"); ok {
		if v, ok2 := val.(uint); ok2 {
			adminID = v
		}
	}

	item, err := h.svc.Update(uint(id64), input, img, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.normalizeImageURL(c, item)
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *NewsHandler) Delete(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
