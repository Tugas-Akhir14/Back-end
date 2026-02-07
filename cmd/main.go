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
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	
	if os.Getenv("RAILWAY_ENVIRONMENT_NAME") == "" {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env tidak ditemukan")
	} else {
		log.Println("✅ .env berhasil dimuat")
	}
}


	config.LoadConfig()


	db := config.InitDB()

	// === MIGRASI & SEED ===
	if err := db.AutoMigrate(
		&auth.Admin{},&hotel.RoomType{},
		&hotel.Room{}, &hotel.Gallery{}, &hotel.News{}, &hotel.VisionMission{},
		&souvenir.Product{}, &souvenir.Category{},&hotel.GuestReview{},
		&book.ProductBook{}, &book.CategoryBook{},
		&cafe.ProductCafe{}, &cafe.CategoryCafe{},&cafe.OrderCafe{},&cafe.OrderItemCafe{},
		&hotel.GuestReview{}, &hotel.Booking{}, &auth.GuestOTP{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	// Tambahkan kolom user_id
	if !db.Migrator().HasColumn(&hotel.Booking{}, "user_id") {
		db.Exec("ALTER TABLE bookings ADD COLUMN user_id BIGINT NOT NULL DEFAULT 1")
		db.Exec("ALTER TABLE bookings ADD CONSTRAINT fk_bookings_user FOREIGN KEY (user_id) REFERENCES admins(id)")
		log.Println("Kolom user_id ditambahkan ke tabel bookings")
	}

	// Tambahkan kolom source dan ota_reference jika belum ada
	if !db.Migrator().HasColumn(&hotel.Booking{}, "source") {
		db.Exec("ALTER TABLE bookings ADD COLUMN source VARCHAR(20) NOT NULL DEFAULT 'web'")
		log.Println("Kolom source ditambahkan ke tabel bookings")
	}
	if !db.Migrator().HasColumn(&hotel.Booking{}, "ota_reference") {
		db.Exec("ALTER TABLE bookings ADD COLUMN ota_reference VARCHAR(100)")
		log.Println("Kolom ota_reference ditambahkan ke tabel bookings")
	}

	seedSuperAdmin(db)

	// === REPO & SERVICE ===
	adminRepo := admin.NewAdminRepository(db)
	adminService := serviceauth.NewAdminService(adminRepo, db, config.GetConfig().JWTSecret)

	// === GIN SETUP ===
	r := gin.Default()

	corsCfg := cors.Config{
		AllowOrigins:     []string{"https://frontend-c2mq3w3fz-pedros-projects-91760395.vercel.app"},
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
port := os.Getenv("PORT")
if port == "" {
	port = "8080"
}

srv := &http.Server{
	Addr:    ":" + port,
	Handler: r,
}


	go func() {
		log.Println("Server berjalan di http://localhost:" + port)
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

// GANTI SELURUH FUNGSI startAutoCheckout JADI INI:
func startAutoCheckout(db *gorm.DB) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    log.Println("Background job: Auto Checkout & Check-in dimulai")

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

            // CHECKOUT otomatis
            var expired []hotel.Booking
            if err := tx.Where("DATE(check_out) < ? AND status IN ?", today, []string{"confirmed", "checked_in"}).
                Find(&expired).Error; err != nil {
                tx.Rollback()
                log.Printf("AutoCheckout: error query expired: %v", err)
                continue
            }

            for _, b := range expired {
                tx.Model(&b).Update("status", "checked_out")
                var room hotel.Room
                if err := tx.First(&room, b.RoomID).Error; err == nil {
                    tx.Model(&room).Update("status", "available")
                }
            }

            // CHECKIN otomatis
            var todayCheckins []hotel.Booking
            if err := tx.Where("DATE(check_in) = ? AND status = ?", today, "confirmed").
                Find(&todayCheckins).Error; err != nil {
                tx.Rollback()
                log.Printf("AutoCheckin: error query today: %v", err)
                continue
            }

            for _, b := range todayCheckins {
                tx.Model(&b).Update("status", "checked_in")
            }

            if err := tx.Commit().Error; err != nil {
                log.Printf("AutoCheckout: commit error: %v", err)
            } else {
                log.Printf("Auto job selesai: %d checkout, %d checkin", len(expired), len(todayCheckins))
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
		Email:       "superadmin@gmail.com",
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
			log.Println("Superadmin berhasil dibuat: superadmin@gmail.com")
		}
	}

}	


