package queue

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestNewSQSQueue_EmptyURL(t *testing.T) {
	ctx := context.Background()
	_, err := NewSQSQueue(ctx, "us-east-1", "")
	if err == nil {
		t.Errorf("expected error when queue URL is empty, got nil")
	}
}

func TestMockQueue(t *testing.T) {
	ctx := context.Background()
	q := NewMockQueue()

	jobID := uuid.New()
	videoID := uuid.New()
	userID := uuid.New()

	msg := &TranscriptionMessage{
		JobID:      jobID,
		VideoID:    videoID,
		UserID:     userID,
		S3VideoKey: "videos/u1/v1/original.mp4",
		Language:   "en-US",
		Attempt:    1,
	}

	// 1. Publish
	err := q.Publish(ctx, msg)
	if err != nil {
		t.Fatalf("failed to publish to mock queue: %v", err)
	}

	if q.Size() != 1 {
		t.Errorf("expected queue size 1, got %d", q.Size())
	}

	// 2. Receive
	received, err := q.Receive(ctx, 10, 5)
	if err != nil {
		t.Fatalf("failed to receive from mock queue: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 received message, got %d", len(received))
	}
	if received[0].Message.JobID != jobID {
		t.Errorf("expected job ID %v, got %v", jobID, received[0].Message.JobID)
	}

	// 3. Delete
	err = q.Delete(ctx, received[0].ReceiptHandle)
	if err != nil {
		t.Fatalf("failed to delete message: %v", err)
	}

	if q.Size() != 0 {
		t.Errorf("expected queue size 0 after deletion, got %d", q.Size())
	}
}
