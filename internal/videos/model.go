package videos

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusUploading  = "uploading"
	StatusUploaded   = "uploaded"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusDeleted    = "deleted"
)

type Video struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	Filename        string     `json:"filename"`
	StorageKey      string     `json:"storage_key"`
	ContentType     string     `json:"content_type"`
	SizeBytes       int64      `json:"size_bytes"`
	DurationSeconds *float64   `json:"duration_seconds,omitempty"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type VideoResponse struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"user_id"`
	Filename        string     `json:"filename"`
	StorageKey      string     `json:"storage_key"`
	ContentType     string     `json:"content_type"`
	SizeBytes       int64      `json:"size_bytes"`
	DurationSeconds *float64   `json:"duration_seconds,omitempty"`
	Status          string     `json:"status"`
	PlaybackURL     string     `json:"playback_url,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type UploadURLResponse struct {
	VideoID    uuid.UUID `json:"video_id"`
	StorageKey string    `json:"storage_key"`
	UploadURL  string    `json:"upload_url"`
	ExpiresIn  int64     `json:"expires_in"` // expiration in seconds
}

type VideoListResponse struct {
	Videos   []*VideoResponse `json:"videos"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (v *Video) ToResponse(playbackURL string) *VideoResponse {
	return &VideoResponse{
		ID:              v.ID,
		UserID:          v.UserID,
		Filename:        v.Filename,
		StorageKey:      v.StorageKey,
		ContentType:     v.ContentType,
		SizeBytes:       v.SizeBytes,
		DurationSeconds: v.DurationSeconds,
		Status:          v.Status,
		PlaybackURL:     playbackURL,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}
