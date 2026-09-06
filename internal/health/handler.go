package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	startTime time.Time
	// readiness checkers can be attached as we add DB, S3, SQS in upcoming phases
}

func NewHandler() *Handler {
	return &Handler{
		startTime: time.Now(),
	}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	router.GET("/health", h.Health)
	router.GET("/ready", h.Ready)
}

// Health handles liveness probe (used by ALB / ECS to check container process responsiveness)
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"uptime":    time.Since(h.startTime).String(),
	})
}

// Ready handles readiness probe (verifies database, queues, and dependencies)
func (h *Handler) Ready(c *gin.Context) {
	// In Phase 2+, we will add DB ping check here
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": gin.H{
			"api": "ok",
		},
	})
}
