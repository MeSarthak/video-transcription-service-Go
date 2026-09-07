package ffmpeg

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestCreateScratchDirAndCleanup(t *testing.T) {
	jobID := uuid.New()

	scratchDir, cleanup, err := CreateScratchDir(jobID)
	if err != nil {
		t.Fatalf("expected scratch dir creation to succeed, got %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(scratchDir)
	if err != nil || !info.IsDir() {
		t.Fatalf("expected scratch directory to exist on disk")
	}

	// Create a dummy file inside
	dummyFile := filepath.Join(scratchDir, "sample.wav")
	if err := os.WriteFile(dummyFile, []byte("audio"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	// Execute cleanup
	cleanup()

	// Verify directory is deleted
	_, err = os.Stat(scratchDir)
	if !os.IsNotExist(err) {
		t.Errorf("expected scratch directory to be removed after cleanup")
	}
}

func TestMockExtractor(t *testing.T) {
	ctx := context.Background()
	extractor := NewMockExtractor(125.5)

	dur, err := extractor.GetDuration(ctx, "fake_video.mp4")
	if err != nil || dur != 125.5 {
		t.Errorf("expected duration 125.5, got %v", dur)
	}

	tempDir := t.TempDir()
	outAudio := filepath.Join(tempDir, "audio.wav")

	err = extractor.ExtractAudio(ctx, "fake_video.mp4", outAudio)
	if err != nil {
		t.Fatalf("mock extraction failed: %v", err)
	}

	data, err := os.ReadFile(outAudio)
	if err != nil || len(data) == 0 {
		t.Fatalf("expected dummy audio file to be written")
	}
}
