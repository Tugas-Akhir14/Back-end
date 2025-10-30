// middleware/middleware.go
package middleware

import (
	"backend/internal/config"
	"backend/internal/models/auth"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token diperlukan"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.GetConfig().JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid"})
			c.Abort()
			return
		}

		userID := uint(claims["user_id"].(float64))
		role := auth.Role(claims["role"].(string))

		c.Set("user_id", userID)
		c.Set("role", string(role))

		// Cek approval untuk admin_*
		if strings.HasPrefix(string(role), "admin_") && role != auth.RoleSuperAdmin {
			var admin auth.Admin
			if err := config.GetDB().Select("is_approved").Where("id = ?", userID).First(&admin).Error; err != nil || !admin.IsApproved {
				c.JSON(http.StatusForbidden, gin.H{"error": "akun belum disetujui"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}