// cmd/main.go
package main

import (
	_ "backend/utils"
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/models/auth"
	"backend/internal/models/book"
	"backend/internal/models/cafe"
	"backend/internal/models/hotel"
	"backend/internal/models/souvenir"
	"backend/internal/repository/admin"
	"backend/internal/service/serviceauth"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	config.LoadConfig()
	db := config.InitDB()

	// === MIGRASI & SEED ===
	if err := db.AutoMigrate(
		&auth.Admin{},
		&hotel.Room{}, &hotel.Gallery{}, &hotel.News{}, &hotel.VisionMission{},
		&souvenir.Product{}, &souvenir.Category{},
		&book.ProductBook{}, &book.CategoryBook{},
		&cafe.ProductCafe{}, &cafe.CategoryCafe{},
		&hotel.GuestReview{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	seedSuperAdmin(db)

	// === REPO & SERVICE ===
	adminRepo := admin.NewAdminRepository(db)
	adminService := serviceauth.NewAdminService(adminRepo, config.GetConfig().JWTSecret)

	// === GIN SETUP ===
	r := gin.Default()
	corsCfg := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsCfg))
	r.MaxMultipartMemory = 8 << 20

	// === ROUTES ===
	handler.SetupRoutes(r, adminService)

	// === STATIC ===
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}
	r.Static("/uploads", "./uploads")

	// === SERVER ===
	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("Failed to run server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
}

func seedSuperAdmin(db *gorm.DB) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("rahasia123"), bcrypt.DefaultCost)
	super := auth.Admin{
		FullName:    "Super Admin",
		Email:       "supperpedrooo@gmail.com",
		PhoneNumber: "08123456789",
		Password:    string(hashed),
		Role:        auth.RoleSuperAdmin,
		IsApproved:  true,
	}

	var count int64
	db.Model(&auth.Admin{}).Where("email = ?", super.Email).Count(&count)
	if count == 0 {
		if err := db.Create(&super).Error; err != nil {
			log.Printf("Gagal seed superadmin: %v", err)
		} else {
			log.Println("Superadmin berhasil dibuat")
		}
	}
}