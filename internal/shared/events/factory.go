package events

import (
	"fmt"
	"time"

	"linke/internal/shared/logger"
	"linke/internal/shared/queue"
	"linke/internal/shared/versioning"

	"gorm.io/gorm"
)

// EventSystemConfig contains configuration for the entire event system
type EventSystemConfig struct {
	// Event Bus Configuration
	EventBusType          string `json:"event_bus_type"` // "in_memory", "redis", "enhanced"
	EnableMetrics         bool   `json:"enable_metrics"`
	EnableDeduplication   bool   `json:"enable_deduplication"`
	EnableCircuitBreaker  bool   `json:"enable_circuit_breaker"`
	EnableAsyncProcessing bool   `json:"enable_async_processing"`

	// Metrics Configuration
	MetricsWindow      time.Duration `json:"metrics_window"`
	MetricsBucketCount int           `json:"metrics_bucket_count"`

	// Deduplication Configuration
	DeduplicationConfig DeduplicationConfig `json:"deduplication_config"`

	// Circuit Breaker Configuration
	CircuitBreakerConfig CircuitBreakerConfig `json:"circuit_breaker_config"`

	// Async Processing Configuration
	AsyncRetryConfig RetryConfig `json:"async_retry_config"`

	// Dead Letter Queue Configuration
	DeadLetterRetryPolicy DeadLetterRetryPolicy `json:"dead_letter_retry_policy"`

	// At-Least-Once Configuration
	AtLeastOnceRetryPolicy AtLeastOnceRetryPolicy `json:"at_least_once_retry_policy"`

	// Event Store Configuration
	EnableEventStore bool `json:"enable_event_store"`

	// Event Versioning Configuration
	EnableVersioning bool `json:"enable_versioning"`
}

// DefaultEventSystemConfig returns a sensible default configuration
func DefaultEventSystemConfig() EventSystemConfig {
	return EventSystemConfig{
		EventBusType:           "enhanced",
		EnableMetrics:          true,
		EnableDeduplication:    true,
		EnableCircuitBreaker:   true,
		EnableAsyncProcessing:  true,
		MetricsWindow:          time.Hour,
		MetricsBucketCount:     12,
		DeduplicationConfig:    DefaultDeduplicationConfig(),
		CircuitBreakerConfig:   DefaultCircuitBreakerConfig(),
		AsyncRetryConfig:       DefaultRetryConfig(),
		DeadLetterRetryPolicy:  DefaultDeadLetterRetryPolicy(),
		AtLeastOnceRetryPolicy: DefaultAtLeastOnceRetryPolicy(),
		EnableEventStore:       true,
		EnableVersioning:       true,
	}
}

// EventSystemComponents contains all the components of the event system
type EventSystemComponents struct {
	EventBus              EventBus
	EventStore            EventStore
	DeadLetterQueue       DeadLetterQueue
	DeadLetterHandler     *DeadLetterHandler
	MetricsCollector      *EventMetricsCollector
	CircuitBreakerManager *CircuitBreakerManager
	DeduplicationManager  *EventDeduplicationManager
	AsyncEventProcessor   *AsyncEventProcessor
	HealthChecker         *EventSystemHealthChecker
	CrossDomainHandlers   *CrossDomainEventHandlers
	NotificationHandler   *NotificationHandler
	EventReplayHandler    *EventReplayHandler
	VersionManager        *versioning.EventVersionManager
	SchemaRegistry        *versioning.EventSchemaRegistry
}

// EventSystemFactory creates and configures event system components
type EventSystemFactory struct {
	config EventSystemConfig
	logger logger.Logger
}

// NewEventSystemFactory creates a new event system factory
func NewEventSystemFactory(config EventSystemConfig) *EventSystemFactory {
	return &EventSystemFactory{
		config: config,
		logger: logger.GetGlobalLogger(),
	}
}

// CreateEventSystem creates a complete event system with all configured components
func (f *EventSystemFactory) CreateEventSystem(db *gorm.DB, taskQueue *queue.TaskQueue) (*EventSystemComponents, error) {
	components := &EventSystemComponents{}

	f.logger.Info("Creating event system",
		logger.Any("config", f.config),
	)

	// Create base event bus
	baseEventBus, err := f.createBaseEventBus()
	if err != nil {
		return nil, fmt.Errorf("failed to create base event bus: %w", err)
	}

	// Create event store if enabled
	if f.config.EnableEventStore && db != nil {
		components.EventStore = NewDatabaseEventStore(db)
		f.logger.Info("Event store enabled")
	}

	// Create metrics collector if enabled
	if f.config.EnableMetrics {
		components.MetricsCollector = NewEventMetricsCollector(
			f.config.MetricsWindow,
			f.config.MetricsBucketCount,
		)
		f.logger.Info("Event metrics enabled")
	}

	// Create circuit breaker manager if enabled
	if f.config.EnableCircuitBreaker {
		components.CircuitBreakerManager = NewCircuitBreakerManager()
		f.logger.Info("Circuit breaker enabled")
	}

	// Create deduplication manager if enabled
	if f.config.EnableDeduplication {
		components.DeduplicationManager = NewEventDeduplicationManager(f.config.DeduplicationConfig)
		f.logger.Info("Event deduplication enabled")
	}

	// Create dead letter queue and handler
	components.DeadLetterQueue = NewInMemoryDeadLetterQueue()
	if taskQueue != nil {
		components.DeadLetterHandler = NewDeadLetterHandler(
			components.DeadLetterQueue,
			baseEventBus,
			taskQueue,
			f.config.DeadLetterRetryPolicy,
		)
	}

	// Create async event processor if enabled
	if f.config.EnableAsyncProcessing && taskQueue != nil {
		components.AsyncEventProcessor = NewAsyncEventProcessor(
			taskQueue,
			components.EventStore,
			baseEventBus,
			f.config.AsyncRetryConfig,
		)
	}

	// Create versioning components if enabled
	if f.config.EnableVersioning {
		components.SchemaRegistry = versioning.NewEventSchemaRegistry()
		if err := components.SchemaRegistry.RegisterDefaultSchemas(); err != nil {
			f.logger.Warn("Failed to register default schemas", logger.ErrorField(err))
		}
		components.VersionManager = components.SchemaRegistry.GetVersionManager()
		f.logger.Info("Event versioning enabled")
	}

	// Wrap event bus with desired capabilities
	eventBus, err := f.wrapEventBus(baseEventBus, components)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap event bus: %w", err)
	}
	components.EventBus = eventBus

	// Create event replay handler if store is available
	if components.EventStore != nil && components.AsyncEventProcessor != nil {
		components.EventReplayHandler = NewEventReplayHandler(
			components.EventStore,
			components.AsyncEventProcessor,
		)
	}

	// Create health checker
	components.HealthChecker = NewEventSystemHealthChecker(
		components.MetricsCollector,
		components.DeadLetterQueue,
		components.CircuitBreakerManager,
	)

	// Create cross-domain handlers - skip for now, will be created via dependency injection
	components.CrossDomainHandlers = nil // NewCrossDomainEventHandlers requires service dependencies

	// Create notification handler
	components.NotificationHandler = NewNotificationHandler()

	f.logger.Info("Event system created successfully",
		logger.Any("components", map[string]bool{
			"event_bus":               components.EventBus != nil,
			"event_store":             components.EventStore != nil,
			"dead_letter_queue":       components.DeadLetterQueue != nil,
			"dead_letter_handler":     components.DeadLetterHandler != nil,
			"metrics_collector":       components.MetricsCollector != nil,
			"circuit_breaker_manager": components.CircuitBreakerManager != nil,
			"deduplication_manager":   components.DeduplicationManager != nil,
			"async_event_processor":   components.AsyncEventProcessor != nil,
			"health_checker":          components.HealthChecker != nil,
			"version_manager":         components.VersionManager != nil,
		}),
	)

	return components, nil
}

// createBaseEventBus creates the base event bus based on configuration
func (f *EventSystemFactory) createBaseEventBus() (EventBus, error) {
	switch f.config.EventBusType {
	case "in_memory":
		return NewInMemoryEventBus(), nil
	case "redis":
		return NewRedisEventBus(), nil
	case "enhanced":
		return NewEnhancedEventBus(), nil
	default:
		return nil, fmt.Errorf("unknown event bus type: %s", f.config.EventBusType)
	}
}

// wrapEventBus wraps the base event bus with additional capabilities
func (f *EventSystemFactory) wrapEventBus(baseEventBus EventBus, components *EventSystemComponents) (EventBus, error) {
	var eventBus EventBus = baseEventBus

	// Wrap with metrics if enabled
	if f.config.EnableMetrics && components.MetricsCollector != nil {
		eventBus = NewMetricsEventBus(eventBus, components.MetricsCollector)
		f.logger.Info("Event bus wrapped with metrics")
	}

	// Wrap with at-least-once delivery if deduplication is enabled
	if f.config.EnableDeduplication && components.DeduplicationManager != nil {
		globalDeduplicator := components.DeduplicationManager.GetOrCreateDeduplicator("global")
		eventBus = NewAtLeastOnceEventBus(eventBus, globalDeduplicator, f.config.AtLeastOnceRetryPolicy)
		f.logger.Info("Event bus wrapped with at-least-once delivery")
	}

	// Wrap with async processing if enabled
	if f.config.EnableAsyncProcessing && components.AsyncEventProcessor != nil {
		eventBus = NewAsyncEventBus(eventBus, components.AsyncEventProcessor)
		f.logger.Info("Event bus wrapped with async processing")
	}

	return eventBus, nil
}

// RegisterHandlers registers all cross-domain event handlers
func (f *EventSystemFactory) RegisterHandlers(components *EventSystemComponents) error {
	if components.EventBus == nil {
		return fmt.Errorf("event bus is required to register handlers")
	}

	var handlers []EventHandler

	// Register cross-domain handlers
	if components.CrossDomainHandlers != nil {
		if err := components.CrossDomainHandlers.RegisterCrossDomainHandlers(components.EventBus); err != nil {
			return fmt.Errorf("failed to register cross-domain handlers: %w", err)
		}
		f.logger.Info("Cross-domain handlers registered")
	}

	// Register notification handler
	if components.NotificationHandler != nil {
		handlers = append(handlers, components.NotificationHandler)
	}

	// Wrap handlers with additional capabilities if enabled
	for i, handler := range handlers {
		wrappedHandler := f.wrapHandler(fmt.Sprintf("handler-%d", i), handler, components)
		handlers[i] = wrappedHandler

		// Subscribe wrapped handler
		if err := components.EventBus.Subscribe(handler.EventTypes(), wrappedHandler); err != nil {
			return fmt.Errorf("failed to subscribe handler: %w", err)
		}
	}

	f.logger.Debug("All event handlers registered successfully",
		logger.Int("handler_count", len(handlers)),
	)

	return nil
}

// wrapHandler wraps an event handler with additional capabilities
func (f *EventSystemFactory) wrapHandler(handlerName string, handler EventHandler, components *EventSystemComponents) EventHandler {
	var wrappedHandler EventHandler = handler

	// Wrap with metrics if enabled
	if f.config.EnableMetrics && components.MetricsCollector != nil {
		wrappedHandler = NewMetricsEventHandler(handlerName, wrappedHandler, components.MetricsCollector)
		f.logger.Debug("Handler wrapped with metrics", logger.String("handler", handlerName))
	}

	// Wrap with deduplication if enabled
	if f.config.EnableDeduplication && components.DeduplicationManager != nil {
		deduplicator := components.DeduplicationManager.GetOrCreateDeduplicator(handlerName)
		wrappedHandler = NewDeduplicatingEventHandler(handlerName, wrappedHandler, deduplicator)
		f.logger.Debug("Handler wrapped with deduplication", logger.String("handler", handlerName))
	}

	// Wrap with circuit breaker if enabled
	if f.config.EnableCircuitBreaker {
		wrappedHandler = NewCircuitBreakerEventHandler(handlerName, wrappedHandler, f.config.CircuitBreakerConfig)
		f.logger.Debug("Handler wrapped with circuit breaker", logger.String("handler", handlerName))
	}

	return wrappedHandler
}

// ConfigureTaskHandlers registers task handlers for async processing
func (f *EventSystemFactory) ConfigureTaskHandlers(processor *queue.TaskProcessor, components *EventSystemComponents) {
	if processor == nil {
		f.logger.Warn("Task processor is nil, skipping task handler registration")
		return
	}

	// Register async event processing handlers
	if components.AsyncEventProcessor != nil {
		RegisterEventHandlers(processor, components.AsyncEventProcessor)
		f.logger.Debug("Async event processing handlers registered")
	}

	// Register dead letter handlers
	if components.DeadLetterHandler != nil {
		RegisterDeadLetterHandlers(processor, components.DeadLetterHandler)
		f.logger.Info("Dead letter handlers registered")
	}
}

// EventSystemBuilder provides a fluent interface for building event systems
type EventSystemBuilder struct {
	config        EventSystemConfig
	db            *gorm.DB
	taskQueue     *queue.TaskQueue
	taskProcessor *queue.TaskProcessor
}

// NewEventSystemBuilder creates a new event system builder
func NewEventSystemBuilder() *EventSystemBuilder {
	return &EventSystemBuilder{
		config: DefaultEventSystemConfig(),
	}
}

// WithConfig sets the configuration
func (b *EventSystemBuilder) WithConfig(config EventSystemConfig) *EventSystemBuilder {
	b.config = config
	return b
}

// WithDatabase sets the database connection
func (b *EventSystemBuilder) WithDatabase(db *gorm.DB) *EventSystemBuilder {
	b.db = db
	return b
}

// WithTaskQueue sets the task queue
func (b *EventSystemBuilder) WithTaskQueue(taskQueue *queue.TaskQueue) *EventSystemBuilder {
	b.taskQueue = taskQueue
	return b
}

// WithTaskProcessor sets the task processor
func (b *EventSystemBuilder) WithTaskProcessor(taskProcessor *queue.TaskProcessor) *EventSystemBuilder {
	b.taskProcessor = taskProcessor
	return b
}

// EnableMetrics enables metrics collection
func (b *EventSystemBuilder) EnableMetrics(window time.Duration, bucketCount int) *EventSystemBuilder {
	b.config.EnableMetrics = true
	b.config.MetricsWindow = window
	b.config.MetricsBucketCount = bucketCount
	return b
}

// EnableDeduplication enables event deduplication
func (b *EventSystemBuilder) EnableDeduplication(config DeduplicationConfig) *EventSystemBuilder {
	b.config.EnableDeduplication = true
	b.config.DeduplicationConfig = config
	return b
}

// EnableCircuitBreaker enables circuit breaker pattern
func (b *EventSystemBuilder) EnableCircuitBreaker(config CircuitBreakerConfig) *EventSystemBuilder {
	b.config.EnableCircuitBreaker = true
	b.config.CircuitBreakerConfig = config
	return b
}

// EnableAsyncProcessing enables asynchronous event processing
func (b *EventSystemBuilder) EnableAsyncProcessing(retryConfig RetryConfig) *EventSystemBuilder {
	b.config.EnableAsyncProcessing = true
	b.config.AsyncRetryConfig = retryConfig
	return b
}

// DisableEventStore disables event store
func (b *EventSystemBuilder) DisableEventStore() *EventSystemBuilder {
	b.config.EnableEventStore = false
	return b
}

// DisableVersioning disables event versioning
func (b *EventSystemBuilder) DisableVersioning() *EventSystemBuilder {
	b.config.EnableVersioning = false
	return b
}

// Build creates the event system
func (b *EventSystemBuilder) Build() (*EventSystemComponents, error) {
	factory := NewEventSystemFactory(b.config)

	components, err := factory.CreateEventSystem(b.db, b.taskQueue)
	if err != nil {
		return nil, fmt.Errorf("failed to create event system: %w", err)
	}

	// Register handlers
	if err := factory.RegisterHandlers(components); err != nil {
		return nil, fmt.Errorf("failed to register handlers: %w", err)
	}

	// Configure task handlers if task processor is available
	if b.taskProcessor != nil {
		factory.ConfigureTaskHandlers(b.taskProcessor, components)
	}

	return components, nil
}

// EventSystemModule provides dependency injection integration
type EventSystemModule struct {
	components *EventSystemComponents
	config     EventSystemConfig
	logger     logger.Logger
}

// NewEventSystemModule creates a new event system module for dependency injection
func NewEventSystemModule(config EventSystemConfig) *EventSystemModule {
	return &EventSystemModule{
		config: config,
		logger: logger.GetGlobalLogger(),
	}
}

// Initialize initializes the event system module
func (m *EventSystemModule) Initialize(db *gorm.DB, taskQueue *queue.TaskQueue, taskProcessor *queue.TaskProcessor) error {
	builder := NewEventSystemBuilder().
		WithConfig(m.config).
		WithDatabase(db).
		WithTaskQueue(taskQueue).
		WithTaskProcessor(taskProcessor)

	components, err := builder.Build()
	if err != nil {
		return fmt.Errorf("failed to build event system: %w", err)
	}

	m.components = components

	// Initialize global instances
	if components.EventBus != nil {
		InitEventBus(components.EventBus)
	}
	if components.MetricsCollector != nil {
		InitEventMetrics(m.config.MetricsWindow, m.config.MetricsBucketCount)
	}
	if components.DeadLetterQueue != nil {
		InitDeadLetterQueue(components.DeadLetterQueue)
	}
	if components.DeduplicationManager != nil {
		InitDeduplicationManager(m.config.DeduplicationConfig)
	}
	if components.CircuitBreakerManager != nil {
		InitCircuitBreakerManager()
	}

	m.logger.Info("Event system module initialized successfully")
	return nil
}

// GetComponents returns the event system components
func (m *EventSystemModule) GetComponents() *EventSystemComponents {
	return m.components
}

// GetEventBus returns the event bus
func (m *EventSystemModule) GetEventBus() EventBus {
	if m.components != nil {
		return m.components.EventBus
	}
	return nil
}

// GetMetricsCollector returns the metrics collector
func (m *EventSystemModule) GetMetricsCollector() *EventMetricsCollector {
	if m.components != nil {
		return m.components.MetricsCollector
	}
	return nil
}

// GetHealthChecker returns the health checker
func (m *EventSystemModule) GetHealthChecker() *EventSystemHealthChecker {
	if m.components != nil {
		return m.components.HealthChecker
	}
	return nil
}

// Shutdown gracefully shuts down the event system
func (m *EventSystemModule) Shutdown() error {
	if m.components == nil {
		return nil
	}

	var errors []error

	// Close event bus
	if m.components.EventBus != nil {
		if err := m.components.EventBus.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close event bus: %w", err))
		}
	}

	// Close deduplication manager
	if m.components.DeduplicationManager != nil {
		if err := m.components.DeduplicationManager.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close deduplication manager: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errors)
	}

	m.logger.Info("Event system module shut down successfully")
	return nil
}

// Global event system module instance
var globalEventSystemModule *EventSystemModule

// InitEventSystemModule initializes the global event system module
func InitEventSystemModule(config EventSystemConfig, db *gorm.DB, taskQueue *queue.TaskQueue, taskProcessor *queue.TaskProcessor) error {
	globalEventSystemModule = NewEventSystemModule(config)
	return globalEventSystemModule.Initialize(db, taskQueue, taskProcessor)
}

// GetEventSystemModule returns the global event system module
func GetEventSystemModule() *EventSystemModule {
	return globalEventSystemModule
}
