// backend/internal/handler/routes.go
package handler

import (
	"backend/internal/config"
	"backend/internal/repository/repohotel"
	"backend/internal/repository/reposouvenir"
	"backend/internal/service/serviceauth"
	"backend/internal/service/hotelservice"
	"backend/internal/service/souvenirservice"
	"backend/middleware"
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/handler/auth"
	"backend/internal/handler/hotel"
	"backend/internal/handler/souvenirhandler"
	"backend/internal/handler/bookhandler"
	"backend/internal/service/bookservice"
	"backend/internal/repository/repobook"

	"backend/internal/handler/cafehandler"
	"backend/internal/service/cafeservice"
	"backend/internal/repository/repocafe"
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

	// === SOUVENIR SERVICES ===
	souvenirProductRepo := reposouvenir.NewProductRepository(db)
	souvenirProductService := souvenirservice.NewProductService(souvenirProductRepo)
	souvenirProductH := souvenirhandler.NewProductHandler(souvenirProductService)

	souvenirCategoryRepo := reposouvenir.NewCategoryRepository(db)
	souvenirCategoryService := souvenirservice.NewCategoryService(souvenirCategoryRepo)
	souvenirCategoryH := souvenirhandler.NewCategoryHandler(souvenirCategoryService)

	// === BOOK SERVICES ===
	bookCategoryRepo := repobook.NewCategoryRepository(db)
	bookProductRepo := repobook.NewProductRepository(db)

	bookCategoryService := bookservice.NewCategoryService(bookCategoryRepo)
	bookProductService := bookservice.NewProductService(bookProductRepo, bookCategoryRepo)

	bookCategoryH := bookhandler.NewCategoryHandler(bookCategoryService)
	bookProductH := bookhandler.NewProductHandler(bookProductService, bookCategoryService)

	// === CAFE SERVICES ===
	cafeCategoryRepo := repocafe.NewCategoryRepository(db)
	cafeProductRepo := repocafe.NewProductRepository(db)

	cafeCategoryService := cafeservice.NewCategoryService(cafeCategoryRepo)
	cafeProductService := cafeservice.NewProductService(cafeProductRepo, cafeCategoryRepo)

	cafeCategoryH := cafehandler.NewCategoryHandler(cafeCategoryService)
	cafeProductH := cafehandler.NewProductHandler(cafeProductService, cafeCategoryService)

	// === STATIC FILES ===
	r.Static("/uploads", "./uploads")

	// === PUBLIC API ===
	public := r.Group("/api")
	{
		// --- BOOK: CATEGORY (PUBLIC) ---
		public.GET("/book-categories", bookCategoryH.GetAll)
		public.GET("/book-categories/:id", bookCategoryH.GetByID)

		// --- BOOK: PRODUCT (PUBLIC) ---
		public.GET("/books", bookProductH.ListBooks)
		public.GET("/books/:id", bookProductH.GetBook)
		public.GET("/books/category/:category_id", bookProductH.GetBooksByCategory)
	}

	admin := r.Group("", middleware.AuthMiddleware()) // HAPUS /api
{
	admin.GET("/pending-admins", adm.GetPending)
	admin.PATCH("/admins/approve/:id", adm.ApproveUser)
	// ... route lain tetap di /api


	// === ADMIN API ===
	admin := r.Group("/api", middleware.AuthMiddleware())
	{
		// --- HOTEL MODULE ---
		admin.POST("/rooms", roomH.Create)
		admin.GET("/rooms", roomH.List)
		admin.GET("/rooms/:id", roomH.GetByID)
		admin.PUT("/rooms/:id", roomH.Update)
		admin.DELETE("/rooms/:id", roomH.Delete)

		admin.POST("/galleries", galleryH.Create)
		admin.GET("/galleries", galleryH.List)
		admin.GET("/galleries/:id", galleryH.GetByID)
		admin.PUT("/galleries/:id", galleryH.Update)
		admin.PUT("/galleries/:id/image", galleryH.UpdateImage)
		admin.DELETE("/galleries/:id", galleryH.Delete)

		admin.GET("/news", newsH.List)
		admin.GET("/news/:id", newsH.GetByID)
		admin.GET("/news/slug/:slug", newsH.GetBySlug)
		admin.POST("/news", newsH.Create)
		admin.PUT("/news/:id", newsH.Update)
		admin.DELETE("/news/:id", newsH.Delete)

		admin.GET("/visi-misi", visionMissionH.Get)
		admin.PUT("/visi-misi", visionMissionH.Upsert)

		// --- SOUVENIR: CATEGORY ---
		admin.POST("/categories", souvenirCategoryH.Create)
		admin.GET("/categories", souvenirCategoryH.GetAll)
		admin.GET("/categories/:id", souvenirCategoryH.GetByID)
		admin.PUT("/categories/:id", souvenirCategoryH.Update)
		admin.DELETE("/categories/:id", souvenirCategoryH.Delete)

		// --- SOUVENIR: PRODUCT ---
		admin.POST("/products", souvenirProductH.CreateProduct)
		admin.GET("/products", souvenirProductH.GetAllProducts)
		admin.GET("/products/:id", souvenirProductH.GetProduct)
		admin.PUT("/products/:id", souvenirProductH.UpdateProduct)
		admin.DELETE("/products/:id", souvenirProductH.DeleteProduct)
		admin.GET("/products/category/:category_id", souvenirProductH.GetProductsByCategory)

		// --- BOOK: CATEGORY (ADMIN) ---
		admin.POST("/book-categories", bookCategoryH.Create)
		admin.PUT("/book-categories/:id", bookCategoryH.Update)
		admin.DELETE("/book-categories/:id", bookCategoryH.Delete)

		// --- BOOK: PRODUCT (ADMIN) ---
		admin.POST("/books", bookProductH.CreateProduct)
		admin.PUT("/books/:id", bookProductH.UpdateProduct)
		admin.DELETE("/books/:id", bookProductH.DeleteProduct)

		// === ADMIN API (tambah cafe) ===
		admin.POST("/cafe-categories", cafeCategoryH.Create)
		admin.PUT("/cafe-categories/:id", cafeCategoryH.Update)
		admin.DELETE("/cafe-categories/:id", cafeCategoryH.Delete)
		admin.GET("/cafe-categories", cafeCategoryH.GetAll)
	

		admin.POST("/cafe-products", cafeProductH.CreateProduct)
		admin.PUT("/cafe-products/:id", cafeProductH.UpdateProduct)
		admin.DELETE("/cafe-products/:id", cafeProductH.DeleteProduct)

				// TAMBAHKAN INI: Route untuk GET
		admin.GET("/cafe-products", cafeProductH.ListProducts)                    // List semua / filter by category
		admin.GET("/cafe-products/:id", cafeProductH.GetProduct)                  // Detail product
		admin.GET("/cafe-categories/:category_id/products", cafeProductH.GetProductsByCategory) 

	}
}
}