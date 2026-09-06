package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that sets appropriate CORS headers based on allowed origins.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowedMap := make(map[string]bool)
	allowAll := false

	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAll = true
			break
		}
		allowedMap[strings.TrimRight(origin, "/")] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin != "" {
			trimmedOrigin := strings.TrimRight(origin, "/")
			if allowAll || allowedMap[trimmedOrigin] {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
				c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
				c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Disposition")
			}
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
