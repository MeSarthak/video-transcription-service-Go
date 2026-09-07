package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-transcription-service/internal/queue"
	"video-transcription-service/internal/videos"
)

type mockJobRepo struct {
	jobs map[uuid.UUID]*Job
}

func newMockJobRepo() *mockJobRepo {
	return &mockJobRepo{
		jobs: make(map[uuid.UUID]*Job),
	}
}

func (m *mockJobRepo) Create(ctx context.Context, job *Job) error {
	m.jobs[job.ID] = job
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	return nil
}

func (m *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*Job, error) {
	j, exists := m.jobs[id]
	if !exists {
		return nil, ErrJobNotFound
	}
	return j, nil
}

func (m *mockJobRepo) GetLatestByVideoID(ctx context.Context, videoID, userID uuid.UUID) (*Job, error) {
	for _, j := range m.jobs {
		if j.VideoID == videoID && j.UserID == userID {
			return j, nil
		}
	}
	return nil, ErrJobNotFound
}

func (m *mockJobRepo) GetNextQueuedJob(ctx context.Context) (*Job, error) {
	for _, j := range m.jobs {
		if j.Status == StatusQueued {
			return j, nil
		}
	}
	return nil, ErrJobNotFound
}

func (m *mockJobRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	j, exists := m.jobs[id]
	if !exists {
		return ErrJobNotFound
	}
	j.Status = status
	j.ErrorMessage = errorMsg
	j.UpdatedAt = time.Now()
	return nil
}

func (m *mockJobRepo) UpdateHeartbeat(ctx context.Context, id uuid.UUID) error {
	j, exists := m.jobs[id]
	if !exists {
		return ErrJobNotFound
	}
	now := time.Now()
	j.LastHeartbeatAt = &now
	j.UpdatedAt = now
	return nil
}

func (m *mockJobRepo) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	j, exists := m.jobs[id]
	if !exists {
		return ErrJobNotFound
	}
	j.Attempts++
	j.UpdatedAt = time.Now()
	return nil
}

// Mock Video Repo for Jobs testing
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
	v.SizeBytes = sizeBytes
	v.ContentType = contentType
	return nil
}

func (m *mockVideoRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	delete(m.videos, id)
	return nil
}

func TestJobService(t *testing.T) {
	ctx := context.Background()
	jobRepo := newMockJobRepo()
	videoRepo := newMockVideoRepo()
	mockQ := queue.NewMockQueue()
	svc := NewService(jobRepo, videoRepo, mockQ)

	userID := uuid.New()
	videoID := uuid.New()

	// Seed an uploaded video
	testVideo := &videos.Video{
		ID:          videoID,
		UserID:      userID,
		Filename:    "presentation.mp4",
		StorageKey:  "videos/u1/v1/presentation.mp4",
		ContentType: "video/mp4",
		SizeBytes:   1024 * 1024 * 5,
		Status:      videos.StatusUploaded,
	}
	_ = videoRepo.Create(ctx, testVideo)

	t.Run("TriggerTranscription creates job, updates video status, and queues message", func(t *testing.T) {
		resp, err := svc.TriggerTranscription(ctx, userID, videoID, "en-US")
		if err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}

		if resp.Status != StatusQueued {
			t.Errorf("expected job status queued, got %s", resp.Status)
		}
		if resp.Language != "en-US" {
			t.Errorf("expected language en-US, got %s", resp.Language)
		}

		// Verify video updated to processing
		v, _ := videoRepo.GetByID(ctx, videoID, userID)
		if v.Status != videos.StatusProcessing {
			t.Errorf("expected video status processing, got %s", v.Status)
		}

		// Verify message in queue
		if mockQ.Size() != 1 {
			t.Errorf("expected 1 message in queue, got %d", mockQ.Size())
		}
	})

	t.Run("TriggerTranscription fails when video is already processing", func(t *testing.T) {
		_, err := svc.TriggerTranscription(ctx, userID, videoID, "en-US")
		if err != ErrVideoAlreadyProcessing {
			t.Errorf("expected ErrVideoAlreadyProcessing, got %v", err)
		}
	})

	t.Run("TriggerTranscription fails when video is still uploading", func(t *testing.T) {
		uploadingVidID := uuid.New()
		_ = videoRepo.Create(ctx, &videos.Video{
			ID:     uploadingVidID,
			UserID: userID,
			Status: videos.StatusUploading,
		})

		_, err := svc.TriggerTranscription(ctx, userID, uploadingVidID, "en-US")
		if err != ErrVideoStillUploading {
			t.Errorf("expected ErrVideoStillUploading, got %v", err)
		}
	})

	t.Run("GetJobStatus returns latest job", func(t *testing.T) {
		jobStatus, err := svc.GetJobStatus(ctx, userID, videoID)
		if err != nil {
			t.Fatalf("expected GetJobStatus to succeed, got %v", err)
		}
		if jobStatus.VideoID != videoID {
			t.Errorf("expected videoID %v, got %v", videoID, jobStatus.VideoID)
		}
	})
}
