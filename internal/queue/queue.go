package queue

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrQueueOpFailed    = errors.New("queue operation failed")
	ErrInvalidMessage   = errors.New("invalid queue message payload")
	ErrQueueUnavailable = errors.New("message queue is unavailable")
)

// TranscriptionMessage represents the JSON payload dispatched over SQS.
type TranscriptionMessage struct {
	JobID      uuid.UUID `json:"job_id"`
	VideoID    uuid.UUID `json:"video_id"`
	UserID     uuid.UUID `json:"user_id"`
	S3VideoKey string    `json:"s3_video_key"`
	Language   string    `json:"language"`
	Attempt    int       `json:"attempt"`
}

// ReceivedMessage encapsulates a message received from SQS with its receipt handle.
type ReceivedMessage struct {
	Message                 *TranscriptionMessage
	ReceiptHandle           string
	ApproximateReceiveCount int
}

// Queue defines the interface for publishing and consuming transcription jobs.
type Queue interface {
	// Publish sends a transcription job message to the queue.
	Publish(ctx context.Context, msg *TranscriptionMessage) error

	// Receive consumes messages from the queue using long-polling.
	Receive(ctx context.Context, maxMessages, waitTimeSeconds int32) ([]*ReceivedMessage, error)

	// Delete removes a successfully processed message from the queue.
	Delete(ctx context.Context, receiptHandle string) error

	// ChangeVisibility extends or modifies the message's visibility timeout.
	ChangeVisibility(ctx context.Context, receiptHandle string, visibilityTimeoutSeconds int32) error
}
