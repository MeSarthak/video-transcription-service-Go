package storage

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	storage Service
}

func NewHandler(storage Service) *Handler {
	return &Handler{storage: storage}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	storageGroup := rg.Group("/storage")
	{
		storageGroup.PUT("/upload", h.Upload)
		storageGroup.GET("/playback", h.Playback)
		storageGroup.HEAD("/playback", h.Playback)
	}
}

func (h *Handler) Upload(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing key query parameter"})
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "video/mp4"
	}

	if err := h.storage.Upload(c.Request.Context(), key, c.Request.Body, contentType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) Playback(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing key query parameter"})
		return
	}

	meta, err := h.storage.HeadObject(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, ErrObjectNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "media file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rc, err := h.storage.GetObject(c.Request.Context(), key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "media file not found"})
		return
	}
	defer rc.Close()

	contentType := meta.ContentType
	if contentType == "" {
		contentType = "video/mp4"
	}

	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")

	// If the reader implements io.ReadSeeker, use Go's standard http.ServeContent which handles
	// RFC-compliant Range requests (HTTP 206 Partial Content), Content-Range, seeking, and HEAD
	if rs, ok := rc.(io.ReadSeeker); ok {
		http.ServeContent(c.Writer, c.Request, filepath.Base(key), meta.LastModified, rs)
		return
	}

	c.DataFromReader(http.StatusOK, meta.SizeBytes, contentType, rc, nil)
}
