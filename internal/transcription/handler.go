package transcription

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-transcription-service/internal/middleware"
	"video-transcription-service/internal/videos"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	videoRoutes := rg.Group("/videos", authMiddleware)
	{
		videoRoutes.GET("/:id/transcription", h.GetTranscription)
		videoRoutes.GET("/:id/transcript/export", h.ExportTranscript)
	}
}

// GetTranscription returns the complete normalized transcript and timestamped subtitle segments.
func (h *Handler) GetTranscription(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	videoIDStr := c.Param("id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video ID format"})
		return
	}

	resp, err := h.service.GetTranscript(c.Request.Context(), userID, videoID)
	if err != nil {
		if errors.Is(err, videos.ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		if errors.Is(err, ErrTranscriptNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transcription is not ready or not found for this video"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve transcript"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ExportTranscript streams the generated .txt, .srt, or .vtt subtitle file.
func (h *Handler) ExportTranscript(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	videoIDStr := c.Param("id")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video ID format"})
		return
	}

	format := c.DefaultQuery("format", "txt")
	content, filename, contentType, err := h.service.ExportTranscript(c.Request.Context(), userID, videoID, format)
	if err != nil {
		if errors.Is(err, videos.ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		if errors.Is(err, ErrTranscriptNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transcript not found for this video"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, contentType, []byte(content))
}
