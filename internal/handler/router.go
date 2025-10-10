package handler

import (
	"backend/internal/config"
	"backend/internal/repository"
	"backend/internal/service"
	"backend/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, adminService service.AdminService) {
	// ==== 404 JSON biar jelas kalau path salah ====
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})

	// ==== Health ====
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// ==== Admin Auth (public) ====
	adm := NewAdminHandler(adminService)
	r.POST("/admins/register", adm.Register)
	r.POST("/admins/login", adm.Login)
	r.GET("/admins/profile", middleware.AuthMiddleware(), adm.GetProfile)

	// ==== Wiring DB & services ====
	db := config.GetDB()

	roomH := NewRoomHandler(
		service.NewRoomService(
			repository.NewRoomRepository(db),
		),
	)
	galleryH := NewGalleryHandler(
		service.NewGalleryService(
			repository.NewGalleryRepository(db),
		),
	)
	newsH := NewNewsHandler(
		service.NewNewsService(
			repository.NewNewsRepository(db),
		),
	)
	visionMissionH := NewVisionMissionHandler(
		service.NewVisionMissionService(
			repository.NewVisionMissionRepository(db),
		),
	)
	// ==== Protected API (WAJIB token) ====
	api := r.Group("/api", middleware.AuthMiddleware())
	{
		// Rooms
		api.POST("/rooms", roomH.Create)
		api.GET("/rooms", roomH.List)
		api.GET("/rooms/:id", roomH.GetByID)
		api.PUT("/rooms/:id", roomH.Update)
		api.DELETE("/rooms/:id", roomH.Delete)

		// Galleries
		api.POST("/galleries", galleryH.Create)
		api.GET("/galleries", galleryH.List)
		api.GET("/galleries/:id", galleryH.GetByID)
		api.PUT("/galleries/:id", galleryH.Update)
		api.PUT("/galleries/:id/image", galleryH.UpdateImage)
		api.DELETE("/galleries/:id", galleryH.Delete)

		// News (CRUD; read di-protect sesuai kebutuhan Anda sekarang)
		api.GET("/news", newsH.List)          // ?page=&page_size=&q=&status=
		api.GET("/news/:id", newsH.GetByID)
		api.GET("/news/slug/:slug", newsH.GetBySlug)
		api.POST("/news", newsH.Create)       // form-data (file "image")
		api.PUT("/news/:id", newsH.Update)    // form-data (image opsional)
		api.DELETE("/news/:id", newsH.Delete)

		// Vision & Mission (protected PUT)
		api.GET("/visi-misi", visionMissionH.Get)
		api.PUT("/visi-misi", visionMissionH.Upsert)
	}
}
