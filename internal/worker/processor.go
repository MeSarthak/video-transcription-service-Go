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

// ProcessJob executes the full video download -> FFmpeg audio extraction -> S3 audio upload -> speech-to-text -> DB persistence pipeline.
func (p *Processor) ProcessJob(ctx context.Context, msg *queue.TranscriptionMessage) error {
	slog.Info("Starting end-to-end transcription processing pipeline",
		slog.String("job_id", msg.JobID.String()),
		slog.String("video_id", msg.VideoID.String()),
	)

	// 1. Create Isolated Scratch Workspace Directory with defer cleanup
	scratchDir, cleanup, err := ffmpeg.CreateScratchDir(msg.JobID)
	if err != nil {
		return fmt.Errorf("failed to create scratch directory: %w", err)
	}
	defer cleanup()

	// 2. Download original video from S3 to local scratch directory
	localVideoPath := filepath.Join(scratchDir, "original_video")
	videoReader, err := p.storage.GetObject(ctx, msg.S3VideoKey)
	if err != nil {
		return fmt.Errorf("failed to download video from S3 (%s): %w", msg.S3VideoKey, err)
	}
	defer videoReader.Close()

	videoFile, err := os.Create(localVideoPath)
	if err != nil {
		return fmt.Errorf("failed to create local video file: %w", err)
	}
	if _, err := io.Copy(videoFile, videoReader); err != nil {
		videoFile.Close()
		return fmt.Errorf("failed to write video to scratch disk: %w", err)
	}
	videoFile.Close()

	// 3. Probe video duration with ffprobe
	duration, _ := p.extractor.GetDuration(ctx, localVideoPath)

	// 4. Extract normalized 16kHz mono audio WAV with FFmpeg
	localAudioPath := filepath.Join(scratchDir, "extracted_audio.wav")
	if err := p.extractor.ExtractAudio(ctx, localVideoPath, localAudioPath); err != nil {
		return fmt.Errorf("FFmpeg audio extraction failed: %w", err)
	}

	// 5. Upload extracted audio to S3
	audioStorageKey := fmt.Sprintf("audio/%s/%s/audio.wav", msg.UserID.String(), msg.VideoID.String())
	audioData, err := os.ReadFile(localAudioPath)
	if err != nil {
		return fmt.Errorf("failed to read extracted audio: %w", err)
	}

	if err := p.storage.Upload(ctx, audioStorageKey, bytes.NewReader(audioData), "audio/wav"); err != nil {
		return fmt.Errorf("failed to upload extracted audio to S3: %w", err)
	}

	// 6. Call Transcription Provider (AWS Transcribe / Mock)
	mediaS3URI := fmt.Sprintf("s3://%s/%s", p.bucket, audioStorageKey)
	transcribeInput := transcribe.TranscriptionInput{
		JobID:          msg.JobID.String(),
		MediaS3URI:     mediaS3URI,
		OutputBucket:   p.bucket,
		OutputKey:      fmt.Sprintf("transcripts/%s/%s/raw_output.json", msg.UserID.String(), msg.VideoID.String()),
		LanguageCode:   msg.Language,
		LocalAudioPath: localAudioPath,
	}

	result, err := p.provider.Transcribe(ctx, transcribeInput)
	if err != nil {
		return fmt.Errorf("transcription provider error: %w", err)
	}

	// 7. Persist Transcripts and Segments inside a Database Transaction
	transcriptID := uuid.New()
	rawS3Key := transcribeInput.OutputKey

	err = p.db.WithTx(ctx, func(tx pgx.Tx) error {
		// Insert or update transcript, returning the actual ID in the table
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
			return fmt.Errorf("failed to insert transcript: %w", err)
		}

		// Delete any existing segments for this transcript
		_, _ = tx.Exec(ctx, `DELETE FROM transcript_segments WHERE transcript_id = $1`, actualTranscriptID)

		// Insert segments
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
				return fmt.Errorf("failed to insert transcript segment #%d: %w", seg.SequenceNumber, err)
			}
		}

		// Update Job Status to Completed
		updateJobSQL := `
			UPDATE transcription_jobs
			SET status = $1, completed_at = NOW(), updated_at = NOW()
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, updateJobSQL, jobs.StatusCompleted, msg.JobID)
		if err != nil {
			return fmt.Errorf("failed to mark job completed: %w", err)
		}

		// Update Video Status to Completed & record duration
		updateVideoSQL := `
			UPDATE videos
			SET status = $1, duration_seconds = COALESCE($2, duration_seconds), updated_at = NOW()
			WHERE id = $3
		`
		var durPtr *float64
		if duration > 0 {
			durPtr = &duration
		}
		_, err = tx.Exec(ctx, updateVideoSQL, videos.StatusCompleted, durPtr, msg.VideoID)
		if err != nil {
			return fmt.Errorf("failed to mark video completed: %w", err)
		}

		return nil
	})

	if err != nil {
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
