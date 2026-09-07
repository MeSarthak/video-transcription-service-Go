package videos

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"video-transcription-service/internal/storage"
)

var (
	ErrEmptyFilename       = errors.New("filename cannot be empty")
	ErrVideoObjectNotFound = errors.New("video file not found in storage; upload may have failed or timed out")
	ErrEmptyVideoFile      = errors.New("uploaded video file is empty (0 bytes)")
)

type Service interface {
	RequestUploadURL(ctx context.Context, userID uuid.UUID, filename, contentType string) (*UploadURLResponse, error)
	CompleteUpload(ctx context.Context, userID, videoID uuid.UUID) (*VideoResponse, error)
	GetVideo(ctx context.Context, userID, videoID uuid.UUID) (*VideoResponse, error)
	ListVideos(ctx context.Context, userID uuid.UUID, page, pageSize int) (*VideoListResponse, error)
	DeleteVideo(ctx context.Context, userID, videoID uuid.UUID) error
}

type service struct {
	repo    Repository
	storage storage.Service
}

func NewService(repo Repository, storage storage.Service) Service {
	return &service{
		repo:    repo,
		storage: storage,
	}
}

func (s *service) RequestUploadURL(ctx context.Context, userID uuid.UUID, filename, contentType string) (*UploadURLResponse, error) {
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return nil, ErrEmptyFilename
	}

	cleanFilename := filepath.Base(filename)
	videoID := uuid.New()
	storageKey := fmt.Sprintf("videos/%s/%s/%s", userID.String(), videoID.String(), cleanFilename)

	if contentType == "" {
		contentType = "video/mp4"
	}

	video := &Video{
		ID:          videoID,
		UserID:      userID,
		Filename:    cleanFilename,
		StorageKey:  storageKey,
		ContentType: contentType,
		Status:      StatusUploading,
		SizeBytes:   0,
	}

	if err := s.repo.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("failed to register video upload: %w", err)
	}

	expiry := 15 * time.Minute
	uploadURL, err := s.storage.GeneratePresignedPutURL(ctx, storageKey, contentType, expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return &UploadURLResponse{
		VideoID:    videoID,
		StorageKey: storageKey,
		UploadURL:  uploadURL,
		ExpiresIn:  int64(expiry.Seconds()),
	}, nil
}

func (s *service) CompleteUpload(ctx context.Context, userID, videoID uuid.UUID) (*VideoResponse, error) {
	video, err := s.repo.GetByID(ctx, videoID, userID)
	if err != nil {
		return nil, err
	}

	// If already uploaded/processing/completed, return current state
	if video.Status != StatusUploading {
		playbackURL, _ := s.storage.GeneratePresignedGetURL(ctx, video.StorageKey, 1*time.Hour)
		return video.ToResponse(playbackURL), nil
	}

	// Verify object in S3 using HeadObject
	meta, err := s.storage.HeadObject(ctx, video.StorageKey)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotFound) {
			return nil, ErrVideoObjectNotFound
		}
		return nil, fmt.Errorf("failed to verify storage object: %w", err)
	}

	if meta.SizeBytes <= 0 {
		return nil, ErrEmptyVideoFile
	}

	contentType := meta.ContentType
	if contentType == "" {
		contentType = video.ContentType
	}

	if err := s.repo.UpdateStatusAndMetadata(ctx, videoID, userID, StatusUploaded, meta.SizeBytes, contentType); err != nil {
		return nil, fmt.Errorf("failed to update video upload status: %w", err)
	}

	video.Status = StatusUploaded
	video.SizeBytes = meta.SizeBytes
	video.ContentType = contentType

	playbackURL, _ := s.storage.GeneratePresignedGetURL(ctx, video.StorageKey, 1*time.Hour)
	return video.ToResponse(playbackURL), nil
}

func (s *service) GetVideo(ctx context.Context, userID, videoID uuid.UUID) (*VideoResponse, error) {
	video, err := s.repo.GetByID(ctx, videoID, userID)
	if err != nil {
		return nil, err
	}

	var playbackURL string
	if video.Status != StatusUploading && video.Status != StatusFailed && video.Status != StatusDeleted {
		playbackURL, _ = s.storage.GeneratePresignedGetURL(ctx, video.StorageKey, 1*time.Hour)
	}

	return video.ToResponse(playbackURL), nil
}

func (s *service) ListVideos(ctx context.Context, userID uuid.UUID, page, pageSize int) (*VideoListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	videoList, total, err := s.repo.ListByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve videos: %w", err)
	}

	var responseList []*VideoResponse
	for _, v := range videoList {
		responseList = append(responseList, v.ToResponse(""))
	}

	return &VideoListResponse{
		Videos:   responseList,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *service) DeleteVideo(ctx context.Context, userID, videoID uuid.UUID) error {
	video, err := s.repo.GetByID(ctx, videoID, userID)
	if err != nil {
		return err
	}

	// Delete from storage (ignore not found errors during deletion)
	if err := s.storage.DeleteObject(ctx, video.StorageKey); err != nil && !errors.Is(err, storage.ErrObjectNotFound) {
		// Log warning but proceed with DB deletion
	}

	return s.repo.Delete(ctx, videoID, userID)
}
