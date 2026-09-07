package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type S3Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
}

// NewS3Storage initializes an S3 client with standard AWS default credentials & region.
func NewS3Storage(ctx context.Context, region, bucket string) (*S3Storage, error) {
	if bucket == "" {
		return nil, errors.New("S3 bucket name cannot be empty")
	}

	var opts []func(*awsconfig.LoadOptions) error
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	slog.Info("AWS S3 Storage initialized",
		slog.String("region", cfg.Region),
		slog.String("bucket", bucket),
	)

	return &S3Storage{
		client:        client,
		presignClient: presignClient,
		bucket:        bucket,
	}, nil
}

// GeneratePresignedPutURL generates a temporary presigned URL for direct browser PUT upload.
func (s *S3Storage) GeneratePresignedPutURL(ctx context.Context, key, contentType string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", errors.New("storage key cannot be empty")
	}

	input := &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}

	if contentType != "" {
		input.ContentType = &contentType
	}

	req, err := s.presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned PUT URL: %w", err)
	}

	return req.URL, nil
}

// GeneratePresignedGetURL generates a temporary presigned URL for direct streaming or download.
func (s *S3Storage) GeneratePresignedGetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", errors.New("storage key cannot be empty")
	}

	input := &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}

	req, err := s.presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned GET URL: %w", err)
	}

	return req.URL, nil
}

// HeadObject retrieves metadata for an object without downloading the entire body.
func (s *S3Storage) HeadObject(ctx context.Context, key string) (*ObjectMetadata, error) {
	if key == "" {
		return nil, errors.New("storage key cannot be empty")
	}

	output, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})

	if err != nil {
		var notFound *types.NotFound
		var apiErr smithy.APIError
		if errors.As(err, &notFound) || (errors.As(err, &apiErr) && (apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey" || strings.Contains(apiErr.ErrorMessage(), "404"))) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("failed to head object %s: %w", key, err)
	}

	meta := &ObjectMetadata{
		Key:       key,
		SizeBytes: 0,
	}

	if output.ContentLength != nil {
		meta.SizeBytes = *output.ContentLength
	}
	if output.ContentType != nil {
		meta.ContentType = *output.ContentType
	}
	if output.ETag != nil {
		meta.ETag = *output.ETag
	}
	if output.LastModified != nil {
		meta.LastModified = *output.LastModified
	}

	return meta, nil
}

// DeleteObject deletes an object from the S3 bucket.
func (s *S3Storage) DeleteObject(ctx context.Context, key string) error {
	if key == "" {
		return errors.New("storage key cannot be empty")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}

	return nil
}

// Upload streams a file directly to S3.
func (s *S3Storage) Upload(ctx context.Context, key string, body io.Reader, contentType string) error {
	if key == "" {
		return errors.New("storage key cannot be empty")
	}

	input := &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   body,
	}

	if contentType != "" {
		input.ContentType = &contentType
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload object %s: %w", key, err)
	}

	return nil
}

// GetObject downloads a stream of the object from S3.
func (s *S3Storage) GetObject(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, errors.New("storage key cannot be empty")
	}

	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("failed to get object %s: %w", key, err)
	}

	return output.Body, nil
}
