package events

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventMetricsCollector(t *testing.T) {
	metrics := NewEventMetricsCollector(time.Hour, 12)

	t.Run("Initial state", func(t *testing.T) {
		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(0), snapshot.TotalEventsPublished)
		assert.Equal(t, int64(0), snapshot.TotalEventsProcessed)
		assert.Equal(t, int64(0), snapshot.TotalEventsFailed)
		assert.Equal(t, float64(0), snapshot.SuccessRate)
	})

	t.Run("Record event published", func(t *testing.T) {
		metrics.RecordEventPublished("test.event")
		metrics.RecordEventPublished("test.event2")

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(2), snapshot.TotalEventsPublished)
		assert.Contains(t, snapshot.EventTypeMetrics, "test.event")
		assert.Contains(t, snapshot.EventTypeMetrics, "test.event2")
		assert.Equal(t, int64(1), snapshot.EventTypeMetrics["test.event"].PublishedCount)
	})

	t.Run("Record event processed", func(t *testing.T) {
		processingTime := time.Millisecond * 100
		metrics.RecordEventProcessed("test.event", "handler1", processingTime)

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.TotalEventsProcessed)
		assert.Equal(t, processingTime, snapshot.AverageProcessingTime)
		assert.Equal(t, processingTime, snapshot.MinProcessingTime)
		assert.Equal(t, processingTime, snapshot.MaxProcessingTime)

		// Check event type metrics
		eventMetrics := snapshot.EventTypeMetrics["test.event"]
		assert.Equal(t, int64(1), eventMetrics.ProcessedCount)
		assert.Equal(t, processingTime, eventMetrics.AverageProcessingTime)

		// Check handler metrics
		handlerMetrics := snapshot.HandlerMetrics["handler1"]
		assert.Equal(t, int64(1), handlerMetrics.ProcessedCount)
		assert.Equal(t, processingTime, handlerMetrics.AverageProcessingTime)
	})

	t.Run("Record event failed", func(t *testing.T) {
		err := fmt.Errorf("test error")
		metrics.RecordEventFailed("test.event", "handler1", err)

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.TotalEventsFailed)
		assert.Equal(t, float64(50), snapshot.SuccessRate) // 1 success, 1 failure = 50%

		// Check event type metrics
		eventMetrics := snapshot.EventTypeMetrics["test.event"]
		assert.Equal(t, int64(1), eventMetrics.FailedCount)

		// Check handler metrics
		handlerMetrics := snapshot.HandlerMetrics["handler1"]
		assert.Equal(t, int64(1), handlerMetrics.FailedCount)
	})

	t.Run("Record event moved to dead letter", func(t *testing.T) {
		metrics.RecordEventMovedToDeadLetter("test.event")

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.TotalEventsInDeadLetter)

		eventMetrics := snapshot.EventTypeMetrics["test.event"]
		assert.Equal(t, int64(1), eventMetrics.DeadLetterCount)
	})

	t.Run("Record circuit breaker trip", func(t *testing.T) {
		metrics.RecordCircuitBreakerTrip("handler1")

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.CircuitBreakerTrips)

		handlerMetrics := snapshot.HandlerMetrics["handler1"]
		assert.Equal(t, int64(1), handlerMetrics.CircuitBreakerTrips)
	})

	t.Run("Reset metrics", func(t *testing.T) {
		metrics.Reset()

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(0), snapshot.TotalEventsPublished)
		assert.Equal(t, int64(0), snapshot.TotalEventsProcessed)
		assert.Equal(t, int64(0), snapshot.TotalEventsFailed)
		assert.Empty(t, snapshot.EventTypeMetrics)
		assert.Empty(t, snapshot.HandlerMetrics)
	})
}

func TestMetricsEventHandler(t *testing.T) {
	metrics := NewEventMetricsCollector(time.Hour, 12)

	callCount := 0
	var handlerError error
	processingDelay := time.Millisecond * 10

	mockHandler := NewEventHandler([]string{"test.event"}, func(ctx context.Context, event Event) error {
		callCount++
		time.Sleep(processingDelay)
		return handlerError
	})

	metricsHandler := NewMetricsEventHandler("test-handler", mockHandler, metrics)

	ctx := context.Background()
	event := NewBaseEvent("test.event", "test-service", map[string]interface{}{
		"test": "data",
	})

	t.Run("Successful event handling should record metrics", func(t *testing.T) {
		callCount = 0
		handlerError = nil

		err := metricsHandler.Handle(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.TotalEventsProcessed)
		assert.Equal(t, int64(0), snapshot.TotalEventsFailed)
		assert.True(t, snapshot.AverageProcessingTime >= processingDelay)
	})

	t.Run("Failed event handling should record failure metrics", func(t *testing.T) {
		callCount = 0
		handlerError = fmt.Errorf("handler error")

		err := metricsHandler.Handle(ctx, event)
		require.Error(t, err)
		assert.Equal(t, 1, callCount)

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.TotalEventsProcessed) // From previous test
		assert.Equal(t, int64(1), snapshot.TotalEventsFailed)
		assert.Equal(t, float64(50), snapshot.SuccessRate)
	})

	t.Run("Event types should be returned correctly", func(t *testing.T) {
		eventTypes := metricsHandler.EventTypes()
		assert.Equal(t, []string{"test.event"}, eventTypes)
	})

	t.Run("Get metrics should return metrics instance", func(t *testing.T) {
		returnedMetrics := metricsHandler.GetMetrics()
		assert.Same(t, metrics, returnedMetrics)
	})
}

func TestMetricsEventBus(t *testing.T) {
	metrics := NewEventMetricsCollector(time.Hour, 12)

	publishCount := 0
	mockEventBus := &MockEventBus{
		PublishFunc: func(ctx context.Context, event Event) error {
			publishCount++
			return nil
		},
		PublishAsyncFunc: func(ctx context.Context, event Event) error {
			publishCount++
			return nil
		},
	}

	metricsEventBus := NewMetricsEventBus(mockEventBus, metrics)

	ctx := context.Background()
	event := NewBaseEvent("test.event", "test-service", map[string]interface{}{
		"test": "data",
	})

	t.Run("Publish should record metrics", func(t *testing.T) {
		publishCount = 0

		err := metricsEventBus.Publish(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, 1, publishCount)

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.TotalEventsPublished)
	})

	t.Run("PublishAsync should record metrics", func(t *testing.T) {
		publishCount = 0

		err := metricsEventBus.PublishAsync(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, 1, publishCount)

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(2), snapshot.TotalEventsPublished) // Previous + this one
	})
}

func TestEventSystemHealthChecker(t *testing.T) {
	metrics := NewEventMetricsCollector(time.Hour, 12)
	deadLetterQueue := NewInMemoryDeadLetterQueue()
	circuitBreakerManager := NewCircuitBreakerManager()

	healthChecker := NewEventSystemHealthChecker(metrics, deadLetterQueue, circuitBreakerManager)

	ctx := context.Background()

	t.Run("Healthy system", func(t *testing.T) {
		// Record some successful events
		for i := 0; i < 10; i++ {
			metrics.RecordEventProcessed("test.event", "handler1", time.Millisecond*10)
		}

		health := healthChecker.CheckHealth(ctx)
		assert.True(t, health.IsHealthy)
		assert.Empty(t, health.Issues)
		assert.True(t, health.EventBusHealth)
		assert.Equal(t, 0, health.CircuitBreakersOpen)
	})

	t.Run("High error rate should be unhealthy", func(t *testing.T) {
		metrics.Reset()

		// Record mostly failures
		metrics.RecordEventProcessed("test.event", "handler1", time.Millisecond*10)
		for i := 0; i < 20; i++ {
			metrics.RecordEventFailed("test.event", "handler1", fmt.Errorf("error"))
		}

		health := healthChecker.CheckHealth(ctx)
		assert.False(t, health.IsHealthy)
		assert.NotEmpty(t, health.Issues)
		assert.Contains(t, health.Issues[0], "High error rate")
	})

	t.Run("Slow processing should be unhealthy", func(t *testing.T) {
		metrics.Reset()

		// Record slow processing
		metrics.RecordEventProcessed("test.event", "handler1", time.Second*10)

		health := healthChecker.CheckHealth(ctx)
		assert.False(t, health.IsHealthy)
		assert.NotEmpty(t, health.Issues)
		assert.Contains(t, health.Issues[0], "Slow processing")
	})

	t.Run("Too many dead letter events should be unhealthy", func(t *testing.T) {
		metrics.Reset()

		// Add many dead letter events
		for i := 0; i < 150; i++ {
			event := NewBaseEvent("test.event", "test-service", map[string]interface{}{
				"id": i,
			})
			dlEvent := &DeadLetterEvent{
				OriginalEventID:   event.EventID(),
				OriginalEventType: event.EventType(),
				Reason:            DeadLetterReasonMaxRetriesExceeded,
				Error:             "test error",
			}
			_ = deadLetterQueue.Add(ctx, dlEvent)
		}

		health := healthChecker.CheckHealth(ctx)
		assert.False(t, health.IsHealthy)
		assert.NotEmpty(t, health.Issues)

		// Find the dead letter queue issue
		found := false
		for _, issue := range health.Issues {
			if strings.Contains(issue, "Too many dead letter events") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain dead letter queue issue")
		assert.True(t, health.DeadLetterQueueSize > 100)
	})

	t.Run("Open circuit breakers should be unhealthy", func(t *testing.T) {
		metrics.Reset()

		// Create a circuit breaker and open it
		config := DefaultCircuitBreakerConfig()
		config.MaxFailures = 1
		cb := circuitBreakerManager.GetOrCreateCircuitBreaker("test-cb", config)

		// Open the circuit breaker
		_ = cb.Execute(ctx, func() error {
			return fmt.Errorf("error")
		})

		health := healthChecker.CheckHealth(ctx)
		assert.False(t, health.IsHealthy)
		assert.NotEmpty(t, health.Issues)

		// Find the circuit breaker issue
		found := false
		for _, issue := range health.Issues {
			if strings.Contains(issue, "circuit breakers are open") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should contain circuit breaker issue")
		assert.Equal(t, 1, health.CircuitBreakersOpen)
	})
}

func TestMetricsBuckets(t *testing.T) {
	// Use very short time windows for testing
	metrics := NewEventMetricsCollector(time.Millisecond*100, 4)

	t.Run("Metrics should be distributed across buckets", func(t *testing.T) {
		// Publish events at different times
		metrics.RecordEventPublished("test.event")
		time.Sleep(time.Millisecond * 30)

		metrics.RecordEventPublished("test.event")
		time.Sleep(time.Millisecond * 30)

		metrics.RecordEventPublished("test.event")

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(3), snapshot.TotalEventsPublished)
		assert.Len(t, snapshot.RecentBuckets, 4)

		// At least one bucket should have events
		totalBucketEvents := int64(0)
		for _, bucket := range snapshot.RecentBuckets {
			totalBucketEvents += bucket.EventsPublished
		}
		assert.True(t, totalBucketEvents > 0)
	})
}

func BenchmarkEventMetrics(b *testing.B) {
	metrics := NewEventMetricsCollector(time.Hour, 12)

	b.Run("RecordEventPublished", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metrics.RecordEventPublished("test.event")
		}
	})

	b.Run("RecordEventProcessed", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metrics.RecordEventProcessed("test.event", "handler1", time.Millisecond*10)
		}
	})

	b.Run("RecordEventFailed", func(b *testing.B) {
		err := fmt.Errorf("test error")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			metrics.RecordEventFailed("test.event", "handler1", err)
		}
	})

	b.Run("GetSnapshot", func(b *testing.B) {
		// Add some data first
		for i := 0; i < 100; i++ {
			metrics.RecordEventPublished("test.event")
			metrics.RecordEventProcessed("test.event", "handler1", time.Millisecond*10)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metrics.GetSnapshot()
		}
	})
}

func TestConcurrentMetrics(t *testing.T) {
	metrics := NewEventMetricsCollector(time.Hour, 12)

	t.Run("Concurrent access should be safe", func(t *testing.T) {
		numGoroutines := 10
		numOperations := 100

		done := make(chan bool, numGoroutines)

		// Start multiple goroutines doing different operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < numOperations; j++ {
					switch j % 5 {
					case 0:
						metrics.RecordEventPublished(fmt.Sprintf("event.%d", id))
					case 1:
						metrics.RecordEventProcessed(fmt.Sprintf("event.%d", id), fmt.Sprintf("handler.%d", id), time.Millisecond*10)
					case 2:
						metrics.RecordEventFailed(fmt.Sprintf("event.%d", id), fmt.Sprintf("handler.%d", id), fmt.Errorf("error"))
					case 3:
						metrics.RecordEventMovedToDeadLetter(fmt.Sprintf("event.%d", id))
					case 4:
						_ = metrics.GetSnapshot()
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify that metrics were recorded
		snapshot := metrics.GetSnapshot()
		assert.True(t, snapshot.TotalEventsPublished > 0)
		assert.True(t, snapshot.TotalEventsProcessed > 0)
		assert.True(t, snapshot.TotalEventsFailed > 0)
	})
}
