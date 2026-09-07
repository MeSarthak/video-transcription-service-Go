package transcribe

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWhisperProvider_Transcribe_Success(t *testing.T) {
	// Create mock server
	mockResponse := `{"text":"Hello world this is a test.","language":"english","segments":[{"id":0,"start":0.0,"end":2.5,"text":"Hello world","avg_logprob":-0.15},{"id":1,"start":2.5,"end":5.0,"text":"this is a test.","avg_logprob":-0.05}]}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
	defer ts.Close()

	// Create temp audio file
	tmpDir := t.TempDir()
	audioFile := filepath.Join(tmpDir, "test.wav")
	if err := os.WriteFile(audioFile, []byte("fake-wav-header-and-samples"), 0644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	provider := NewWhisperProvider("test-key", ts.URL, "whisper-large-v3")

	result, err := provider.Transcribe(context.Background(), TranscriptionInput{
		JobID:          "test-job-123",
		LocalAudioPath: audioFile,
		LanguageCode:   "en-US",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if result.FullText != "Hello world this is a test." {
		t.Errorf("unexpected full text: %s", result.FullText)
	}

	if len(result.Segments) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(result.Segments))
	}

	if result.Segments[0].Text != "Hello world" || result.Segments[0].StartTime != 0.0 || result.Segments[0].EndTime != 2.5 {
		t.Errorf("segment 0 mismatch: %+v", result.Segments[0])
	}
}

func TestWhisperProvider_Transcribe_MissingFile(t *testing.T) {
	provider := NewWhisperProvider("test-key", "https://api.groq.com", "whisper-large-v3")

	_, err := provider.Transcribe(context.Background(), TranscriptionInput{
		JobID:          "test-job",
		LocalAudioPath: "",
	})

	if err == nil {
		t.Errorf("expected error when LocalAudioPath is empty, got nil")
	}
}

func TestWhisperProvider_Transcribe_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error": "rate limit exceeded"}`, http.StatusTooManyRequests)
	}))
	defer ts.Close()

	tmpDir := t.TempDir()
	audioFile := filepath.Join(tmpDir, "test.wav")
	_ = os.WriteFile(audioFile, []byte("fake-wav"), 0644)

	provider := NewWhisperProvider("test-key", ts.URL, "whisper-large-v3")

	_, err := provider.Transcribe(context.Background(), TranscriptionInput{
		JobID:          "test-job",
		LocalAudioPath: audioFile,
	})

	if err == nil {
		t.Errorf("expected error on HTTP 429, got nil")
	}
}
