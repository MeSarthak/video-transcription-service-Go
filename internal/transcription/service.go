package transcription

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"video-transcription-service/internal/videos"
)

type Service interface {
	GetTranscript(ctx context.Context, userID, videoID uuid.UUID) (*TranscriptResponse, error)
	ExportTranscript(ctx context.Context, userID, videoID uuid.UUID, format string) (content, filename, contentType string, err error)
}

type service struct {
	repo      Repository
	videoRepo videos.Repository
}

func NewService(repo Repository, videoRepo videos.Repository) Service {
	return &service{
		repo:      repo,
		videoRepo: videoRepo,
	}
}

func (s *service) GetTranscript(ctx context.Context, userID, videoID uuid.UUID) (*TranscriptResponse, error) {
	// Verify user owns the video
	_, err := s.videoRepo.GetByID(ctx, videoID, userID)
	if err != nil {
		return nil, err
	}

	t, segments, err := s.repo.GetByVideoID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	return ToResponse(t, segments), nil
}

func (s *service) ExportTranscript(ctx context.Context, userID, videoID uuid.UUID, format string) (string, string, string, error) {
	// Verify user owns the video
	video, err := s.videoRepo.GetByID(ctx, videoID, userID)
	if err != nil {
		return "", "", "", err
	}

	t, segments, err := s.repo.GetByVideoID(ctx, videoID)
	if err != nil {
		return "", "", "", err
	}

	baseName := strings.TrimSuffix(video.Filename, filepath.Ext(video.Filename))
	if baseName == "" {
		baseName = "transcript"
	}

	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "srt":
		return FormatSRT(segments), baseName + ".srt", "application/x-subrip", nil
	case "vtt":
		return FormatVTT(segments), baseName + ".vtt", "text/vtt; charset=utf-8", nil
	case "txt", "":
		return FormatTXT(t.FullText), baseName + ".txt", "text/plain; charset=utf-8", nil
	default:
		return "", "", "", fmt.Errorf("unsupported export format '%s' (supported: txt, srt, vtt)", format)
	}
}
