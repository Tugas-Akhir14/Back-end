package hotel

import (
	"net/http"

	"backend/internal/models/hotel"
	"backend/internal/service/hotelservice"
	"github.com/gin-gonic/gin"
)

type VisionMissionHandler struct {
	svc hotelservice.VisionMissionService
}

func NewVisionMissionHandler(s hotelservice.VisionMissionService) *VisionMissionHandler {
	return &VisionMissionHandler{svc: s}
}

// GET /api/visi-misi  (public)
func (h *VisionMissionHandler) Get(c *gin.Context) {
	row, err := h.svc.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "visi & misi belum diset"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": row})
}

// PUT /api/admin/visi-misi  (protected)
func (h *VisionMissionHandler) Upsert(c *gin.Context) {
	var req hotel.UpsertVisionMissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload tidak valid: " + err.Error()})
		return
	}

	// Ambil user id dari context (sesuaikan dengan middleware JWT kamu)
	var uid *uint
	if v, ok := c.Get("user_id"); ok {
		if num, ok2 := v.(uint); ok2 {
			uid = &num
		}
	}

	row, err := h.svc.Upsert(c.Request.Context(), req, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": row})
}
