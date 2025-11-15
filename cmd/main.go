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
	"context"
	"log"
	"net/http"
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
		&auth.Admin{},&hotel.RoomType{},
		&hotel.Room{}, &hotel.Gallery{}, &hotel.News{}, &hotel.VisionMission{},
		&souvenir.Product{}, &souvenir.Category{},
		&book.ProductBook{}, &book.CategoryBook{},
		&cafe.ProductCafe{}, &cafe.CategoryCafe{},
		&hotel.GuestReview{}, &hotel.Booking{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	// Tambahkan kolom user_id
	if !db.Migrator().HasColumn(&hotel.Booking{}, "user_id") {
		db.Exec("ALTER TABLE bookings ADD COLUMN user_id BIGINT NOT NULL DEFAULT 1")
		db.Exec("ALTER TABLE bookings ADD CONSTRAINT fk_bookings_user FOREIGN KEY (user_id) REFERENCES admins(id)")
		log.Println("Kolom user_id ditambahkan ke tabel bookings")
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
	handler.SetupRoutes(r, adminService, db)

	// === STATIC ===
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}
	r.Static("/uploads", "./uploads")

	// === SERVER ===
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		log.Println("Server berjalan di http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// === BACKGROUND JOB ===
	go startAutoCheckout(db)

	// === GRACEFUL SHUTDOWN ===
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server gracefully stopped")
	}
}

// === BACKGROUND JOB: Auto Checkout & Check-in ===
func startAutoCheckout(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("Background job: Auto Checkout dimulai")

// cmd/main.go → di dalam startAutoCheckout
for {
	select {
	case <-ticker.C:
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		tx := db.Begin()
		if tx.Error != nil {
			log.Printf("AutoCheckout: gagal mulai transaksi: %v", tx.Error)
			continue
		}

		// CHECKOUT
		var expired []hotel.Booking
		if err := tx.Where("DATE(check_out) < ? AND status IN ?", today, []string{
			hotel.BookingStatusConfirmed.String(),
			hotel.BookingStatusCheckedIn.String(),
		}).Find(&expired).Error; err != nil {
			tx.Rollback()
			log.Printf("AutoCheckout: error query expired: %v", err)
			continue
		}

		for _, b := range expired {
			tx.Model(&b).Update("status", hotel.BookingStatusCheckedOut.String())
			var room hotel.Room
			if err := tx.First(&room, b.RoomID).Error; err == nil {
				tx.Model(&room).Update("status", hotel.RoomStatusAvailable.String())
			}
		}

		// CHECKIN
		var todayCheckins []hotel.Booking
		if err := tx.Where("DATE(check_in) = ? AND status = ?", today, hotel.BookingStatusConfirmed.String()).
			Find(&todayCheckins).Error; err != nil {
			tx.Rollback()
			log.Printf("AutoCheckin: error query today: %v", err)
			continue
		}

		for _, b := range todayCheckins {
			tx.Model(&b).Update("status", hotel.BookingStatusCheckedIn.String())
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("AutoCheckout: commit error: %v", err)
		}
	}
}
}

func seedSuperAdmin(db *gorm.DB) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("rahasia123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Gagal hash password: %v", err)
		return
	}

	super := auth.Admin{
		FullName:    "Super Admin",
		Email:       "supperpedrooo@gmail.com",
		PhoneNumber: "08123456789",
		Password:    string(hashed),
		Role:        auth.RoleSuperAdmin,
		IsApproved:  true,
	}

	var count int64
	if err := db.Model(&auth.Admin{}).Where("email = ?", super.Email).Count(&count).Error; err != nil {
		log.Printf("Gagal cek admin: %v", err)
		return
	}

	if count == 0 {
		if err := db.Create(&super).Error; err != nil {
			log.Printf("Gagal seed superadmin: %v", err)
		} else {
			log.Println("Superadmin berhasil dibuat: supperpedrooo@gmail.com")
		}
	}
}