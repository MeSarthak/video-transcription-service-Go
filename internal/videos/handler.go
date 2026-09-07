package videos

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"video-transcription-service/internal/middleware"
)

type RequestUploadURLInput struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type"`
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	videos := rg.Group("/videos", authMiddleware)
	{
		videos.POST("/upload-url", h.RequestUploadURL)
		videos.POST("/:id/complete-upload", h.CompleteUpload)
		videos.GET("", h.ListVideos)
		videos.GET("/:id", h.GetVideo)
		videos.DELETE("/:id", h.DeleteVideo)
	}
}

// RequestUploadURL handles requests to generate a presigned S3 PUT URL.
func (h *Handler) RequestUploadURL(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input RequestUploadURLInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.RequestUploadURL(c.Request.Context(), userID, input.Filename, input.ContentType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// CompleteUpload verifies the S3 upload via HeadObject and marks the video as uploaded.
func (h *Handler) CompleteUpload(c *gin.Context) {
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

	resp, err := h.service.CompleteUpload(c.Request.Context(), userID, videoID)
	if err != nil {
		if errors.Is(err, ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		if errors.Is(err, ErrVideoObjectNotFound) || errors.Is(err, ErrEmptyVideoFile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListVideos lists the authenticated user's uploaded videos with pagination.
func (h *Handler) ListVideos(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := h.service.ListVideos(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list videos"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetVideo returns metadata and a streaming playback URL for a single video.
func (h *Handler) GetVideo(c *gin.Context) {
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

	resp, err := h.service.GetVideo(c.Request.Context(), userID, videoID)
	if err != nil {
		if errors.Is(err, ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve video"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteVideo removes a video record and its storage object.
func (h *Handler) DeleteVideo(c *gin.Context) {
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

	if err := h.service.DeleteVideo(c.Request.Context(), userID, videoID); err != nil {
		if errors.Is(err, ErrVideoNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete video"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "video deleted successfully"})
}
