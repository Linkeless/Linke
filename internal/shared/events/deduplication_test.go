package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryEventDeduplicator(t *testing.T) {
	config := DeduplicationConfig{
		Strategy:        DeduplicationByEventID,
		TTL:             time.Minute,
		CleanupInterval: time.Second * 30,
		UseSignature:    false,
	}

	deduplicator := NewInMemoryEventDeduplicator(config)
	defer deduplicator.Close()

	ctx := context.Background()

	// Create test event
	event := NewBaseEvent("test.event", "test-service", map[string]any{
		"test": "data",
	})

	t.Run("First event should not be duplicate", func(t *testing.T) {
		isDuplicate, err := deduplicator.IsDuplicate(ctx, event)
		require.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Mark event as processed", func(t *testing.T) {
		err := deduplicator.MarkProcessed(ctx, event)
		require.NoError(t, err)
	})

	t.Run("Same event should now be duplicate", func(t *testing.T) {
		isDuplicate, err := deduplicator.IsDuplicate(ctx, event)
		require.NoError(t, err)
		assert.True(t, isDuplicate)
	})

	t.Run("Different event should not be duplicate", func(t *testing.T) {
		differentEvent := NewBaseEvent("test.event2", "test-service", map[string]any{
			"test": "data2",
		})

		isDuplicate, err := deduplicator.IsDuplicate(ctx, differentEvent)
		require.NoError(t, err)
		assert.False(t, isDuplicate)
	})

	t.Run("Get processed count", func(t *testing.T) {
		count, err := deduplicator.GetProcessedCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestDeduplicationStrategies(t *testing.T) {
	ctx := context.Background()

	t.Run("DeduplicationByContent", func(t *testing.T) {
		config := DeduplicationConfig{
			Strategy:        DeduplicationByContent,
			TTL:             time.Minute,
			CleanupInterval: time.Second * 30,
			UseSignature:    true,
		}

		deduplicator := NewInMemoryEventDeduplicator(config)
		defer deduplicator.Close()

		// Create two events with same content but different IDs
		event1 := NewBaseEvent("test.event", "test-service", map[string]any{
			"test": "data",
		})
		event1.ID = "id1"

		event2 := NewBaseEvent("test.event", "test-service", map[string]any{
			"test": "data",
		})
		event2.ID = "id2"
		// Ensure both events have the same timestamp for content-based deduplication
		event2.Time = event1.Time

		// First event should not be duplicate
		isDuplicate, err := deduplicator.IsDuplicate(ctx, event1)
		require.NoError(t, err)
		assert.False(t, isDuplicate)

		// Mark as processed
		err = deduplicator.MarkProcessed(ctx, event1)
		require.NoError(t, err)

		// Second event with same content should be duplicate
		isDuplicate, err = deduplicator.IsDuplicate(ctx, event2)
		require.NoError(t, err)
		assert.True(t, isDuplicate)
	})
}

func TestDeduplicationCleanup(t *testing.T) {
	config := DeduplicationConfig{
		Strategy:        DeduplicationByEventID,
		TTL:             time.Millisecond * 100, // Very short TTL for testing
		CleanupInterval: time.Millisecond * 50,
		UseSignature:    false,
	}

	deduplicator := NewInMemoryEventDeduplicator(config)
	defer deduplicator.Close()

	ctx := context.Background()

	// Create and process event
	event := NewBaseEvent("test.event", "test-service", map[string]any{
		"test": "data",
	})

	err := deduplicator.MarkProcessed(ctx, event)
	require.NoError(t, err)

	// Should be duplicate initially
	isDuplicate, err := deduplicator.IsDuplicate(ctx, event)
	require.NoError(t, err)
	assert.True(t, isDuplicate)

	// Wait for cleanup
	time.Sleep(time.Millisecond * 200)

	// Should not be duplicate after cleanup
	isDuplicate, err = deduplicator.IsDuplicate(ctx, event)
	require.NoError(t, err)
	assert.False(t, isDuplicate)
}

func TestDeduplicatingEventHandler(t *testing.T) {
	config := DefaultDeduplicationConfig()
	deduplicator := NewInMemoryEventDeduplicator(config)
	defer deduplicator.Close()

	// Create mock handler
	callCount := 0
	mockHandler := NewEventHandler([]string{"test.event"}, func(ctx context.Context, event Event) error {
		callCount++
		return nil
	})

	// Create deduplicating handler
	dedupHandler := NewDeduplicatingEventHandler("test-handler", mockHandler, deduplicator)

	ctx := context.Background()
	event := NewBaseEvent("test.event", "test-service", map[string]any{
		"test": "data",
	})

	t.Run("First call should be processed", func(t *testing.T) {
		err := dedupHandler.Handle(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("Second call should be skipped (duplicate)", func(t *testing.T) {
		err := dedupHandler.Handle(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, 1, callCount) // Should still be 1
	})

	t.Run("Different event should be processed", func(t *testing.T) {
		differentEvent := NewBaseEvent("test.event", "test-service", map[string]any{
			"test": "different data",
		})

		err := dedupHandler.Handle(ctx, differentEvent)
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})
}

func TestAtLeastOnceEventBus(t *testing.T) {
	// Create mock event bus
	publishCount := 0
	var publishError error
	mockEventBus := &MockEventBus{
		PublishFunc: func(ctx context.Context, event Event) error {
			publishCount++
			return publishError
		},
	}

	// Create deduplicator
	config := DefaultDeduplicationConfig()
	deduplicator := NewInMemoryEventDeduplicator(config)
	defer deduplicator.Close()

	// Create at-least-once event bus
	retryPolicy := AtLeastOnceRetryPolicy{
		MaxRetries:   2,
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Millisecond * 10,
		Multiplier:   2.0,
		EnableJitter: false,
	}
	aloBus := NewAtLeastOnceEventBus(mockEventBus, deduplicator, retryPolicy)

	ctx := context.Background()
	event := NewBaseEvent("test.event", "test-service", map[string]any{
		"test": "data",
	})

	t.Run("Successful publish on first try", func(t *testing.T) {
		publishCount = 0
		publishError = nil

		err := aloBus.Publish(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, 1, publishCount)
	})

	t.Run("Retry on failure", func(t *testing.T) {
		// Create different event for this test
		retryEvent := NewBaseEvent("test.retry", "test-service", map[string]any{
			"test": "retry data",
		})

		publishCount = 0
		publishError = fmt.Errorf("temporary failure")

		err := aloBus.Publish(ctx, retryEvent)
		require.Error(t, err)
		assert.Equal(t, 3, publishCount) // Initial + 2 retries
	})

	t.Run("Success after retry", func(t *testing.T) {
		// Create different event for this test
		retrySuccessEvent := NewBaseEvent("test.retry.success", "test-service", map[string]any{
			"test": "retry success data",
		})

		publishCount = 0
		callCount := 0
		mockEventBus.PublishFunc = func(ctx context.Context, event Event) error {
			callCount++
			if callCount < 2 {
				return fmt.Errorf("temporary failure")
			}
			return nil
		}

		err := aloBus.Publish(ctx, retrySuccessEvent)
		require.NoError(t, err)
		assert.Equal(t, 2, callCount)
	})
}

// MockEventBus is a mock implementation of EventBus for testing
type MockEventBus struct {
	PublishFunc      func(ctx context.Context, event Event) error
	PublishAsyncFunc func(ctx context.Context, event Event) error
	SubscribeFunc    func(eventTypes []string, handler EventHandler) error
	UnsubscribeFunc  func(eventTypes []string, handler EventHandler) error
	CloseFunc        func() error
}

func (m *MockEventBus) Publish(ctx context.Context, event Event) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, event)
	}
	return nil
}

func (m *MockEventBus) PublishAsync(ctx context.Context, event Event) error {
	if m.PublishAsyncFunc != nil {
		return m.PublishAsyncFunc(ctx, event)
	}
	return nil
}

func (m *MockEventBus) Subscribe(eventTypes []string, handler EventHandler) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(eventTypes, handler)
	}
	return nil
}

func (m *MockEventBus) Unsubscribe(eventTypes []string, handler EventHandler) error {
	if m.UnsubscribeFunc != nil {
		return m.UnsubscribeFunc(eventTypes, handler)
	}
	return nil
}

func (m *MockEventBus) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestEventDeduplicationManager(t *testing.T) {
	config := DefaultDeduplicationConfig()
	manager := NewEventDeduplicationManager(config)
	defer manager.Close()

	t.Run("Create new deduplicator", func(t *testing.T) {
		dedup1 := manager.GetOrCreateDeduplicator("test1")
		assert.NotNil(t, dedup1)

		// Should return same instance
		dedup2 := manager.GetOrCreateDeduplicator("test1")
		assert.Same(t, dedup1, dedup2)
	})

	t.Run("Get non-existent deduplicator", func(t *testing.T) {
		dedup, exists := manager.GetDeduplicator("non-existent")
		assert.Nil(t, dedup)
		assert.False(t, exists)
	})

	t.Run("Get existing deduplicator", func(t *testing.T) {
		dedup, exists := manager.GetDeduplicator("test1")
		assert.NotNil(t, dedup)
		assert.True(t, exists)
	})
}

func BenchmarkEventDeduplication(b *testing.B) {
	config := DefaultDeduplicationConfig()
	deduplicator := NewInMemoryEventDeduplicator(config)
	defer deduplicator.Close()

	ctx := context.Background()

	b.Run("IsDuplicate", func(b *testing.B) {
		event := NewBaseEvent("benchmark.event", "test-service", map[string]any{
			"test": "data",
		})

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = deduplicator.IsDuplicate(ctx, event)
		}
	})

	b.Run("MarkProcessed", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			event := NewBaseEvent("benchmark.event", "test-service", map[string]any{
				"test": "data",
				"id":   i,
			})
			_ = deduplicator.MarkProcessed(ctx, event)
		}
	})
}
