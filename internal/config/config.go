// internal/config/config.go
package config

import (
	"log"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
}

var (
	cfg  Config
	db   *gorm.DB
	once sync.Once
)

// LoadConfig membaca .env
func LoadConfig() Config {
	viper.SetConfigFile(".env")
	if err := viper.ReadInConfig(); err != nil {
		panic("Error reading config file: " + err.Error())
	}

	cfg = Config{
		DatabaseURL: viper.GetString("DATABASE_URL"),
		JWTSecret:   viper.GetString("JWT_SECRET"),
	}
	return cfg
}

// GetConfig mengembalikan config yang sudah di-load
func GetConfig() Config {
	return cfg
}

// InitDB inisialisasi DB sekali
func InitDB() *gorm.DB {
	once.Do(func() {
		dsn := cfg.DatabaseURL
		var dial gorm.Dialector

		low := strings.ToLower(dsn)
		switch {
		case strings.HasPrefix(low, "postgres://") || strings.HasPrefix(low, "postgresql://"):
			log.Fatal("PostgreSQL belum didukung")
		case strings.HasSuffix(low, ".db") || strings.HasPrefix(low, "sqlite://"):
			log.Fatal("SQLite belum didukung")
		default:
			dial = mysql.Open(dsn)
		}

		var err error
		db, err = gorm.Open(dial, &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
		}
	})
	return db
}

// GetDB mengembalikan instance DB
func GetDB() *gorm.DB {
	if db == nil {
		return InitDB()
	}
	return db
}