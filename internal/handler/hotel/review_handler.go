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
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")
	role := auth.Role(c.GetString("role"))

	if role != auth.RoleGuest {
		c.JSON(http.StatusForbidden, gin.H{"error": "Hanya tamu yang bisa mengirim ulasan"})
		return
	}

	err := h.service.Create(hotelservice.CreateReviewInput{
		Rating:    input.Rating,
		Comment:   input.Comment,
		GuestName: input.GuestName,
	}, c.ClientIP(), userID)

	if err != nil {
		if err.Error() == "Anda hanya diperbolehkan mengirim satu ulasan saja" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Ulasan berhasil dikirim! Terima kasih atas masukan Anda.",
	})
}

// PUBLIC: Ambil semua ulasan (langsung tampil semua)
func (h *ReviewHandler) GetAll(c *gin.Context) {
    reviews, err := h.service.GetAll()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil ulasan"})
        return
    }
    c.JSON(http.StatusOK, reviews)
}

// GUEST: Ambil ulasan saya sendiri
func (h *ReviewHandler) GetMyReviews(c *gin.Context) {
    userID := c.GetUint("user_id")
    reviews, err := h.service.GetMyReviews(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, reviews)
}

// GUEST: Edit ulasan sendiri
func (h *ReviewHandler) Update(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    userID := c.GetUint("user_id")

    var input struct {
        Rating  *int    `json:"rating,omitempty"`
        Comment *string `json:"comment,omitempty"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if input.Rating == nil && input.Comment == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada data yang diupdate"})
        return
    }

    err := h.service.Update(uint(id), hotelservice.UpdateReviewInput{
        Rating:  input.Rating,
        Comment: input.Comment,
    }, userID)

    if err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Ulasan berhasil diupdate"})
}

// GUEST atau ADMIN: Hapus ulasan
func (h *ReviewHandler) Delete(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    userID := c.GetUint("user_id")
    role := auth.Role(c.GetString("role"))

    isAdmin := role == auth.RoleAdminHotel || role == auth.RoleSuperAdmin

    err := h.service.Delete(uint(id), userID, isAdmin)
    if err != nil {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Ulasan berhasil dihapus"})
}