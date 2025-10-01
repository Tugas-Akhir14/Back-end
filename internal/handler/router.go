package handler

import (

	"backend/internal/service"
	"backend/middleware"

	"github.com/gin-gonic/gin"
	// Import the middleware package here
)

func SetupRoutes(r *gin.Engine, adminService service.AdminService) {
    adminHandler := NewAdminHandler(adminService)
    r.POST("/admins/register", adminHandler.Register)
    r.POST("/admins/login", adminHandler.Login)
    // Route untuk mengambil profil admin, dilindungi oleh middleware
    r.GET("/admins/profile", middleware.AuthMiddleware(), adminHandler.GetProfile) // Middleware is used here
}
