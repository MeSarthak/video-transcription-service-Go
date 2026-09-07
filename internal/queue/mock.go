package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// MockQueue is an in-memory thread-safe implementation of queue.Queue for unit tests and local emulation.
type MockQueue struct {
	mu       sync.Mutex
	messages []*ReceivedMessage
}

func NewMockQueue() *MockQueue {
	return &MockQueue{
		messages: make([]*ReceivedMessage, 0),
	}
}

func (m *MockQueue) Publish(ctx context.Context, msg *TranscriptionMessage) error {
	if msg == nil {
		return ErrInvalidMessage
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	receiptHandle := fmt.Sprintf("receipt-%s", uuid.New().String())
	m.messages = append(m.messages, &ReceivedMessage{
		Message:                 msg,
		ReceiptHandle:           receiptHandle,
		ApproximateReceiveCount: 1,
	})

	return nil
}

func (m *MockQueue) Receive(ctx context.Context, maxMessages, waitTimeSeconds int32) ([]*ReceivedMessage, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.messages) == 0 {
		return []*ReceivedMessage{}, nil
	}

	n := int(maxMessages)
	if n > len(m.messages) {
		n = len(m.messages)
	}

	result := make([]*ReceivedMessage, n)
	copy(result, m.messages[:n])

	return result, nil
}

func (m *MockQueue) Delete(ctx context.Context, receiptHandle string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, msg := range m.messages {
		if msg.ReceiptHandle == receiptHandle {
			m.messages = append(m.messages[:i], m.messages[i+1:]...)
			return nil
		}
	}

	return nil
}

func (m *MockQueue) ChangeVisibility(ctx context.Context, receiptHandle string, visibilityTimeoutSeconds int32) error {
	// No-op in memory
	return nil
}

func (m *MockQueue) Size() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}
