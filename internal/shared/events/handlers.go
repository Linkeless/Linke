package events

import (
	"context"
	"fmt"

	"linke/internal/shared/logger"
)

// CrossDomainEventHandlers contains handlers for cross-domain event communication
type CrossDomainEventHandlers struct {
	logger logger.Logger
}

// NewCrossDomainEventHandlers creates new cross-domain event handlers
func NewCrossDomainEventHandlers() *CrossDomainEventHandlers {
	return &CrossDomainEventHandlers{
		logger: logger.GetGlobalLogger(),
	}
}

// PaymentCompletedHandler handles payment completion events
// When PaymentCompleted → Update order status and activate subscription
func (h *CrossDomainEventHandlers) PaymentCompletedHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypePaymentCompleted},
		func(ctx context.Context, event Event) error {
			paymentEvent, ok := event.(*PaymentEvent)
			if !ok {
				return fmt.Errorf("expected PaymentEvent, got %T", event)
			}

			h.logger.Info("Processing payment completed event",
				logger.String("payment_id", paymentEvent.PaymentID),
				logger.Any("amount", paymentEvent.Amount),
				logger.Uint("user_id", paymentEvent.UserID),
			)

			// TODO: Integrate with actual services when implementing
			// This is where you would:
			// 1. Update order status to "paid"
			// 2. Activate subscription
			// 3. Send confirmation notifications

			// For now, we'll publish follow-up events to demonstrate the flow
			if orderData, ok := paymentEvent.EventData().(map[string]interface{}); ok {
				if orderIDFloat, exists := orderData["order_id"]; exists {
					if orderID, ok := orderIDFloat.(float64); ok {
						// Create order paid event
						orderPaidEvent := NewOrderEvent(
							EventTypeOrderPaid,
							uint(orderID),
							paymentEvent.UserID,
							map[string]interface{}{
								"payment_id": paymentEvent.PaymentID,
								"amount":     paymentEvent.Amount,
								"paid_at":    paymentEvent.EventTime(),
							},
						)

						// Publish order paid event (this would trigger subscription activation)
						if err := Publish(ctx, orderPaidEvent); err != nil {
							h.logger.Error("Failed to publish order paid event",
								logger.String("payment_id", paymentEvent.PaymentID),
								logger.ErrorField(err),
							)
							return fmt.Errorf("failed to publish order paid event: %w", err)
						}
					}
				}
			}

			h.logger.Info("Payment completed event processed successfully",
				logger.String("payment_id", paymentEvent.PaymentID),
			)

			return nil
		},
	)
}

// OrderPaidHandler handles order paid events
// When OrderPaid → Create invoice and update subscription
func (h *CrossDomainEventHandlers) OrderPaidHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeOrderPaid},
		func(ctx context.Context, event Event) error {
			orderEvent, ok := event.(*OrderEvent)
			if !ok {
				return fmt.Errorf("expected OrderEvent, got %T", event)
			}

			h.logger.Info("Processing order paid event",
				logger.Uint("order_id", orderEvent.OrderID),
				logger.Uint("user_id", orderEvent.UserID),
			)

			// TODO: Integrate with actual services
			// This is where you would:
			// 1. Create invoice
			// 2. Activate subscription
			// 3. Update user access

			// Simulate invoice creation
			orderData := orderEvent.EventData().(map[string]interface{})
			amount, _ := orderData["amount"].(float64)

			// Create invoice event
			invoiceEvent := NewInvoiceEvent(
				EventTypeInvoiceCreated,
				0, // This would be the actual invoice ID from the service
				orderEvent.OrderID,
				orderEvent.UserID,
				amount,
				map[string]interface{}{
					"order_id":   orderEvent.OrderID,
					"user_id":    orderEvent.UserID,
					"amount":     amount,
					"status":     "generated",
					"created_at": orderEvent.EventTime(),
				},
			)

			// Publish invoice created event
			if err := Publish(ctx, invoiceEvent); err != nil {
				h.logger.Error("Failed to publish invoice created event",
					logger.Uint("order_id", orderEvent.OrderID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to publish invoice created event: %w", err)
			}

			// Create subscription activation event
			subscriptionEvent := NewSubscriptionEvent(
				EventTypeSubscriptionActivated,
				0, // This would be the actual subscription ID from the service
				orderEvent.UserID,
				map[string]interface{}{
					"order_id":     orderEvent.OrderID,
					"user_id":      orderEvent.UserID,
					"activated_at": orderEvent.EventTime(),
					"status":       "active",
				},
			)

			// Publish subscription activated event
			if err := Publish(ctx, subscriptionEvent); err != nil {
				h.logger.Error("Failed to publish subscription activated event",
					logger.Uint("order_id", orderEvent.OrderID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to publish subscription activated event: %w", err)
			}

			h.logger.Info("Order paid event processed successfully",
				logger.Uint("order_id", orderEvent.OrderID),
			)

			return nil
		},
	)
}

// SubscriptionExpiredHandler handles subscription expiry events
// When SubscriptionExpired → Update user access and send notifications
func (h *CrossDomainEventHandlers) SubscriptionExpiredHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeSubscriptionExpired},
		func(ctx context.Context, event Event) error {
			subscriptionEvent, ok := event.(*SubscriptionEvent)
			if !ok {
				return fmt.Errorf("expected SubscriptionEvent, got %T", event)
			}

			h.logger.Info("Processing subscription expired event",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
			)

			// TODO: Integrate with actual services
			// This is where you would:
			// 1. Update user access permissions
			// 2. Send expiry notifications
			// 3. Update user status if needed

			// Create user status change event
			userEvent := NewUserEvent(
				EventTypeUserStatusChanged,
				subscriptionEvent.UserID,
				map[string]interface{}{
					"user_id":         subscriptionEvent.UserID,
					"old_status":      "active",
					"new_status":      "expired",
					"reason":          "subscription_expired",
					"subscription_id": subscriptionEvent.SubscriptionID,
					"changed_at":      subscriptionEvent.EventTime(),
				},
			)

			// Publish user status change event
			if err := Publish(ctx, userEvent); err != nil {
				h.logger.Error("Failed to publish user status change event",
					logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to publish user status change event: %w", err)
			}

			h.logger.Info("Subscription expired event processed successfully",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
			)

			return nil
		},
	)
}

// UserDeletedHandler handles user deletion events
// When UserDeleted → Cancel active subscriptions
func (h *CrossDomainEventHandlers) UserDeletedHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeUserDeleted},
		func(ctx context.Context, event Event) error {
			userEvent, ok := event.(*UserEvent)
			if !ok {
				return fmt.Errorf("expected UserEvent, got %T", event)
			}

			h.logger.Info("Processing user deleted event",
				logger.Uint("user_id", userEvent.UserID),
			)

			// TODO: Integrate with actual services
			// This is where you would:
			// 1. Cancel all active subscriptions for the user
			// 2. Clean up user data according to privacy policies
			// 3. Send confirmation notifications

			// For demonstration, create subscription cancellation events
			// In real implementation, you'd query the subscription service for active subscriptions
			userData := userEvent.EventData().(map[string]interface{})
			if activeSubscriptions, exists := userData["active_subscriptions"]; exists {
				if subscriptions, ok := activeSubscriptions.([]interface{}); ok {
					for _, sub := range subscriptions {
						if subMap, ok := sub.(map[string]interface{}); ok {
							if subIDFloat, exists := subMap["id"]; exists {
								if subID, ok := subIDFloat.(float64); ok {
									// Create subscription cancellation event
									subscriptionEvent := NewSubscriptionEvent(
										EventTypeSubscriptionCancelled,
										uint(subID),
										userEvent.UserID,
										map[string]interface{}{
											"subscription_id": uint(subID),
											"user_id":         userEvent.UserID,
											"reason":          "user_deleted",
											"cancelled_at":    userEvent.EventTime(),
										},
									)

									// Publish subscription cancelled event
									if err := Publish(ctx, subscriptionEvent); err != nil {
										h.logger.Error("Failed to publish subscription cancelled event",
											logger.Uint("user_id", userEvent.UserID),
											logger.Uint("subscription_id", uint(subID)),
											logger.ErrorField(err),
										)
										// Continue processing other subscriptions
									}
								}
							}
						}
					}
				}
			}

			h.logger.Info("User deleted event processed successfully",
				logger.Uint("user_id", userEvent.UserID),
			)

			return nil
		},
	)
}

// PaymentFailedHandler handles payment failure events
func (h *CrossDomainEventHandlers) PaymentFailedHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypePaymentFailed},
		func(ctx context.Context, event Event) error {
			paymentEvent, ok := event.(*PaymentEvent)
			if !ok {
				return fmt.Errorf("expected PaymentEvent, got %T", event)
			}

			h.logger.Info("Processing payment failed event",
				logger.String("payment_id", paymentEvent.PaymentID),
				logger.Uint("user_id", paymentEvent.UserID),
			)

			// TODO: Integrate with actual services
			// This is where you would:
			// 1. Update order status to "failed"
			// 2. Send failure notifications
			// 3. Possibly retry payment or offer alternatives

			// Extract order ID from payment data
			if orderData, ok := paymentEvent.EventData().(map[string]interface{}); ok {
				if orderIDFloat, exists := orderData["order_id"]; exists {
					if orderID, ok := orderIDFloat.(float64); ok {
						// Create order cancellation event
						orderEvent := NewOrderEvent(
							EventTypeOrderCancelled,
							uint(orderID),
							paymentEvent.UserID,
							map[string]interface{}{
								"order_id":     uint(orderID),
								"user_id":      paymentEvent.UserID,
								"reason":       "payment_failed",
								"payment_id":   paymentEvent.PaymentID,
								"cancelled_at": paymentEvent.EventTime(),
							},
						)

						// Publish order cancelled event
						if err := Publish(ctx, orderEvent); err != nil {
							h.logger.Error("Failed to publish order cancelled event",
								logger.String("payment_id", paymentEvent.PaymentID),
								logger.ErrorField(err),
							)
							return fmt.Errorf("failed to publish order cancelled event: %w", err)
						}
					}
				}
			}

			h.logger.Info("Payment failed event processed successfully",
				logger.String("payment_id", paymentEvent.PaymentID),
			)

			return nil
		},
	)
}

// InvoiceOverdueHandler handles overdue invoice events
func (h *CrossDomainEventHandlers) InvoiceOverdueHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeInvoiceOverdue},
		func(ctx context.Context, event Event) error {
			invoiceEvent, ok := event.(*InvoiceEvent)
			if !ok {
				return fmt.Errorf("expected InvoiceEvent, got %T", event)
			}

			h.logger.Info("Processing invoice overdue event",
				logger.Uint("invoice_id", invoiceEvent.InvoiceID),
				logger.Uint("user_id", invoiceEvent.UserID),
			)

			// TODO: Integrate with actual services
			// This is where you would:
			// 1. Suspend subscription if applicable
			// 2. Send overdue notifications
			// 3. Initiate collection processes

			// Extract subscription ID from invoice data
			if invoiceData, ok := invoiceEvent.EventData().(map[string]interface{}); ok {
				if subscriptionIDFloat, exists := invoiceData["subscription_id"]; exists {
					if subscriptionID, ok := subscriptionIDFloat.(float64); ok {
						// Create subscription suspension event
						subscriptionEvent := NewSubscriptionEvent(
							EventTypeSubscriptionSuspended,
							uint(subscriptionID),
							invoiceEvent.UserID,
							map[string]interface{}{
								"subscription_id": uint(subscriptionID),
								"user_id":         invoiceEvent.UserID,
								"reason":          "invoice_overdue",
								"invoice_id":      invoiceEvent.InvoiceID,
								"suspended_at":    invoiceEvent.EventTime(),
							},
						)

						// Publish subscription suspended event
						if err := Publish(ctx, subscriptionEvent); err != nil {
							h.logger.Error("Failed to publish subscription suspended event",
								logger.Uint("invoice_id", invoiceEvent.InvoiceID),
								logger.ErrorField(err),
							)
							return fmt.Errorf("failed to publish subscription suspended event: %w", err)
						}
					}
				}
			}

			h.logger.Info("Invoice overdue event processed successfully",
				logger.Uint("invoice_id", invoiceEvent.InvoiceID),
			)

			return nil
		},
	)
}

// RegisterCrossDomainHandlers registers all cross-domain event handlers with the event bus
func (h *CrossDomainEventHandlers) RegisterCrossDomainHandlers(eventBus EventBus) error {
	handlers := []EventHandler{
		h.PaymentCompletedHandler(),
		h.OrderPaidHandler(),
		h.SubscriptionExpiredHandler(),
		h.UserDeletedHandler(),
		h.PaymentFailedHandler(),
		h.InvoiceOverdueHandler(),
	}

	for _, handler := range handlers {
		if err := eventBus.Subscribe(handler.EventTypes(), handler); err != nil {
			h.logger.Error("Failed to register cross-domain event handler",
				logger.Any("event_types", handler.EventTypes()),
				logger.ErrorField(err),
			)
			return fmt.Errorf("failed to register event handler: %w", err)
		}
	}

	h.logger.Info("Cross-domain event handlers registered successfully",
		logger.Int("handler_count", len(handlers)),
	)

	return nil
}

// NotificationHandler handles events that require user notifications
type NotificationHandler struct {
	logger logger.Logger
	id     string
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{
		logger: logger.GetGlobalLogger(),
		id:     generateEventID(), // Generate unique ID for handler
	}
}

// Handle processes events that require notifications
func (h *NotificationHandler) Handle(ctx context.Context, event Event) error {
	h.logger.Info("Processing notification event",
		logger.String("event_type", event.EventType()),
		logger.String("event_id", event.EventID()),
	)

	// TODO: Integrate with actual notification service
	// This is where you would:
	// 1. Determine notification recipients
	// 2. Choose notification channels (email, SMS, push)
	// 3. Format notification content
	// 4. Send notifications

	switch event.EventType() {
	case EventTypePaymentCompleted:
		// Send payment confirmation
		h.logger.Info("Would send payment confirmation notification")
	case EventTypeSubscriptionExpired:
		// Send expiry notification
		h.logger.Info("Would send subscription expiry notification")
	case EventTypeInvoiceOverdue:
		// Send overdue notice
		h.logger.Info("Would send invoice overdue notification")
	case EventTypeOrderPaid:
		// Send order confirmation
		h.logger.Info("Would send order confirmation notification")
	default:
		h.logger.Debug("No notification required for event type",
			logger.String("event_type", event.EventType()),
		)
	}

	return nil
}

// EventTypes returns the event types this handler processes
func (h *NotificationHandler) EventTypes() []string {
	return []string{
		EventTypePaymentCompleted,
		EventTypeSubscriptionExpired,
		EventTypeInvoiceOverdue,
		EventTypeOrderPaid,
		EventTypeUserCreated,
		EventTypeSubscriptionActivated,
	}
}

// ID returns the unique identifier for this handler
func (h *NotificationHandler) ID() string {
	return h.id
}
