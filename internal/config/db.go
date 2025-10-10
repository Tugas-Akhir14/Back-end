package config

import (
	"log"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

// InitDB membaca .env via Viper, menentukan driver, dan membuka koneksi.
// Panggil sekali saat start (cmd/main.go).
func InitDB() *gorm.DB {
	once.Do(func() {
		cfg := LoadConfig()                // pastikan .env sudah dibaca
		driver := viper.GetString("DB_DRIVER")
		dsn := cfg.DatabaseURL

		var dial gorm.Dialector

		// Jika DB_DRIVER kosong, coba tebak dari DSN
		if driver == "" {
			low := strings.ToLower(dsn)
			switch {
			case strings.HasPrefix(low, "postgres://") || strings.HasPrefix(low, "postgresql://"):
				driver = "postgres"
			case strings.HasSuffix(low, ".db") || strings.HasPrefix(low, "sqlite://"):
				driver = "sqlite"
			default:
				driver = "mysql" // asumsi default
			}
		}

		switch driver {
		
	
		case "mysql":
			dial = mysql.Open(dsn)
		default:
			log.Fatalf("unknown DB_DRIVER: %s", driver)
		}

		var err error
		db, err = gorm.Open(dial, &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
	})

	return db
}

// GetDB mengembalikan instance *gorm.DB; auto-init kalau belum.
func GetDB() *gorm.DB {
	if db == nil {
		return InitDB()
	}
	return db
}
