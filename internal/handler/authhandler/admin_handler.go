// internal/handler/auth/admin_handler.go
package authhandler

import (
	"backend/internal/models/auth"
	"backend/internal/service/serviceauth"
	"backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	service serviceauth.AdminService
}

func NewAdminHandler(service serviceauth.AdminService) *AdminHandler {
	return &AdminHandler{service}
}

func (h *AdminHandler) Register(c *gin.Context) {
	var req serviceauth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.service.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Kirim email jika perlu approval
	if !admin.IsApproved && admin.Role != "guest" {
		go utils.SendApprovalPendingEmail(admin.Email, admin.FullName)
	}

	msg := "Registrasi berhasil."
	if !admin.IsApproved {
		msg += " Menunggu persetujuan Superadmin."
	}

	c.JSON(http.StatusCreated, gin.H{"message": msg, "data": admin})
}



func (h *AdminHandler) Login(c *gin.Context) {
	var login struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.Login(login.Email, login.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login berhasil",
		"token":   resp.Token,
		"user":    resp.User,
	})
}

func (h *AdminHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	admin, err := h.service.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": admin})
}

func (h *AdminHandler) ApproveUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	role := auth.Role(c.GetString("role"))
	if err := h.service.ApproveUser(uint(id), role); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Kirim email sukses
	admin, _ := h.service.GetProfile(uint(id))
	if admin != nil {
		go utils.SendApprovalSuccessEmail(admin.Email, admin.FullName)
	}

	c.JSON(http.StatusOK, gin.H{"message": "User disetujui"})
}

func (h *AdminHandler) GetPending(c *gin.Context) {
	role := auth.Role(c.GetString("role"))
	admins, err := h.service.GetPendingAdmins(role)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": admins})
}