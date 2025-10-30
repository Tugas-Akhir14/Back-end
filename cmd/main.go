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
	// Load config & inisialisasi DB sekali
	config.LoadConfig()
	db := config.InitDB() // <-- PAKAI INI, BUKAN gorm.Open

	// === MIGRASE SOUVENIR DULU ===
	if err := db.AutoMigrate(&souvenir.Category{}); err != nil {
		log.Fatalf("Migrate Category failed: %v", err)
	}

	// === SEEDER KATEGORI ===
	var count int64
	db.Model(&souvenir.Category{}).Count(&count)
	if count == 0 {
		defaultCats := []souvenir.Category{
			{Nama: "Uncategorized", Slug: "uncategorized"},
			{Nama: "Kaos", Slug: "kaos"},
			{Nama: "Aksesoris", Slug: "aksesoris"},
		}
		for _, c := range defaultCats {
			db.Create(&c)
		}
		log.Println("Seeder: Default categories created")
	}

	// === UPDATE PRODUCT INVALID ===
	db.Exec(`
		UPDATE products 
		SET category_id = 1 
		WHERE category_id IS NULL 
		   OR category_id NOT IN (SELECT id FROM categories)
	`)

	// === MIGRASE SEMUA MODEL ===
	if err := db.AutoMigrate(
		&auth.Admin{},
		&hotel.Room{},
		&hotel.Gallery{},
		&hotel.News{},
		&hotel.VisionMission{},
		&souvenir.Product{},
		&book.CategoryBook{},
		&book.ProductBook{},
		&cafe.CategoryCafe{},
		&cafe.ProductCafe{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	// === SEED SUPERADMIN ===
	seedSuperAdmin(db)

	// === INDEX HOTEL ===
	m := db.Migrator()
	oldIdx := []string{"idx_rooms_number", "uix_rooms_number", "rooms_number_unique", "Number"}
	for _, name := range oldIdx {
		if m.HasIndex(&hotel.Room{}, name) {
			_ = m.DropIndex(&hotel.Room{}, name)
		}
	}
	if !m.HasIndex(&hotel.Room{}, "ux_room_number_deleted_at") {
		_ = m.CreateIndex(&hotel.Room{}, "ux_room_number_deleted_at")
	}

	// === WIRING ===
	adminRepo := admin.NewAdminRepository(db)
	adminService := serviceauth.NewAdminService(adminRepo, config.GetConfig().JWTSecret)

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

	handler.SetupRoutes(r, adminService)

	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

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

// === SEEDER SUPERADMIN ===
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