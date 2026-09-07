package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrJobNotFound = errors.New("transcription job not found")
)

type Repository interface {
	Create(ctx context.Context, job *Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*Job, error)
	GetLatestByVideoID(ctx context.Context, videoID, userID uuid.UUID) (*Job, error)
	GetNextQueuedJob(ctx context.Context) (*Job, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error
	UpdateHeartbeat(ctx context.Context, id uuid.UUID) error
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, job *Job) error {
	query := `
		INSERT INTO transcription_jobs (id, video_id, user_id, status, provider, language, attempts)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`

	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	if job.Status == "" {
		job.Status = StatusQueued
	}
	if job.Provider == "" {
		job.Provider = "aws_transcribe"
	}
	if job.Language == "" {
		job.Language = "en-US"
	}

	err := r.pool.QueryRow(
		ctx,
		query,
		job.ID,
		job.VideoID,
		job.UserID,
		job.Status,
		job.Provider,
		job.Language,
		job.Attempts,
	).Scan(&job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transcription job: %w", err)
	}

	return nil
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*Job, error) {
	query := `
		SELECT id, video_id, user_id, status, provider, language, error_message, attempts, last_heartbeat_at, started_at, completed_at, created_at, updated_at
		FROM transcription_jobs
		WHERE id = $1
	`

	var job Job
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.VideoID,
		&job.UserID,
		&job.Status,
		&job.Provider,
		&job.Language,
		&job.ErrorMessage,
		&job.Attempts,
		&job.LastHeartbeatAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to get job by id: %w", err)
	}

	return &job, nil
}

func (r *pgRepository) GetLatestByVideoID(ctx context.Context, videoID, userID uuid.UUID) (*Job, error) {
	query := `
		SELECT id, video_id, user_id, status, provider, language, error_message, attempts, last_heartbeat_at, started_at, completed_at, created_at, updated_at
		FROM transcription_jobs
		WHERE video_id = $1 AND user_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var job Job
	err := r.pool.QueryRow(ctx, query, videoID, userID).Scan(
		&job.ID,
		&job.VideoID,
		&job.UserID,
		&job.Status,
		&job.Provider,
		&job.Language,
		&job.ErrorMessage,
		&job.Attempts,
		&job.LastHeartbeatAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, fmt.Errorf("failed to get latest job for video: %w", err)
	}

	return &job, nil
}

func (r *pgRepository) GetNextQueuedJob(ctx context.Context) (*Job, error) {
	query := `
		SELECT id, video_id, user_id, status, provider, language, error_message, attempts, last_heartbeat_at, started_at, completed_at, created_at, updated_at
		FROM transcription_jobs
		WHERE status = 'queued'
		ORDER BY created_at ASC
		LIMIT 1
	`

	var job Job
	err := r.pool.QueryRow(ctx, query).Scan(
		&job.ID,
		&job.VideoID,
		&job.UserID,
		&job.Status,
		&job.Provider,
		&job.Language,
		&job.ErrorMessage,
		&job.Attempts,
		&job.LastHeartbeatAt,
		&job.StartedAt,
		&job.CompletedAt,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrJobNotFound
		}
		return nil, err
	}

	return &job, nil
}

func (r *pgRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	var query string
	now := time.Now().UTC()

	switch status {
	case StatusProcessing:
		query = `
			UPDATE transcription_jobs
			SET status = $1, error_message = $2, started_at = $3, last_heartbeat_at = $3, updated_at = $3
			WHERE id = $4
		`
		_, err := r.pool.Exec(ctx, query, status, errorMsg, now, id)
		return err

	case StatusCompleted:
		query = `
			UPDATE transcription_jobs
			SET status = $1, error_message = $2, completed_at = $3, updated_at = $3
			WHERE id = $4
		`
		_, err := r.pool.Exec(ctx, query, status, errorMsg, now, id)
		return err

	default:
		query = `
			UPDATE transcription_jobs
			SET status = $1, error_message = $2, updated_at = $3
			WHERE id = $4
		`
		_, err := r.pool.Exec(ctx, query, status, errorMsg, now, id)
		return err
	}
}

func (r *pgRepository) UpdateHeartbeat(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE transcription_jobs
		SET last_heartbeat_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *pgRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE transcription_jobs
		SET attempts = attempts + 1, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}
