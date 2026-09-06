package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Checker represents a health check function for a specific subsystem.
type Checker func(ctx context.Context) error

type Handler struct {
	startTime time.Time
	checkers  map[string]Checker
}

func NewHandler(checkers map[string]Checker) *Handler {
	if checkers == nil {
		checkers = make(map[string]Checker)
	}
	return &Handler{
		startTime: time.Now(),
		checkers:  checkers,
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.Health)
	router.GET("/ready", h.Ready)
}

// Health handles liveness probe (checks process responsiveness)
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(h.startTime).String(),
	})
}

// Ready handles readiness probe (verifies database, queues, and dependencies)
func (h *Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	checks := make(map[string]string)
	checks["api"] = "ok"

	allHealthy := true
	for name, checker := range h.checkers {
		if err := checker(ctx); err != nil {
			checks[name] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			checks[name] = "ok"
		}
	}

	statusCode := http.StatusOK
	statusText := "ready"
	if !allHealthy {
		statusCode = http.StatusServiceUnavailable
		statusText = "unhealthy"
	}

	c.JSON(statusCode, gin.H{
		"status":    statusText,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(h.startTime).String(),
		"checks":    checks,
	})
}
