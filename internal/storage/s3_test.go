package storage

import (
	"bytes"
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

// MockStorage is an in-memory implementation of storage.Service for unit tests.
type MockStorage struct {
	Objects map[string][]byte
	Meta    map[string]*ObjectMetadata
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		Objects: make(map[string][]byte),
		Meta:    make(map[string]*ObjectMetadata),
	}
}

func (m *MockStorage) GeneratePresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", ErrStorageOpFailed
	}
	return "https://mock-s3.amazonaws.com/test-bucket/" + key + "?signed=true", nil
}

func (m *MockStorage) GeneratePresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", ErrStorageOpFailed
	}
	return "https://mock-s3.amazonaws.com/test-bucket/" + key + "?playback=true", nil
}

func (m *MockStorage) HeadObject(ctx context.Context, key string) (*ObjectMetadata, error) {
	data, exists := m.Objects[key]
	if !exists {
		return nil, ErrObjectNotFound
	}
	meta, exists := m.Meta[key]
	if !exists {
		return &ObjectMetadata{
			Key:          key,
			SizeBytes:    int64(len(data)),
			ContentType:  "video/mp4",
			LastModified: time.Now(),
		}, nil
	}
	return meta, nil
}

func (m *MockStorage) DeleteObject(ctx context.Context, key string) error {
	delete(m.Objects, key)
	delete(m.Meta, key)
	return nil
}

func (m *MockStorage) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	m.Objects[key] = data
	m.Meta[key] = &ObjectMetadata{
		Key:          key,
		SizeBytes:    int64(len(data)),
		ContentType:  contentType,
		LastModified: time.Now(),
	}
	return nil
}

func (m *MockStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	data, exists := m.Objects[key]
	if !exists {
		return nil, ErrObjectNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
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
	if err != nil || !strings.Contains(putURL, key) {
		t.Errorf("invalid presigned put url: %v", putURL)
	}

	getURL, err := svc.GeneratePresignedGetURL(ctx, key, 15*time.Minute)
	if err != nil || !strings.Contains(getURL, key) {
		t.Errorf("invalid presigned get url: %v", getURL)
	}

	// 4. GetObject
	reader, err := svc.GetObject(ctx, key)
	if err != nil {
		t.Fatalf("get object failed: %v", err)
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
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
