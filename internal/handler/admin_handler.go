package handler

import (
    "backend/internal/models"
    "backend/internal/service"
    "github.com/gin-gonic/gin"
    "net/http"
    "log"
    "fmt"
    "strconv"
)

type AdminHandler struct {
    service service.AdminService
}

func NewAdminHandler(service service.AdminService) *AdminHandler {
    return &AdminHandler{service}
}

func (h *AdminHandler) Register(c *gin.Context) {
    var admin models.Admin

    // Bind form-data ke struct Admin
    if err := c.ShouldBind(&admin); err != nil {
        log.Printf("Error binding form data: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
        return
    }

    // Validasi apakah password dan confirm_password cocok
    if admin.Password != admin.ConfirmPassword {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
        return
    }

    // Panggil service untuk mendaftarkan admin
    err := h.service.Register(&admin, admin.Password)
    if err != nil {
        log.Printf("Error registering admin: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register admin: " + err.Error()})
        return
    }

    // Return success response
    c.JSON(http.StatusCreated, gin.H{"message": "Admin registered successfully"})
}


func (h *AdminHandler) Login(c *gin.Context) {
    var loginData struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required,min=6"`
    }
    if err := c.ShouldBindJSON(&loginData); err != nil {
        log.Printf("Login binding error: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
        return
    }

    token, err := h.service.Login(loginData.Email, loginData.Password)
    if err != nil {
        log.Printf("Login error: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Login successful",
        "token":   token,
    })
}

func (h *AdminHandler) GetProfile(c *gin.Context) {
    // Mengambil user_id dari konteks yang diset oleh middleware
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Mengonversi userID ke uint (karena ID disimpan sebagai string dalam konteks)
    id, err := strconv.ParseUint(fmt.Sprint(userID), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ID pengguna tidak valid"})
        return
    }

    // Mengambil profil admin berdasarkan ID
    admin, err := h.service.GetProfile(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    // Mengembalikan profil admin dalam format JSON
    c.JSON(http.StatusOK, gin.H{
        "message": "Profil berhasil diambil",
        "data":    admin,
    })
}
