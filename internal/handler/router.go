// backend/internal/handler/routes.go
package handler

import (
	"backend/internal/config"
	"backend/internal/repository/repohotel"
	"backend/internal/repository/reposouvenir" // TAMBAHAN
	"backend/internal/service/serviceauth"
	"backend/internal/service/hotelservice"
	"backend/internal/service/souvenirservice" // TAMBAHAN
	"backend/middleware"
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/handler/auth"
	"backend/internal/handler/hotel"
	"backend/internal/handler/souvenirhandler" // TAMBAHAN
)

func SetupRoutes(r *gin.Engine, adminService serviceauth.AdminService) {
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// === AUTH ===
	adm := auth.NewAdminHandler(adminService)
	r.POST("/admins/register", adm.Register)
	r.POST("/admins/login", adm.Login)
	r.GET("/admins/profile", middleware.AuthMiddleware(), adm.GetProfile)

	// === DATABASE ===
	db := config.GetDB()

	// === HOTEL SERVICES ===
	roomH := hotel.NewRoomHandler(hotelservice.NewRoomService(repohotel.NewRoomRepository(db)))
	galleryH := hotel.NewGalleryHandler(hotelservice.NewGalleryService(repohotel.NewGalleryRepository(db)))
	newsH := hotel.NewNewsHandler(hotelservice.NewNewsService(repohotel.NewNewsRepository(db)))
	visionMissionH := hotel.NewVisionMissionHandler(hotelservice.NewVisionMissionService(repohotel.NewVisionMissionRepository(db)))

	// === SOUVENIR SERVICES (BARU) ===
	// Product
	productRepo := reposouvenir.NewProductRepository(db)
	productService := souvenirservice.NewProductService(productRepo)
	productH := souvenirhandler.NewProductHandler(productService)

	// Category
	categoryRepo := reposouvenir.NewCategoryRepository(db)
	categoryService := souvenirservice.NewCategoryService(categoryRepo)
	categoryH := souvenirhandler.NewCategoryHandler(categoryService)

	// === STATIC FILES ===
	r.Static("/uploads", "./uploads")

	// === API GROUP (PROTECTED) ===
	api := r.Group("/api", middleware.AuthMiddleware())
	{
		// --- HOTEL MODULE ---
		api.POST("/rooms", roomH.Create)
		api.GET("/rooms", roomH.List)
		api.GET("/rooms/:id", roomH.GetByID)
		api.PUT("/rooms/:id", roomH.Update)
		api.DELETE("/rooms/:id", roomH.Delete)

		api.POST("/galleries", galleryH.Create)
		api.GET("/galleries", galleryH.List)
		api.GET("/galleries/:id", galleryH.GetByID)
		api.PUT("/galleries/:id", galleryH.Update)
		api.PUT("/galleries/:id/image", galleryH.UpdateImage)
		api.DELETE("/galleries/:id", galleryH.Delete)

		api.GET("/news", newsH.List)
		api.GET("/news/:id", newsH.GetByID)
		api.GET("/news/slug/:slug", newsH.GetBySlug)
		api.POST("/news", newsH.Create)
		api.PUT("/news/:id", newsH.Update)
		api.DELETE("/news/:id", newsH.Delete)

		api.GET("/visi-misi", visionMissionH.Get)
		api.PUT("/visi-misi", visionMissionH.Upsert)

		// --- SOUVENIR: CATEGORY ---
		api.POST("/categories", categoryH.Create)
		api.GET("/categories", categoryH.GetAll)
		api.GET("/categories/:id", categoryH.GetByID)
		api.PUT("/categories/:id", categoryH.Update)
		api.DELETE("/categories/:id", categoryH.Delete)

		// --- SOUVENIR: PRODUCT ---
		api.POST("/products", productH.CreateProduct)
		api.GET("/products", productH.GetAllProducts)
		api.GET("/products/:id", productH.GetProduct)
		api.PUT("/products/:id", productH.UpdateProduct)
		api.DELETE("/products/:id", productH.DeleteProduct)

		// Di dalam grup /api
		api.GET("/products/category/:category_id", productH.GetProductsByCategory)
	}
}