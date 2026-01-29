package config

import (
	"log"
	"os"
	"sync"
	"strings"
	"net/url"    
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

func LoadConfig() {
	once.Do(func() {

		instance = &Config{
			DatabaseURL: os.Getenv("MYSQL_URL"),
			JWTSecret:   os.Getenv("JWT_SECRET"),
			SMTPHost:    os.Getenv("SMTP_HOST"),
			SMTPPort:    os.Getenv("SMTP_PORT"),
			SMTPUser:    os.Getenv("SMTP_USER"),
			SMTPPass:    os.Getenv("SMTP_PASS"),
		}

		if instance.DatabaseURL == "" {
			log.Fatal("MYSQL_URL belum diset di environment")
		}

		log.Println("Config loaded successfully")
		log.Printf("SMTP Config: %s:%s → %s", instance.SMTPHost, instance.SMTPPort, instance.SMTPUser)
	})
}

func GetConfig() *Config {
	if instance == nil {
		log.Fatal("Config belum di-load")
	}
	return instance
}

func InitDB() *gorm.DB {
	dbOnce.Do(func() {

		rawURL := GetConfig().DatabaseURL
		if rawURL == "" {
			log.Fatal("MYSQL_URL kosong")
		}

		u, err := url.Parse(rawURL)
		if err != nil {
			log.Fatalf("Invalid MYSQL_URL: %v", err)
		}

		user := u.User.Username()
		pass, _ := u.User.Password()
		host := u.Host
		dbname := strings.TrimPrefix(u.Path, "/")

		dsn := user + ":" + pass + "@tcp(" + host + ")/" + dbname + "?parseTime=true"

		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Gagal konek database: %v", err)
		}

		log.Println("Database connected!")
	})

	return db
}



func GetDB() *gorm.DB {
	if db == nil {
		return InitDB()
	}
	return db
}



