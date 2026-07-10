package middlewares

import (
	"dfrecap/config"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWTClaims struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Mail     string `json:"mail"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		var tokenStr string
		if header == "" {
			cookie, err := c.Cookie("Auth")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token gerekli"})
				c.Abort()
				return
			}
			tokenStr = cookie
		} else {
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz token formatı"})
				c.Abort()
				return
			}
			tokenStr = parts[1]
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			config.Log.Error("JWT_SECRET tanımlı değil")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Sunucu yapılandırma hatası"})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("beklenmeyen imzalama metodu: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil {
			config.Log.Warn("Token doğrulanamadı", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz veya süresi dolmuş token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz token"})
			c.Abort()
			return
		}

		c.Set("id", claims.ID)
		c.Set("username", claims.Username)
		c.Set("mail", claims.Mail)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlem için sadece yöneticiler yetkilidir"})
			c.Abort()
			return
		}
		c.Next()
	}
}
