package cache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"linke/internal/shared/events"
	"linke/internal/shared/logger"
)

// InvalidationRule defines how cache invalidation should be handled for specific events
type InvalidationRule struct {
	EventTypes       []string                `json:"event_types"`
	KeyPatterns      []string                `json:"key_patterns"`
	InvalidationType InvalidationType        `json:"invalidation_type"`
	Condition        func(events.Event) bool `json:"-"` // Custom condition function
}

// InvalidationType defines different cache invalidation strategies
type InvalidationType string

const (
	InvalidationTypePattern InvalidationType = "pattern" // Delete by pattern
	InvalidationTypeExact   InvalidationType = "exact"   // Delete exact keys
	InvalidationTypeCascade InvalidationType = "cascade" // Cascade to related keys
	InvalidationTypeTag     InvalidationType = "tag"     // Delete by cache tags
	InvalidationTypePartial InvalidationType = "partial" // Partial invalidation
)

// CacheInvalidationConfig configures event-driven cache invalidation
type CacheInvalidationConfig struct {
	Enabled       bool               `json:"enabled"`
	Rules         []InvalidationRule `json:"rules"`
	AsyncMode     bool               `json:"async_mode"`
	BatchSize     int                `json:"batch_size"`
	BufferTimeout time.Duration      `json:"buffer_timeout"`
}

// EventDrivenInvalidator handles cache invalidation based on domain events
type EventDrivenInvalidator struct {
	cache     Cache
	cacheKeys *AllCacheKeys
	config    *CacheInvalidationConfig
	logger    logger.Logger

	// Rule mapping for efficient lookup
	eventRules map[string][]InvalidationRule

	// Async processing
	invalidationQueue chan *invalidationJob
	stopChan          chan struct{}
	wg                sync.WaitGroup

	// Metrics
	metrics struct {
		EventsProcessed    int64
		InvalidationsCount int64
		ErrorsCount        int64
	}
	metricsMu sync.RWMutex
}

type invalidationJob struct {
	event events.Event
	rule  InvalidationRule
}

// NewEventDrivenInvalidator creates a new event-driven cache invalidator
func NewEventDrivenInvalidator(
	cache Cache,
	cacheKeys *AllCacheKeys,
	config *CacheInvalidationConfig,
	logger logger.Logger,
) *EventDrivenInvalidator {
	if config == nil {
		config = &CacheInvalidationConfig{
			Enabled:       true,
			AsyncMode:     true,
			BatchSize:     100,
			BufferTimeout: 1 * time.Second,
		}
	}

	invalidator := &EventDrivenInvalidator{
		cache:             cache,
		cacheKeys:         cacheKeys,
		config:            config,
		logger:            logger,
		eventRules:        make(map[string][]InvalidationRule),
		invalidationQueue: make(chan *invalidationJob, 1000),
		stopChan:          make(chan struct{}),
	}

	// Build event rules mapping
	invalidator.buildEventRulesMapping()

	// Add default invalidation rules
	invalidator.addDefaultRules()

	// Start async processor if enabled
	if config.AsyncMode {
		invalidator.startAsyncProcessor()
	}

	return invalidator
}

// Handle processes events and triggers cache invalidation
func (edi *EventDrivenInvalidator) Handle(ctx context.Context, event events.Event) error {
	if !edi.config.Enabled {
		return nil
	}

	edi.updateMetrics(func() {
		edi.metrics.EventsProcessed++
	})

	rules, exists := edi.eventRules[event.EventType()]
	if !exists {
		// No rules for this event type
		return nil
	}

	for _, rule := range rules {
		// Check custom condition if exists
		if rule.Condition != nil && !rule.Condition(event) {
			continue
		}

		if edi.config.AsyncMode {
			// Queue for async processing
			select {
			case edi.invalidationQueue <- &invalidationJob{event: event, rule: rule}:
			default:
				// Queue full, process synchronously
				edi.processInvalidation(ctx, event, rule)
			}
		} else {
			// Process synchronously
			edi.processInvalidation(ctx, event, rule)
		}
	}

	return nil
}

// EventTypes returns the event types this handler is interested in
func (edi *EventDrivenInvalidator) EventTypes() []string {
	var eventTypes []string
	for eventType := range edi.eventRules {
		eventTypes = append(eventTypes, eventType)
	}
	return eventTypes
}

// AddRule adds a custom invalidation rule
func (edi *EventDrivenInvalidator) AddRule(rule InvalidationRule) {
	for _, eventType := range rule.EventTypes {
		edi.eventRules[eventType] = append(edi.eventRules[eventType], rule)
	}
}

// Stop gracefully shuts down the invalidator
func (edi *EventDrivenInvalidator) Stop() {
	if edi.config.AsyncMode {
		close(edi.stopChan)
		edi.wg.Wait()
	}
}

// GetMetrics returns invalidation metrics
func (edi *EventDrivenInvalidator) GetMetrics() map[string]int64 {
	edi.metricsMu.RLock()
	defer edi.metricsMu.RUnlock()

	return map[string]int64{
		"events_processed":    edi.metrics.EventsProcessed,
		"invalidations_count": edi.metrics.InvalidationsCount,
		"errors_count":        edi.metrics.ErrorsCount,
	}
}

// Private methods

func (edi *EventDrivenInvalidator) processInvalidation(ctx context.Context, event events.Event, rule InvalidationRule) {
	defer func() {
		edi.updateMetrics(func() {
			edi.metrics.InvalidationsCount++
		})
	}()

	switch rule.InvalidationType {
	case InvalidationTypePattern:
		edi.invalidateByPattern(ctx, event, rule)
	case InvalidationTypeExact:
		edi.invalidateExactKeys(ctx, event, rule)
	case InvalidationTypeCascade:
		edi.invalidateCascade(ctx, event, rule)
	case InvalidationTypeTag:
		edi.invalidateByTag(ctx, event, rule)
	case InvalidationTypePartial:
		edi.invalidatePartial(ctx, event, rule)
	default:
		edi.logger.Warn("Unknown invalidation type",
			logger.String("type", string(rule.InvalidationType)))
	}
}

func (edi *EventDrivenInvalidator) invalidateByPattern(ctx context.Context, event events.Event, rule InvalidationRule) {
	for _, pattern := range rule.KeyPatterns {
		// Replace placeholders in pattern with actual values from event
		actualPattern := edi.replacePlaceholders(pattern, event)

		if err := edi.cache.DeleteByPattern(ctx, actualPattern); err != nil {
			edi.handleInvalidationError(err, "pattern", actualPattern)
		} else {
			edi.logger.Debug("Invalidated cache by pattern",
				logger.String("pattern", actualPattern),
				logger.String("event_type", event.EventType()))
		}
	}
}

func (edi *EventDrivenInvalidator) invalidateExactKeys(ctx context.Context, event events.Event, rule InvalidationRule) {
	for _, keyTemplate := range rule.KeyPatterns {
		// Replace placeholders to get exact key
		exactKey := edi.replacePlaceholders(keyTemplate, event)

		if err := edi.cache.Delete(ctx, exactKey); err != nil {
			edi.handleInvalidationError(err, "exact", exactKey)
		} else {
			edi.logger.Debug("Invalidated exact cache key",
				logger.String("key", exactKey),
				logger.String("event_type", event.EventType()))
		}
	}
}

func (edi *EventDrivenInvalidator) invalidateCascade(ctx context.Context, event events.Event, rule InvalidationRule) {
	// Cascade invalidation involves invalidating related keys
	// For example, when a user is updated, invalidate user caches AND related subscription caches

	switch event.EventType() {
	case events.EventTypeUserUpdated, events.EventTypeUserDeleted:
		if userEvent, ok := event.(*events.UserEvent); ok {
			patterns := []string{
				edi.cacheKeys.User.UserPattern(userEvent.UserID),
				edi.cacheKeys.Subscription.UserSubscription(userEvent.UserID),
				edi.cacheKeys.Subscription.UserActiveSubscriptions(userEvent.UserID),
				edi.cacheKeys.Payment.UserPayments(userEvent.UserID),
			}

			for _, pattern := range patterns {
				if err := edi.cache.DeleteByPattern(ctx, pattern); err != nil {
					edi.handleInvalidationError(err, "cascade", pattern)
				}
			}
		}

	case events.EventTypeSubscriptionUpdated, events.EventTypeSubscriptionActivated:
		if subEvent, ok := event.(*events.SubscriptionEvent); ok {
			patterns := []string{
				edi.cacheKeys.Subscription.UserSubscription(subEvent.UserID),
				edi.cacheKeys.Subscription.UserActiveSubscriptions(subEvent.UserID),
				edi.cacheKeys.User.UserPattern(subEvent.UserID), // User cache might contain subscription info
			}

			for _, pattern := range patterns {
				if err := edi.cache.DeleteByPattern(ctx, pattern); err != nil {
					edi.handleInvalidationError(err, "cascade", pattern)
				}
			}
		}
	}
}

func (edi *EventDrivenInvalidator) invalidateByTag(ctx context.Context, event events.Event, rule InvalidationRule) {
	// This would require a cache implementation that supports tagging
	// For now, we'll use pattern-based invalidation as a fallback
	edi.invalidateByPattern(ctx, event, rule)
}

func (edi *EventDrivenInvalidator) invalidatePartial(ctx context.Context, event events.Event, rule InvalidationRule) {
	// Partial invalidation selectively invalidates specific fields or sections
	// This is more complex and would require structured cache keys

	// For now, implement as pattern-based invalidation
	edi.invalidateByPattern(ctx, event, rule)
}

func (edi *EventDrivenInvalidator) replacePlaceholders(pattern string, event events.Event) string {
	result := pattern

	// Replace common placeholders based on event type
	switch e := event.(type) {
	case *events.UserEvent:
		result = strings.ReplaceAll(result, "{user_id}", fmt.Sprintf("%d", e.UserID))

	case *events.SubscriptionEvent:
		result = strings.ReplaceAll(result, "{user_id}", fmt.Sprintf("%d", e.UserID))
		result = strings.ReplaceAll(result, "{subscription_id}", fmt.Sprintf("%d", e.SubscriptionID))

	case *events.PaymentEvent:
		result = strings.ReplaceAll(result, "{user_id}", fmt.Sprintf("%d", e.UserID))
		result = strings.ReplaceAll(result, "{payment_id}", e.PaymentID)

	case *events.OrderEvent:
		result = strings.ReplaceAll(result, "{user_id}", fmt.Sprintf("%d", e.UserID))
		result = strings.ReplaceAll(result, "{order_id}", fmt.Sprintf("%d", e.OrderID))

	case *events.InvoiceEvent:
		result = strings.ReplaceAll(result, "{user_id}", fmt.Sprintf("%d", e.UserID))
		result = strings.ReplaceAll(result, "{invoice_id}", fmt.Sprintf("%d", e.InvoiceID))
		result = strings.ReplaceAll(result, "{order_id}", fmt.Sprintf("%d", e.OrderID))
	}

	return result
}

func (edi *EventDrivenInvalidator) buildEventRulesMapping() {
	for _, rule := range edi.config.Rules {
		for _, eventType := range rule.EventTypes {
			edi.eventRules[eventType] = append(edi.eventRules[eventType], rule)
		}
	}
}

func (edi *EventDrivenInvalidator) addDefaultRules() {
	// User-related invalidation rules
	edi.AddRule(InvalidationRule{
		EventTypes:       []string{events.EventTypeUserUpdated, events.EventTypeUserDeleted},
		KeyPatterns:      []string{CachePrefixUser + ":{user_id}:*"},
		InvalidationType: InvalidationTypePattern,
	})

	// Subscription-related invalidation rules
	edi.AddRule(InvalidationRule{
		EventTypes: []string{
			events.EventTypeSubscriptionCreated,
			events.EventTypeSubscriptionUpdated,
			events.EventTypeSubscriptionActivated,
			events.EventTypeSubscriptionCancelled,
		},
		KeyPatterns:      []string{CachePrefixSubscription + "*{user_id}*"},
		InvalidationType: InvalidationTypeCascade,
	})

	// Payment-related invalidation rules
	edi.AddRule(InvalidationRule{
		EventTypes: []string{
			events.EventTypePaymentCompleted,
			events.EventTypePaymentFailed,
			events.EventTypePaymentRefunded,
		},
		KeyPatterns:      []string{CachePrefixPayment + "*{payment_id}*", CachePrefixPayment + "*{user_id}*"},
		InvalidationType: InvalidationTypePattern,
	})

	// Order-related invalidation rules
	edi.AddRule(InvalidationRule{
		EventTypes: []string{
			events.EventTypeOrderCreated,
			events.EventTypeOrderUpdated,
			events.EventTypeOrderPaid,
			events.EventTypeOrderCancelled,
		},
		KeyPatterns:      []string{CachePrefixSubscription + "*{order_id}*", CachePrefixSubscription + "*{user_id}*"},
		InvalidationType: InvalidationTypePattern,
	})

	// Plan invalidation rules (affects everyone)
	edi.AddRule(InvalidationRule{
		EventTypes:       []string{"plan.updated", "plan.deleted"},
		KeyPatterns:      []string{CachePrefixPlan + "*"},
		InvalidationType: InvalidationTypePattern,
	})
}

func (edi *EventDrivenInvalidator) startAsyncProcessor() {
	edi.wg.Add(1)
	go func() {
		defer edi.wg.Done()

		for {
			select {
			case job := <-edi.invalidationQueue:
				edi.processInvalidation(context.Background(), job.event, job.rule)

			case <-edi.stopChan:
				// Drain remaining jobs
				for {
					select {
					case job := <-edi.invalidationQueue:
						edi.processInvalidation(context.Background(), job.event, job.rule)
					default:
						return
					}
				}
			}
		}
	}()
}

func (edi *EventDrivenInvalidator) handleInvalidationError(err error, invalidationType, key string) {
	edi.updateMetrics(func() {
		edi.metrics.ErrorsCount++
	})

	edi.logger.Error("Cache invalidation failed",
		logger.String("type", invalidationType),
		logger.String("key", key),
		logger.Error2("error", err))
}

func (edi *EventDrivenInvalidator) updateMetrics(updateFunc func()) {
	edi.metricsMu.Lock()
	updateFunc()
	edi.metricsMu.Unlock()
}

// SmartInvalidator provides intelligent cache invalidation based on access patterns
type SmartInvalidator struct {
	cache         Cache
	accessTracker *AccessTracker
	logger        logger.Logger
}

// AccessTracker tracks cache access patterns for intelligent invalidation
type AccessTracker struct {
	mu           sync.RWMutex
	accessCounts map[string]int64
	lastAccess   map[string]time.Time
}

// NewSmartInvalidator creates a smart cache invalidator
func NewSmartInvalidator(cache Cache, logger logger.Logger) *SmartInvalidator {
	return &SmartInvalidator{
		cache: cache,
		accessTracker: &AccessTracker{
			accessCounts: make(map[string]int64),
			lastAccess:   make(map[string]time.Time),
		},
		logger: logger,
	}
}

// TrackAccess records access to a cache key
func (si *SmartInvalidator) TrackAccess(key string) {
	si.accessTracker.mu.Lock()
	si.accessTracker.accessCounts[key]++
	si.accessTracker.lastAccess[key] = time.Now()
	si.accessTracker.mu.Unlock()
}

// InvalidateIntelligently invalidates cache based on access patterns
func (si *SmartInvalidator) InvalidateIntelligently(ctx context.Context, pattern string, threshold int64) error {
	si.accessTracker.mu.RLock()
	keysToInvalidate := make([]string, 0)

	for key, count := range si.accessTracker.accessCounts {
		if strings.Contains(key, pattern) && count < threshold {
			keysToInvalidate = append(keysToInvalidate, key)
		}
	}
	si.accessTracker.mu.RUnlock()

	// Invalidate low-access keys
	for _, key := range keysToInvalidate {
		if err := si.cache.Delete(ctx, key); err != nil {
			si.logger.Error("Smart invalidation failed",
				logger.String("key", key),
				logger.Error2("error", err))
		}
	}

	si.logger.Info("Smart invalidation completed",
		logger.String("pattern", pattern),
		logger.Int("invalidated_count", len(keysToInvalidate)))

	return nil
}
