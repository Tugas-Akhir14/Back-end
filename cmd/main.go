// cmd/main.go
package main

import (
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
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()
	db, err := gorm.Open(mysql.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	// === MIGRASE SOUVENIR DULU (karena ada seeder) ===
	if err := db.AutoMigrate(&souvenir.Category{}); err != nil {
		log.Fatalf("Migrate Category failed: %v", err)
	}

	// === SEEDER: Pastikan ada kategori default ===
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

	// === UPDATE PRODUCT YANG CATEGORY_ID INVALID ===
	db.Exec(`
		UPDATE products 
		SET category_id = 1 
		WHERE category_id IS NULL 
		   OR category_id NOT IN (SELECT id FROM categories)
	`)

	// === MIGRASE SEMUA MODEL SETELAH CATEGORY ===
	if err := db.AutoMigrate(
		&auth.Admin{},
		&hotel.Room{},
		&hotel.Gallery{},
		&hotel.News{},
		&hotel.VisionMission{},
		&souvenir.Product{},
		&book.CategoryBook{},   // ← DIPINDAH KE SINI
		&book.ProductBook{},
		&cafe.CategoryCafe{},
		&cafe.ProductCafe{},  // ← DIPINDAH KE SINI
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

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
	adminService := serviceauth.NewAdminService(adminRepo, cfg.JWTSecret)

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

	// === SETUP ROUTES (termasuk public & admin) ===
	handler.SetupRoutes(r, adminService)

	// === UPLOADS FOLDER ===
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// === RUN SERVER ===
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