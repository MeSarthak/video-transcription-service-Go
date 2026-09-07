package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	ErrObjectNotFound = errors.New("object not found in storage")
	ErrStorageOpFailed = errors.New("storage operation failed")
)

type ObjectMetadata struct {
	Key          string
	SizeBytes    int64
	ContentType  string
	ETag         string
	LastModified time.Time
}

// Service defines the interface for object storage operations.
type Service interface {
	// GeneratePresignedPutURL generates a temporary presigned URL for direct client upload.
	GeneratePresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error)

	// GeneratePresignedGetURL generates a temporary presigned URL for direct client download or streaming.
	GeneratePresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// HeadObject retrieves metadata for an object without downloading its body.
	HeadObject(ctx context.Context, key string) (*ObjectMetadata, error)

	// DeleteObject removes an object from storage.
	DeleteObject(ctx context.Context, key string) error

	// Upload streams content directly to storage (useful for server-side uploads like extracted audio).
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error

	// GetObject returns a readable stream of the object body.
	GetObject(ctx context.Context, key string) (io.ReadCloser, error)
}
