package storage

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MockStorage is a disk-backed and in-memory implementation of storage.Service for tests and local emulation.
type MockStorage struct {
	mu      sync.RWMutex
	baseDir string
	Objects map[string][]byte
	Meta    map[string]*ObjectMetadata
}

func NewMockStorage() *MockStorage {
	baseDir := os.Getenv("STORAGE_DIR")
	if baseDir == "" {
		baseDir = filepath.Join(os.TempDir(), "transcription_storage")
	}
	_ = os.MkdirAll(baseDir, 0755)

	return &MockStorage{
		baseDir: baseDir,
		Objects: make(map[string][]byte),
		Meta:    make(map[string]*ObjectMetadata),
	}
}

func (m *MockStorage) filePath(key string) string {
	return filepath.Join(m.baseDir, filepath.FromSlash(key))
}

func (m *MockStorage) GeneratePresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", ErrStorageOpFailed
	}
	return "/api/v1/storage/upload?key=" + url.QueryEscape(key), nil
}

func (m *MockStorage) GeneratePresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", ErrStorageOpFailed
	}
	return "/api/v1/storage/playback?key=" + url.QueryEscape(key), nil
}

func (m *MockStorage) HeadObject(ctx context.Context, key string) (*ObjectMetadata, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check in-memory metadata
	if meta, exists := m.Meta[key]; exists {
		return meta, nil
	}

	// Check disk
	fp := m.filePath(key)
	fi, err := os.Stat(fp)
	if err == nil {
		return &ObjectMetadata{
			Key:          key,
			SizeBytes:    fi.Size(),
			ContentType:  "video/mp4",
			LastModified: fi.ModTime(),
		}, nil
	}

	// Check in-memory byte slice
	if data, exists := m.Objects[key]; exists {
		return &ObjectMetadata{
			Key:          key,
			SizeBytes:    int64(len(data)),
			ContentType:  "video/mp4",
			LastModified: time.Now(),
		}, nil
	}

	return nil, ErrObjectNotFound
}

func (m *MockStorage) DeleteObject(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Objects, key)
	delete(m.Meta, key)
	if err := os.Remove(m.filePath(key)); err != nil && !os.IsNotExist(err) {
		return err
	}
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

	// Also persist to disk for worker process sharing
	fp := m.filePath(key)
	_ = os.MkdirAll(filepath.Dir(fp), 0755)
	_ = os.WriteFile(fp, data, 0644)

	return nil
}

type readSeekCloser struct {
	io.ReadSeeker
}

func (r *readSeekCloser) Close() error {
	return nil
}

func (m *MockStorage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check disk first for persisted file (returns *os.File which implements io.ReadSeekCloser)
	fp := m.filePath(key)
	if f, err := os.Open(fp); err == nil {
		return f, nil
	}

	// Check in-memory
	if data, exists := m.Objects[key]; exists {
		return &readSeekCloser{bytes.NewReader(data)}, nil
	}

	return nil, ErrObjectNotFound
}
