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

	// Pastikan service Register return *auth.AdminResponse
	adminResp, err := h.service.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// PESAN YANG BENAR SESUAI ROLE
	var message string
	if req.Role == auth.RoleGuest {
		message = "Registrasi berhasil! Silakan cek email Anda untuk kode OTP (berlaku 5 menit)."
	} else {
		message = "Registrasi berhasil. Menunggu persetujuan Superadmin."
	}

	// PENTING: KIRIM adminResp (AdminResponse), BUKAN model admin mentah!
	c.JSON(http.StatusCreated, gin.H{
		"message": message,
		"data":    adminResp, // Ini yang aman & sesuai frontend
	})

	
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

func (h *AdminHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req serviceauth.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.service.UpdateProfile(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profil berhasil diperbarui",
		"data":    admin,
	})
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req serviceauth.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ChangePassword(userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}

// Tambah method baru di AdminHandler
func (h *AdminHandler) VerifyOTP(c *gin.Context) {
    var req struct {
        Email string `json:"email" binding:"required,email"`
        OTP   string `json:"otp" binding:"required,len=6"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.service.VerifyGuestOTP(req.Email, req.OTP); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "OTP berhasil diverifikasi! Akun guest Anda sudah aktif dan bisa login.",
    })
}