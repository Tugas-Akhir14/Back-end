package main

import (
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/models"
	"backend/internal/repository"
	"backend/internal/service"
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

    // 1) Migrate dulu
    if err := db.AutoMigrate(&models.Admin{}, &models.Room{}, &models.Gallery{}, &models.News{}, &models.VisionMission{}); err != nil {
        log.Fatalf("AutoMigrate failed: %v", err)
    }

    // 2) Drop index lama, buat index baru (composite)
    m := db.Migrator()

    // Kadang nama index unik lama beda-beda. Coba drop yang umum:
    oldIdx := []string{
        "idx_rooms_number",       // yang muncul di error kamu
        "uix_rooms_number",
        "rooms_number_unique",
        "rooms_number_key",
        "Number",                 // GORM bisa simpan index berdasar field
    }
    for _, name := range oldIdx {
        if m.HasIndex(&models.Room{}, name) {
            _ = m.DropIndex(&models.Room{}, name)
        }
    }

    // Buat composite unique index sesuai tag di struct
    // (harus ada tag `uniqueIndex:ux_room_number_deleted_at` di Number & DeletedAt)
    if !m.HasIndex(&models.Room{}, "ux_room_number_deleted_at") {
        if err := m.CreateIndex(&models.Room{}, "ux_room_number_deleted_at"); err != nil {
            log.Printf("Create composite index failed: %v", err)
        }
    }
    
    // --- sisa kode kamu (wiring, router, dll) ---
    adminRepo := repository.NewAdminRepository(db)
    adminService := service.NewAdminService(adminRepo, cfg.JWTSecret)

    r := gin.Default()

    // CORS: izinkan hanya frontend-mu
    corsCfg := cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true, // <- penting jika pakai cookie/credentials
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
