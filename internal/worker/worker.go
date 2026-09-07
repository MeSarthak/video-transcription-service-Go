package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"video-transcription-service/internal/jobs"
	"video-transcription-service/internal/queue"
	"video-transcription-service/internal/videos"
)

const (
	DefaultMaxAttempts       = 3
	HeartbeatIntervalSeconds = 20
	VisibilityTimeoutSeconds = 60
)

// JobProcessor defines the function signature for executing transcription work on a job.
type JobProcessor func(ctx context.Context, msg *queue.TranscriptionMessage) error

type Worker struct {
	queue        queue.Queue
	jobRepo      jobs.Repository
	videoRepo    videos.Repository
	processor    JobProcessor
	maxAttempts  int
	pollInterval time.Duration
}

func NewWorker(
	q queue.Queue,
	jobRepo jobs.Repository,
	videoRepo videos.Repository,
	processor JobProcessor,
) *Worker {
	return &Worker{
		queue:        q,
		jobRepo:      jobRepo,
		videoRepo:    videoRepo,
		processor:    processor,
		maxAttempts:  DefaultMaxAttempts,
		pollInterval: 1 * time.Second,
	}
}

// Start runs the continuous worker polling loop until the context is canceled.
func (w *Worker) Start(ctx context.Context) error {
	slog.Info("Starting transcription worker loop...")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker context canceled, stopping polling loop")
			return ctx.Err()
		default:
			// 1. Poll SQS with long polling
			messages, err := w.queue.Receive(ctx, 1, 5)
			if err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("Failed to receive messages from queue", slog.Any("error", err))
			}

			if len(messages) > 0 {
				for _, msg := range messages {
					w.processMessage(ctx, msg)
				}
				continue
			}

			// 2. Fallback check for local development: check PostgreSQL directly for queued jobs
			nextJob, err := w.jobRepo.GetNextQueuedJob(ctx)
			if err == nil && nextJob != nil {
				slog.Info("Found queued job in database (local queue dispatch)", slog.String("job_id", nextJob.ID.String()))
				syntheticMsg := &queue.ReceivedMessage{
					Message: &queue.TranscriptionMessage{
						JobID:    nextJob.ID,
						VideoID:  nextJob.VideoID,
						UserID:   nextJob.UserID,
						Language: nextJob.Language,
						Attempt:  nextJob.Attempts + 1,
					},
					ReceiptHandle:           fmt.Sprintf("db-job-%s", nextJob.ID.String()),
					ApproximateReceiveCount: nextJob.Attempts + 1,
				}
				w.processMessage(ctx, syntheticMsg)
				continue
			}

			// Idle sleep before next poll
			time.Sleep(w.pollInterval)
		}
	}
}

func (w *Worker) processMessage(ctx context.Context, received *queue.ReceivedMessage) {
	msg := received.Message
	receiptHandle := received.ReceiptHandle

	slog.Info("Processing transcription job",
		slog.String("job_id", msg.JobID.String()),
		slog.String("video_id", msg.VideoID.String()),
		slog.Int("attempt", received.ApproximateReceiveCount),
	)

	// 1. Idempotency check: verify job status in DB
	job, err := w.jobRepo.GetByID(ctx, msg.JobID)
	if err != nil {
		slog.Error("Failed to fetch job from database", slog.Any("error", err), slog.String("job_id", msg.JobID.String()))
		return
	}

	if job.Status == jobs.StatusCompleted {
		slog.Info("Job already completed, skipping and deleting duplicate message", slog.String("job_id", msg.JobID.String()))
		_ = w.queue.Delete(ctx, receiptHandle)
		return
	}

	// 2. Retry limits check
	if received.ApproximateReceiveCount > w.maxAttempts {
		errMsg := fmt.Sprintf("Maximum retry attempts (%d) exceeded", w.maxAttempts)
		slog.Error("Job failed maximum attempts, routing to DLQ / marking failed",
			slog.String("job_id", msg.JobID.String()),
			slog.String("error", errMsg),
		)
		_ = w.jobRepo.UpdateStatus(ctx, msg.JobID, jobs.StatusFailed, &errMsg)
		_ = w.videoRepo.UpdateStatusAndMetadata(ctx, msg.VideoID, msg.UserID, videos.StatusFailed, 0, "")
		_ = w.queue.Delete(ctx, receiptHandle)
		return
	}

	// 3. Mark job as processing & increment attempts
	_ = w.jobRepo.IncrementAttempts(ctx, msg.JobID)
	_ = w.jobRepo.UpdateStatus(ctx, msg.JobID, jobs.StatusProcessing, nil)
	_ = w.videoRepo.UpdateStatusAndMetadata(ctx, msg.VideoID, msg.UserID, videos.StatusProcessing, 0, "")

	// 4. Start Background Heartbeat & Visibility Timeout Goroutine
	stopHeartbeat := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(HeartbeatIntervalSeconds * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopHeartbeat:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Extend SQS Visibility Timeout
				if err := w.queue.ChangeVisibility(ctx, receiptHandle, VisibilityTimeoutSeconds); err != nil {
					slog.Warn("Failed to extend SQS visibility timeout", slog.Any("error", err), slog.String("job_id", msg.JobID.String()))
				}
				// Update DB Heartbeat
				if err := w.jobRepo.UpdateHeartbeat(ctx, msg.JobID); err != nil {
					slog.Warn("Failed to update database job heartbeat", slog.Any("error", err), slog.String("job_id", msg.JobID.String()))
				}
			}
		}
	}()

	// 5. Execute Job Processor
	var procErr error
	if w.processor != nil {
		procErr = w.processor(ctx, msg)
	}

	// 6. Stop Heartbeat Goroutine
	close(stopHeartbeat)
	wg.Wait()

	// 7. Handle Processor Result
	if procErr != nil {
		errMsg := procErr.Error()
		slog.Error("Job processing failed", slog.String("job_id", msg.JobID.String()), slog.Any("error", procErr))
		_ = w.jobRepo.UpdateStatus(ctx, msg.JobID, jobs.StatusFailed, &errMsg)
		_ = w.videoRepo.UpdateStatusAndMetadata(ctx, msg.VideoID, msg.UserID, videos.StatusFailed, 0, "")
		time.Sleep(w.pollInterval)
		return
	}

	// 8. Delete message on success
	if err := w.queue.Delete(ctx, receiptHandle); err != nil {
		slog.Warn("Failed to delete completed SQS message", slog.Any("error", err))
	} else {
		slog.Info("Job processed and acknowledged successfully", slog.String("job_id", msg.JobID.String()))
	}
}
