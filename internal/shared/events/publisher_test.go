package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryEventBus(t *testing.T) {
	t.Run("Subscribe and Publish work correctly", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		var receivedEvents []Event
		var mu sync.Mutex

		handler := NewEventHandler(
			[]string{EventTypeUserCreated},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				receivedEvents = append(receivedEvents, event)
				return nil
			},
		)

		// Subscribe
		err := bus.Subscribe([]string{EventTypeUserCreated}, handler)
		require.NoError(t, err)

		// Publish event
		event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
			"email": "test@example.com",
		})

		err = bus.Publish(context.Background(), event)
		require.NoError(t, err)

		// Wait a bit for async processing
		time.Sleep(10 * time.Millisecond)

		// Check received events
		mu.Lock()
		assert.Len(t, receivedEvents, 1)
		assert.Equal(t, event, receivedEvents[0])
		mu.Unlock()
	})

	t.Run("Multiple handlers for same event type", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		var receivedEvents1, receivedEvents2 []Event
		var mu1, mu2 sync.Mutex

		handler1 := NewEventHandler(
			[]string{EventTypeUserCreated},
			func(ctx context.Context, event Event) error {
				mu1.Lock()
				defer mu1.Unlock()
				receivedEvents1 = append(receivedEvents1, event)
				return nil
			},
		)

		handler2 := NewEventHandler(
			[]string{EventTypeUserCreated},
			func(ctx context.Context, event Event) error {
				mu2.Lock()
				defer mu2.Unlock()
				receivedEvents2 = append(receivedEvents2, event)
				return nil
			},
		)

		// Subscribe both handlers
		err := bus.Subscribe([]string{EventTypeUserCreated}, handler1)
		require.NoError(t, err)
		err = bus.Subscribe([]string{EventTypeUserCreated}, handler2)
		require.NoError(t, err)

		// Publish event
		event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
			"email": "test@example.com",
		})

		err = bus.Publish(context.Background(), event)
		require.NoError(t, err)

		// Wait a bit for async processing
		time.Sleep(10 * time.Millisecond)

		// Check both handlers received the event
		mu1.Lock()
		assert.Len(t, receivedEvents1, 1)
		mu1.Unlock()

		mu2.Lock()
		assert.Len(t, receivedEvents2, 1)
		mu2.Unlock()
	})

	t.Run("PublishAsync works correctly", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		var receivedEvents []Event
		var mu sync.Mutex

		handler := NewEventHandler(
			[]string{EventTypeUserCreated},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				receivedEvents = append(receivedEvents, event)
				return nil
			},
		)

		// Subscribe
		err := bus.Subscribe([]string{EventTypeUserCreated}, handler)
		require.NoError(t, err)

		// Publish event asynchronously
		event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
			"email": "test@example.com",
		})

		err = bus.PublishAsync(context.Background(), event)
		require.NoError(t, err)

		// Wait a bit for async processing
		time.Sleep(50 * time.Millisecond)

		// Check received events
		mu.Lock()
		assert.Len(t, receivedEvents, 1)
		assert.Equal(t, event, receivedEvents[0])
		mu.Unlock()
	})

	t.Run("Unsubscribe works correctly", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		var receivedEvents []Event
		var mu sync.Mutex

		handler := NewEventHandler(
			[]string{EventTypeUserCreated},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				receivedEvents = append(receivedEvents, event)
				return nil
			},
		)

		// Subscribe
		err := bus.Subscribe([]string{EventTypeUserCreated}, handler)
		require.NoError(t, err)

		// Unsubscribe
		err = bus.Unsubscribe([]string{EventTypeUserCreated}, handler)
		require.NoError(t, err)

		// Publish event
		event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
			"email": "test@example.com",
		})

		err = bus.Publish(context.Background(), event)
		require.NoError(t, err)

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Check no events were received
		mu.Lock()
		assert.Len(t, receivedEvents, 0)
		mu.Unlock()
	})

	t.Run("GetHandlerCount works correctly", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		assert.Equal(t, 0, bus.GetHandlerCount(EventTypeUserCreated))

		handler1 := NewEventHandler([]string{EventTypeUserCreated}, func(ctx context.Context, event Event) error { return nil })
		handler2 := NewEventHandler([]string{EventTypeUserCreated}, func(ctx context.Context, event Event) error { return nil })

		err := bus.Subscribe([]string{EventTypeUserCreated}, handler1)
		require.NoError(t, err)
		assert.Equal(t, 1, bus.GetHandlerCount(EventTypeUserCreated))

		err = bus.Subscribe([]string{EventTypeUserCreated}, handler2)
		require.NoError(t, err)
		assert.Equal(t, 2, bus.GetHandlerCount(EventTypeUserCreated))

		err = bus.Unsubscribe([]string{EventTypeUserCreated}, handler1)
		require.NoError(t, err)
		assert.Equal(t, 1, bus.GetHandlerCount(EventTypeUserCreated))
	})

	t.Run("ListEventTypes works correctly", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		handler := NewEventHandler(
			[]string{EventTypeUserCreated, EventTypeUserUpdated},
			func(ctx context.Context, event Event) error { return nil },
		)

		err := bus.Subscribe([]string{EventTypeUserCreated, EventTypeUserUpdated}, handler)
		require.NoError(t, err)

		eventTypes := bus.ListEventTypes()
		assert.Len(t, eventTypes, 2)
		assert.Contains(t, eventTypes, EventTypeUserCreated)
		assert.Contains(t, eventTypes, EventTypeUserUpdated)
	})

	t.Run("No handlers for event type", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
			"email": "test@example.com",
		})

		// Should not error when no handlers are registered
		err := bus.Publish(context.Background(), event)
		require.NoError(t, err)
	})
}

func TestEnhancedEventBus(t *testing.T) {
	t.Run("SubscribeWithConfig works correctly", func(t *testing.T) {
		bus := NewEnhancedEventBus()
		defer bus.Close()

		var receivedEvents []Event
		var mu sync.Mutex

		handler := NewEventHandler(
			[]string{EventTypeUserCreated},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				receivedEvents = append(receivedEvents, event)
				return nil
			},
		)

		config := DefaultSubscriberConfig()
		err := bus.SubscribeWithConfig("test-subscriber", []string{EventTypeUserCreated}, handler, config)
		require.NoError(t, err)

		// Publish event
		event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
			"email": "test@example.com",
		})

		err = bus.Publish(context.Background(), event)
		require.NoError(t, err)

		// Wait a bit for processing
		time.Sleep(10 * time.Millisecond)

		// Check received events
		mu.Lock()
		assert.Len(t, receivedEvents, 1)
		mu.Unlock()

		// Check subscriber stats
		stats := bus.GetSubscriberStats()
		assert.Len(t, stats, 1)
		assert.Contains(t, stats, "test-subscriber")
		assert.Equal(t, int64(1), stats["test-subscriber"].ProcessedCount)
	})

	t.Run("UnsubscribeByID works correctly", func(t *testing.T) {
		bus := NewEnhancedEventBus()
		defer bus.Close()

		handler := NewEventHandler(
			[]string{EventTypeUserCreated},
			func(ctx context.Context, event Event) error { return nil },
		)

		config := DefaultSubscriberConfig()
		err := bus.SubscribeWithConfig("test-subscriber", []string{EventTypeUserCreated}, handler, config)
		require.NoError(t, err)

		// Check subscriber exists
		stats := bus.GetSubscriberStats()
		assert.Len(t, stats, 1)

		// Unsubscribe by ID
		err = bus.UnsubscribeByID("test-subscriber")
		require.NoError(t, err)

		// Check subscriber is removed
		stats = bus.GetSubscriberStats()
		assert.Len(t, stats, 0)
	})

	t.Run("HealthCheck works correctly", func(t *testing.T) {
		bus := NewEnhancedEventBus()
		defer bus.Close()

		handler := NewEventHandler(
			[]string{EventTypeUserCreated},
			func(ctx context.Context, event Event) error { return nil },
		)

		config := DefaultSubscriberConfig()
		err := bus.SubscribeWithConfig("test-subscriber", []string{EventTypeUserCreated}, handler, config)
		require.NoError(t, err)

		health := bus.HealthCheck()
		assert.Len(t, health, 1)
		assert.Contains(t, health, "test-subscriber")
		assert.True(t, health["test-subscriber"]) // Should be healthy for new subscriber
	})
}

func TestEventMiddleware(t *testing.T) {
	t.Run("LoggingMiddleware processes events", func(t *testing.T) {
		middleware := LoggingMiddleware()

		var processedEvent Event
		next := func(ctx context.Context, event Event) error {
			processedEvent = event
			return nil
		}

		event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
			"email": "test@example.com",
		})

		err := middleware.Process(context.Background(), event, next)
		require.NoError(t, err)
		assert.Equal(t, event, processedEvent)
	})

	t.Run("RetryMiddleware retries on failure", func(t *testing.T) {
		middleware := RetryMiddleware(2)

		attemptCount := 0
		next := func(ctx context.Context, event Event) error {
			attemptCount++
			if attemptCount <= 2 {
				return assert.AnError
			}
			return nil
		}

		event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
			"email": "test@example.com",
		})

		err := middleware.Process(context.Background(), event, next)
		require.NoError(t, err)
		assert.Equal(t, 3, attemptCount) // Initial attempt + 2 retries
	})

	t.Run("EventProcessingMiddleware adds metadata", func(t *testing.T) {
		middleware := EventProcessingMiddleware()

		ctx := context.WithValue(context.Background(), "correlation_id", "test-correlation-123")

		var processedEvent Event
		next := func(ctx context.Context, event Event) error {
			processedEvent = event
			return nil
		}

		event := NewBaseEvent("test.event", "test-service", nil)

		err := middleware.Process(ctx, event, next)
		require.NoError(t, err)

		// Check correlation ID was set
		correlationID := extractCorrelationID(processedEvent)
		assert.Equal(t, "test-correlation-123", correlationID)

		// Check processed_at metadata was set
		if baseEvent, ok := processedEvent.(*BaseEvent); ok {
			processedAt, exists := baseEvent.GetMetadata("processed_at")
			assert.True(t, exists)
			assert.IsType(t, time.Time{}, processedAt)
		}
	})
}

// Benchmark tests
func BenchmarkEventBus(b *testing.B) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	handler := NewEventHandler(
		[]string{EventTypeUserCreated},
		func(ctx context.Context, event Event) error { return nil },
	)

	err := bus.Subscribe([]string{EventTypeUserCreated}, handler)
	require.NoError(b, err)

	event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
		"email": "test@example.com",
	})

	b.Run("Publish", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bus.Publish(context.Background(), event)
		}
	})

	b.Run("PublishAsync", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bus.PublishAsync(context.Background(), event)
		}
	})
}

func BenchmarkEventHandling(b *testing.B) {
	var processedCount int
	handler := NewEventHandler(
		[]string{EventTypeUserCreated},
		func(ctx context.Context, event Event) error {
			processedCount++
			return nil
		},
	)

	event := NewUserEvent(EventTypeUserCreated, 123, map[string]interface{}{
		"email": "test@example.com",
	})

	b.Run("DirectHandlerCall", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = handler.Handle(context.Background(), event)
		}
	})
}
