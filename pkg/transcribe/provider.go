package transcribe

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
)

var (
	ErrTranscriptionFailed  = errors.New("transcription job failed")
	ErrTranscriptionTimeout = errors.New("transcription job timed out")
)

type SegmentResult struct {
	SequenceNumber int     `json:"sequence_number"`
	StartTime      float64 `json:"start_time"`
	EndTime        float64 `json:"end_time"`
	Text           string  `json:"text"`
	Confidence     float64 `json:"confidence"`
}

type TranscriptionResult struct {
	FullText      string          `json:"full_text"`
	Language      string          `json:"language"`
	RawJSONOutput []byte          `json:"raw_json_output"`
	Segments      []SegmentResult `json:"segments"`
}

type TranscriptionInput struct {
	JobID        string
	MediaS3URI   string // e.g. s3://bucket/audio/user/video/audio.wav
	OutputBucket string
	OutputKey    string
	LanguageCode string // e.g. "en-US"
}

// Provider defines the interface for speech-to-text service providers.
type Provider interface {
	Transcribe(ctx context.Context, input TranscriptionInput) (*TranscriptionResult, error)
}

type AWSTranscribeProvider struct {
	client     *transcribe.Client
	httpClient *http.Client
}

func NewAWSTranscribeProvider(ctx context.Context, region string) (*AWSTranscribeProvider, error) {
	var opts []func(*awsconfig.LoadOptions) error
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration for Transcribe: %w", err)
	}

	client := transcribe.NewFromConfig(cfg)

	return &AWSTranscribeProvider{
		client:     client,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Transcribe starts an AWS Transcribe job, polls until completion, and downloads/parses the result.
func (p *AWSTranscribeProvider) Transcribe(ctx context.Context, input TranscriptionInput) (*TranscriptionResult, error) {
	jobName := fmt.Sprintf("transcription-%s", input.JobID)
	lang := input.LanguageCode
	if lang == "" {
		lang = "en-US"
	}

	mediaFormat := types.MediaFormatWav
	if strings.HasSuffix(strings.ToLower(input.MediaS3URI), ".mp3") {
		mediaFormat = types.MediaFormatMp3
	}

	startInput := &transcribe.StartTranscriptionJobInput{
		TranscriptionJobName: &jobName,
		LanguageCode:         types.LanguageCode(lang),
		MediaFormat:          mediaFormat,
		Media: &types.Media{
			MediaFileUri: &input.MediaS3URI,
		},
	}

	if input.OutputBucket != "" {
		startInput.OutputBucketName = &input.OutputBucket
		if input.OutputKey != "" {
			startInput.OutputKey = &input.OutputKey
		}
	}

	slog.Info("Starting AWS Transcribe job",
		slog.String("job_name", jobName),
		slog.String("media_uri", input.MediaS3URI),
		slog.String("language", lang),
	)

	_, err := p.client.StartTranscriptionJob(ctx, startInput)
	if err != nil {
		return nil, fmt.Errorf("failed to start AWS Transcribe job: %w", err)
	}

	// Poll Transcribe job status with backoff
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			getJobInput := &transcribe.GetTranscriptionJobInput{
				TranscriptionJobName: &jobName,
			}

			output, err := p.client.GetTranscriptionJob(ctx, getJobInput)
			if err != nil {
				return nil, fmt.Errorf("failed to get transcription job status: %w", err)
			}

			status := output.TranscriptionJob.TranscriptionJobStatus
			slog.Debug("AWS Transcribe polling status", slog.String("job_name", jobName), slog.String("status", string(status)))

			switch status {
			case types.TranscriptionJobStatusCompleted:
				transcriptFileURI := output.TranscriptionJob.Transcript.TranscriptFileUri
				if transcriptFileURI == nil {
					return nil, errors.New("transcribe completed but no transcript URI returned")
				}

				// Download transcript JSON
				rawJSON, err := p.downloadTranscriptFile(ctx, *transcriptFileURI)
				if err != nil {
					return nil, fmt.Errorf("failed to download transcript JSON output: %w", err)
				}

				fullText, segments, err := ParseAWSOutput(rawJSON)
				if err != nil {
					return nil, fmt.Errorf("failed to parse transcript output: %w", err)
				}

				return &TranscriptionResult{
					FullText:      fullText,
					Language:      lang,
					RawJSONOutput: rawJSON,
					Segments:      segments,
				}, nil

			case types.TranscriptionJobStatusFailed:
				failureReason := "unknown error"
				if output.TranscriptionJob.FailureReason != nil {
					failureReason = *output.TranscriptionJob.FailureReason
				}
				return nil, fmt.Errorf("%w: %s", ErrTranscriptionFailed, failureReason)

			case types.TranscriptionJobStatusInProgress, types.TranscriptionJobStatusQueued:
				// Continue polling
				continue
			}
		}
	}
}

func (p *AWSTranscribeProvider) downloadTranscriptFile(ctx context.Context, uri string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP GET %s failed with status: %d", uri, resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// MockTranscribeProvider is an in-memory test provider for unit tests and local development.
type MockTranscribeProvider struct {
	MockResult *TranscriptionResult
	MockError  error
}

func NewMockTranscribeProvider() *MockTranscribeProvider {
	return &MockTranscribeProvider{
		MockResult: &TranscriptionResult{
			FullText: "Welcome to this video tutorial. In this video we will learn backend architecture.",
			Language: "en-US",
			Segments: []SegmentResult{
				{
					SequenceNumber: 1,
					StartTime:      0.0,
					EndTime:        2.5,
					Text:           "Welcome to this video tutorial.",
					Confidence:     0.98,
				},
				{
					SequenceNumber: 2,
					StartTime:      2.6,
					EndTime:        6.0,
					Text:           "In this video we will learn backend architecture.",
					Confidence:     0.97,
				},
			},
			RawJSONOutput: []byte(`{"mock": true}`),
		},
	}
}

func (m *MockTranscribeProvider) Transcribe(ctx context.Context, input TranscriptionInput) (*TranscriptionResult, error) {
	if m.MockError != nil {
		return nil, m.MockError
	}
	return m.MockResult, nil
}
