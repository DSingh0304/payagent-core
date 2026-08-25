package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AgentAuth validates the X-API-Key header to secure internal agent endpoints
func AgentAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED"})
			c.Abort()
			return
		}
		c.Next()
	}
}
