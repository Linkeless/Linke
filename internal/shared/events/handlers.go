package events

import (
	"context"
	"fmt"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/shared/queue"

	// Service interfaces
	userInterfaces "linke/internal/domains/user/usecases/interfaces"
	subscriptionInterfaces "linke/internal/domains/subscription/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	invoiceInterfaces "linke/internal/domains/invoice/usecases/interfaces"
	serverInterfaces "linke/internal/domains/server/usecases/interfaces"
)

// CrossDomainEventHandlers contains handlers for cross-domain event communication
// This structure manages all cross-domain business flows through event-driven architecture
type CrossDomainEventHandlers struct {
	// Core dependencies
	logger     logger.Logger
	cacheStore EventCacheStore
	taskQueue  *queue.TaskQueue

	// Service dependencies - injected via constructor
	userService             userInterfaces.UserService
	userSubscriptionService subscriptionInterfaces.UserSubscriptionService
	subscriptionOrderService subscriptionInterfaces.SubscriptionOrderService
	paymentService          paymentInterfaces.PaymentService
	invoiceService          invoiceInterfaces.InvoiceService
	shadowsocksServerService serverInterfaces.ShadowsocksServerService

	// Idempotency tracking
	processedEvents map[string]time.Time // In-memory cache for processed events
}

// NewCrossDomainEventHandlers creates new cross-domain event handlers with all required services
func NewCrossDomainEventHandlers(
	userService userInterfaces.UserService,
	userSubscriptionService subscriptionInterfaces.UserSubscriptionService,
	subscriptionOrderService subscriptionInterfaces.SubscriptionOrderService,
	paymentService paymentInterfaces.PaymentService,
	invoiceService invoiceInterfaces.InvoiceService,
	shadowsocksServerService serverInterfaces.ShadowsocksServerService,
	cacheStore EventCacheStore,
	taskQueue *queue.TaskQueue,
) *CrossDomainEventHandlers {
	return &CrossDomainEventHandlers{
		logger:                   logger.GetGlobalLogger(),
		cacheStore:               cacheStore,
		taskQueue:                taskQueue,
		userService:              userService,
		userSubscriptionService:  userSubscriptionService,
		subscriptionOrderService: subscriptionOrderService,
		paymentService:           paymentService,
		invoiceService:           invoiceService,
		shadowsocksServerService: shadowsocksServerService,
		processedEvents:          make(map[string]time.Time),
	}
}

// Idempotency helper method
func (h *CrossDomainEventHandlers) isEventProcessed(eventID string) bool {
	cacheKey := fmt.Sprintf("event_processed:%s", eventID)

	// Check in-memory cache first
	if processedAt, exists := h.processedEvents[eventID]; exists {
		// Clean up old entries (older than 1 hour)
		if time.Since(processedAt) < time.Hour {
			return true
		}
		delete(h.processedEvents, eventID)
	}

	// Check distributed cache
	if h.cacheStore != nil {
		if exists, _ := h.cacheStore.Exists(context.Background(), cacheKey); exists {
			return true
		}
	}

	return false
}

func (h *CrossDomainEventHandlers) markEventProcessed(eventID string) {
	cacheKey := fmt.Sprintf("event_processed:%s", eventID)
	now := time.Now()

	// Mark in memory
	h.processedEvents[eventID] = now

	// Mark in distributed cache with 1 hour TTL
	if h.cacheStore != nil {
		h.cacheStore.Set(context.Background(), cacheKey, "1", time.Hour)
	}
}

// PaymentCompletedHandler handles payment completion events
// Business Flow: Payment Completed → Process Order → Activate Subscription → Generate Invoice → Send Notifications
func (h *CrossDomainEventHandlers) PaymentCompletedHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypePaymentCompleted},
		func(ctx context.Context, event Event) error {
			paymentEvent, ok := event.(*PaymentEvent)
			if !ok {
				return fmt.Errorf("expected PaymentEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(paymentEvent.EventID()) {
				h.logger.Info("Payment completed event already processed, skipping",
					logger.String("payment_id", paymentEvent.PaymentID),
					logger.String("event_id", paymentEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing payment completed event",
				logger.String("payment_id", paymentEvent.PaymentID),
				logger.Any("amount", paymentEvent.Amount),
				logger.Uint("user_id", paymentEvent.UserID),
				logger.String("event_id", paymentEvent.EventID()),
			)

			// Extract order ID from payment data
			orderData, ok := paymentEvent.EventData().(map[string]any)
			if !ok {
				return fmt.Errorf("invalid payment event data format")
			}

			orderIDFloat, exists := orderData["order_id"]
			if !exists {
				return fmt.Errorf("order_id not found in payment event data")
			}

			orderID, ok := orderIDFloat.(float64)
			if !ok {
				return fmt.Errorf("invalid order_id format in payment event")
			}

			// Step 1: Process order payment success
			if err := h.subscriptionOrderService.ProcessOrderPaymentSuccess(ctx, uint(orderID)); err != nil {
				h.logger.Error("Failed to process order payment success",
					logger.String("payment_id", paymentEvent.PaymentID),
					logger.Uint("order_id", uint(orderID)),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to process order payment success: %w", err)
			}

			// Step 2: Get order details to proceed with business flow
			order, err := h.subscriptionOrderService.GetSubscriptionOrder(ctx, uint(orderID))
			if err != nil {
				h.logger.Error("Failed to get subscription order",
					logger.Uint("order_id", uint(orderID)),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get subscription order: %w", err)
			}

			// Step 3: Create/activate subscription if needed
			if order.OrderType == "new" || order.OrderType == "renewal" {
				// Check if user already has an active subscription for this plan
				existingSubscription, _ := h.userSubscriptionService.GetActiveUserSubscription(ctx, order.UserID, order.SubscriptionPlanID)
				
				if existingSubscription == nil {
					// Create new subscription
					createReq := &subscriptionInterfaces.CreateSubscriptionRequest{
						UserID:             order.UserID,
						SubscriptionPlanID: order.SubscriptionPlanID,
						StartDate:          time.Now().Format(time.RFC3339),
					}

					subscription, err := h.userSubscriptionService.CreateUserSubscription(ctx, createReq)
					if err != nil {
						h.logger.Error("Failed to create user subscription",
							logger.Uint("user_id", order.UserID),
							logger.Uint("plan_id", order.SubscriptionPlanID),
							logger.ErrorField(err),
						)
						return fmt.Errorf("failed to create user subscription: %w", err)
					}

					// Publish subscription created event
					subscriptionCreatedEvent := NewSubscriptionEvent(
						EventTypeSubscriptionCreated,
						subscription.ID,
						subscription.UserID,
						map[string]any{
							"subscription_id": subscription.ID,
							"user_id":         subscription.UserID,
							"plan_id":         subscription.SubscriptionPlanID,
							"order_id":        order.ID,
							"status":          subscription.Status,
							"start_date":      subscription.StartDate,
						},
					)
					if err := Publish(ctx, subscriptionCreatedEvent); err != nil {
						h.logger.Error("Failed to publish subscription created event",
							logger.Uint("subscription_id", subscription.ID),
							logger.ErrorField(err),
						)
						// Don't fail the whole process for event publishing failure
					}
				} else {
					// Renew existing subscription
					if err := h.userSubscriptionService.RenewUserSubscription(ctx, existingSubscription.ID); err != nil {
						h.logger.Error("Failed to renew user subscription",
							logger.Uint("subscription_id", existingSubscription.ID),
							logger.ErrorField(err),
						)
						return fmt.Errorf("failed to renew user subscription: %w", err)
					}

					// Publish subscription renewed event
					subscriptionRenewedEvent := NewSubscriptionEvent(
						EventTypeSubscriptionRenewed,
						existingSubscription.ID,
						existingSubscription.UserID,
						map[string]any{
							"subscription_id": existingSubscription.ID,
							"user_id":         existingSubscription.UserID,
							"plan_id":         existingSubscription.SubscriptionPlanID,
							"order_id":        order.ID,
						},
					)
					if err := Publish(ctx, subscriptionRenewedEvent); err != nil {
						h.logger.Error("Failed to publish subscription renewed event",
							logger.Uint("subscription_id", existingSubscription.ID),
							logger.ErrorField(err),
						)
					}
				}
			}

			// Step 4: Create invoice if not already exists
			existingInvoices, _, err := h.invoiceService.GetInvoices(ctx, &invoiceInterfaces.GetInvoicesRequest{
				UserID: order.UserID,
				Limit:  1,
			})
			if err == nil && len(existingInvoices) == 0 {
				// Create invoice for the order
				invoice, err := h.invoiceService.CreateInvoiceFromOrder(ctx, order.ID, &invoiceInterfaces.CreateInvoiceRequest{
					UserID:              order.UserID,
					SubscriptionOrderID: order.ID,
					Amount:              order.TotalAmount,
					Currency:            order.Currency,
					BillingName:         fmt.Sprintf("User %d", order.UserID), // Would need to get actual user name
					BillingEmail:        fmt.Sprintf("user%d@example.com", order.UserID), // Would need actual email
					Description:         fmt.Sprintf("Subscription Order #%s", order.OrderNumber),
					AutoSend:            false, // Don't auto-send, let event handler decide
				})
				if err != nil {
					h.logger.Error("Failed to create invoice from order",
						logger.Uint("order_id", order.ID),
						logger.ErrorField(err),
					)
					// Don't fail the whole process for invoice creation failure
				} else {
					// Publish invoice created event
					invoiceCreatedEvent := NewInvoiceEvent(
						EventTypeInvoiceCreated,
						invoice.ID,
						order.ID,
						order.UserID,
						order.TotalAmount,
						map[string]any{
							"invoice_id": invoice.ID,
							"order_id":   order.ID,
							"user_id":    order.UserID,
							"amount":     order.TotalAmount,
							"status":     invoice.Status,
						},
					)
					if err := Publish(ctx, invoiceCreatedEvent); err != nil {
						h.logger.Error("Failed to publish invoice created event",
							logger.Uint("invoice_id", invoice.ID),
							logger.ErrorField(err),
						)
					}
				}
			}

			// Mark event as processed
			h.markEventProcessed(paymentEvent.EventID())

			h.logger.Info("Payment completed event processed successfully",
				logger.String("payment_id", paymentEvent.PaymentID),
				logger.Uint("order_id", uint(orderID)),
			)

			return nil
		},
	)
}

// OrderPaidHandler handles order paid events
// Business Flow: Order Paid → Create Invoice → Activate Subscription → Send Notifications
func (h *CrossDomainEventHandlers) OrderPaidHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeOrderPaid},
		func(ctx context.Context, event Event) error {
			orderEvent, ok := event.(*OrderEvent)
			if !ok {
				return fmt.Errorf("expected OrderEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(orderEvent.EventID()) {
				h.logger.Info("Order paid event already processed, skipping",
					logger.Uint("order_id", orderEvent.OrderID),
					logger.String("event_id", orderEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing order paid event",
				logger.Uint("order_id", orderEvent.OrderID),
				logger.Uint("user_id", orderEvent.UserID),
				logger.String("event_id", orderEvent.EventID()),
			)

			// Step 1: Get order details
			order, err := h.subscriptionOrderService.GetSubscriptionOrder(ctx, orderEvent.OrderID)
			if err != nil {
				h.logger.Error("Failed to get subscription order",
					logger.Uint("order_id", orderEvent.OrderID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get subscription order: %w", err)
			}

			// Step 2: Create invoice for the paid order
			invoice, err := h.invoiceService.CreateInvoiceFromOrder(ctx, order.ID, &invoiceInterfaces.CreateInvoiceRequest{
				UserID:              order.UserID,
				SubscriptionOrderID: order.ID,
				Amount:              order.TotalAmount,
				Currency:            order.Currency,
				BillingName:         fmt.Sprintf("User %d", order.UserID),
				BillingEmail:        fmt.Sprintf("user%d@example.com", order.UserID),
				Description:         fmt.Sprintf("Subscription Order #%s", order.OrderNumber),
				AutoSend:            true,
			})
			if err != nil {
				h.logger.Error("Failed to create invoice from order",
					logger.Uint("order_id", order.ID),
					logger.ErrorField(err),
				)
				// Don't fail the whole process for invoice creation failure
			} else {
				// Publish invoice created event
				invoiceCreatedEvent := NewInvoiceEvent(
					EventTypeInvoiceCreated,
					invoice.ID,
					order.ID,
					order.UserID,
					order.TotalAmount,
					map[string]any{
						"invoice_id": invoice.ID,
						"order_id":   order.ID,
						"user_id":    order.UserID,
						"amount":     order.TotalAmount,
						"status":     invoice.Status,
					},
				)
				if err := Publish(ctx, invoiceCreatedEvent); err != nil {
					h.logger.Error("Failed to publish invoice created event",
						logger.Uint("invoice_id", invoice.ID),
						logger.ErrorField(err),
					)
				}
			}

			// Step 3: Activate subscription
			existingSubscription, _ := h.userSubscriptionService.GetActiveUserSubscription(ctx, order.UserID, order.SubscriptionPlanID)
			if existingSubscription == nil {
				// Create new subscription
				createReq := &subscriptionInterfaces.CreateSubscriptionRequest{
					UserID:             order.UserID,
					SubscriptionPlanID: order.SubscriptionPlanID,
					StartDate:          time.Now().Format(time.RFC3339),
				}

				subscription, err := h.userSubscriptionService.CreateUserSubscription(ctx, createReq)
				if err != nil {
					h.logger.Error("Failed to create user subscription from paid order",
						logger.Uint("user_id", order.UserID),
						logger.Uint("plan_id", order.SubscriptionPlanID),
						logger.ErrorField(err),
					)
					// Don't fail the whole process
				} else {
					// Publish subscription activated event
					subscriptionActivatedEvent := NewSubscriptionEvent(
						EventTypeSubscriptionActivated,
						subscription.ID,
						subscription.UserID,
						map[string]any{
							"subscription_id": subscription.ID,
							"user_id":         subscription.UserID,
							"plan_id":         subscription.SubscriptionPlanID,
							"order_id":        order.ID,
							"status":          subscription.Status,
							"activated_at":    time.Now(),
						},
					)
					if err := Publish(ctx, subscriptionActivatedEvent); err != nil {
						h.logger.Error("Failed to publish subscription activated event",
							logger.Uint("subscription_id", subscription.ID),
							logger.ErrorField(err),
						)
					}
				}
			} else {
				// Extend existing subscription
				if err := h.userSubscriptionService.RenewUserSubscription(ctx, existingSubscription.ID); err != nil {
					h.logger.Error("Failed to renew user subscription from paid order",
						logger.Uint("subscription_id", existingSubscription.ID),
						logger.ErrorField(err),
					)
				} else {
					// Publish subscription activated event for renewal
					subscriptionActivatedEvent := NewSubscriptionEvent(
						EventTypeSubscriptionActivated,
						existingSubscription.ID,
						existingSubscription.UserID,
						map[string]any{
							"subscription_id": existingSubscription.ID,
							"user_id":         existingSubscription.UserID,
							"plan_id":         existingSubscription.SubscriptionPlanID,
							"order_id":        order.ID,
							"renewed":         true,
							"activated_at":    time.Now(),
						},
					)
					if err := Publish(ctx, subscriptionActivatedEvent); err != nil {
						h.logger.Error("Failed to publish subscription activated event",
							logger.Uint("subscription_id", existingSubscription.ID),
							logger.ErrorField(err),
						)
					}
				}
			}

			// Mark event as processed
			h.markEventProcessed(orderEvent.EventID())

			h.logger.Info("Order paid event processed successfully",
				logger.Uint("order_id", orderEvent.OrderID),
				logger.Uint("user_id", orderEvent.UserID),
			)

			return nil
		},
	)
}

// SubscriptionCreatedHandler handles subscription creation events
// Business Flow: Subscription Created → Initialize User Access → Generate Welcome Invoice → Send Notifications
func (h *CrossDomainEventHandlers) SubscriptionCreatedHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeSubscriptionCreated},
		func(ctx context.Context, event Event) error {
			subscriptionEvent, ok := event.(*SubscriptionEvent)
			if !ok {
				return fmt.Errorf("expected SubscriptionEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(subscriptionEvent.EventID()) {
				h.logger.Info("Subscription created event already processed, skipping",
					logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
					logger.String("event_id", subscriptionEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing subscription created event",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
				logger.String("event_id", subscriptionEvent.EventID()),
			)

			// Step 1: Get subscription details
			subscription, err := h.userSubscriptionService.GetUserSubscription(ctx, subscriptionEvent.SubscriptionID)
			if err != nil {
				h.logger.Error("Failed to get user subscription",
					logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get user subscription: %w", err)
			}

			// Step 2: Update user status to active if needed
			user, err := h.userService.GetUserByID(ctx, subscription.UserID)
			if err != nil {
				h.logger.Error("Failed to get user",
					logger.Uint("user_id", subscription.UserID),
					logger.ErrorField(err),
				)
			} else {
				// Activate user if they are inactive
				if user.Status != "active" {
					if _, err := h.userService.UpdateUserStatus(ctx, user.ID, "active"); err != nil {
						h.logger.Error("Failed to activate user",
							logger.Uint("user_id", user.ID),
							logger.ErrorField(err),
						)
						// Don't fail the whole process
					} else {
						h.logger.Info("User activated successfully",
							logger.Uint("user_id", user.ID),
						)
					}
				}
			}

			// Step 3: Configure server access (if server groups are specified)
			if subscription.ServerGroupIDs != "" {
				groupIDs := subscription.GetServerGroupIDs()
				if len(groupIDs) > 0 {
					h.logger.Info("User has access to server groups",
						logger.Uint("user_id", subscription.UserID),
						logger.Any("server_group_ids", groupIDs),
					)
					// Server access is managed through subscription entity itself
					// No additional service calls needed
				}
			}

			// Step 4: Send welcome notification (async)
			if h.taskQueue != nil {
				notificationData := map[string]any{
					"user_id":         subscription.UserID,
					"subscription_id": subscription.ID,
					"type":            "welcome",
					"template":        "subscription_welcome",
				}
				task := queue.NewTask("notification", notificationData)
				if err := h.taskQueue.Enqueue(ctx, "notification", task); err != nil {
					h.logger.Error("Failed to enqueue welcome notification",
						logger.Uint("user_id", subscription.UserID),
						logger.ErrorField(err),
					)
					// Don't fail the process for notification failure
				}
			}

			// Mark event as processed
			h.markEventProcessed(subscriptionEvent.EventID())

			h.logger.Info("Subscription created event processed successfully",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
			)

			return nil
		},
	)
}

// SubscriptionExpiredHandler handles subscription expiry events
// Business Flow: Subscription Expired → Disable User Access → Send Expiry Notifications → Update User Status
func (h *CrossDomainEventHandlers) SubscriptionExpiredHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeSubscriptionExpired},
		func(ctx context.Context, event Event) error {
			subscriptionEvent, ok := event.(*SubscriptionEvent)
			if !ok {
				return fmt.Errorf("expected SubscriptionEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(subscriptionEvent.EventID()) {
				h.logger.Info("Subscription expired event already processed, skipping",
					logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
					logger.String("event_id", subscriptionEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing subscription expired event",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
				logger.String("event_id", subscriptionEvent.EventID()),
			)

			// Step 1: Check if user has any other active subscriptions
			activeSubscriptions, err := h.userSubscriptionService.GetUserActiveSubscriptions(ctx, subscriptionEvent.UserID)
			if err != nil {
				h.logger.Error("Failed to get user active subscriptions",
					logger.Uint("user_id", subscriptionEvent.UserID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get user active subscriptions: %w", err)
			}

			// Filter out the expired subscription from active list
			hasOtherActive := false
			for _, sub := range activeSubscriptions {
				if sub.ID != subscriptionEvent.SubscriptionID && sub.IsActiveForService() {
					hasOtherActive = true
					break
				}
			}

			// Step 2: Update user status to inactive if no other active subscriptions
			if !hasOtherActive {
				user, err := h.userService.GetUserByID(ctx, subscriptionEvent.UserID)
				if err != nil {
					h.logger.Error("Failed to get user",
						logger.Uint("user_id", subscriptionEvent.UserID),
						logger.ErrorField(err),
					)
				} else if user.Status == "active" {
					// Set user status to expired
					if _, err := h.userService.UpdateUserStatus(ctx, user.ID, "expired"); err != nil {
						h.logger.Error("Failed to update user status to expired",
							logger.Uint("user_id", user.ID),
							logger.ErrorField(err),
						)
						// Don't fail the whole process
					} else {
						// Publish user status change event
						userEvent := NewUserEvent(
							EventTypeUserStatusChanged,
							subscriptionEvent.UserID,
							map[string]any{
								"user_id":         subscriptionEvent.UserID,
								"old_status":      "active",
								"new_status":      "expired",
								"reason":          "subscription_expired",
								"subscription_id": subscriptionEvent.SubscriptionID,
								"changed_at":      subscriptionEvent.EventTime(),
							},
						)
						if err := Publish(ctx, userEvent); err != nil {
							h.logger.Error("Failed to publish user status change event",
								logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
								logger.ErrorField(err),
							)
						}
					}
				}
			} else {
				h.logger.Info("User has other active subscriptions, not updating status",
					logger.Uint("user_id", subscriptionEvent.UserID),
					logger.Int("active_subscription_count", len(activeSubscriptions)),
				)
			}

			// Step 3: Send expiry notification (async)
			if h.taskQueue != nil {
				notificationData := map[string]any{
					"user_id":         subscriptionEvent.UserID,
					"subscription_id": subscriptionEvent.SubscriptionID,
					"type":            "expiry",
					"template":        "subscription_expired",
					"has_other_active": hasOtherActive,
				}
				task := queue.NewTask("notification", notificationData)
				if err := h.taskQueue.Enqueue(ctx, "notification", task); err != nil {
					h.logger.Error("Failed to enqueue expiry notification",
						logger.Uint("user_id", subscriptionEvent.UserID),
						logger.ErrorField(err),
					)
					// Don't fail the process for notification failure
				}
			}

			// Mark event as processed
			h.markEventProcessed(subscriptionEvent.EventID())

			h.logger.Info("Subscription expired event processed successfully",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
				logger.Bool("has_other_active", hasOtherActive),
			)

			return nil
		},
	)
}

// UserRegisteredHandler handles user registration events
// Business Flow: User Registered → Initialize User Profile → Send Welcome Email → Create Initial Configuration
func (h *CrossDomainEventHandlers) UserRegisteredHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeUserRegistered, EventTypeUserCreated},
		func(ctx context.Context, event Event) error {
			userEvent, ok := event.(*UserEvent)
			if !ok {
				return fmt.Errorf("expected UserEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(userEvent.EventID()) {
				h.logger.Info("User registered event already processed, skipping",
					logger.Uint("user_id", userEvent.UserID),
					logger.String("event_id", userEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing user registered event",
				logger.Uint("user_id", userEvent.UserID),
				logger.String("event_type", userEvent.EventType()),
				logger.String("event_id", userEvent.EventID()),
			)

			// Step 1: Get user details
			user, err := h.userService.GetUserByID(ctx, userEvent.UserID)
			if err != nil {
				h.logger.Error("Failed to get user",
					logger.Uint("user_id", userEvent.UserID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get user: %w", err)
			}

			// Step 2: Send welcome email notification (async)
			if h.taskQueue != nil {
				welcomeData := map[string]any{
					"user_id":    user.ID,
					"email":      user.Email,
					"username":   user.Username,
					"name":       user.Name,
					"type":       "welcome",
					"template":   "user_welcome",
				}
				task := queue.NewTask("email", welcomeData)
				if err := h.taskQueue.Enqueue(ctx, "email", task); err != nil {
					h.logger.Error("Failed to enqueue welcome email",
						logger.Uint("user_id", user.ID),
						logger.ErrorField(err),
					)
					// Don't fail the process for email failure
				}
			}

			// Step 3: Initialize user configuration cache
			if h.cacheStore != nil {
				userConfigKey := fmt.Sprintf("user_config:%d", user.ID)
				initialConfig := map[string]any{
					"user_id":          user.ID,
					"preferences":      map[string]any{},
					"notification_settings": map[string]bool{
						"email_notifications": true,
						"sms_notifications":   false,
						"push_notifications":  true,
					},
					"created_at": time.Now(),
				}
				if err := h.cacheStore.SetJSON(ctx, userConfigKey, initialConfig, 24*time.Hour); err != nil {
					h.logger.Error("Failed to initialize user configuration cache",
						logger.Uint("user_id", user.ID),
						logger.ErrorField(err),
					)
					// Don't fail the process
				}
			}

			// Mark event as processed
			h.markEventProcessed(userEvent.EventID())

			h.logger.Info("User registered event processed successfully",
				logger.Uint("user_id", userEvent.UserID),
				logger.String("email", user.Email),
			)

			return nil
		},
	)
}

// UserDeletedHandler handles user deletion events
// Business Flow: User Deleted → Cancel Active Subscriptions → Clean Up Data → Send Confirmation
func (h *CrossDomainEventHandlers) UserDeletedHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeUserDeleted},
		func(ctx context.Context, event Event) error {
			userEvent, ok := event.(*UserEvent)
			if !ok {
				return fmt.Errorf("expected UserEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(userEvent.EventID()) {
				h.logger.Info("User deleted event already processed, skipping",
					logger.Uint("user_id", userEvent.UserID),
					logger.String("event_id", userEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing user deleted event",
				logger.Uint("user_id", userEvent.UserID),
				logger.String("event_id", userEvent.EventID()),
			)

			// Step 1: Get and cancel all active subscriptions for the user
			activeSubscriptions, err := h.userSubscriptionService.GetUserActiveSubscriptions(ctx, userEvent.UserID)
			if err != nil {
				h.logger.Error("Failed to get user active subscriptions",
					logger.Uint("user_id", userEvent.UserID),
					logger.ErrorField(err),
				)
				// Continue with cleanup even if we can't get subscriptions
			} else {
				for _, subscription := range activeSubscriptions {
					if subscription.IsActive() {
						// Cancel the subscription
						if err := h.userSubscriptionService.CancelUserSubscription(ctx, subscription.ID, "user_deleted", false); err != nil {
							h.logger.Error("Failed to cancel user subscription",
								logger.Uint("user_id", userEvent.UserID),
								logger.Uint("subscription_id", subscription.ID),
								logger.ErrorField(err),
							)
							// Continue with other subscriptions
						} else {
							// Publish subscription cancelled event
							subscriptionEvent := NewSubscriptionEvent(
								EventTypeSubscriptionCancelled,
								subscription.ID,
								userEvent.UserID,
								map[string]any{
									"subscription_id": subscription.ID,
									"user_id":         userEvent.UserID,
									"reason":          "user_deleted",
									"cancelled_at":    userEvent.EventTime(),
								},
							)

							if err := Publish(ctx, subscriptionEvent); err != nil {
								h.logger.Error("Failed to publish subscription cancelled event",
									logger.Uint("user_id", userEvent.UserID),
									logger.Uint("subscription_id", subscription.ID),
									logger.ErrorField(err),
								)
							}
						}
					}
				}
			}

			// Step 2: Clean up user data from cache
			if h.cacheStore != nil {
				// Remove user configuration
				userConfigKey := fmt.Sprintf("user_config:%d", userEvent.UserID)
				if err := h.cacheStore.Delete(ctx, userConfigKey); err != nil {
					h.logger.Error("Failed to delete user config from cache",
						logger.Uint("user_id", userEvent.UserID),
						logger.ErrorField(err),
					)
				}

				// Remove user cache entries
				userCachePattern := fmt.Sprintf("user:*:%d", userEvent.UserID)
				if err := h.cacheStore.DeletePattern(ctx, userCachePattern); err != nil {
					h.logger.Error("Failed to delete user cache pattern",
						logger.Uint("user_id", userEvent.UserID),
						logger.String("pattern", userCachePattern),
						logger.ErrorField(err),
					)
				}
			}

			// Step 3: Send confirmation notification (async)
			if h.taskQueue != nil {
				confirmationData := map[string]any{
					"user_id": userEvent.UserID,
					"type":    "user_deletion_confirmation",
					"template": "user_deleted",
				}
				task := queue.NewTask("notification", confirmationData)
				if err := h.taskQueue.Enqueue(ctx, "notification", task); err != nil {
					h.logger.Error("Failed to enqueue deletion confirmation",
						logger.Uint("user_id", userEvent.UserID),
						logger.ErrorField(err),
					)
					// Don't fail the process
				}
			}

			// Mark event as processed
			h.markEventProcessed(userEvent.EventID())

			h.logger.Info("User deleted event processed successfully",
				logger.Uint("user_id", userEvent.UserID),
				logger.Int("cancelled_subscriptions", len(activeSubscriptions)),
			)

			return nil
		},
	)
}

// PaymentFailedHandler handles payment failure events
// Business Flow: Payment Failed → Cancel Order → Send Failure Notification → Suggest Retry Options
func (h *CrossDomainEventHandlers) PaymentFailedHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypePaymentFailed},
		func(ctx context.Context, event Event) error {
			paymentEvent, ok := event.(*PaymentEvent)
			if !ok {
				return fmt.Errorf("expected PaymentEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(paymentEvent.EventID()) {
				h.logger.Info("Payment failed event already processed, skipping",
					logger.String("payment_id", paymentEvent.PaymentID),
					logger.String("event_id", paymentEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing payment failed event",
				logger.String("payment_id", paymentEvent.PaymentID),
				logger.Uint("user_id", paymentEvent.UserID),
				logger.String("event_id", paymentEvent.EventID()),
			)

			// Extract order ID from payment data
			orderData, ok := paymentEvent.EventData().(map[string]any)
			if !ok {
				return fmt.Errorf("invalid payment event data format")
			}

			orderIDFloat, exists := orderData["order_id"]
			if !exists {
				return fmt.Errorf("order_id not found in payment event data")
			}

			orderID, ok := orderIDFloat.(float64)
			if !ok {
				return fmt.Errorf("invalid order_id format in payment event")
			}

			// Step 1: Cancel the subscription order
			if err := h.subscriptionOrderService.CancelSubscriptionOrder(ctx, uint(orderID), "payment_failed"); err != nil {
				h.logger.Error("Failed to cancel subscription order",
					logger.String("payment_id", paymentEvent.PaymentID),
					logger.Uint("order_id", uint(orderID)),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to cancel subscription order: %w", err)
			}

			// Step 2: Get order and user details for notification
			order, err := h.subscriptionOrderService.GetSubscriptionOrder(ctx, uint(orderID))
			if err != nil {
				h.logger.Error("Failed to get subscription order for failed payment",
					logger.Uint("order_id", uint(orderID)),
					logger.ErrorField(err),
				)
				// Continue with notification even if we can't get order details
			}

			// Step 3: Send payment failure notification (async)
			if h.taskQueue != nil {
				var userID uint
				if order != nil {
					userID = order.UserID
				} else {
					userID = paymentEvent.UserID
				}

				user, err := h.userService.GetUserByID(ctx, userID)
				if err == nil && user.Email != "" {
					failureData := map[string]any{
						"payment_id":       paymentEvent.PaymentID,
						"user_id":          user.ID,
						"order_id":         uint(orderID),
						"to_email":         user.Email,
						"user_name":        user.Name,
						"payment_amount":   paymentEvent.Amount,
						"failure_reason":   orderData["failure_reason"],
						"retry_available":  true,
						"type":             "payment_failed",
						"template":         "payment_failed",
					}
					task := queue.NewTask("email", failureData)
					if err := h.taskQueue.Enqueue(ctx, "email", task); err != nil {
						h.logger.Error("Failed to enqueue payment failure notification",
							logger.String("payment_id", paymentEvent.PaymentID),
							logger.ErrorField(err),
						)
					}
				}
			}

			// Step 4: Publish order cancelled event
			orderEvent := NewOrderEvent(
				EventTypeOrderCancelled,
				uint(orderID),
				paymentEvent.UserID,
				map[string]any{
					"order_id":       uint(orderID),
					"user_id":        paymentEvent.UserID,
					"reason":         "payment_failed",
					"payment_id":     paymentEvent.PaymentID,
					"cancelled_at":   paymentEvent.EventTime(),
					"failure_reason": orderData["failure_reason"],
				},
			)

			if err := Publish(ctx, orderEvent); err != nil {
				h.logger.Error("Failed to publish order cancelled event",
					logger.String("payment_id", paymentEvent.PaymentID),
					logger.ErrorField(err),
				)
				// Don't fail the whole process for event publishing failure
			}

			// Mark event as processed
			h.markEventProcessed(paymentEvent.EventID())

			h.logger.Info("Payment failed event processed successfully",
				logger.String("payment_id", paymentEvent.PaymentID),
				logger.Uint("order_id", uint(orderID)),
			)

			return nil
		},
	)
}

// InvoiceGeneratedHandler handles invoice generation events
// Business Flow: Invoice Generated → Send Invoice Email → Update User Billing Status → Schedule Follow-up
func (h *CrossDomainEventHandlers) InvoiceGeneratedHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeInvoiceGenerated, EventTypeInvoiceCreated},
		func(ctx context.Context, event Event) error {
			invoiceEvent, ok := event.(*InvoiceEvent)
			if !ok {
				return fmt.Errorf("expected InvoiceEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(invoiceEvent.EventID()) {
				h.logger.Info("Invoice generated event already processed, skipping",
					logger.Uint("invoice_id", invoiceEvent.InvoiceID),
					logger.String("event_id", invoiceEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing invoice generated event",
				logger.Uint("invoice_id", invoiceEvent.InvoiceID),
				logger.Uint("user_id", invoiceEvent.UserID),
				logger.String("event_id", invoiceEvent.EventID()),
			)

			// Step 1: Get invoice details
			invoice, err := h.invoiceService.GetInvoice(ctx, invoiceEvent.InvoiceID)
			if err != nil {
				h.logger.Error("Failed to get invoice",
					logger.Uint("invoice_id", invoiceEvent.InvoiceID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get invoice: %w", err)
			}

			// Step 2: Get user details for email sending
			user, err := h.userService.GetUserByID(ctx, invoice.UserID)
			if err != nil {
				h.logger.Error("Failed to get user for invoice",
					logger.Uint("user_id", invoice.UserID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get user: %w", err)
			}

			// Step 3: Generate and send invoice PDF via email (async)
			if h.taskQueue != nil && user.Email != "" {
				emailData := map[string]any{
					"invoice_id":     invoice.ID,
					"user_id":        user.ID,
					"to_email":       user.Email,
					"user_name":      user.Name,
					"invoice_number": invoice.InvoiceNumber,
					"amount":         invoice.Amount,
					"currency":       invoice.Currency,
					"due_date":       invoice.DueAt,
					"type":           "invoice",
					"template":       "invoice_generated",
					"generate_pdf":   true,
				}
				task := queue.NewTask("email", emailData)
				if err := h.taskQueue.Enqueue(ctx, "email", task); err != nil {
					h.logger.Error("Failed to enqueue invoice email",
						logger.Uint("invoice_id", invoice.ID),
						logger.String("user_email", user.Email),
						logger.ErrorField(err),
					)
					// Don't fail the process for email failure
				}
			}

			// Step 4: Update user billing status cache
			if h.cacheStore != nil {
				billingStatusKey := fmt.Sprintf("user_billing:%d", user.ID)
				billingStatus := map[string]any{
					"user_id":              user.ID,
					"latest_invoice_id":    invoice.ID,
					"latest_invoice_amount": invoice.Amount,
					"latest_invoice_status": invoice.Status,
					"last_invoice_date":    invoice.CreatedAt,
					"updated_at":           time.Now(),
				}
				if err := h.cacheStore.SetJSON(ctx, billingStatusKey, billingStatus, 24*time.Hour); err != nil {
					h.logger.Error("Failed to update user billing status cache",
						logger.Uint("user_id", user.ID),
						logger.ErrorField(err),
					)
				}
			}

			// Mark event as processed
			h.markEventProcessed(invoiceEvent.EventID())

			h.logger.Info("Invoice generated event processed successfully",
				logger.Uint("invoice_id", invoiceEvent.InvoiceID),
				logger.String("invoice_number", invoice.InvoiceNumber),
				logger.String("user_email", user.Email),
			)

			return nil
		},
	)
}

// InvoiceOverdueHandler handles overdue invoice events
// Business Flow: Invoice Overdue → Suspend Related Subscription → Send Overdue Notices → Escalate Collection
func (h *CrossDomainEventHandlers) InvoiceOverdueHandler() EventHandler {
	return NewEventHandler(
		[]string{EventTypeInvoiceOverdue},
		func(ctx context.Context, event Event) error {
			invoiceEvent, ok := event.(*InvoiceEvent)
			if !ok {
				return fmt.Errorf("expected InvoiceEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(invoiceEvent.EventID()) {
				h.logger.Info("Invoice overdue event already processed, skipping",
					logger.Uint("invoice_id", invoiceEvent.InvoiceID),
					logger.String("event_id", invoiceEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing invoice overdue event",
				logger.Uint("invoice_id", invoiceEvent.InvoiceID),
				logger.Uint("user_id", invoiceEvent.UserID),
				logger.String("event_id", invoiceEvent.EventID()),
			)

			// Step 1: Get invoice details
			invoice, err := h.invoiceService.GetInvoice(ctx, invoiceEvent.InvoiceID)
			if err != nil {
				h.logger.Error("Failed to get overdue invoice",
					logger.Uint("invoice_id", invoiceEvent.InvoiceID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get invoice: %w", err)
			}

			// Step 2: Find and suspend related subscription
			if invoice.SubscriptionOrderID > 0 {
				// Get the order to find subscription
				order, err := h.subscriptionOrderService.GetSubscriptionOrder(ctx, invoice.SubscriptionOrderID)
				if err != nil {
					h.logger.Error("Failed to get subscription order for overdue invoice",
						logger.Uint("order_id", invoice.SubscriptionOrderID),
						logger.ErrorField(err),
					)
				} else {
					// Find active subscription for this user and plan
					activeSubscription, err := h.userSubscriptionService.GetActiveUserSubscription(ctx, order.UserID, order.SubscriptionPlanID)
					if err == nil && activeSubscription != nil && activeSubscription.IsActive() {
						// Pause the subscription due to overdue invoice
						pauseReq := &subscriptionInterfaces.PauseSubscriptionRequest{
							Reason: fmt.Sprintf("Invoice #%s overdue", invoice.InvoiceNumber),
						}
						if _, err := h.userSubscriptionService.PauseUserSubscription(ctx, activeSubscription.ID, pauseReq, 0); err != nil {
							h.logger.Error("Failed to pause subscription for overdue invoice",
								logger.Uint("subscription_id", activeSubscription.ID),
								logger.Uint("invoice_id", invoice.ID),
								logger.ErrorField(err),
							)
						} else {
							// Publish subscription suspended event
							subscriptionEvent := NewSubscriptionEvent(
								EventTypeSubscriptionSuspended,
								activeSubscription.ID,
								invoiceEvent.UserID,
								map[string]any{
									"subscription_id": activeSubscription.ID,
									"user_id":         invoiceEvent.UserID,
									"reason":          "invoice_overdue",
									"invoice_id":      invoiceEvent.InvoiceID,
									"suspended_at":    invoiceEvent.EventTime(),
								},
							)

							if err := Publish(ctx, subscriptionEvent); err != nil {
								h.logger.Error("Failed to publish subscription suspended event",
									logger.Uint("invoice_id", invoiceEvent.InvoiceID),
									logger.ErrorField(err),
								)
							}
						}
					}
				}
			}

			// Step 3: Send overdue notification (async)
			if h.taskQueue != nil {
				user, err := h.userService.GetUserByID(ctx, invoice.UserID)
				if err == nil && user.Email != "" {
					overdueData := map[string]any{
						"invoice_id":     invoice.ID,
						"user_id":        user.ID,
						"to_email":       user.Email,
						"user_name":      user.Name,
						"invoice_number": invoice.InvoiceNumber,
						"amount":         invoice.Amount,
						"currency":       invoice.Currency,
						"due_date":       invoice.DueAt,
						"days_overdue":   time.Since(*invoice.DueAt).Hours() / 24,
						"type":           "overdue_notice",
						"template":       "invoice_overdue",
					}
					task := queue.NewTask("email", overdueData)
					if err := h.taskQueue.Enqueue(ctx, "email", task); err != nil {
						h.logger.Error("Failed to enqueue overdue notice",
							logger.Uint("invoice_id", invoice.ID),
							logger.ErrorField(err),
						)
					}
				}
			}

			// Mark event as processed
			h.markEventProcessed(invoiceEvent.EventID())

			h.logger.Info("Invoice overdue event processed successfully",
				logger.Uint("invoice_id", invoiceEvent.InvoiceID),
				logger.String("invoice_number", invoice.InvoiceNumber),
			)

			return nil
		},
	)
}

// TrafficLimitExceededHandler handles traffic limit exceeded events
// Business Flow: Traffic Limit Exceeded → Suspend User Access → Send Usage Alert → Suggest Upgrade
func (h *CrossDomainEventHandlers) TrafficLimitExceededHandler() EventHandler {
	return NewEventHandler(
		[]string{"subscription.traffic_limit_exceeded"},
		func(ctx context.Context, event Event) error {
			subscriptionEvent, ok := event.(*SubscriptionEvent)
			if !ok {
				return fmt.Errorf("expected SubscriptionEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(subscriptionEvent.EventID()) {
				h.logger.Info("Traffic limit exceeded event already processed, skipping",
					logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
					logger.String("event_id", subscriptionEvent.EventID()),
				)
				return nil
			}

			h.logger.Info("Processing traffic limit exceeded event",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
				logger.String("event_id", subscriptionEvent.EventID()),
			)

			// Step 1: Get subscription details
			subscription, err := h.userSubscriptionService.GetUserSubscription(ctx, subscriptionEvent.SubscriptionID)
			if err != nil {
				h.logger.Error("Failed to get subscription for traffic limit check",
					logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get subscription: %w", err)
			}

			// Step 2: Get traffic usage stats (for logging)
			_, err = h.userSubscriptionService.GetSubscriptionTrafficStats(ctx, subscription.ID)
			if err != nil {
				h.logger.Error("Failed to get traffic stats",
					logger.Uint("subscription_id", subscription.ID),
					logger.ErrorField(err),
				)
			} else {
				h.logger.Info("Traffic usage stats",
					logger.Uint("subscription_id", subscription.ID),
					logger.Int64("traffic_used", subscription.TrafficUsed),
					logger.Int64("traffic_limit", subscription.TrafficLimit),
					logger.Float64("usage_percentage", subscription.GetTrafficUsagePercentage()),
				)
			}

			// Step 3: Send usage alert notification (async)
			if h.taskQueue != nil {
				user, err := h.userService.GetUserByID(ctx, subscription.UserID)
				if err == nil && user.Email != "" {
					usageAlertData := map[string]any{
						"subscription_id":   subscription.ID,
						"user_id":           user.ID,
						"to_email":          user.Email,
						"user_name":         user.Name,
						"traffic_used":      subscription.TrafficUsed,
						"traffic_limit":     subscription.TrafficLimit,
						"usage_percentage":  subscription.GetTrafficUsagePercentage(),
						"remaining_traffic": subscription.GetRemainingTraffic(),
						"reset_date":        subscription.TrafficResetDate,
						"type":              "traffic_limit_exceeded",
						"template":          "traffic_limit_exceeded",
						"suspended":         subscription.TrafficSuspended,
					}
					task := queue.NewTask("email", usageAlertData)
					if err := h.taskQueue.Enqueue(ctx, "email", task); err != nil {
						h.logger.Error("Failed to enqueue traffic limit exceeded notification",
							logger.Uint("user_id", user.ID),
							logger.ErrorField(err),
						)
					}
				}
			}

			// Step 4: Create usage alert record (async)
			if h.taskQueue != nil {
				alertData := map[string]any{
					"subscription_id": subscription.ID,
					"user_id":         subscription.UserID,
					"alert_type":      "traffic_limit_exceeded",
					"threshold":       100.0, // 100% usage
					"current_usage":   subscription.GetTrafficUsagePercentage(),
					"traffic_used":    subscription.TrafficUsed,
					"traffic_limit":   subscription.TrafficLimit,
					"triggered_at":    time.Now(),
				}
				task := queue.NewTask("data_processing", alertData)
				if err := h.taskQueue.Enqueue(ctx, "data_processing", task); err != nil {
					h.logger.Error("Failed to enqueue usage alert creation",
						logger.Uint("subscription_id", subscription.ID),
						logger.ErrorField(err),
					)
				}
			}

			// Mark event as processed
			h.markEventProcessed(subscriptionEvent.EventID())

			h.logger.Info("Traffic limit exceeded event processed successfully",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
				logger.Bool("traffic_suspended", subscription.TrafficSuspended),
			)

			return nil
		},
	)
}

// TrafficUsageWarningHandler handles traffic usage warning events (80%, 90%)
// Business Flow: Usage Warning → Send Warning Email → Suggest Actions → Track Alert
func (h *CrossDomainEventHandlers) TrafficUsageWarningHandler() EventHandler {
	return NewEventHandler(
		[]string{"subscription.traffic_usage_warning"},
		func(ctx context.Context, event Event) error {
			subscriptionEvent, ok := event.(*SubscriptionEvent)
			if !ok {
				return fmt.Errorf("expected SubscriptionEvent, got %T", event)
			}

			// Idempotency check
			if h.isEventProcessed(subscriptionEvent.EventID()) {
				h.logger.Info("Traffic usage warning event already processed, skipping",
					logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
					logger.String("event_id", subscriptionEvent.EventID()),
				)
				return nil
			}

			eventData := subscriptionEvent.EventData().(map[string]any)
			threshold, _ := eventData["threshold"].(float64)
			usagePercentage, _ := eventData["usage_percentage"].(float64)

			h.logger.Info("Processing traffic usage warning event",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
				logger.Float64("threshold", threshold),
				logger.Float64("usage_percentage", usagePercentage),
				logger.String("event_id", subscriptionEvent.EventID()),
			)

			// Step 1: Get subscription details
			subscription, err := h.userSubscriptionService.GetUserSubscription(ctx, subscriptionEvent.SubscriptionID)
			if err != nil {
				h.logger.Error("Failed to get subscription for usage warning",
					logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
					logger.ErrorField(err),
				)
				return fmt.Errorf("failed to get subscription: %w", err)
			}

			// Step 2: Send usage warning notification (async)
			if h.taskQueue != nil {
				user, err := h.userService.GetUserByID(ctx, subscription.UserID)
				if err == nil && user.Email != "" {
					warningData := map[string]any{
						"subscription_id":   subscription.ID,
						"user_id":           user.ID,
						"to_email":          user.Email,
						"user_name":         user.Name,
						"traffic_used":      subscription.TrafficUsed,
						"traffic_limit":     subscription.TrafficLimit,
						"usage_percentage":  usagePercentage,
						"warning_threshold": threshold,
						"remaining_traffic": subscription.GetRemainingTraffic(),
						"reset_date":        subscription.TrafficResetDate,
						"type":              "traffic_usage_warning",
						"template":          fmt.Sprintf("traffic_warning_%d", int(threshold)),
					}
					task := queue.NewTask("email", warningData)
					if err := h.taskQueue.Enqueue(ctx, "email", task); err != nil {
						h.logger.Error("Failed to enqueue traffic usage warning notification",
							logger.Uint("user_id", user.ID),
							logger.ErrorField(err),
						)
					}
				}
			}

			// Step 3: Create usage alert record (async)
			if h.taskQueue != nil {
				alertData := map[string]any{
					"subscription_id": subscription.ID,
					"user_id":         subscription.UserID,
					"alert_type":      "traffic_usage_warning",
					"threshold":       threshold,
					"current_usage":   usagePercentage,
					"traffic_used":    subscription.TrafficUsed,
					"traffic_limit":   subscription.TrafficLimit,
					"triggered_at":    time.Now(),
				}
				task := queue.NewTask("data_processing", alertData)
				if err := h.taskQueue.Enqueue(ctx, "data_processing", task); err != nil {
					h.logger.Error("Failed to enqueue usage warning alert creation",
						logger.Uint("subscription_id", subscription.ID),
						logger.ErrorField(err),
					)
				}
			}

			// Mark event as processed
			h.markEventProcessed(subscriptionEvent.EventID())

			h.logger.Info("Traffic usage warning event processed successfully",
				logger.Uint("subscription_id", subscriptionEvent.SubscriptionID),
				logger.Uint("user_id", subscriptionEvent.UserID),
				logger.Float64("threshold", threshold),
				logger.Float64("usage_percentage", usagePercentage),
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
		h.SubscriptionCreatedHandler(),
		h.SubscriptionExpiredHandler(),
		h.UserRegisteredHandler(),
		h.UserDeletedHandler(),
		h.PaymentFailedHandler(),
		h.InvoiceGeneratedHandler(),
		h.InvoiceOverdueHandler(),
		h.TrafficLimitExceededHandler(),
		h.TrafficUsageWarningHandler(),
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
