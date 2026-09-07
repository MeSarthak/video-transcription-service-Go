package storage

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestNewS3Storage_EmptyBucket(t *testing.T) {
	ctx := context.Background()
	_, err := NewS3Storage(ctx, "us-east-1", "")
	if err == nil {
		t.Errorf("expected error when bucket is empty, got nil")
	}
}

func TestMockStorageService(t *testing.T) {
	var svc Service = NewMockStorage()
	ctx := context.Background()

	key := "videos/user1/vid1/original.mp4"
	content := "fake video payload data"

	// 1. Upload
	err := svc.Upload(ctx, key, strings.NewReader(content), "video/mp4")
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	// 2. HeadObject
	meta, err := svc.HeadObject(ctx, key)
	if err != nil {
		t.Fatalf("head object failed: %v", err)
	}
	if meta.SizeBytes != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), meta.SizeBytes)
	}

	// 3. Presigned URLs
	putURL, err := svc.GeneratePresignedPutURL(ctx, key, "video/mp4", 15*time.Minute)
	if err != nil || !strings.Contains(putURL, "upload") {
		t.Errorf("invalid presigned put url: %v", putURL)
	}

	getURL, err := svc.GeneratePresignedGetURL(ctx, key, 15*time.Minute)
	if err != nil || !strings.Contains(getURL, "playback") {
		t.Errorf("invalid presigned get url: %v", getURL)
	}

	// 4. GetObject
	reader, err := svc.GetObject(ctx, key)
	if err != nil {
		t.Fatalf("get object failed: %v", err)
	}
	data, _ := io.ReadAll(reader)
	reader.Close()
	if string(data) != content {
		t.Errorf("content mismatch: got %s", string(data))
	}

	// 5. DeleteObject
	if err := svc.DeleteObject(ctx, key); err != nil {
		t.Fatalf("delete object failed: %v", err)
	}

	// 6. Head after delete should fail
	_, err = svc.HeadObject(ctx, key)
	if err != ErrObjectNotFound {
		t.Errorf("expected ErrObjectNotFound, got %v", err)
	}
}
