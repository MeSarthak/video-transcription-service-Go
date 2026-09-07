package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSQueue struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSQueue initializes an SQS queue client with default AWS configuration.
func NewSQSQueue(ctx context.Context, region, queueURL string) (*SQSQueue, error) {
	if queueURL == "" {
		return nil, errors.New("SQS queue URL cannot be empty")
	}

	var opts []func(*awsconfig.LoadOptions) error
	if region != "" {
		opts = append(opts, awsconfig.WithRegion(region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration for SQS: %w", err)
	}

	client := sqs.NewFromConfig(cfg)

	slog.Info("AWS SQS Queue initialized",
		slog.String("region", cfg.Region),
		slog.String("queue_url", queueURL),
	)

	return &SQSQueue{
		client:   client,
		queueURL: queueURL,
	}, nil
}

// Publish serializes and sends a transcription message to the SQS queue.
func (q *SQSQueue) Publish(ctx context.Context, msg *TranscriptionMessage) error {
	if msg == nil {
		return ErrInvalidMessage
	}

	bodyBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal transcription message: %w", err)
	}

	bodyStr := string(bodyBytes)
	input := &sqs.SendMessageInput{
		QueueUrl:    &q.queueURL,
		MessageBody: &bodyStr,
	}

	_, err = q.client.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to publish message to SQS: %w", err)
	}

	slog.Debug("Message published to SQS",
		slog.String("job_id", msg.JobID.String()),
		slog.String("video_id", msg.VideoID.String()),
	)

	return nil
}

// Receive long-polls for available messages from the SQS queue.
func (q *SQSQueue) Receive(ctx context.Context, maxMessages, waitTimeSeconds int32) ([]*ReceivedMessage, error) {
	if maxMessages <= 0 || maxMessages > 10 {
		maxMessages = 10
	}
	if waitTimeSeconds < 0 || waitTimeSeconds > 20 {
		waitTimeSeconds = 20
	}

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            &q.queueURL,
		MaxNumberOfMessages: maxMessages,
		WaitTimeSeconds:     waitTimeSeconds,
		MessageSystemAttributeNames: []types.MessageSystemAttributeName{
			types.MessageSystemAttributeNameApproximateReceiveCount,
		},
	}

	output, err := q.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages from SQS: %w", err)
	}

	var received []*ReceivedMessage
	for _, raw := range output.Messages {
		if raw.Body == nil || raw.ReceiptHandle == nil {
			continue
		}

		var msg TranscriptionMessage
		if err := json.Unmarshal([]byte(*raw.Body), &msg); err != nil {
			slog.Warn("Failed to unmarshal SQS message body, skipping", slog.Any("error", err))
			continue
		}

		receiveCount := 1
		if countStr, exists := raw.Attributes[string(types.MessageSystemAttributeNameApproximateReceiveCount)]; exists {
			if count, err := strconv.Atoi(countStr); err == nil {
				receiveCount = count
			}
		}

		received = append(received, &ReceivedMessage{
			Message:                 &msg,
			ReceiptHandle:           *raw.ReceiptHandle,
			ApproximateReceiveCount: receiveCount,
		})
	}

	return received, nil
}

// Delete removes a message from SQS after successful processing.
func (q *SQSQueue) Delete(ctx context.Context, receiptHandle string) error {
	if receiptHandle == "" {
		return errors.New("receipt handle cannot be empty")
	}

	input := &sqs.DeleteMessageInput{
		QueueUrl:      &q.queueURL,
		ReceiptHandle: &receiptHandle,
	}

	_, err := q.client.DeleteMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	return nil
}

// ChangeVisibility modifies the message's visibility timeout in SQS (used by worker heartbeat).
func (q *SQSQueue) ChangeVisibility(ctx context.Context, receiptHandle string, visibilityTimeoutSeconds int32) error {
	if receiptHandle == "" {
		return errors.New("receipt handle cannot be empty")
	}

	input := &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          &q.queueURL,
		ReceiptHandle:     &receiptHandle,
		VisibilityTimeout: visibilityTimeoutSeconds,
	}

	_, err := q.client.ChangeMessageVisibility(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to change SQS message visibility: %w", err)
	}

	return nil
}
