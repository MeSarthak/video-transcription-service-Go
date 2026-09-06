package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger returns a Gin middleware that logs HTTP requests with structured slog.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		bytesOut := c.Writer.Size()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		fields := []any{
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", clientIP),
			slog.Int("bytes_out", bytesOut),
		}

		if rawQuery != "" {
			fields = append(fields, slog.String("query", rawQuery))
		}

		if errorMessage != "" {
			fields = append(fields, slog.String("error", errorMessage))
		}

		if status >= 500 {
			slog.Error("HTTP Request Error", fields...)
		} else if status >= 400 {
			slog.Warn("HTTP Client Request Warning", fields...)
		} else {
			slog.Info("HTTP Request", fields...)
		}
	}
}
