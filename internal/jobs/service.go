package jobs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"video-transcription-service/internal/queue"
	"video-transcription-service/internal/videos"
)

var (
	ErrVideoStillUploading    = errors.New("video is still uploading; complete upload before triggering transcription")
	ErrVideoAlreadyProcessing = errors.New("video is already being transcribed")
)

type Service interface {
	TriggerTranscription(ctx context.Context, userID, videoID uuid.UUID, language string) (*JobResponse, error)
	GetJobStatus(ctx context.Context, userID, videoID uuid.UUID) (*JobResponse, error)
}

type service struct {
	jobRepo   Repository
	videoRepo videos.Repository
	queue     queue.Queue
}

func NewService(jobRepo Repository, videoRepo videos.Repository, q queue.Queue) Service {
	return &service{
		jobRepo:   jobRepo,
		videoRepo: videoRepo,
		queue:     q,
	}
}

func (s *service) TriggerTranscription(ctx context.Context, userID, videoID uuid.UUID, language string) (*JobResponse, error) {
	// 1. Validate video exists and check status
	video, err := s.videoRepo.GetByID(ctx, videoID, userID)
	if err != nil {
		return nil, err
	}

	if video.Status == videos.StatusUploading {
		return nil, ErrVideoStillUploading
	}
	if video.Status == videos.StatusProcessing {
		return nil, ErrVideoAlreadyProcessing
	}

	language = strings.TrimSpace(language)
	if language == "" {
		language = "en-US"
	}

	// 2. Create Job in database
	job := &Job{
		ID:       uuid.New(),
		VideoID:  videoID,
		UserID:   userID,
		Status:   StatusQueued,
		Provider: "aws_transcribe",
		Language: language,
		Attempts: 0,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job record: %w", err)
	}

	// 3. Update video status to processing
	if err := s.videoRepo.UpdateStatusAndMetadata(ctx, videoID, userID, videos.StatusProcessing, video.SizeBytes, video.ContentType); err != nil {
		return nil, fmt.Errorf("failed to update video status to processing: %w", err)
	}

	// 4. Publish message to SQS
	msg := &queue.TranscriptionMessage{
		JobID:      job.ID,
		VideoID:    videoID,
		UserID:     userID,
		S3VideoKey: video.StorageKey,
		Language:   language,
		Attempt:    1,
	}

	if err := s.queue.Publish(ctx, msg); err != nil {
		// If queuing fails, mark job as failed
		errMsg := fmt.Sprintf("failed to queue message: %v", err)
		_ = s.jobRepo.UpdateStatus(ctx, job.ID, StatusFailed, &errMsg)
		_ = s.videoRepo.UpdateStatusAndMetadata(ctx, videoID, userID, videos.StatusFailed, video.SizeBytes, video.ContentType)
		return nil, fmt.Errorf("failed to dispatch transcription job: %w", err)
	}

	return job.ToResponse(), nil
}

func (s *service) GetJobStatus(ctx context.Context, userID, videoID uuid.UUID) (*JobResponse, error) {
	// Verify user owns the video
	_, err := s.videoRepo.GetByID(ctx, videoID, userID)
	if err != nil {
		return nil, err
	}

	job, err := s.jobRepo.GetLatestByVideoID(ctx, videoID, userID)
	if err != nil {
		return nil, err
	}

	return job.ToResponse(), nil
}
