package videos

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-transcription-service/internal/storage"
)

type mockVideoRepo struct {
	videos map[uuid.UUID]*Video
}

func newMockVideoRepo() *mockVideoRepo {
	return &mockVideoRepo{
		videos: make(map[uuid.UUID]*Video),
	}
}

func (m *mockVideoRepo) Create(ctx context.Context, video *Video) error {
	m.videos[video.ID] = video
	video.CreatedAt = time.Now()
	video.UpdatedAt = time.Now()
	return nil
}

func (m *mockVideoRepo) GetByID(ctx context.Context, id, userID uuid.UUID) (*Video, error) {
	v, exists := m.videos[id]
	if !exists || v.UserID != userID {
		return nil, ErrVideoNotFound
	}
	return v, nil
}

func (m *mockVideoRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Video, int64, error) {
	var list []*Video
	for _, v := range m.videos {
		if v.UserID == userID && v.Status != StatusDeleted {
			list = append(list, v)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockVideoRepo) UpdateStatusAndMetadata(ctx context.Context, id, userID uuid.UUID, status string, sizeBytes int64, contentType string) error {
	v, exists := m.videos[id]
	if !exists || v.UserID != userID {
		return ErrVideoNotFound
	}
	v.Status = status
	v.SizeBytes = sizeBytes
	v.ContentType = contentType
	v.UpdatedAt = time.Now()
	return nil
}

func (m *mockVideoRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	v, exists := m.videos[id]
	if !exists || v.UserID != userID {
		return ErrVideoNotFound
	}
	delete(m.videos, id)
	return nil
}

func TestVideoService(t *testing.T) {
	ctx := context.Background()
	repo := newMockVideoRepo()
	mockStore := storage.NewMockStorage()
	svc := NewService(repo, mockStore)

	userID := uuid.New()

	t.Run("RequestUploadURL creates record and presigned PUT URL", func(t *testing.T) {
		resp, err := svc.RequestUploadURL(ctx, userID, "my_lecture.mp4", "video/mp4")
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}

		if resp.VideoID == uuid.Nil {
			t.Errorf("expected valid video ID")
		}
		if !strings.Contains(resp.UploadURL, "upload") && !strings.Contains(resp.UploadURL, "signed=true") {
			t.Errorf("expected presigned upload URL, got %s", resp.UploadURL)
		}

		// Verify record in repo
		video, err := repo.GetByID(ctx, resp.VideoID, userID)
		if err != nil {
			t.Fatalf("expected video in repo, got %v", err)
		}
		if video.Status != StatusUploading {
			t.Errorf("expected status %s, got %s", StatusUploading, video.Status)
		}
	})

	t.Run("CompleteUpload fails when object is missing in S3", func(t *testing.T) {
		reqResp, _ := svc.RequestUploadURL(ctx, userID, "missing.mp4", "video/mp4")
		_, err := svc.CompleteUpload(ctx, userID, reqResp.VideoID)
		if err != ErrVideoObjectNotFound {
			t.Errorf("expected ErrVideoObjectNotFound, got %v", err)
		}
	})

	t.Run("CompleteUpload fails when object is 0 bytes", func(t *testing.T) {
		reqResp, _ := svc.RequestUploadURL(ctx, userID, "empty.mp4", "video/mp4")
		_ = mockStore.Upload(ctx, reqResp.StorageKey, strings.NewReader(""), "video/mp4")

		_, err := svc.CompleteUpload(ctx, userID, reqResp.VideoID)
		if err != ErrEmptyVideoFile {
			t.Errorf("expected ErrEmptyVideoFile, got %v", err)
		}
	})

	t.Run("CompleteUpload succeeds when valid object uploaded", func(t *testing.T) {
		reqResp, _ := svc.RequestUploadURL(ctx, userID, "valid.mp4", "video/mp4")
		payload := "valid video file content"
		_ = mockStore.Upload(ctx, reqResp.StorageKey, strings.NewReader(payload), "video/mp4")

		completeResp, err := svc.CompleteUpload(ctx, userID, reqResp.VideoID)
		if err != nil {
			t.Fatalf("expected CompleteUpload to succeed, got %v", err)
		}

		if completeResp.Status != StatusUploaded {
			t.Errorf("expected status uploaded, got %s", completeResp.Status)
		}
		if completeResp.SizeBytes != int64(len(payload)) {
			t.Errorf("expected size %d, got %d", len(payload), completeResp.SizeBytes)
		}
		if completeResp.PlaybackURL == "" {
			t.Errorf("expected playback URL to be generated")
		}
	})

	t.Run("GetVideo returns playback URL", func(t *testing.T) {
		reqResp, _ := svc.RequestUploadURL(ctx, userID, "stream_test.mp4", "video/mp4")
		_ = mockStore.Upload(ctx, reqResp.StorageKey, strings.NewReader("test"), "video/mp4")
		_, _ = svc.CompleteUpload(ctx, userID, reqResp.VideoID)

		video, err := svc.GetVideo(ctx, userID, reqResp.VideoID)
		if err != nil {
			t.Fatalf("expected GetVideo to succeed, got %v", err)
		}
		if video.PlaybackURL == "" {
			t.Errorf("expected playback URL")
		}
	})

	t.Run("DeleteVideo removes from storage and database", func(t *testing.T) {
		reqResp, _ := svc.RequestUploadURL(ctx, userID, "delete_test.mp4", "video/mp4")
		_ = mockStore.Upload(ctx, reqResp.StorageKey, strings.NewReader("to be deleted"), "video/mp4")
		_, _ = svc.CompleteUpload(ctx, userID, reqResp.VideoID)

		err := svc.DeleteVideo(ctx, userID, reqResp.VideoID)
		if err != nil {
			t.Fatalf("expected delete to succeed, got %v", err)
		}

		_, err = svc.GetVideo(ctx, userID, reqResp.VideoID)
		if err != ErrVideoNotFound {
			t.Errorf("expected video not found after deletion, got %v", err)
		}
	})
}
