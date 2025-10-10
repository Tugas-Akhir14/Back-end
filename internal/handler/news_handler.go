package handler

import (
	"backend/internal/models"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type NewsHandler struct {
	svc service.NewsService
}

func NewNewsHandler(s service.NewsService) *NewsHandler {
	return &NewsHandler{svc: s}
}

// Public: GET /news?page=1&page_size=10&q=...&status=published
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
	c.JSON(http.StatusOK, gin.H{"data": data, "total": total, "page": page, "page_size": ps})
}

// Public: GET /news/:id
func (h *NewsHandler) GetByID(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := h.svc.GetByID(uint(id64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// Public: GET /news/slug/:slug
func (h *NewsHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	item, err := h.svc.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "news not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// Protected (form-data): POST /news
func (h *NewsHandler) Create(c *gin.Context) {
	var input models.News
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	img, _ := c.FormFile("image") // optional

	// contoh ambil adminID dari middleware (sesuaikan claim mu)
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
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

// Protected (form-data): PUT /news/:id
func (h *NewsHandler) Update(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	var input models.News
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	img, _ := c.FormFile("image") // optional

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
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// Protected: DELETE /news/:id
func (h *NewsHandler) Delete(c *gin.Context) {
	id64, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.Delete(uint(id64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
