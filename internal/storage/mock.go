package storage

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"
)

// MockStorage is an in-memory implementation of storage.Service for tests and local emulation.
type MockStorage struct {
	mu      sync.RWMutex
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
	m.mu.RLock()
	defer m.mu.RUnlock()

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
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Objects, key)
	delete(m.Meta, key)
	return nil
}

func (m *MockStorage) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

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
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.Objects[key]
	if !exists {
		return nil, ErrObjectNotFound
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}
