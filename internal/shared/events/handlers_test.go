package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrossDomainEventHandlers(t *testing.T) {
	t.Run("PaymentCompletedHandler processes payment completed events", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		handlers := NewCrossDomainEventHandlers()

		var publishedEvents []Event
		var mu sync.Mutex

		// Capture all events published to the bus
		testHandler := NewEventHandler(
			[]string{EventTypeOrderPaid},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				publishedEvents = append(publishedEvents, event)
				return nil
			},
		)

		err := bus.Subscribe([]string{EventTypeOrderPaid}, testHandler)
		require.NoError(t, err)

		// Set the global event bus for the test
		InitEventBus(bus)

		// Create payment completed event
		paymentEvent := NewPaymentEvent(
			EventTypePaymentCompleted,
			"payment_123",
			99.99,
			456,
			map[string]any{
				"order_id": float64(789), // JSON unmarshaling converts numbers to float64
				"method":   "credit_card",
			},
		)

		// Handle the payment completed event
		handler := handlers.PaymentCompletedHandler()
		err = handler.Handle(context.Background(), paymentEvent)
		require.NoError(t, err)

		// Wait for async processing
		time.Sleep(10 * time.Millisecond)

		// Check that order paid event was published
		mu.Lock()
		assert.Len(t, publishedEvents, 1)
		if len(publishedEvents) > 0 {
			orderEvent, ok := publishedEvents[0].(*OrderEvent)
			assert.True(t, ok)
			assert.Equal(t, EventTypeOrderPaid, orderEvent.EventType())
			assert.Equal(t, uint(789), orderEvent.OrderID)
			assert.Equal(t, uint(456), orderEvent.UserID)
		}
		mu.Unlock()
	})

	t.Run("OrderPaidHandler processes order paid events", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		handlers := NewCrossDomainEventHandlers()

		var publishedEvents []Event
		var mu sync.Mutex

		// Capture all events published to the bus
		testHandler := NewEventHandler(
			[]string{EventTypeInvoiceCreated, EventTypeSubscriptionActivated},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				publishedEvents = append(publishedEvents, event)
				return nil
			},
		)

		err := bus.Subscribe([]string{EventTypeInvoiceCreated, EventTypeSubscriptionActivated}, testHandler)
		require.NoError(t, err)

		// Set the global event bus for the test
		InitEventBus(bus)

		// Create order paid event
		orderEvent := NewOrderEvent(
			EventTypeOrderPaid,
			789,
			456,
			map[string]any{
				"payment_id": "payment_123",
				"amount":     99.99,
			},
		)

		// Handle the order paid event
		handler := handlers.OrderPaidHandler()
		err = handler.Handle(context.Background(), orderEvent)
		require.NoError(t, err)

		// Wait for async processing
		time.Sleep(10 * time.Millisecond)

		// Check that invoice created and subscription activated events were published
		mu.Lock()
		assert.Len(t, publishedEvents, 2)

		var invoiceEvent, subscriptionEvent Event
		for _, event := range publishedEvents {
			switch event.EventType() {
			case EventTypeInvoiceCreated:
				invoiceEvent = event
			case EventTypeSubscriptionActivated:
				subscriptionEvent = event
			}
		}

		assert.NotNil(t, invoiceEvent)
		assert.NotNil(t, subscriptionEvent)
		mu.Unlock()
	})

	t.Run("SubscriptionExpiredHandler processes subscription expired events", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		handlers := NewCrossDomainEventHandlers()

		var publishedEvents []Event
		var mu sync.Mutex

		// Capture all events published to the bus
		testHandler := NewEventHandler(
			[]string{EventTypeUserStatusChanged},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				publishedEvents = append(publishedEvents, event)
				return nil
			},
		)

		err := bus.Subscribe([]string{EventTypeUserStatusChanged}, testHandler)
		require.NoError(t, err)

		// Set the global event bus for the test
		InitEventBus(bus)

		// Create subscription expired event
		subscriptionEvent := NewSubscriptionEvent(
			EventTypeSubscriptionExpired,
			123,
			456,
			map[string]any{
				"expired_at": time.Now(),
			},
		)

		// Handle the subscription expired event
		handler := handlers.SubscriptionExpiredHandler()
		err = handler.Handle(context.Background(), subscriptionEvent)
		require.NoError(t, err)

		// Wait for async processing
		time.Sleep(10 * time.Millisecond)

		// Check that user status changed event was published
		mu.Lock()
		assert.Len(t, publishedEvents, 1)
		if len(publishedEvents) > 0 {
			userEvent, ok := publishedEvents[0].(*UserEvent)
			assert.True(t, ok)
			assert.Equal(t, EventTypeUserStatusChanged, userEvent.EventType())
			assert.Equal(t, uint(456), userEvent.UserID)
		}
		mu.Unlock()
	})

	t.Run("UserDeletedHandler processes user deleted events", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		handlers := NewCrossDomainEventHandlers()

		var publishedEvents []Event
		var mu sync.Mutex

		// Capture all events published to the bus
		testHandler := NewEventHandler(
			[]string{EventTypeSubscriptionCancelled},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				publishedEvents = append(publishedEvents, event)
				return nil
			},
		)

		err := bus.Subscribe([]string{EventTypeSubscriptionCancelled}, testHandler)
		require.NoError(t, err)

		// Set the global event bus for the test
		InitEventBus(bus)

		// Create user deleted event with active subscriptions
		userEvent := NewUserEvent(
			EventTypeUserDeleted,
			456,
			map[string]any{
				"email": "test@example.com",
				"active_subscriptions": []any{
					map[string]any{
						"id":     float64(123),
						"status": "active",
					},
				},
			},
		)

		// Handle the user deleted event
		handler := handlers.UserDeletedHandler()
		err = handler.Handle(context.Background(), userEvent)
		require.NoError(t, err)

		// Wait for async processing
		time.Sleep(10 * time.Millisecond)

		// Check that subscription cancelled event was published
		mu.Lock()
		assert.Len(t, publishedEvents, 1)
		if len(publishedEvents) > 0 {
			subscriptionEvent, ok := publishedEvents[0].(*SubscriptionEvent)
			assert.True(t, ok)
			assert.Equal(t, EventTypeSubscriptionCancelled, subscriptionEvent.EventType())
			assert.Equal(t, uint(123), subscriptionEvent.SubscriptionID)
			assert.Equal(t, uint(456), subscriptionEvent.UserID)
		}
		mu.Unlock()
	})

	t.Run("PaymentFailedHandler processes payment failed events", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		handlers := NewCrossDomainEventHandlers()

		var publishedEvents []Event
		var mu sync.Mutex

		// Capture all events published to the bus
		testHandler := NewEventHandler(
			[]string{EventTypeOrderCancelled},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				publishedEvents = append(publishedEvents, event)
				return nil
			},
		)

		err := bus.Subscribe([]string{EventTypeOrderCancelled}, testHandler)
		require.NoError(t, err)

		// Set the global event bus for the test
		InitEventBus(bus)

		// Create payment failed event
		paymentEvent := NewPaymentEvent(
			EventTypePaymentFailed,
			"payment_123",
			99.99,
			456,
			map[string]any{
				"order_id": float64(789),
				"reason":   "insufficient_funds",
			},
		)

		// Handle the payment failed event
		handler := handlers.PaymentFailedHandler()
		err = handler.Handle(context.Background(), paymentEvent)
		require.NoError(t, err)

		// Wait for async processing
		time.Sleep(10 * time.Millisecond)

		// Check that order cancelled event was published
		mu.Lock()
		assert.Len(t, publishedEvents, 1)
		if len(publishedEvents) > 0 {
			orderEvent, ok := publishedEvents[0].(*OrderEvent)
			assert.True(t, ok)
			assert.Equal(t, EventTypeOrderCancelled, orderEvent.EventType())
			assert.Equal(t, uint(789), orderEvent.OrderID)
			assert.Equal(t, uint(456), orderEvent.UserID)
		}
		mu.Unlock()
	})

	t.Run("RegisterCrossDomainHandlers registers all handlers", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		handlers := NewCrossDomainEventHandlers()

		err := handlers.RegisterCrossDomainHandlers(bus)
		require.NoError(t, err)

		// Check that handlers are registered for expected event types
		expectedEventTypes := []string{
			EventTypePaymentCompleted,
			EventTypeOrderPaid,
			EventTypeSubscriptionExpired,
			EventTypeUserDeleted,
			EventTypePaymentFailed,
			EventTypeInvoiceOverdue,
		}

		for _, eventType := range expectedEventTypes {
			count := bus.GetHandlerCount(eventType)
			assert.Greater(t, count, 0, "Expected handler for event type: %s", eventType)
		}
	})
}

func TestNotificationHandler(t *testing.T) {
	t.Run("Handle processes notification events correctly", func(t *testing.T) {
		handler := NewNotificationHandler()

		testCases := []struct {
			name      string
			eventType string
			shouldLog bool
		}{
			{"PaymentCompleted", EventTypePaymentCompleted, true},
			{"SubscriptionExpired", EventTypeSubscriptionExpired, true},
			{"InvoiceOverdue", EventTypeInvoiceOverdue, true},
			{"OrderPaid", EventTypeOrderPaid, true},
			{"UserCreated", EventTypeUserCreated, true},
			{"SubscriptionActivated", EventTypeSubscriptionActivated, true},
			{"UnhandledEvent", "unhandled.event", false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				event := NewBaseEvent(tc.eventType, "test-service", map[string]any{
					"test": "data",
				})

				// This test mainly ensures the handler doesn't error
				// In a real implementation, you'd test actual notification sending
				err := handler.Handle(context.Background(), event)
				require.NoError(t, err)
			})
		}
	})

	t.Run("EventTypes returns correct event types", func(t *testing.T) {
		handler := NewNotificationHandler()
		eventTypes := handler.EventTypes()

		expectedTypes := []string{
			EventTypePaymentCompleted,
			EventTypeSubscriptionExpired,
			EventTypeInvoiceOverdue,
			EventTypeOrderPaid,
			EventTypeUserCreated,
			EventTypeSubscriptionActivated,
		}

		assert.ElementsMatch(t, expectedTypes, eventTypes)
	})
}

func TestEventHandlerErrorCases(t *testing.T) {
	t.Run("PaymentCompletedHandler handles wrong event type", func(t *testing.T) {
		handlers := NewCrossDomainEventHandlers()
		handler := handlers.PaymentCompletedHandler()

		// Pass wrong event type
		userEvent := NewUserEvent(EventTypeUserCreated, 123, map[string]any{})

		err := handler.Handle(context.Background(), userEvent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected PaymentEvent")
	})

	t.Run("OrderPaidHandler handles wrong event type", func(t *testing.T) {
		handlers := NewCrossDomainEventHandlers()
		handler := handlers.OrderPaidHandler()

		// Pass wrong event type
		userEvent := NewUserEvent(EventTypeUserCreated, 123, map[string]any{})

		err := handler.Handle(context.Background(), userEvent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected OrderEvent")
	})

	t.Run("SubscriptionExpiredHandler handles wrong event type", func(t *testing.T) {
		handlers := NewCrossDomainEventHandlers()
		handler := handlers.SubscriptionExpiredHandler()

		// Pass wrong event type
		userEvent := NewUserEvent(EventTypeUserCreated, 123, map[string]any{})

		err := handler.Handle(context.Background(), userEvent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected SubscriptionEvent")
	})
}

// Benchmark tests
func BenchmarkCrossDomainHandlers(b *testing.B) {
	bus := NewInMemoryEventBus()
	defer bus.Close()

	// Set up a minimal event bus for benchmarking
	InitEventBus(bus)

	handlers := NewCrossDomainEventHandlers()
	paymentHandler := handlers.PaymentCompletedHandler()

	paymentEvent := NewPaymentEvent(
		EventTypePaymentCompleted,
		"payment_123",
		99.99,
		456,
		map[string]any{
			"order_id": float64(789),
			"method":   "credit_card",
		},
	)

	b.Run("PaymentCompletedHandler", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = paymentHandler.Handle(context.Background(), paymentEvent)
		}
	})
}

func TestHandlerChaining(t *testing.T) {
	t.Run("Payment -> Order -> Invoice -> Subscription chain", func(t *testing.T) {
		bus := NewInMemoryEventBus()
		defer bus.Close()

		handlers := NewCrossDomainEventHandlers()

		var allEvents []Event
		var mu sync.Mutex

		// Capture all events
		allEventHandler := NewEventHandler(
			[]string{
				EventTypeOrderPaid,
				EventTypeInvoiceCreated,
				EventTypeSubscriptionActivated,
			},
			func(ctx context.Context, event Event) error {
				mu.Lock()
				defer mu.Unlock()
				allEvents = append(allEvents, event)
				return nil
			},
		)

		err := bus.Subscribe([]string{
			EventTypeOrderPaid,
			EventTypeInvoiceCreated,
			EventTypeSubscriptionActivated,
		}, allEventHandler)
		require.NoError(t, err)

		// Set the global event bus
		InitEventBus(bus)

		// Start the chain with payment completed
		paymentEvent := NewPaymentEvent(
			EventTypePaymentCompleted,
			"payment_123",
			99.99,
			456,
			map[string]any{
				"order_id": float64(789),
				"method":   "credit_card",
			},
		)

		// Process payment completed event
		paymentHandler := handlers.PaymentCompletedHandler()
		err = paymentHandler.Handle(context.Background(), paymentEvent)
		require.NoError(t, err)

		// Wait for async processing
		time.Sleep(20 * time.Millisecond)

		// Should have published order paid event
		mu.Lock()
		assert.Len(t, allEvents, 1)
		orderEvent := allEvents[0]
		assert.Equal(t, EventTypeOrderPaid, orderEvent.EventType())
		mu.Unlock()

		// Now process the order paid event to trigger invoice and subscription events
		orderHandler := handlers.OrderPaidHandler()
		err = orderHandler.Handle(context.Background(), orderEvent)
		require.NoError(t, err)

		// Wait for async processing
		time.Sleep(20 * time.Millisecond)

		// Should now have invoice created and subscription activated events
		mu.Lock()
		assert.GreaterOrEqual(t, len(allEvents), 3) // Original order event + invoice + subscription

		var hasInvoiceEvent, hasSubscriptionEvent bool
		for _, event := range allEvents {
			switch event.EventType() {
			case EventTypeInvoiceCreated:
				hasInvoiceEvent = true
			case EventTypeSubscriptionActivated:
				hasSubscriptionEvent = true
			}
		}

		assert.True(t, hasInvoiceEvent, "Expected invoice created event")
		assert.True(t, hasSubscriptionEvent, "Expected subscription activated event")
		mu.Unlock()
	})
}
