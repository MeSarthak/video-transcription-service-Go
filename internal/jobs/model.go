package jobs

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusQueued     = "queued"
	StatusProcessing = "processing"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
)

type Job struct {
	ID              uuid.UUID  `json:"id"`
	VideoID         uuid.UUID  `json:"video_id"`
	UserID          uuid.UUID  `json:"user_id"`
	Status          string     `json:"status"`
	Provider        string     `json:"provider"`
	Language        string     `json:"language"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	Attempts        int        `json:"attempts"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type JobResponse struct {
	ID              uuid.UUID  `json:"id"`
	VideoID         uuid.UUID  `json:"video_id"`
	UserID          uuid.UUID  `json:"user_id"`
	Status          string     `json:"status"`
	Provider        string     `json:"provider"`
	Language        string     `json:"language"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
	Attempts        int        `json:"attempts"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (j *Job) ToResponse() *JobResponse {
	return &JobResponse{
		ID:              j.ID,
		VideoID:         j.VideoID,
		UserID:          j.UserID,
		Status:          j.Status,
		Provider:        j.Provider,
		Language:        j.Language,
		ErrorMessage:    j.ErrorMessage,
		Attempts:        j.Attempts,
		LastHeartbeatAt: j.LastHeartbeatAt,
		StartedAt:       j.StartedAt,
		CompletedAt:     j.CompletedAt,
		CreatedAt:       j.CreatedAt,
		UpdatedAt:       j.UpdatedAt,
	}
}
