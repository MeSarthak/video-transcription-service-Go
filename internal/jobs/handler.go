package jobs

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-transcription-service/internal/middleware"
	"video-transcription-service/internal/videos"
)

type TranscribeRequest struct {
	Language string `json:"language"`
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	videoRoutes := rg.Group("/videos", authMiddleware)
	{
		videoRoutes.POST("/:id/transcribe", h.TriggerTranscription)
		videoRoutes.GET("/:id/transcription/status", h.GetJobStatus)
	}
}

// TriggerTranscription initiates an asynchronous transcription job.
func (h *Handler) TriggerTranscription(c *gin.Context) {
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

	var req TranscribeRequest
	_ = c.ShouldBindJSON(&req)

	resp, err := h.service.TriggerTranscription(c.Request.Context(), userID, videoID, req.Language)
	if err != nil {
		if errors.Is(err, videos.ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		if errors.Is(err, ErrVideoStillUploading) || errors.Is(err, ErrVideoAlreadyProcessing) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, resp)
}

// GetJobStatus returns the current processing status of a video's transcription job.
func (h *Handler) GetJobStatus(c *gin.Context) {
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

	resp, err := h.service.GetJobStatus(c.Request.Context(), userID, videoID)
	if err != nil {
		if errors.Is(err, videos.ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		if errors.Is(err, ErrJobNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no transcription job found for this video"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve job status"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
