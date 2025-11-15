// internal/handler/routes.go
package handler

import (
	"backend/internal/models/auth"
	"backend/internal/repository/repohotel"
	"backend/internal/repository/reposouvenir"
	"backend/internal/service/serviceauth"
	"backend/internal/service/hotelservice"
	"backend/internal/service/souvenirservice"
	"backend/middleware"
	"net/http"

	"github.com/gin-gonic/gin"

	"backend/internal/handler/authhandler"
	"backend/internal/handler/hotel"
	"backend/internal/handler/souvenirhandler"
	"backend/internal/handler/bookhandler"
	"backend/internal/service/bookservice"
	"backend/internal/repository/repobook"

	"backend/internal/handler/cafehandler"
	"backend/internal/service/cafeservice"
	"backend/internal/repository/repocafe"

	"backend/internal/repository/admin"
	"gorm.io/gorm"
)

// TAMBAH PARAMETER db
func SetupRoutes(r *gin.Engine, adminService serviceauth.AdminService, db *gorm.DB) {
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
	})
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// === AUTH ===
	adm := authhandler.NewAdminHandler(adminService)
	r.POST("/admins/register", adm.Register)
	r.POST("/admins/login", adm.Login)
	r.GET("/admins/profile", middleware.AuthMiddleware(), adm.GetProfile)

	// INISIALISASI ADMIN REPOSITORY
	adminRepo := admin.NewAdminRepository(db)

	// === REPO & HANDLER ===
	galleryH := hotel.NewGalleryHandler(hotelservice.NewGalleryService(repohotel.NewGalleryRepository(db)))
	newsH := hotel.NewNewsHandler(hotelservice.NewNewsService(repohotel.NewNewsRepository(db)))
	visionMissionH := hotel.NewVisionMissionHandler(hotelservice.NewVisionMissionService(repohotel.NewVisionMissionRepository(db)))

	souvenirProductH := souvenirhandler.NewProductHandler(souvenirservice.NewProductService(reposouvenir.NewProductRepository(db)))
	souvenirCategoryH := souvenirhandler.NewCategoryHandler(souvenirservice.NewCategoryService(reposouvenir.NewCategoryRepository(db)))

	// BOOK
	bookProductRepo := repobook.NewProductRepository(db)
	bookCategoryRepo := repobook.NewCategoryRepository(db)
	bookProductService := bookservice.NewProductService(bookProductRepo, bookCategoryRepo)
	bookProductH := bookhandler.NewProductHandler(bookProductService, bookservice.NewCategoryService(bookCategoryRepo))

	// CAFE
	cafeProductRepo := repocafe.NewProductRepository(db)
	cafeCategoryRepo := repocafe.NewCategoryRepository(db)
	cafeProductService := cafeservice.NewProductService(cafeProductRepo, cafeCategoryRepo)
	cafeProductH := cafehandler.NewProductHandler(cafeProductService, cafeservice.NewCategoryService(cafeCategoryRepo))

	bookCategoryH := bookhandler.NewCategoryHandler(bookservice.NewCategoryService(bookCategoryRepo))
	cafeCategoryH := cafehandler.NewCategoryHandler(cafeservice.NewCategoryService(cafeCategoryRepo))

	// REVIEW HOTEL
	reviewRepo := repohotel.NewReviewRepository(db)
	reviewService := hotelservice.NewReviewService(reviewRepo, adminRepo)
	reviewH := hotel.NewReviewHandler(reviewService)

	// BOOKING → PASS db
	bookingRepo := repohotel.NewBookingRepository(db)
	bookingService := hotelservice.NewBookingService(bookingRepo, repohotel.NewRoomRepository(db), db)
	bookingH := hotel.NewBookingHandler(bookingService)


		// Di dalam SetupRoutes
	roomRepo := repohotel.NewRoomRepository(db)
	roomTypeRepo := repohotel.NewRoomTypeRepository(db)

	roomService := hotelservice.NewRoomService(roomRepo, roomTypeRepo, db)
	roomTypeService := hotelservice.NewRoomTypeService(roomTypeRepo)

	roomH := hotel.NewRoomHandler(roomService)
	roomTypeH := hotel.NewRoomTypeHandler(roomTypeService)
	// === PUBLIC API ===
	public := r.Group("/public")
	{
		public.GET("/rooms", roomH.ListPublic)
		public.GET("/gallery", galleryH.ListPublic)
		public.GET("/gallery/:id", galleryH.GetByID)
		public.GET("/rooms/:id/gallery", galleryH.ListByRoom)
		public.GET("/reviews", reviewH.GetApproved)
		public.GET("/news", newsH.ListPublic)
		public.GET("/news/:id", newsH.GetPublicByID)
		public.GET("/news/slug/:slug", newsH.GetPublicBySlug)
		public.GET("/visi-misi", visionMissionH.GetPublic)
		public.POST("/reviews", middleware.AuthMiddleware(), middleware.RoleMiddleware(auth.RoleGuest), reviewH.Create)

		// FITUR BARU
	

		public.GET("/souvenirs", souvenirProductH.ListPublic)
		public.GET("/souvenirs/:id", souvenirProductH.GetPublicByID)
		public.GET("/souvenirs/category/:category_id", souvenirProductH.GetPublicByCategory)

		public.GET("/books", bookProductH.ListPublic)
		public.GET("/books/:id", bookProductH.GetPublicByID)
		public.GET("/books/category/:category_id", bookProductH.GetPublicByCategory)

		public.GET("/cafe", cafeProductH.ListPublic)
		public.GET("/cafe/:id", cafeProductH.GetPublicByID)
		public.GET("/cafe/category/:category_id", cafeProductH.GetPublicByCategory)

		public.POST("/bookings", middleware.AuthMiddleware(), middleware.RoleMiddleware(auth.RoleGuest), bookingH.Create)
		public.POST("/guest-bookings", middleware.AuthMiddleware(), middleware.RoleMiddleware(auth.RoleGuest), bookingH.GuestBook)
		public.GET("/availability", bookingH.CheckAvailability)
	}

	// === ADMIN API ===
	adminGroup := r.Group("/api", middleware.AuthMiddleware())

	// SUPERADMIN
	super := adminGroup.Group("", middleware.RoleMiddleware(auth.RoleSuperAdmin))
	{
		super.GET("/pending-admins", adm.GetPending)
		super.PATCH("/admins/approve/:id", adm.ApproveUser)
	}

	// HOTEL
	hotelGroup := adminGroup.Group("", middleware.RoleMiddleware(auth.RoleAdminHotel, auth.RoleSuperAdmin))
	{
			// Room CRUD
		hotelGroup.POST("/rooms", roomH.Create)
		hotelGroup.GET("/rooms", roomH.List)
		hotelGroup.GET("/rooms/:id", roomH.GetByID)
		hotelGroup.PUT("/rooms/:id", roomH.Update)
		hotelGroup.DELETE("/rooms/:id", roomH.Delete)

		// RoomType CRUD
		hotelGroup.POST("/room-types", roomTypeH.Create)
		hotelGroup.GET("/room-types", roomTypeH.List)
		hotelGroup.GET("/room-types/:id", roomTypeH.GetByID)
		hotelGroup.PUT("/room-types/:id", roomTypeH.Update)
		hotelGroup.DELETE("/room-types/:id", roomTypeH.Delete)

		hotelGroup.POST("/galleries", galleryH.Create)
		hotelGroup.GET("/galleries", galleryH.List)
		hotelGroup.GET("/galleries/:id", galleryH.GetByID)
		hotelGroup.PUT("/galleries/:id", galleryH.Update)
		hotelGroup.PUT("/galleries/:id/image", galleryH.UpdateImage)
		hotelGroup.DELETE("/galleries/:id", galleryH.Delete)

		hotelGroup.GET("/news", newsH.List)
		hotelGroup.GET("/news/:id", newsH.GetByID)
		hotelGroup.GET("/news/slug/:slug", newsH.GetBySlug)
		hotelGroup.POST("/news", newsH.Create)
		hotelGroup.PUT("/news/:id", newsH.Update)
		hotelGroup.DELETE("/news/:id", newsH.Delete)

		hotelGroup.GET("/visi-misi", visionMissionH.Get)
		hotelGroup.PUT("/visi-misi", visionMissionH.Upsert)

		hotelGroup.GET("/reviews/pending", reviewH.GetPending)
		hotelGroup.PUT("/reviews/:id/approve", reviewH.Approve)
		hotelGroup.DELETE("/reviews/:id", reviewH.Delete)
		hotelGroup.POST("/reviews", reviewH.Create)

		// Admin
		hotelGroup.GET("/bookings", bookingH.List)
		hotelGroup.PATCH("/bookings/:id/confirm", bookingH.Confirm)
		hotelGroup.PATCH("/bookings/:id/cancel", bookingH.Cancel)
	}

	
	// SOUVENIR
	souvenirGroup := adminGroup.Group("", middleware.RoleMiddleware(auth.RoleAdminSouvenir, auth.RoleSuperAdmin))
	{
		souvenirGroup.POST("/categories", souvenirCategoryH.Create)
		souvenirGroup.GET("/categories", souvenirCategoryH.GetAll)
		souvenirGroup.GET("/categories/:id", souvenirCategoryH.GetByID)
		souvenirGroup.PUT("/categories/:id", souvenirCategoryH.Update)
		souvenirGroup.DELETE("/categories/:id", souvenirCategoryH.Delete)

		souvenirGroup.POST("/products", souvenirProductH.CreateProduct)
		souvenirGroup.GET("/products", souvenirProductH.GetAllProducts)
		souvenirGroup.GET("/products/:id", souvenirProductH.GetProduct)
		souvenirGroup.PUT("/products/:id", souvenirProductH.UpdateProduct)
		souvenirGroup.DELETE("/products/:id", souvenirProductH.DeleteProduct)
		souvenirGroup.GET("/products/category/:category_id", souvenirProductH.GetProductsByCategory)
	}

	// BOOK
	bookGroup := adminGroup.Group("", middleware.RoleMiddleware(auth.RoleAdminBuku, auth.RoleSuperAdmin))
	{
		bookGroup.POST("/book-categories", bookCategoryH.Create)
		bookGroup.GET("/book-categories", bookCategoryH.GetAll)
		bookGroup.GET("/book-categories/:id", bookCategoryH.GetByID)
		bookGroup.PUT("/book-categories/:id", bookCategoryH.Update)
		bookGroup.DELETE("/book-categories/:id", bookCategoryH.Delete)

		bookGroup.POST("/books", bookProductH.CreateProduct)
		bookGroup.GET("/books", bookProductH.ListBooks)
		bookGroup.PUT("/books/:id", bookProductH.UpdateProduct)
		bookGroup.DELETE("/books/:id", bookProductH.DeleteProduct)
	}

	// CAFE
	cafeGroup := adminGroup.Group("", middleware.RoleMiddleware(auth.RoleAdminCafe, auth.RoleSuperAdmin))
	{
		cafeGroup.POST("/cafe-categories", cafeCategoryH.Create)
		cafeGroup.PUT("/cafe-categories/:id", cafeCategoryH.Update)
		cafeGroup.DELETE("/cafe-categories/:id", cafeCategoryH.Delete)
		cafeGroup.GET("/cafe-categories", cafeCategoryH.GetAll)

		cafeGroup.POST("/cafe-products", cafeProductH.CreateProduct)
		cafeGroup.PUT("/cafe-products/:id", cafeProductH.UpdateProduct)
		cafeGroup.DELETE("/cafe-products/:id", cafeProductH.DeleteProduct)
		cafeGroup.GET("/cafe-products", cafeProductH.ListProducts)
		cafeGroup.GET("/cafe-products/:id", cafeProductH.GetProduct)
		cafeGroup.GET("/cafe-categories/:category_id/products", cafeProductH.GetProductsByCategory)
	}
}