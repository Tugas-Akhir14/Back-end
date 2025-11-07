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
		if !strings.HasPrefix(authHeader, "Bearer ") {
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

func RoleMiddleware(allowed ...auth.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleStr := c.GetString("role")
		role := auth.Role(roleStr)

		for _, r := range allowed {
			if role == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "akses ditolak: role tidak diizinkan"})
		c.Abort()
	}
}

// Optional: Helper untuk cek role di handler
func GetUserRole(c *gin.Context) auth.Role {
	roleStr, _ := c.Get("role")
	return auth.Role(roleStr.(string))
}	