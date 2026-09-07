package transcription

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"video-transcription-service/internal/videos"
)

type mockTranscriptRepo struct {
	transcripts map[uuid.UUID]*Transcript
	segments    map[uuid.UUID][]*TranscriptSegment
}

func newMockTranscriptRepo() *mockTranscriptRepo {
	return &mockTranscriptRepo{
		transcripts: make(map[uuid.UUID]*Transcript),
		segments:    make(map[uuid.UUID][]*TranscriptSegment),
	}
}

func (m *mockTranscriptRepo) GetByVideoID(ctx context.Context, videoID uuid.UUID) (*Transcript, []*TranscriptSegment, error) {
	t, exists := m.transcripts[videoID]
	if !exists {
		return nil, nil, ErrTranscriptNotFound
	}
	return t, m.segments[t.ID], nil
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
	return nil
}

func (m *mockVideoRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	delete(m.videos, id)
	return nil
}

func TestTranscriptionService(t *testing.T) {
	ctx := context.Background()
	transcriptRepo := newMockTranscriptRepo()
	videoRepo := newMockVideoRepo()
	svc := NewService(transcriptRepo, videoRepo)

	userID := uuid.New()
	videoID := uuid.New()
	transcriptID := uuid.New()

	// Seed video
	_ = videoRepo.Create(ctx, &videos.Video{
		ID:       videoID,
		UserID:   userID,
		Filename: "talk.mp4",
		Status:   videos.StatusCompleted,
	})

	// Seed transcript & segments
	transcriptRepo.transcripts[videoID] = &Transcript{
		ID:        transcriptID,
		VideoID:   videoID,
		Language:  "en-US",
		FullText:  "Welcome to the video transcription course.",
		CreatedAt: time.Now(),
	}
	transcriptRepo.segments[transcriptID] = []*TranscriptSegment{
		{
			SequenceNumber: 1,
			StartTime:      0.0,
			EndTime:        3.5,
			Text:           "Welcome to the video transcription course.",
			Confidence:     0.99,
		},
	}

	t.Run("GetTranscript returns full transcript with segments", func(t *testing.T) {
		resp, err := svc.GetTranscript(ctx, userID, videoID)
		if err != nil {
			t.Fatalf("expected GetTranscript to succeed, got %v", err)
		}

		if resp.FullText != "Welcome to the video transcription course." {
			t.Errorf("unexpected full text: %s", resp.FullText)
		}
		if len(resp.Segments) != 1 {
			t.Errorf("expected 1 segment, got %d", len(resp.Segments))
		}
	})

	t.Run("ExportTranscript generates SRT format", func(t *testing.T) {
		content, filename, contentType, err := svc.ExportTranscript(ctx, userID, videoID, "srt")
		if err != nil {
			t.Fatalf("export failed: %v", err)
		}

		if filename != "talk.srt" {
			t.Errorf("expected talk.srt, got %s", filename)
		}
		if contentType != "application/x-subrip" {
			t.Errorf("expected application/x-subrip, got %s", contentType)
		}
		if !strings.Contains(content, "00:00:00,000 --> 00:00:03,500") {
			t.Errorf("missing timestamps in SRT export: %s", content)
		}
	})

	t.Run("ExportTranscript generates VTT format", func(t *testing.T) {
		content, filename, contentType, err := svc.ExportTranscript(ctx, userID, videoID, "vtt")
		if err != nil {
			t.Fatalf("export failed: %v", err)
		}

		if filename != "talk.vtt" {
			t.Errorf("expected talk.vtt, got %s", filename)
		}
		if contentType != "text/vtt; charset=utf-8" {
			t.Errorf("expected text/vtt; charset=utf-8, got %s", contentType)
		}
		if !strings.HasPrefix(content, "WEBVTT") {
			t.Errorf("expected WEBVTT header: %s", content)
		}
	})
}
