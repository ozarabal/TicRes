package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware mengembalikan Gin HandlerFunc
// Kita butuh jwtSecret untuk memvalidasi tanda tangan token
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Ambil Header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort() // Stop! Jangan lanjut ke handler berikutnya
			return
		}

		// 2. Cek Format (Harus "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. Parse & Validasi Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validasi algoritma signing (Wajib!)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 4. Ambil Claims (Data dalam token)
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Ambil user_id (perhatikan tipe datanya, jwt biasanya menaruh angka sebagai float64)
			userID := claims["user_id"]
			
			// 5. Simpan ke Context Gin
			// Ini kuncinya! Handler selanjutnya bisa akses user_id lewat c.Get("userID")
			c.Set("userID", userID)
			
			c.Next() // Lanjut ke handler berikutnya (misal: Booking Ticket)
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
		}
	}
}