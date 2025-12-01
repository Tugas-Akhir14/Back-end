// internal/config/config.go
package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	SMTPHost    string
	SMTPPort    string
	SMTPUser    string
	SMTPPass    string
}

var (
	instance *Config
	db       *gorm.DB
	once     sync.Once
	dbOnce   sync.Once
)

// LoadConfig — panggil sekali dari main.go
func LoadConfig() {
	once.Do(func() {
		viper.SetConfigFile(".env")
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading .env file: %v", err)
		}

		instance = &Config{
			DatabaseURL: viper.GetString("DATABASE_URL"),
			JWTSecret:   viper.GetString("JWT_SECRET"),
			SMTPHost:    viper.GetString("SMTP_HOST"),
			SMTPPort:    viper.GetString("SMTP_PORT"),
			SMTPUser:    viper.GetString("SMTP_USER"),
			SMTPPass:    viper.GetString("SMTP_PASS"),
		}

		log.Println("Config loaded successfully")
		log.Printf("SMTP Config: %s:%s → %s", instance.SMTPHost, instance.SMTPPort, instance.SMTPUser)
	})
}

// GetConfig — aman dipanggil kapan saja setelah LoadConfig()
func GetConfig() *Config {
	if instance == nil {
		log.Fatal("Config belum di-load! Panggil config.LoadConfig() dulu")
	}
	return instance
}

// InitDB — inisialisasi DB sekali
func InitDB() *gorm.DB {
	dbOnce.Do(func() {
		dsn := GetConfig().DatabaseURL
		var err error
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Gagal konek database: %v", err)
		}
		log.Println("Database connected!")
	})
	return db
}

// GetDB — untuk dipakai di handler kalau perlu
func GetDB() *gorm.DB {
	if db == nil {
		return InitDB()
	}
	return db
}