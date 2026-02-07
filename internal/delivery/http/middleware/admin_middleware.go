package middleware

import (
	"net/http"

	"ticres/pkg/logger"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			logger.Warn("middleware: admin check failed - no role in context",
				logger.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if userRole != "admin" {
			logger.Warn("middleware: admin access denied",
				logger.Any("role", userRole),
				logger.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		logger.Debug("middleware: admin access granted",
			logger.String("path", c.Request.URL.Path),
		)
		c.Next()
	}
}