package transcription

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTranscriptNotFound = errors.New("transcript not found for this video")
)

type Repository interface {
	GetByVideoID(ctx context.Context, videoID uuid.UUID) (*Transcript, []*TranscriptSegment, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) GetByVideoID(ctx context.Context, videoID uuid.UUID) (*Transcript, []*TranscriptSegment, error) {
	// 1. Fetch transcript header
	queryTranscript := `
		SELECT id, video_id, job_id, language, full_text, raw_s3_key, created_at, updated_at
		FROM transcripts
		WHERE video_id = $1
	`

	var t Transcript
	err := r.pool.QueryRow(ctx, queryTranscript, videoID).Scan(
		&t.ID,
		&t.VideoID,
		&t.JobID,
		&t.Language,
		&t.FullText,
		&t.RawS3Key,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrTranscriptNotFound
		}
		return nil, nil, fmt.Errorf("failed to get transcript: %w", err)
	}

	// 2. Fetch segments ordered by sequence_number
	querySegments := `
		SELECT id, transcript_id, sequence_number, start_time, end_time, text, confidence
		FROM transcript_segments
		WHERE transcript_id = $1
		ORDER BY sequence_number ASC
	`

	rows, err := r.pool.Query(ctx, querySegments, t.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get transcript segments: %w", err)
	}
	defer rows.Close()

	var segments []*TranscriptSegment
	for rows.Next() {
		var s TranscriptSegment
		if err := rows.Scan(
			&s.ID,
			&s.TranscriptID,
			&s.SequenceNumber,
			&s.StartTime,
			&s.EndTime,
			&s.Text,
			&s.Confidence,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan segment row: %w", err)
		}
		segments = append(segments, &s)
	}

	return &t, segments, nil
}
