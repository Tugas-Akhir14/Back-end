package config

import "os"

var (
	// pastikan di .env ada JWT_SECRET yang sama untuk semua
	JWTSecret = os.Getenv("JWT_SECRET")
)

// opsional: fallback dev
func init() {
	if JWTSecret == "" {
		JWTSecret = "dev_only_change_me"
	}
}
