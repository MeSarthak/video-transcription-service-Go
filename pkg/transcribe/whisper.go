package transcribe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type WhisperProvider struct {
	apiKey     string
	apiURL     string
	model      string
	httpClient *http.Client
}

func NewWhisperProvider(apiKey, apiURL, model string) *WhisperProvider {
	if apiURL == "" {
		apiURL = "https://api.groq.com/openai/v1/audio/transcriptions"
	}
	if model == "" {
		if strings.Contains(apiURL, "groq") {
			model = "whisper-large-v3"
		} else {
			model = "whisper-1"
		}
	}
	return &WhisperProvider{
		apiKey: apiKey,
		apiURL: apiURL,
		model:  model,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

type whisperVerboseResponse struct {
	Text     string `json:"text"`
	Language string `json:"language"`
	Segments []struct {
		ID         int     `json:"id"`
		Start      float64 `json:"start"`
		End        float64 `json:"end"`
		Text       string  `json:"text"`
		AvgLogprob float64 `json:"avg_logprob"`
	} `json:"segments"`
}

func (w *WhisperProvider) Transcribe(ctx context.Context, input TranscriptionInput) (*TranscriptionResult, error) {
	audioPath := input.LocalAudioPath
	if audioPath == "" {
		return nil, fmt.Errorf("local audio path is required for Whisper speech-to-text")
	}

	fi, err := os.Stat(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access audio file %s: %w", audioPath, err)
	}

	// If audio file exceeds 20MB, compress to 32kbps mono MP3 via FFmpeg so it stays well within the 25MB API limit
	sendPath := audioPath
	if fi.Size() > 20*1024*1024 {
		mp3Path := filepath.Join(filepath.Dir(audioPath), "compressed_audio.mp3")
		slog.Info("Compressing audio for Whisper API", slog.Int64("original_size", fi.Size()), slog.String("target", mp3Path))
		cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", audioPath, "-codec:a", "libmp3lame", "-b:a", "32k", "-ac", "1", mp3Path)
		if cErr := cmd.Run(); cErr == nil {
			sendPath = mp3Path
			defer os.Remove(mp3Path)
		} else {
			slog.Warn("Audio compression warning, attempting original file", slog.Any("error", cErr))
		}
	}

	file, err := os.Open(sendPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(sendPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart form file: %w", err)
	}
	if _, err = io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy audio bytes to form: %w", err)
	}

	_ = writer.WriteField("model", w.model)
	_ = writer.WriteField("response_format", "verbose_json")
	if input.LanguageCode != "" {
		lang := strings.ToLower(strings.Split(input.LanguageCode, "-")[0])
		_ = writer.WriteField("language", lang)
	}

	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.apiURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create whisper HTTP request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	slog.Info("Sending audio to Whisper Speech-to-Text API",
		slog.String("url", w.apiURL),
		slog.String("model", w.model),
		slog.String("job_id", input.JobID),
	)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("whisper API HTTP call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read whisper API response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("whisper API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var whisperResp whisperVerboseResponse
	if err := json.Unmarshal(respBody, &whisperResp); err != nil {
		return nil, fmt.Errorf("failed to parse whisper verbose JSON response: %w", err)
	}

	segments := make([]SegmentResult, len(whisperResp.Segments))
	for i, seg := range whisperResp.Segments {
		conf := 0.95
		if seg.AvgLogprob != 0 {
			conf = math.Exp(seg.AvgLogprob)
			if conf > 1.0 {
				conf = 0.99
			} else if conf < 0.1 {
				conf = 0.5
			}
		}
		segments[i] = SegmentResult{
			SequenceNumber: i + 1,
			StartTime:      seg.Start,
			EndTime:        seg.End,
			Text:           strings.TrimSpace(seg.Text),
			Confidence:     conf,
		}
	}

	lang := whisperResp.Language
	if lang == "" {
		lang = input.LanguageCode
	}

	slog.Info("Whisper transcription successfully completed",
		slog.String("job_id", input.JobID),
		slog.Int("segments_count", len(segments)),
		slog.String("language", lang),
	)

	return &TranscriptionResult{
		FullText:      strings.TrimSpace(whisperResp.Text),
		Language:      lang,
		RawJSONOutput: respBody,
		Segments:      segments,
	}, nil
}
