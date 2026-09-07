package ffmpeg

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrFFmpegNotFound   = errors.New("ffmpeg executable not found in PATH")
	ErrFFprobeNotFound  = errors.New("ffprobe executable not found in PATH")
	ErrExtractionFailed = errors.New("audio extraction failed")
)

type Extractor interface {
	ExtractAudio(ctx context.Context, inputVideoPath, outputAudioPath string) error
	GetDuration(ctx context.Context, mediaPath string) (float64, error)
}

type CLIFFmpeg struct {
	ffmpegPath  string
	ffprobePath string
	timeout     time.Duration
}

func NewCLIFFmpeg(timeout time.Duration) *CLIFFmpeg {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	ffmpegPath, _ := exec.LookPath("ffmpeg")
	ffprobePath, _ := exec.LookPath("ffprobe")

	return &CLIFFmpeg{
		ffmpegPath:  ffmpegPath,
		ffprobePath: ffprobePath,
		timeout:     timeout,
	}
}

// ExtractAudio extracts normalized 16kHz mono PCM WAV audio from a video file.
func (f *CLIFFmpeg) ExtractAudio(ctx context.Context, inputVideoPath, outputAudioPath string) error {
	if f.ffmpegPath == "" {
		return ErrFFmpegNotFound
	}

	execCtx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	// ffmpeg -y -i <input> -vn -acodec pcm_s16le -ar 16000 -ac 1 <output.wav>
	args := []string{
		"-y",
		"-i", inputVideoPath,
		"-vn",
		"-acodec", "pcm_s16le",
		"-ar", "16000",
		"-ac", "1",
		outputAudioPath,
	}

	cmd := exec.CommandContext(execCtx, f.ffmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	slog.Info("Running FFmpeg audio extraction",
		slog.String("input", inputVideoPath),
		slog.String("output", outputAudioPath),
	)

	if err := cmd.Run(); err != nil {
		errStr := stderr.String()
		if strings.Contains(errStr, "Output file does not contain any stream") || strings.Contains(errStr, "does not contain any stream") {
			slog.Warn("No audio stream detected in video, generating silent audio track fallback", slog.String("input", inputVideoPath))
			dur, _ := f.GetDuration(ctx, inputVideoPath)
			if dur <= 0 {
				dur = 1.0
			}
			silentArgs := []string{
				"-y",
				"-f", "lavfi",
				"-i", "anullsrc=r=16000:cl=mono",
				"-t", fmt.Sprintf("%.3f", dur),
				"-acodec", "pcm_s16le",
				outputAudioPath,
			}
			silentCmd := exec.CommandContext(execCtx, f.ffmpegPath, silentArgs...)
			var silentStderr bytes.Buffer
			silentCmd.Stderr = &silentStderr
			sErr := silentCmd.Run()
			if sErr == nil {
				return nil
			}
			slog.Error("Failed to generate silent audio fallback", slog.Any("error", sErr), slog.String("stderr", silentStderr.String()))
		}
		return fmt.Errorf("%w: %v, stderr: %s", ErrExtractionFailed, err, errStr)
	}

	return nil
}

// GetDuration probes the media duration in seconds using ffprobe.
func (f *CLIFFmpeg) GetDuration(ctx context.Context, mediaPath string) (float64, error) {
	if f.ffprobePath == "" {
		return 0, ErrFFprobeNotFound
	}

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// ffprobe -v error -show_entries format=duration -of default=noprint_wrappers=1:nokey=1 <mediaPath>
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		mediaPath,
	}

	cmd := exec.CommandContext(execCtx, f.ffprobePath, args...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe duration check failed: %v, stderr: %s", err, stderr.String())
	}

	durStr := strings.TrimSpace(out.String())
	dur, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration '%s': %w", durStr, err)
	}

	return dur, nil
}

// CreateScratchDir creates an isolated workspace directory for a job and returns a cleanup func.
func CreateScratchDir(jobID uuid.UUID) (string, func(), error) {
	baseTemp := os.TempDir()
	jobDir := filepath.Join(baseTemp, "transcription-jobs", jobID.String())

	if err := os.MkdirAll(jobDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create scratch directory: %w", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(jobDir); err != nil {
			slog.Warn("Failed to clean up scratch directory", slog.String("dir", jobDir), slog.Any("error", err))
		} else {
			slog.Debug("Scratch directory cleaned up", slog.String("dir", jobDir))
		}
	}

	return jobDir, cleanup, nil
}

// MockExtractor is an in-memory/file test double for FFmpeg operations.
type MockExtractor struct {
	ExtractedFiles map[string]string
	DurationResult float64
}

func NewMockExtractor(dur float64) *MockExtractor {
	return &MockExtractor{
		ExtractedFiles: make(map[string]string),
		DurationResult: dur,
	}
}

func (m *MockExtractor) ExtractAudio(ctx context.Context, inputVideoPath, outputAudioPath string) error {
	m.ExtractedFiles[inputVideoPath] = outputAudioPath
	// Create a dummy output WAV file for tests
	return os.WriteFile(outputAudioPath, []byte("RIFFdummywavheaderandpcmdata"), 0644)
}

func (m *MockExtractor) GetDuration(ctx context.Context, mediaPath string) (float64, error) {
	return m.DurationResult, nil
}
