package videos

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrVideoNotFound      = errors.New("video not found")
	ErrInvalidVideoStatus = errors.New("invalid video status")
)

type Repository interface {
	Create(ctx context.Context, video *Video) error
	GetByID(ctx context.Context, id, userID uuid.UUID) (*Video, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Video, int64, error)
	UpdateStatusAndMetadata(ctx context.Context, id, userID uuid.UUID, status string, sizeBytes int64, contentType string) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, video *Video) error {
	query := `
		INSERT INTO videos (id, user_id, filename, storage_key, content_type, size_bytes, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	if video.ID == uuid.Nil {
		video.ID = uuid.New()
	}
	if video.Status == "" {
		video.Status = StatusUploading
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		video.ID,
		video.UserID,
		video.Filename,
		video.StorageKey,
		video.ContentType,
		video.SizeBytes,
		video.Status,
	).Scan(&video.CreatedAt, &video.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create video record: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*Video, error) {
	query := `
		SELECT id, user_id, filename, storage_key, content_type, size_bytes, duration_seconds, status, created_at, updated_at
		FROM videos
		WHERE id = $1 AND user_id = $2
	`

	var video Video
	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&video.ID,
		&video.UserID,
		&video.Filename,
		&video.StorageKey,
		&video.ContentType,
		&video.SizeBytes,
		&video.DurationSeconds,
		&video.Status,
		&video.CreatedAt,
		&video.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVideoNotFound
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	return &video, nil
}

func (r *pgRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Video, int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM videos
		WHERE user_id = $1 AND status != 'deleted'
	`
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count videos: %w", err)
	}

	query := `
		SELECT id, user_id, filename, storage_key, content_type, size_bytes, duration_seconds, status, created_at, updated_at
		FROM videos
		WHERE user_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list videos: %w", err)
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		var video Video
		if err := rows.Scan(
			&video.ID,
			&video.UserID,
			&video.Filename,
			&video.StorageKey,
			&video.ContentType,
			&video.SizeBytes,
			&video.DurationSeconds,
			&video.Status,
			&video.CreatedAt,
			&video.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan video row: %w", err)
		}
		videos = append(videos, &video)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error reading video rows: %w", err)
	}

	return videos, total, nil
}

func (r *pgRepository) UpdateStatusAndMetadata(ctx context.Context, id, userID uuid.UUID, status string, sizeBytes int64, contentType string) error {
	query := `
		UPDATE videos
		SET status = $1, size_bytes = $2, content_type = $3, updated_at = NOW()
		WHERE id = $4 AND user_id = $5
	`

	cmdTag, err := r.pool.Exec(ctx, query, status, sizeBytes, contentType, id, userID)
	if err != nil {
		return fmt.Errorf("failed to update video metadata: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrVideoNotFound
	}

	return nil
}

func (r *pgRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	query := `
		DELETE FROM videos
		WHERE id = $1 AND user_id = $2
	`

	cmdTag, err := r.pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrVideoNotFound
	}

	return nil
}
