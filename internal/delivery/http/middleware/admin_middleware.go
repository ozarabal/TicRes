package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


func AdminMiddleware(jwtSecret string) gin.HandlerFunc{
	return func(c *gin.Context){
		
		// mengambil context
		userRole, exists := c.Get("role")
		if !exists{
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// cek userrole apakah admin?
		if userRole != "admin"{
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		
		c.Next()
	}
}