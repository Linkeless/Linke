package events

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      3,
		ResetTimeout:     time.Millisecond * 100,
		SuccessThreshold: 2,
		MonitoringWindow: time.Minute,
		HalfOpenMaxCalls: 2,
	}

	cb := NewCircuitBreaker("test-cb", config)
	ctx := context.Background()

	t.Run("Initial state should be closed", func(t *testing.T) {
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("Successful execution", func(t *testing.T) {
		err := cb.Execute(ctx, func() error {
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("Failures should open circuit breaker", func(t *testing.T) {
		// Generate enough failures to open the circuit
		for i := 0; i < config.MaxFailures; i++ {
			err := cb.Execute(ctx, func() error {
				return fmt.Errorf("test error %d", i)
			})
			require.Error(t, err)
		}

		assert.Equal(t, CircuitBreakerOpen, cb.GetState())
	})

	t.Run("Requests should be rejected when open", func(t *testing.T) {
		err := cb.Execute(ctx, func() error {
			return nil
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker")
		assert.Contains(t, err.Error(), "OPEN")
	})

	t.Run("Should transition to half-open after timeout", func(t *testing.T) {
		// Wait for reset timeout
		time.Sleep(config.ResetTimeout + time.Millisecond*10)

		// This should transition to half-open
		err := cb.Execute(ctx, func() error {
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, CircuitBreakerHalfOpen, cb.GetState())
	})

	t.Run("Success in half-open should close circuit", func(t *testing.T) {
		// Need enough successes to close
		for i := 0; i < config.SuccessThreshold-1; i++ {
			err := cb.Execute(ctx, func() error {
				return nil
			})
			require.NoError(t, err)
		}

		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})

	t.Run("Reset should work", func(t *testing.T) {
		// Open the circuit again
		for i := 0; i < config.MaxFailures; i++ {
			_ = cb.Execute(ctx, func() error {
				return fmt.Errorf("test error")
			})
		}
		assert.Equal(t, CircuitBreakerOpen, cb.GetState())

		// Reset
		cb.Reset()
		assert.Equal(t, CircuitBreakerClosed, cb.GetState())
	})
}

func TestCircuitBreakerEventHandler(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	config.MaxFailures = 2 // Lower threshold for testing
	config.ResetTimeout = time.Millisecond * 50

	callCount := 0
	var handlerError error

	mockHandler := NewEventHandler([]string{"test.event"}, func(ctx context.Context, event Event) error {
		callCount++
		return handlerError
	})

	cbHandler := NewCircuitBreakerEventHandler("test-cb-handler", mockHandler, config)
	ctx := context.Background()

	event := NewBaseEvent("test.event", "test-service", map[string]interface{}{
		"test": "data",
	})

	t.Run("Successful event handling", func(t *testing.T) {
		callCount = 0
		handlerError = nil

		err := cbHandler.Handle(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})

	t.Run("Failed event handling should open circuit", func(t *testing.T) {
		callCount = 0
		handlerError = fmt.Errorf("handler error")

		// Generate failures to open circuit
		for i := 0; i < config.MaxFailures; i++ {
			err := cbHandler.Handle(ctx, event)
			require.Error(t, err)
		}

		stats := cbHandler.GetCircuitBreakerStats()
		assert.Equal(t, CircuitBreakerOpen, stats.State)
	})

	t.Run("Events should be rejected when circuit is open", func(t *testing.T) {
		originalCallCount := callCount

		err := cbHandler.Handle(ctx, event)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "circuit breaker")

		// Handler should not be called
		assert.Equal(t, originalCallCount, callCount)
	})

	t.Run("Circuit should recover after timeout", func(t *testing.T) {
		// Wait for reset timeout
		time.Sleep(config.ResetTimeout + time.Millisecond*10)

		handlerError = nil // Fix the error
		callCount = 0

		// This should work and close the circuit
		err := cbHandler.Handle(ctx, event)
		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
	})
}

func TestCircuitBreakerManager(t *testing.T) {
	manager := NewCircuitBreakerManager()

	config1 := DefaultCircuitBreakerConfig()
	config2 := DefaultCircuitBreakerConfig()
	config2.MaxFailures = 10

	t.Run("Create new circuit breakers", func(t *testing.T) {
		cb1 := manager.GetOrCreateCircuitBreaker("cb1", config1)
		assert.NotNil(t, cb1)

		cb2 := manager.GetOrCreateCircuitBreaker("cb2", config2)
		assert.NotNil(t, cb2)
		assert.NotSame(t, cb1, cb2)

		// Should return same instance
		cb1Again := manager.GetOrCreateCircuitBreaker("cb1", config1)
		assert.Same(t, cb1, cb1Again)
	})

	t.Run("Get existing circuit breaker", func(t *testing.T) {
		cb, exists := manager.GetCircuitBreaker("cb1")
		assert.NotNil(t, cb)
		assert.True(t, exists)
	})

	t.Run("Get non-existent circuit breaker", func(t *testing.T) {
		cb, exists := manager.GetCircuitBreaker("non-existent")
		assert.Nil(t, cb)
		assert.False(t, exists)
	})

	t.Run("Get all stats", func(t *testing.T) {
		stats := manager.GetAllStats()
		assert.Len(t, stats, 2)
		assert.Contains(t, stats, "cb1")
		assert.Contains(t, stats, "cb2")
	})

	t.Run("Health check", func(t *testing.T) {
		health := manager.HealthCheck()
		assert.Len(t, health, 2)
		assert.True(t, health["cb1"])
		assert.True(t, health["cb2"])
	})

	t.Run("Reset all", func(t *testing.T) {
		manager.ResetAll()
		// All circuit breakers should be in closed state
		health := manager.HealthCheck()
		for _, isHealthy := range health {
			assert.True(t, isHealthy)
		}
	})

	t.Run("Remove circuit breaker", func(t *testing.T) {
		manager.RemoveCircuitBreaker("cb1")
		_, exists := manager.GetCircuitBreaker("cb1")
		assert.False(t, exists)
	})
}

func TestCircuitBreakerWithMetrics(t *testing.T) {
	// Test integration with metrics
	metrics := NewEventMetricsCollector(time.Hour, 12)

	config := DefaultCircuitBreakerConfig()
	config.MaxFailures = 2

	callCount := 0
	var handlerError error

	mockHandler := NewEventHandler([]string{"test.event"}, func(ctx context.Context, event Event) error {
		callCount++
		return handlerError
	})

	// Wrap with both circuit breaker and metrics
	cbHandler := NewCircuitBreakerEventHandler("test-handler", mockHandler, config)
	metricsHandler := NewMetricsEventHandler("test-handler", cbHandler, metrics)

	ctx := context.Background()
	event := NewBaseEvent("test.event", "test-service", map[string]interface{}{
		"test": "data",
	})

	t.Run("Successful processing should update metrics", func(t *testing.T) {
		callCount = 0
		handlerError = nil

		err := metricsHandler.Handle(ctx, event)
		require.NoError(t, err)

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.TotalEventsProcessed)
		assert.Equal(t, int64(0), snapshot.TotalEventsFailed)
	})

	t.Run("Circuit breaker failures should update metrics", func(t *testing.T) {
		callCount = 0
		handlerError = fmt.Errorf("handler error")

		// Generate failures to open circuit
		for i := 0; i < config.MaxFailures; i++ {
			err := metricsHandler.Handle(ctx, event)
			require.Error(t, err)
		}

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, int64(1), snapshot.TotalEventsProcessed) // From previous test
		assert.Equal(t, int64(2), snapshot.TotalEventsFailed)    // From failures
	})

	t.Run("Circuit breaker rejections should also be tracked", func(t *testing.T) {
		originalFailedCount := metrics.GetSnapshot().TotalEventsFailed

		// Try to process when circuit is open
		err := metricsHandler.Handle(ctx, event)
		require.Error(t, err)

		snapshot := metrics.GetSnapshot()
		assert.Equal(t, originalFailedCount+1, snapshot.TotalEventsFailed)
	})
}

func BenchmarkCircuitBreaker(b *testing.B) {
	config := DefaultCircuitBreakerConfig()
	cb := NewCircuitBreaker("bench-cb", config)
	ctx := context.Background()

	b.Run("Successful execution", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cb.Execute(ctx, func() error {
				return nil
			})
		}
	})

	b.Run("Failed execution", func(b *testing.B) {
		cb.Reset() // Ensure we start with closed circuit
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = cb.Execute(ctx, func() error {
				return fmt.Errorf("error")
			})
		}
	})
}

func TestCircuitBreakerStateTransitions(t *testing.T) {
	config := CircuitBreakerConfig{
		MaxFailures:      2,
		ResetTimeout:     time.Millisecond * 50,
		SuccessThreshold: 2,
		MonitoringWindow: time.Minute,
		HalfOpenMaxCalls: 2,
	}

	cb := NewCircuitBreaker("state-test", config)
	ctx := context.Background()

	// Start: CLOSED
	assert.Equal(t, CircuitBreakerClosed, cb.GetState())

	// CLOSED -> OPEN (failures)
	for i := 0; i < config.MaxFailures; i++ {
		_ = cb.Execute(ctx, func() error {
			return fmt.Errorf("error")
		})
	}
	assert.Equal(t, CircuitBreakerOpen, cb.GetState())

	// OPEN -> HALF_OPEN (timeout)
	time.Sleep(config.ResetTimeout + time.Millisecond*10)
	_ = cb.Execute(ctx, func() error {
		return nil
	})
	assert.Equal(t, CircuitBreakerHalfOpen, cb.GetState())

	// HALF_OPEN -> CLOSED (success)
	_ = cb.Execute(ctx, func() error {
		return nil
	})
	assert.Equal(t, CircuitBreakerClosed, cb.GetState())

	// Test HALF_OPEN -> OPEN (failure)
	// First, open the circuit again
	for i := 0; i < config.MaxFailures; i++ {
		_ = cb.Execute(ctx, func() error {
			return fmt.Errorf("error")
		})
	}
	assert.Equal(t, CircuitBreakerOpen, cb.GetState())

	// Wait and try to transition to half-open, then fail
	time.Sleep(config.ResetTimeout + time.Millisecond*10)
	_ = cb.Execute(ctx, func() error {
		return fmt.Errorf("error")
	})
	assert.Equal(t, CircuitBreakerOpen, cb.GetState())
}
