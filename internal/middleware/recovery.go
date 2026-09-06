package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from any panics, logs the stack trace with slog, and writes a 500 JSON error.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				slog.Error("Panic recovered in HTTP handler",
					slog.Any("error", err),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
					slog.String("stack", stack),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "internal server error",
					"message": fmt.Sprintf("%v", err),
				})
			}
		}()
		c.Next()
	}
}
