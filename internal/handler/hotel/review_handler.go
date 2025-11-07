package hotel

import (
	"backend/internal/models/auth"
	"backend/internal/service/hotelservice"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ReviewHandler struct {
	service hotelservice.ReviewService
}

func NewReviewHandler(s hotelservice.ReviewService) *ReviewHandler {
	return &ReviewHandler{s}
}

func (h *ReviewHandler) Create(c *gin.Context) {
	var input struct {
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
		Comment   string `json:"comment" binding:"required,min=10"`
		GuestName string `json:"guest_name"`
		Honeypot  string `json:"honeypot"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Honeypot != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "spam"})
		return
	}

	adminIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login diperlukan"})
		return
	}
	adminID := adminIDVal.(uint)

	roleVal, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Role tidak ditemukan"})
		return
	}
	role := auth.Role(roleVal.(string))

	if role != auth.RoleGuest {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya tamu (guest) yang dapat mengirim ulasan"})
		return
	}

	err := h.service.Create(hotelservice.CreateInput{
		Rating:    input.Rating,
		Comment:   input.Comment,
		GuestName: input.GuestName,
	}, c.ClientIP(), adminID)

	if err != nil {
		if err.Error() == "rate limit exceeded" {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Tunggu 3 menit sebelum kirim ulasan lagi",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan ulasan: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Ulasan terkirim! Menunggu moderasi.",
	})
}

func (h *ReviewHandler) GetApproved(c *gin.Context) {
	revs, err := h.service.GetApproved()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil ulasan"})
		return
	}
	c.JSON(http.StatusOK, revs)
}

func (h *ReviewHandler) GetPending(c *gin.Context) {
	revs, err := h.service.GetPending()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil ulasan pending"})
		return
	}
	c.JSON(http.StatusOK, revs)
}

func (h *ReviewHandler) Approve(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	if err := h.service.Approve(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ulasan tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Ulasan disetujui"})
}

func (h *ReviewHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}
	if err := h.service.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus ulasan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Ulasan dihapus"})
}