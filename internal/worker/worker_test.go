package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-transcription-service/internal/jobs"
	"video-transcription-service/internal/queue"
	"video-transcription-service/internal/videos"
)

type mockJobRepo struct {
	jobs map[uuid.UUID]*jobs.Job
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{
		jobs: make(map[uuid.UUID]*jobs.Job),
	}
}

func (m *mockJobRepo) Create(ctx context.Context, job *jobs.Job) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*jobs.Job, error) {
	j, exists := m.jobs[id]
	if !exists {
		return nil, jobs.ErrJobNotFound
	}
	return j, nil
}

func (m *mockJobRepo) GetLatestByVideoID(ctx context.Context, videoID, userID uuid.UUID) (*jobs.Job, error) {
	for _, j := range m.jobs {
		if j.VideoID == videoID && j.UserID == userID {
			return j, nil
		}
	}
	return nil, jobs.ErrJobNotFound
}

func (m *mockJobRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	j, exists := m.jobs[id]
	if !exists {
		return jobs.ErrJobNotFound
	}
	j.Status = status
	j.ErrorMessage = errorMsg
	return nil
}

func (m *mockJobRepo) UpdateHeartbeat(ctx context.Context, id uuid.UUID) error {
	j, exists := m.jobs[id]
	if !exists {
		return jobs.ErrJobNotFound
	}
	now := time.Now()
	j.LastHeartbeatAt = &now
	return nil
}

func (m *mockJobRepo) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	j, exists := m.jobs[id]
	if !exists {
		return jobs.ErrJobNotFound
	}
	j.Attempts++
	return nil
}

type mockVideoRepo struct {
	videos map[uuid.UUID]*videos.Video
}

func newMockVideoRepo() *mockVideoRepo {
	return &mockVideoRepo{
		videos: make(map[uuid.UUID]*videos.Video),
	}
}

func (m *mockVideoRepo) Create(ctx context.Context, video *videos.Video) error {
	m.videos[video.ID] = video
	return nil
}

func (m *mockVideoRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*videos.Video, error) {
	v, exists := m.videos[id]
	if !exists || v.UserID != userID {
		return nil, videos.ErrVideoNotFound
	}
	return v, nil
}

func (m *mockVideoRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*videos.Video, int64, error) {
	return nil, 0, nil
}

func (m *mockVideoRepo) UpdateStatusAndMetadata(ctx context.Context, id, userID uuid.UUID, status string, sizeBytes int64, contentType string) error {
	v, exists := m.videos[id]
	if !exists || v.UserID != userID {
		return videos.ErrVideoNotFound
	}
	v.Status = status
	return nil
}

func (m *mockVideoRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	delete(m.videos, id)
	return nil
}

func TestWorkerEngine(t *testing.T) {
	ctx := context.Background()

	t.Run("Worker processes message successfully and deletes from queue", func(t *testing.T) {
		q := queue.NewMockQueue()
		jobRepo := newMockJobRepo()
		videoRepo := newMockVideoRepo()

		jobID := uuid.New()
		videoID := uuid.New()
		userID := uuid.New()

		_ = jobRepo.Create(ctx, &jobs.Job{ID: jobID, VideoID: videoID, UserID: userID, Status: jobs.StatusQueued})
		_ = videoRepo.Create(ctx, &videos.Video{ID: videoID, UserID: userID, Status: videos.StatusProcessing})

		_ = q.Publish(ctx, &queue.TranscriptionMessage{
			JobID:      jobID,
			VideoID:    videoID,
			UserID:     userID,
			S3VideoKey: "videos/u1/v1/video.mp4",
		})

		var processedCount int32
		processor := func(ctx context.Context, msg *queue.TranscriptionMessage) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		}

		w := NewWorker(q, jobRepo, videoRepo, processor)

		workerCtx, cancel := context.WithCancel(ctx)
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_ = w.Start(workerCtx)

		if atomic.LoadInt32(&processedCount) != 1 {
			t.Errorf("expected 1 processed job, got %d", processedCount)
		}
		if q.Size() != 0 {
			t.Errorf("expected queue to be empty after successful processing, got %d", q.Size())
		}
	})

	t.Run("Worker skips already completed job (idempotency)", func(t *testing.T) {
		q := queue.NewMockQueue()
		jobRepo := newMockJobRepo()
		videoRepo := newMockVideoRepo()

		jobID := uuid.New()
		videoID := uuid.New()
		userID := uuid.New()

		_ = jobRepo.Create(ctx, &jobs.Job{ID: jobID, VideoID: videoID, UserID: userID, Status: jobs.StatusCompleted})
		_ = q.Publish(ctx, &queue.TranscriptionMessage{
			JobID:   jobID,
			VideoID: videoID,
			UserID:  userID,
		})

		var processedCount int32
		processor := func(ctx context.Context, msg *queue.TranscriptionMessage) error {
			atomic.AddInt32(&processedCount, 1)
			return nil
		}

		w := NewWorker(q, jobRepo, videoRepo, processor)

		workerCtx, cancel := context.WithCancel(ctx)
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_ = w.Start(workerCtx)

		if atomic.LoadInt32(&processedCount) != 0 {
			t.Errorf("expected completed job to be skipped without processing, got %d", processedCount)
		}
		if q.Size() != 0 {
			t.Errorf("expected duplicate message to be acknowledged and deleted from queue")
		}
	})

	t.Run("Worker handles processor error and marks job failed", func(t *testing.T) {
		q := queue.NewMockQueue()
		jobRepo := newMockJobRepo()
		videoRepo := newMockVideoRepo()

		jobID := uuid.New()
		videoID := uuid.New()
		userID := uuid.New()

		_ = jobRepo.Create(ctx, &jobs.Job{ID: jobID, VideoID: videoID, UserID: userID, Status: jobs.StatusQueued})
		_ = videoRepo.Create(ctx, &videos.Video{ID: videoID, UserID: userID, Status: videos.StatusProcessing})

		_ = q.Publish(ctx, &queue.TranscriptionMessage{
			JobID:   jobID,
			VideoID: videoID,
			UserID:  userID,
		})

		processor := func(ctx context.Context, msg *queue.TranscriptionMessage) error {
			return errors.New("FFmpeg extraction failure")
		}

		w := NewWorker(q, jobRepo, videoRepo, processor)

		workerCtx, cancel := context.WithCancel(ctx)
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		_ = w.Start(workerCtx)

		job, _ := jobRepo.GetByID(ctx, jobID)
		if job.Status != jobs.StatusFailed {
			t.Errorf("expected job status failed, got %s", job.Status)
		}
	})
}
