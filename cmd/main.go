package main

import (
    "backend/internal/config"
    "backend/internal/handler"
    "backend/internal/models"
    "backend/internal/repository"
    "backend/internal/service"
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "log"
    "os"
    "syscall"
    "os/signal"
)

func main() {

    
    // Load konfigurasi
    cfg := config.LoadConfig()
    log.Printf("DATABASE_URL: %s", cfg.DatabaseURL)

    // Koneksi database MySQL
    db, err := gorm.Open(mysql.Open(cfg.DatabaseURL), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect database: %v", err)
    }

    // Auto-migrate tabel customers
    db.AutoMigrate(&models.Admin{})

    // Inisialisasi repository dan service
    adminRepo := repository.NewAdminRepository(db)
    adminService := service.NewAdminService(adminRepo, cfg.JWTSecret)

    // Setup Gin router
    r := gin.Default()

    // Tambahkan middleware CORS
    r.Use(cors.Default())

    // Set batas ukuran file upload
    r.MaxMultipartMemory = 8 << 20 // 8 MiB
    handler.SetupRoutes(r, adminService)

    // Buat folder uploads
    if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
        log.Fatalf("Failed to create uploads directory: %v", err)
    }

    log.Println("========================================")
    log.Println("🚀 Server running at http://localhost:8080")
    log.Println("Tekan CTRL+C untuk menghentikan server")
    log.Println("========================================")

    log.Printf("Connected to database: %s", cfg.DatabaseURL)


    // Setup graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        if err := r.Run(":8080"); err != nil {
            log.Fatalf("❌ Failed to run server: %v", err)
        }
    }()

    <-quit
    log.Println("Shutting down server...")
}
