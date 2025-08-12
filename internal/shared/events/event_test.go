package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBaseEvent(t *testing.T) {
	t.Run("NewBaseEvent creates event with correct fields", func(t *testing.T) {
		eventType := "test.event"
		source := "test-service"
		data := map[string]any{
			"test_field": "test_value",
		}

		event := NewBaseEvent(eventType, source, data)

		assert.Equal(t, eventType, event.EventType())
		assert.Equal(t, source, event.EventSource())
		assert.Equal(t, data, event.EventData())
		assert.Equal(t, "1.0", event.EventVersion())
		assert.NotEmpty(t, event.EventID())
		assert.True(t, time.Since(event.EventTime()) < time.Second)
	})

	t.Run("SetMetadata and GetMetadata work correctly", func(t *testing.T) {
		event := NewBaseEvent("test.event", "test-service", nil)

		// Set metadata
		event.SetMetadata("key1", "value1")
		event.SetMetadata("key2", 42)

		// Get metadata
		value1, exists1 := event.GetMetadata("key1")
		assert.True(t, exists1)
		assert.Equal(t, "value1", value1)

		value2, exists2 := event.GetMetadata("key2")
		assert.True(t, exists2)
		assert.Equal(t, 42, value2)

		// Non-existent key
		_, exists3 := event.GetMetadata("nonexistent")
		assert.False(t, exists3)
	})
}

func TestDomainEvents(t *testing.T) {
	t.Run("NewUserEvent creates valid user event", func(t *testing.T) {
		userID := uint(123)
		eventType := EventTypeUserCreated
		data := map[string]any{
			"email": "test@example.com",
			"name":  "Test User",
		}

		event := NewUserEvent(eventType, userID, data)

		assert.Equal(t, eventType, event.EventType())
		assert.Equal(t, "user-service", event.EventSource())
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, data, event.EventData())
	})

	t.Run("NewPaymentEvent creates valid payment event", func(t *testing.T) {
		paymentID := "payment_123"
		amount := 99.99
		userID := uint(456)
		eventType := EventTypePaymentCompleted
		data := map[string]any{
			"currency": "CNY",
			"method":   "credit_card",
		}

		event := NewPaymentEvent(eventType, paymentID, amount, userID, data)

		assert.Equal(t, eventType, event.EventType())
		assert.Equal(t, "payment-service", event.EventSource())
		assert.Equal(t, paymentID, event.PaymentID)
		assert.Equal(t, amount, event.Amount)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, data, event.EventData())
	})

	t.Run("NewSubscriptionEvent creates valid subscription event", func(t *testing.T) {
		subscriptionID := uint(789)
		userID := uint(456)
		eventType := EventTypeSubscriptionActivated
		data := map[string]any{
			"plan_id": "premium",
			"period":  "monthly",
		}

		event := NewSubscriptionEvent(eventType, subscriptionID, userID, data)

		assert.Equal(t, eventType, event.EventType())
		assert.Equal(t, "subscription-service", event.EventSource())
		assert.Equal(t, subscriptionID, event.SubscriptionID)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, data, event.EventData())
	})

	t.Run("NewOrderEvent creates valid order event", func(t *testing.T) {
		orderID := uint(999)
		userID := uint(456)
		eventType := EventTypeOrderCreated
		data := map[string]any{
			"total_amount": 199.99,
			"items_count":  3,
		}

		event := NewOrderEvent(eventType, orderID, userID, data)

		assert.Equal(t, eventType, event.EventType())
		assert.Equal(t, "order-service", event.EventSource())
		assert.Equal(t, orderID, event.OrderID)
		assert.Equal(t, userID, event.UserID)
		assert.Equal(t, data, event.EventData())
	})
}

func TestEventSerialization(t *testing.T) {
	t.Run("SerializeEvent and DeserializeEvent work correctly", func(t *testing.T) {
		originalEvent := NewUserEvent(
			EventTypeUserCreated,
			123,
			map[string]any{
				"email": "test@example.com",
				"name":  "Test User",
			},
		)
		originalEvent.SetMetadata("correlation_id", "test-correlation-123")

		// Serialize
		data, err := SerializeEvent(originalEvent)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Deserialize
		envelope, err := DeserializeEvent(data)
		require.NoError(t, err)
		assert.NotNil(t, envelope)
		assert.NotNil(t, envelope.Event)
		assert.True(t, time.Since(envelope.CreatedAt) < time.Second)

		// Verify the deserialized event has the same properties
		// Note: The exact type might not be preserved in deserialization,
		// but the interface methods should work
		assert.Equal(t, originalEvent.EventType(), envelope.Event.EventType())
		assert.Equal(t, originalEvent.EventID(), envelope.Event.EventID())
		assert.Equal(t, originalEvent.EventSource(), envelope.Event.EventSource())
	})
}

func TestEventHandler(t *testing.T) {
	t.Run("EventHandlerFunc works correctly", func(t *testing.T) {
		var capturedEvent Event
		var capturedContext context.Context

		handler := NewEventHandler(
			[]string{EventTypeUserCreated, EventTypeUserUpdated},
			func(ctx context.Context, event Event) error {
				capturedContext = ctx
				capturedEvent = event
				return nil
			},
		)

		assert.Equal(t, []string{EventTypeUserCreated, EventTypeUserUpdated}, handler.EventTypes())

		// Test handling an event
		ctx := context.Background()
		testEvent := NewUserEvent(EventTypeUserCreated, 123, map[string]any{"test": "data"})

		err := handler.Handle(ctx, testEvent)
		require.NoError(t, err)

		assert.Equal(t, ctx, capturedContext)
		assert.Equal(t, testEvent, capturedEvent)
	})
}

func TestEventEnvelope(t *testing.T) {
	t.Run("NewEventEnvelope creates envelope with correct properties", func(t *testing.T) {
		event := NewUserEvent(EventTypeUserCreated, 123, map[string]any{"test": "data"})
		envelope := NewEventEnvelope(event)

		assert.Equal(t, event, envelope.Event)
		assert.NotNil(t, envelope.Context)
		assert.NotNil(t, envelope.Headers)
		assert.True(t, time.Since(envelope.CreatedAt) < time.Second)
	})
}

func TestEventConstants(t *testing.T) {
	t.Run("All event constants are properly defined", func(t *testing.T) {
		// User events
		assert.Equal(t, "user.created", EventTypeUserCreated)
		assert.Equal(t, "user.registered", EventTypeUserRegistered)
		assert.Equal(t, "user.updated", EventTypeUserUpdated)
		assert.Equal(t, "user.deleted", EventTypeUserDeleted)
		assert.Equal(t, "user.status_changed", EventTypeUserStatusChanged)

		// Payment events
		assert.Equal(t, "payment.created", EventTypePaymentCreated)
		assert.Equal(t, "payment.completed", EventTypePaymentCompleted)
		assert.Equal(t, "payment.failed", EventTypePaymentFailed)
		assert.Equal(t, "payment.refunded", EventTypePaymentRefunded)

		// Subscription events
		assert.Equal(t, "subscription.created", EventTypeSubscriptionCreated)
		assert.Equal(t, "subscription.activated", EventTypeSubscriptionActivated)
		assert.Equal(t, "subscription.expired", EventTypeSubscriptionExpired)
		assert.Equal(t, "subscription.cancelled", EventTypeSubscriptionCancelled)

		// Order events
		assert.Equal(t, "order.created", EventTypeOrderCreated)
		assert.Equal(t, "order.paid", EventTypeOrderPaid)
		assert.Equal(t, "order.cancelled", EventTypeOrderCancelled)

		// Invoice events
		assert.Equal(t, "invoice.created", EventTypeInvoiceCreated)
		assert.Equal(t, "invoice.generated", EventTypeInvoiceGenerated)
		assert.Equal(t, "invoice.paid", EventTypeInvoicePaid)
	})
}

func TestCorrelationID(t *testing.T) {
	t.Run("SetCorrelationID and extractCorrelationID work correctly", func(t *testing.T) {
		event := NewBaseEvent("test.event", "test-service", nil)
		correlationID := "test-correlation-123"

		// Set correlation ID
		SetCorrelationID(event, correlationID)

		// Extract correlation ID
		extracted := extractCorrelationID(event)
		assert.Equal(t, correlationID, extracted)
	})

	t.Run("extractCorrelationID returns empty string for missing correlation ID", func(t *testing.T) {
		event := NewBaseEvent("test.event", "test-service", nil)

		extracted := extractCorrelationID(event)
		assert.Empty(t, extracted)
	})
}

// Benchmark tests
func BenchmarkEventCreation(b *testing.B) {
	b.Run("BaseEvent creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewBaseEvent("test.event", "test-service", map[string]any{
				"test_field": "test_value",
				"counter":    i,
			})
		}
	})

	b.Run("UserEvent creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewUserEvent(EventTypeUserCreated, uint(i), map[string]any{
				"email": "test@example.com",
				"name":  "Test User",
			})
		}
	})
}

func BenchmarkEventSerialization(b *testing.B) {
	event := NewUserEvent(EventTypeUserCreated, 123, map[string]any{
		"email": "test@example.com",
		"name":  "Test User",
	})

	b.Run("SerializeEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = SerializeEvent(event)
		}
	})

	data, _ := SerializeEvent(event)
	b.Run("DeserializeEvent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = DeserializeEvent(data)
		}
	})
}
