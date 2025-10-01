package middleware

import (
	"backend/internal/config"
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token diperlukan"})
			c.Abort()
		 return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Header Authorization tidak valid"})
			c.Abort()
			return
		}

		// buang spasi/kutip yang sering kebawa dari copy-paste
		tokenString := strings.Trim(parts[1], " \t\r\n\"")

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// pastikan HS256 family
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("metode penandatangan tidak didukung")
			}
			return []byte(config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			// Saat debug: boleh kirim detail error biar jelas (hapus di production)
			// c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid", "detail": err.Error(), "claims": claims})
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token tidak valid atau sudah kedaluwarsa"})
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Next()
	}
}
