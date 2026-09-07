package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"video-transcription-service/internal/database"
	"video-transcription-service/internal/jobs"
	"video-transcription-service/internal/queue"
	"video-transcription-service/internal/storage"
	"video-transcription-service/internal/videos"
	"video-transcription-service/pkg/ffmpeg"
	"video-transcription-service/pkg/transcribe"
)

type Processor struct {
	db        *database.Database
	storage   storage.Service
	extractor ffmpeg.Extractor
	provider  transcribe.Provider
	jobRepo   jobs.Repository
	videoRepo videos.Repository
	bucket    string
}

func NewProcessor(
	db *database.Database,
	storage storage.Service,
	extractor ffmpeg.Extractor,
	provider transcribe.Provider,
	jobRepo jobs.Repository,
	videoRepo videos.Repository,
	bucket string,
) *Processor {
	return &Processor{
		db:        db,
		storage:   storage,
		extractor: extractor,
		provider:  provider,
		jobRepo:   jobRepo,
		videoRepo: videoRepo,
		bucket:    bucket,
	}
}

// ProcessJob coordinates the end-to-end transcription pipeline.
func (p *Processor) ProcessJob(ctx context.Context, msg *queue.TranscriptionMessage) error {
	slog.Info("Starting end-to-end transcription processing pipeline",
		slog.String("job_id", msg.JobID.String()),
		slog.String("video_id", msg.VideoID.String()),
	)

	// 1. Create isolated scratch workspace
	scratchDir, cleanup, err := ffmpeg.CreateScratchDir(msg.JobID)
	if err != nil {
		return fmt.Errorf("failed to create scratch directory: %w", err)
	}
	defer cleanup()

	// 2. Download original video to local scratch directory
	localVideoPath, err := p.downloadVideo(ctx, msg.S3VideoKey, scratchDir)
	if err != nil {
		return err
	}

	// 3. Extract and probe audio
	localAudioPath, duration, err := p.extractAudio(ctx, localVideoPath, scratchDir)
	if err != nil {
		return err
	}

	// 4. Upload extracted audio to storage
	audioStorageKey, err := p.uploadAudio(ctx, msg.UserID, msg.VideoID, localAudioPath)
	if err != nil {
		return err
	}

	// 5. Transcribe audio with speech-to-text provider
	result, rawKey, err := p.transcribeAudio(ctx, msg, audioStorageKey, localAudioPath)
	if err != nil {
		return err
	}

	// 6. Persist transcripts, segments, and updated statuses inside a database transaction
	if err := p.persistResults(ctx, msg, result, rawKey, duration); err != nil {
		return fmt.Errorf("failed to persist transcription transaction: %w", err)
	}

	slog.Info("Transcription pipeline completed successfully",
		slog.String("job_id", msg.JobID.String()),
		slog.String("video_id", msg.VideoID.String()),
		slog.Int("segments_count", len(result.Segments)),
		slog.Float64("duration_sec", duration),
	)

	return nil
}

func (p *Processor) downloadVideo(ctx context.Context, s3Key, scratchDir string) (string, error) {
	localVideoPath := filepath.Join(scratchDir, "original_video")
	reader, err := p.storage.GetObject(ctx, s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to download video from storage (%s): %w", s3Key, err)
	}
	defer reader.Close()

	file, err := os.Create(localVideoPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local video file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("failed to write video to scratch disk: %w", err)
	}
	return localVideoPath, nil
}

func (p *Processor) extractAudio(ctx context.Context, localVideoPath, scratchDir string) (string, float64, error) {
	duration, _ := p.extractor.GetDuration(ctx, localVideoPath)

	localAudioPath := filepath.Join(scratchDir, "extracted_audio.wav")
	if err := p.extractor.ExtractAudio(ctx, localVideoPath, localAudioPath); err != nil {
		return "", 0, fmt.Errorf("FFmpeg audio extraction failed: %w", err)
	}
	return localAudioPath, duration, nil
}

func (p *Processor) uploadAudio(ctx context.Context, userID, videoID uuid.UUID, localAudioPath string) (string, error) {
	audioStorageKey := fmt.Sprintf("audio/%s/%s/audio.wav", userID.String(), videoID.String())
	audioData, err := os.ReadFile(localAudioPath)
	if err != nil {
		return "", fmt.Errorf("failed to read extracted audio: %w", err)
	}

	if err := p.storage.Upload(ctx, audioStorageKey, bytes.NewReader(audioData), "audio/wav"); err != nil {
		return "", fmt.Errorf("failed to upload extracted audio to storage: %w", err)
	}
	return audioStorageKey, nil
}

func (p *Processor) transcribeAudio(ctx context.Context, msg *queue.TranscriptionMessage, audioStorageKey, localAudioPath string) (*transcribe.TranscriptionResult, string, error) {
	mediaS3URI := fmt.Sprintf("s3://%s/%s", p.bucket, audioStorageKey)
	rawOutputKey := fmt.Sprintf("transcripts/%s/%s/raw_output.json", msg.UserID.String(), msg.VideoID.String())

	transcribeInput := transcribe.TranscriptionInput{
		JobID:          msg.JobID.String(),
		MediaS3URI:     mediaS3URI,
		OutputBucket:   p.bucket,
		OutputKey:      rawOutputKey,
		LanguageCode:   msg.Language,
		LocalAudioPath: localAudioPath,
	}

	result, err := p.provider.Transcribe(ctx, transcribeInput)
	if err != nil {
		return nil, "", fmt.Errorf("transcription provider error: %w", err)
	}
	return result, rawOutputKey, nil
}

func (p *Processor) persistResults(ctx context.Context, msg *queue.TranscriptionMessage, result *transcribe.TranscriptionResult, rawS3Key string, duration float64) error {
	transcriptID := uuid.New()

	return p.db.WithTx(ctx, func(tx pgx.Tx) error {
		// 1. Upsert transcript
		var actualTranscriptID uuid.UUID
		insertTranscriptSQL := `
			INSERT INTO transcripts (id, video_id, job_id, language, full_text, raw_s3_key)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (video_id) DO UPDATE
			SET job_id = $3, language = $4, full_text = $5, raw_s3_key = $6, updated_at = NOW()
			RETURNING id
		`
		err := tx.QueryRow(ctx, insertTranscriptSQL, transcriptID, msg.VideoID, msg.JobID, result.Language, result.FullText, rawS3Key).Scan(&actualTranscriptID)
		if err != nil {
			return fmt.Errorf("failed to upsert transcript: %w", err)
		}

		// 2. Clear old segments
		if _, err := tx.Exec(ctx, `DELETE FROM transcript_segments WHERE transcript_id = $1`, actualTranscriptID); err != nil {
			return fmt.Errorf("failed to delete old segments: %w", err)
		}

		// 3. Batch insert new segments
		insertSegmentSQL := `
			INSERT INTO transcript_segments (id, transcript_id, sequence_number, start_time, end_time, text, confidence)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
		for _, seg := range result.Segments {
			_, err := tx.Exec(ctx, insertSegmentSQL,
				uuid.New(),
				actualTranscriptID,
				seg.SequenceNumber,
				seg.StartTime,
				seg.EndTime,
				seg.Text,
				seg.Confidence,
			)
			if err != nil {
				return fmt.Errorf("failed to insert segment #%d: %w", seg.SequenceNumber, err)
			}
		}

		// 4. Mark job as completed
		updateJobSQL := `
			UPDATE transcription_jobs
			SET status = $1, completed_at = NOW(), updated_at = NOW()
			WHERE id = $2
		`
		if _, err := tx.Exec(ctx, updateJobSQL, jobs.StatusCompleted, msg.JobID); err != nil {
			return fmt.Errorf("failed to mark job completed: %w", err)
		}

		// 5. Mark video as completed with duration
		var durPtr *float64
		if duration > 0 {
			durPtr = &duration
		}
		updateVideoSQL := `
			UPDATE videos
			SET status = $1, duration_seconds = COALESCE($2, duration_seconds), updated_at = NOW()
			WHERE id = $3
		`
		if _, err := tx.Exec(ctx, updateVideoSQL, videos.StatusCompleted, durPtr, msg.VideoID); err != nil {
			return fmt.Errorf("failed to mark video completed: %w", err)
		}

		return nil
	})
}
