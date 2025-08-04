package events

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"linke/internal/shared/logger"
)

// EventMetricsCollector provides comprehensive metrics for event processing
type EventMetricsCollector struct {
	// Counters
	totalEventsPublished     int64
	totalEventsProcessed     int64
	totalEventsFailed        int64
	totalEventsInDeadLetter  int64
	totalCircuitBreakerTrips int64

	// Processing time tracking
	totalProcessingTime int64 // nanoseconds
	minProcessingTime   int64 // nanoseconds
	maxProcessingTime   int64 // nanoseconds

	// Event type specific metrics
	eventTypeMetrics map[string]*EventTypeMetrics

	// Handler specific metrics
	handlerMetrics map[string]*HandlerMetrics

	// Time-based metrics
	metricsWindow    time.Duration
	buckets          []*MetricsBucket
	currentBucketIdx int

	mutex  sync.RWMutex
	logger logger.Logger
}

// EventTypeMetrics tracks metrics for a specific event type
type EventTypeMetrics struct {
	EventType             string        `json:"event_type"`
	PublishedCount        int64         `json:"published_count"`
	ProcessedCount        int64         `json:"processed_count"`
	FailedCount           int64         `json:"failed_count"`
	DeadLetterCount       int64         `json:"dead_letter_count"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	LastProcessedAt       time.Time     `json:"last_processed_at"`
}

// HandlerMetrics tracks metrics for a specific event handler
type HandlerMetrics struct {
	HandlerName           string        `json:"handler_name"`
	ProcessedCount        int64         `json:"processed_count"`
	FailedCount           int64         `json:"failed_count"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	CircuitBreakerTrips   int64         `json:"circuit_breaker_trips"`
	LastProcessedAt       time.Time     `json:"last_processed_at"`
}

// MetricsBucket represents metrics for a specific time window
type MetricsBucket struct {
	Timestamp       time.Time     `json:"timestamp"`
	EventsPublished int64         `json:"events_published"`
	EventsProcessed int64         `json:"events_processed"`
	EventsFailed    int64         `json:"events_failed"`
	AverageLatency  time.Duration `json:"average_latency"`
}

// EventMetricsSnapshot provides a point-in-time view of all metrics
type EventMetricsSnapshot struct {
	Timestamp               time.Time                    `json:"timestamp"`
	TotalEventsPublished    int64                        `json:"total_events_published"`
	TotalEventsProcessed    int64                        `json:"total_events_processed"`
	TotalEventsFailed       int64                        `json:"total_events_failed"`
	TotalEventsInDeadLetter int64                        `json:"total_events_in_dead_letter"`
	CircuitBreakerTrips     int64                        `json:"circuit_breaker_trips"`
	AverageProcessingTime   time.Duration                `json:"average_processing_time"`
	MinProcessingTime       time.Duration                `json:"min_processing_time"`
	MaxProcessingTime       time.Duration                `json:"max_processing_time"`
	SuccessRate             float64                      `json:"success_rate"`
	EventTypeMetrics        map[string]*EventTypeMetrics `json:"event_type_metrics"`
	HandlerMetrics          map[string]*HandlerMetrics   `json:"handler_metrics"`
	RecentBuckets           []*MetricsBucket             `json:"recent_buckets"`
}

// NewEventMetricsCollector creates a new event metrics instance
func NewEventMetricsCollector(metricsWindow time.Duration, bucketCount int) *EventMetricsCollector {
	buckets := make([]*MetricsBucket, bucketCount)
	for i := range buckets {
		buckets[i] = &MetricsBucket{
			Timestamp: time.Now().Add(-time.Duration(bucketCount-i) * metricsWindow / time.Duration(bucketCount)),
		}
	}

	return &EventMetricsCollector{
		eventTypeMetrics: make(map[string]*EventTypeMetrics),
		handlerMetrics:   make(map[string]*HandlerMetrics),
		metricsWindow:    metricsWindow,
		buckets:          buckets,
		logger:           logger.GetGlobalLogger(),
	}
}

// RecordEventPublished records that an event was published
func (em *EventMetricsCollector) RecordEventPublished(eventType string) {
	atomic.AddInt64(&em.totalEventsPublished, 1)

	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Update event type metrics
	if _, exists := em.eventTypeMetrics[eventType]; !exists {
		em.eventTypeMetrics[eventType] = &EventTypeMetrics{
			EventType: eventType,
		}
	}
	em.eventTypeMetrics[eventType].PublishedCount++

	// Update current bucket
	em.getCurrentBucket().EventsPublished++
}

// RecordEventProcessed records that an event was successfully processed
func (em *EventMetricsCollector) RecordEventProcessed(eventType, handlerName string, processingTime time.Duration) {
	atomic.AddInt64(&em.totalEventsProcessed, 1)

	// Update processing time statistics
	processingNanos := processingTime.Nanoseconds()
	atomic.AddInt64(&em.totalProcessingTime, processingNanos)

	// Update min processing time
	for {
		current := atomic.LoadInt64(&em.minProcessingTime)
		if current == 0 || processingNanos < current {
			if atomic.CompareAndSwapInt64(&em.minProcessingTime, current, processingNanos) {
				break
			}
		} else {
			break
		}
	}

	// Update max processing time
	for {
		current := atomic.LoadInt64(&em.maxProcessingTime)
		if processingNanos > current {
			if atomic.CompareAndSwapInt64(&em.maxProcessingTime, current, processingNanos) {
				break
			}
		} else {
			break
		}
	}

	em.mutex.Lock()
	defer em.mutex.Unlock()

	now := time.Now()

	// Update event type metrics
	if _, exists := em.eventTypeMetrics[eventType]; !exists {
		em.eventTypeMetrics[eventType] = &EventTypeMetrics{
			EventType: eventType,
		}
	}
	typeMetrics := em.eventTypeMetrics[eventType]
	typeMetrics.ProcessedCount++
	typeMetrics.LastProcessedAt = now
	// Update average processing time
	if typeMetrics.ProcessedCount == 1 {
		typeMetrics.AverageProcessingTime = processingTime
	} else {
		typeMetrics.AverageProcessingTime = time.Duration(
			(int64(typeMetrics.AverageProcessingTime) + processingNanos) / 2,
		)
	}

	// Update handler metrics
	if _, exists := em.handlerMetrics[handlerName]; !exists {
		em.handlerMetrics[handlerName] = &HandlerMetrics{
			HandlerName: handlerName,
		}
	}
	handlerMetrics := em.handlerMetrics[handlerName]
	handlerMetrics.ProcessedCount++
	handlerMetrics.LastProcessedAt = now
	// Update average processing time
	if handlerMetrics.ProcessedCount == 1 {
		handlerMetrics.AverageProcessingTime = processingTime
	} else {
		handlerMetrics.AverageProcessingTime = time.Duration(
			(int64(handlerMetrics.AverageProcessingTime) + processingNanos) / 2,
		)
	}

	// Update current bucket
	bucket := em.getCurrentBucket()
	bucket.EventsProcessed++
	if bucket.EventsProcessed == 1 {
		bucket.AverageLatency = processingTime
	} else {
		bucket.AverageLatency = time.Duration(
			(int64(bucket.AverageLatency) + processingNanos) / 2,
		)
	}
}

// RecordEventFailed records that an event processing failed
func (em *EventMetricsCollector) RecordEventFailed(eventType, handlerName string, err error) {
	atomic.AddInt64(&em.totalEventsFailed, 1)

	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Update event type metrics
	if _, exists := em.eventTypeMetrics[eventType]; !exists {
		em.eventTypeMetrics[eventType] = &EventTypeMetrics{
			EventType: eventType,
		}
	}
	em.eventTypeMetrics[eventType].FailedCount++

	// Update handler metrics
	if _, exists := em.handlerMetrics[handlerName]; !exists {
		em.handlerMetrics[handlerName] = &HandlerMetrics{
			HandlerName: handlerName,
		}
	}
	em.handlerMetrics[handlerName].FailedCount++

	// Update current bucket
	em.getCurrentBucket().EventsFailed++
}

// RecordEventMovedToDeadLetter records that an event was moved to dead letter queue
func (em *EventMetricsCollector) RecordEventMovedToDeadLetter(eventType string) {
	atomic.AddInt64(&em.totalEventsInDeadLetter, 1)

	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Update event type metrics
	if _, exists := em.eventTypeMetrics[eventType]; !exists {
		em.eventTypeMetrics[eventType] = &EventTypeMetrics{
			EventType: eventType,
		}
	}
	em.eventTypeMetrics[eventType].DeadLetterCount++
}

// RecordCircuitBreakerTrip records that a circuit breaker was tripped
func (em *EventMetricsCollector) RecordCircuitBreakerTrip(handlerName string) {
	atomic.AddInt64(&em.totalCircuitBreakerTrips, 1)

	em.mutex.Lock()
	defer em.mutex.Unlock()

	// Update handler metrics
	if _, exists := em.handlerMetrics[handlerName]; !exists {
		em.handlerMetrics[handlerName] = &HandlerMetrics{
			HandlerName: handlerName,
		}
	}
	em.handlerMetrics[handlerName].CircuitBreakerTrips++
}

// GetSnapshot returns a snapshot of current metrics
func (em *EventMetricsCollector) GetSnapshot() *EventMetricsSnapshot {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	totalPublished := atomic.LoadInt64(&em.totalEventsPublished)
	totalProcessed := atomic.LoadInt64(&em.totalEventsProcessed)
	totalFailed := atomic.LoadInt64(&em.totalEventsFailed)
	totalDeadLetter := atomic.LoadInt64(&em.totalEventsInDeadLetter)
	circuitBreakerTrips := atomic.LoadInt64(&em.totalCircuitBreakerTrips)
	totalProcessingTime := atomic.LoadInt64(&em.totalProcessingTime)
	minProcessingTime := atomic.LoadInt64(&em.minProcessingTime)
	maxProcessingTime := atomic.LoadInt64(&em.maxProcessingTime)

	// Calculate success rate
	var successRate float64
	if totalProcessed+totalFailed > 0 {
		successRate = float64(totalProcessed) / float64(totalProcessed+totalFailed) * 100
	}

	// Calculate average processing time
	var avgProcessingTime time.Duration
	if totalProcessed > 0 {
		avgProcessingTime = time.Duration(totalProcessingTime / totalProcessed)
	}

	// Copy event type metrics
	eventTypeMetricsCopy := make(map[string]*EventTypeMetrics)
	for k, v := range em.eventTypeMetrics {
		eventTypeMetricsCopy[k] = &EventTypeMetrics{
			EventType:             v.EventType,
			PublishedCount:        v.PublishedCount,
			ProcessedCount:        v.ProcessedCount,
			FailedCount:           v.FailedCount,
			DeadLetterCount:       v.DeadLetterCount,
			AverageProcessingTime: v.AverageProcessingTime,
			LastProcessedAt:       v.LastProcessedAt,
		}
	}

	// Copy handler metrics
	handlerMetricsCopy := make(map[string]*HandlerMetrics)
	for k, v := range em.handlerMetrics {
		handlerMetricsCopy[k] = &HandlerMetrics{
			HandlerName:           v.HandlerName,
			ProcessedCount:        v.ProcessedCount,
			FailedCount:           v.FailedCount,
			AverageProcessingTime: v.AverageProcessingTime,
			CircuitBreakerTrips:   v.CircuitBreakerTrips,
			LastProcessedAt:       v.LastProcessedAt,
		}
	}

	// Copy recent buckets
	recentBuckets := make([]*MetricsBucket, len(em.buckets))
	for i, bucket := range em.buckets {
		recentBuckets[i] = &MetricsBucket{
			Timestamp:       bucket.Timestamp,
			EventsPublished: bucket.EventsPublished,
			EventsProcessed: bucket.EventsProcessed,
			EventsFailed:    bucket.EventsFailed,
			AverageLatency:  bucket.AverageLatency,
		}
	}

	return &EventMetricsSnapshot{
		Timestamp:               time.Now(),
		TotalEventsPublished:    totalPublished,
		TotalEventsProcessed:    totalProcessed,
		TotalEventsFailed:       totalFailed,
		TotalEventsInDeadLetter: totalDeadLetter,
		CircuitBreakerTrips:     circuitBreakerTrips,
		AverageProcessingTime:   avgProcessingTime,
		MinProcessingTime:       time.Duration(minProcessingTime),
		MaxProcessingTime:       time.Duration(maxProcessingTime),
		SuccessRate:             successRate,
		EventTypeMetrics:        eventTypeMetricsCopy,
		HandlerMetrics:          handlerMetricsCopy,
		RecentBuckets:           recentBuckets,
	}
}

// getCurrentBucket returns the current metrics bucket, rotating if necessary
func (em *EventMetricsCollector) getCurrentBucket() *MetricsBucket {
	now := time.Now()
	currentBucket := em.buckets[em.currentBucketIdx]

	// Check if we need to rotate to a new bucket
	bucketDuration := em.metricsWindow / time.Duration(len(em.buckets))
	if now.Sub(currentBucket.Timestamp) > bucketDuration {
		// Rotate to next bucket
		em.currentBucketIdx = (em.currentBucketIdx + 1) % len(em.buckets)
		em.buckets[em.currentBucketIdx] = &MetricsBucket{
			Timestamp: now,
		}
		currentBucket = em.buckets[em.currentBucketIdx]
	}

	return currentBucket
}

// Reset resets all metrics
func (em *EventMetricsCollector) Reset() {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	atomic.StoreInt64(&em.totalEventsPublished, 0)
	atomic.StoreInt64(&em.totalEventsProcessed, 0)
	atomic.StoreInt64(&em.totalEventsFailed, 0)
	atomic.StoreInt64(&em.totalEventsInDeadLetter, 0)
	atomic.StoreInt64(&em.totalCircuitBreakerTrips, 0)
	atomic.StoreInt64(&em.totalProcessingTime, 0)
	atomic.StoreInt64(&em.minProcessingTime, 0)
	atomic.StoreInt64(&em.maxProcessingTime, 0)

	em.eventTypeMetrics = make(map[string]*EventTypeMetrics)
	em.handlerMetrics = make(map[string]*HandlerMetrics)

	// Reset buckets
	for i := range em.buckets {
		em.buckets[i] = &MetricsBucket{
			Timestamp: time.Now().Add(-time.Duration(len(em.buckets)-i) * em.metricsWindow / time.Duration(len(em.buckets))),
		}
	}
	em.currentBucketIdx = 0

	em.logger.Info("Event metrics reset")
}

// MetricsEventHandler wraps an event handler to collect metrics
type MetricsEventHandler struct {
	handler     EventHandler
	handlerName string
	metrics     *EventMetricsCollector
	logger      logger.Logger
	id          string
}

// NewMetricsEventHandler creates a new metrics-collecting event handler
func NewMetricsEventHandler(handlerName string, handler EventHandler, metrics *EventMetricsCollector) *MetricsEventHandler {
	return &MetricsEventHandler{
		handler:     handler,
		handlerName: handlerName,
		metrics:     metrics,
		logger:      logger.GetGlobalLogger(),
		id:          generateEventID(),
	}
}

// Handle implements the EventHandler interface with metrics collection
func (meh *MetricsEventHandler) Handle(ctx context.Context, event Event) error {
	startTime := time.Now()

	err := meh.handler.Handle(ctx, event)

	processingTime := time.Since(startTime)

	if err != nil {
		meh.metrics.RecordEventFailed(event.EventType(), meh.handlerName, err)
	} else {
		meh.metrics.RecordEventProcessed(event.EventType(), meh.handlerName, processingTime)
	}

	return err
}

// EventTypes returns the event types this handler processes
func (meh *MetricsEventHandler) EventTypes() []string {
	return meh.handler.EventTypes()
}

// ID returns the unique identifier for this handler
func (meh *MetricsEventHandler) ID() string {
	return meh.id
}

// GetMetrics returns the metrics instance
func (meh *MetricsEventHandler) GetMetrics() *EventMetricsCollector {
	return meh.metrics
}

// MetricsEventBus wraps an event bus to collect metrics
type MetricsEventBus struct {
	EventBus
	metrics *EventMetricsCollector
	logger  logger.Logger
}

// NewMetricsEventBus creates a new metrics-collecting event bus
func NewMetricsEventBus(eventBus EventBus, metrics *EventMetricsCollector) *MetricsEventBus {
	return &MetricsEventBus{
		EventBus: eventBus,
		metrics:  metrics,
		logger:   logger.GetGlobalLogger(),
	}
}

// Publish publishes an event and records metrics
func (meb *MetricsEventBus) Publish(ctx context.Context, event Event) error {
	// Record that the event was published
	meb.metrics.RecordEventPublished(event.EventType())

	return meb.EventBus.Publish(ctx, event)
}

// PublishAsync publishes an event asynchronously and records metrics
func (meb *MetricsEventBus) PublishAsync(ctx context.Context, event Event) error {
	// Record that the event was published
	meb.metrics.RecordEventPublished(event.EventType())

	return meb.EventBus.PublishAsync(ctx, event)
}

// HealthMetrics provides health check information for the event system
type HealthMetrics struct {
	IsHealthy             bool          `json:"is_healthy"`
	LastHealthCheck       time.Time     `json:"last_health_check"`
	EventBusHealth        bool          `json:"event_bus_health"`
	DeadLetterQueueSize   int64         `json:"dead_letter_queue_size"`
	CircuitBreakersOpen   int           `json:"circuit_breakers_open"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	ErrorRate             float64       `json:"error_rate"`
	Issues                []string      `json:"issues,omitempty"`
}

// EventSystemHealthChecker performs health checks on the event system
type EventSystemHealthChecker struct {
	metrics               *EventMetricsCollector
	deadLetterQueue       DeadLetterQueue
	circuitBreakerManager *CircuitBreakerManager
	logger                logger.Logger
}

// NewEventSystemHealthChecker creates a new health checker
func NewEventSystemHealthChecker(
	metrics *EventMetricsCollector,
	deadLetterQueue DeadLetterQueue,
	circuitBreakerManager *CircuitBreakerManager,
) *EventSystemHealthChecker {
	return &EventSystemHealthChecker{
		metrics:               metrics,
		deadLetterQueue:       deadLetterQueue,
		circuitBreakerManager: circuitBreakerManager,
		logger:                logger.GetGlobalLogger(),
	}
}

// CheckHealth performs a comprehensive health check
func (hc *EventSystemHealthChecker) CheckHealth(ctx context.Context) *HealthMetrics {
	health := &HealthMetrics{
		LastHealthCheck: time.Now(),
		IsHealthy:       true,
		Issues:          []string{},
	}

	// Check metrics
	if hc.metrics != nil {
		snapshot := hc.metrics.GetSnapshot()
		health.AverageProcessingTime = snapshot.AverageProcessingTime
		health.ErrorRate = 100 - snapshot.SuccessRate

		// Check for high error rate
		if snapshot.SuccessRate < 95 && snapshot.TotalEventsProcessed+snapshot.TotalEventsFailed > 10 {
			health.IsHealthy = false
			health.Issues = append(health.Issues, fmt.Sprintf("High error rate: %.2f%%", 100-snapshot.SuccessRate))
		}

		// Check for slow processing
		if health.AverageProcessingTime > time.Second*5 {
			health.IsHealthy = false
			health.Issues = append(health.Issues, fmt.Sprintf("Slow processing: %v", health.AverageProcessingTime))
		}
	}

	// Check dead letter queue
	if hc.deadLetterQueue != nil {
		stats, err := hc.deadLetterQueue.GetStats(ctx)
		if err != nil {
			health.IsHealthy = false
			health.Issues = append(health.Issues, fmt.Sprintf("Dead letter queue error: %v", err))
		} else {
			health.DeadLetterQueueSize = stats.TotalEvents

			// Check for too many dead letter events
			if stats.TotalEvents > 100 {
				health.IsHealthy = false
				health.Issues = append(health.Issues, fmt.Sprintf("Too many dead letter events: %d", stats.TotalEvents))
			}
		}
	}

	// Check circuit breakers
	if hc.circuitBreakerManager != nil {
		cbHealth := hc.circuitBreakerManager.HealthCheck()
		openCount := 0
		for _, isHealthy := range cbHealth {
			if !isHealthy {
				openCount++
			}
		}

		health.CircuitBreakersOpen = openCount
		if openCount > 0 {
			health.IsHealthy = false
			health.Issues = append(health.Issues, fmt.Sprintf("%d circuit breakers are open", openCount))
		}
	}

	// Event bus health (simple check)
	health.EventBusHealth = true // Assume healthy if no issues detected

	return health
}

// Global metrics instance
var globalEventMetrics *EventMetricsCollector

// InitEventMetrics initializes the global event metrics
func InitEventMetrics(metricsWindow time.Duration, bucketCount int) {
	globalEventMetrics = NewEventMetricsCollector(metricsWindow, bucketCount)
}

// GetEventMetrics returns the global event metrics
func GetEventMetrics() *EventMetricsCollector {
	if globalEventMetrics == nil {
		globalEventMetrics = NewEventMetricsCollector(time.Hour, 12) // Default: 1 hour window with 12 buckets (5-minute buckets)
	}
	return globalEventMetrics
}

// WrapHandlerWithMetrics wraps an event handler to collect metrics
func WrapHandlerWithMetrics(handlerName string, handler EventHandler, metrics ...*EventMetricsCollector) EventHandler {
	var metricsInstance *EventMetricsCollector
	if len(metrics) > 0 {
		metricsInstance = metrics[0]
	} else {
		metricsInstance = GetEventMetrics()
	}

	return NewMetricsEventHandler(handlerName, handler, metricsInstance)
}

// WrapEventBusWithMetrics wraps an event bus to collect metrics
func WrapEventBusWithMetrics(eventBus EventBus, metrics ...*EventMetricsCollector) EventBus {
	var metricsInstance *EventMetricsCollector
	if len(metrics) > 0 {
		metricsInstance = metrics[0]
	} else {
		metricsInstance = GetEventMetrics()
	}

	return NewMetricsEventBus(eventBus, metricsInstance)
}
