package hotel

import (
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

    if err := h.service.Create(hotelservice.CreateInput{
        Rating:    input.Rating,
        Comment:   input.Comment,
        GuestName: input.GuestName,
    }, c.ClientIP()); err != nil {
        if err.Error() == "rate limit exceeded" {
           c.JSON(http.StatusTooManyRequests, gin.H{
        "error": "Kamu sudah mengirim ulasan dalam 3 menit terakhir"})
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal simpan"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Ulasan terkirim! Menunggu moderasi.",
    })
}

func (h *ReviewHandler) GetApproved(c *gin.Context) {
    revs, _ := h.service.GetApproved()
    c.JSON(http.StatusOK, revs)
}

func (h *ReviewHandler) GetPending(c *gin.Context) {
    revs, _ := h.service.GetPending()
    c.JSON(http.StatusOK, revs)
}

func (h *ReviewHandler) Approve(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    if err := h.service.Approve(uint(id)); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Disetujui"})
}

func (h *ReviewHandler) Delete(c *gin.Context) {
    id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
    h.service.Delete(uint(id))
    c.JSON(http.StatusOK, gin.H{"message": "Dihapus"})
}