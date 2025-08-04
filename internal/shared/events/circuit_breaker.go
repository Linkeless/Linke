package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/logger"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "CLOSED"
	case CircuitBreakerOpen:
		return "OPEN"
	case CircuitBreakerHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig contains configuration for a circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures      int           `json:"max_failures"`        // Number of failures before opening
	ResetTimeout     time.Duration `json:"reset_timeout"`       // Time to wait before trying half-open
	SuccessThreshold int           `json:"success_threshold"`   // Successes needed in half-open to close
	MonitoringWindow time.Duration `json:"monitoring_window"`   // Window for counting failures
	HalfOpenMaxCalls int           `json:"half_open_max_calls"` // Max calls allowed in half-open state
}

// DefaultCircuitBreakerConfig returns a sensible default configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:      5,
		ResetTimeout:     time.Minute * 1,
		SuccessThreshold: 3,
		MonitoringWindow: time.Minute * 5,
		HalfOpenMaxCalls: 3,
	}
}

// CircuitBreaker implements the circuit breaker pattern for event handlers
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitBreakerState
	failures        []time.Time
	successes       int
	lastFailureTime time.Time
	halfOpenCalls   int
	mutex           sync.RWMutex
	logger          logger.Logger
	name            string
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config:    config,
		state:     CircuitBreakerClosed,
		failures:  make([]time.Time, 0),
		successes: 0,
		logger:    logger.GetGlobalLogger(),
		name:      name,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.allowRequest() {
		return fmt.Errorf("circuit breaker %s is OPEN", cb.name)
	}

	err := fn()
	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// allowRequest determines if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// Check if reset timeout has elapsed
		if time.Since(cb.lastFailureTime) >= cb.config.ResetTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.halfOpenCalls = 0
			cb.successes = 0
			cb.logger.Info("Circuit breaker transitioning to HALF_OPEN",
				logger.String("name", cb.name),
			)
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		// Allow limited requests
		if cb.halfOpenCalls < cb.config.HalfOpenMaxCalls {
			cb.halfOpenCalls++
			return true
		}
		return false
	default:
		return false
	}
}

// recordFailure records a failure and updates circuit breaker state
func (cb *CircuitBreaker) recordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	cb.lastFailureTime = now
	cb.failures = append(cb.failures, now)

	// Clean old failures outside monitoring window
	cb.cleanOldFailures(now)

	switch cb.state {
	case CircuitBreakerClosed:
		if len(cb.failures) >= cb.config.MaxFailures {
			cb.state = CircuitBreakerOpen
			cb.logger.Warn("Circuit breaker opening due to failures",
				logger.String("name", cb.name),
				logger.Int("failure_count", len(cb.failures)),
				logger.Int("max_failures", cb.config.MaxFailures),
			)
		}
	case CircuitBreakerHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.state = CircuitBreakerOpen
		cb.halfOpenCalls = 0
		cb.successes = 0
		cb.logger.Warn("Circuit breaker reopening due to failure in HALF_OPEN state",
			logger.String("name", cb.name),
		)
	}
}

// recordSuccess records a success and updates circuit breaker state
func (cb *CircuitBreaker) recordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case CircuitBreakerHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			successCount := cb.successes // Store before reset for logging
			cb.state = CircuitBreakerClosed
			cb.successes = 0
			cb.halfOpenCalls = 0
			cb.failures = make([]time.Time, 0) // Reset failures
			cb.logger.Info("Circuit breaker closing due to successful requests",
				logger.String("name", cb.name),
				logger.Int("success_count", successCount),
			)
		}
	case CircuitBreakerClosed:
		// In closed state, clean old failures on success
		cb.cleanOldFailures(time.Now())
	}
}

// cleanOldFailures removes failures outside the monitoring window
func (cb *CircuitBreaker) cleanOldFailures(now time.Time) {
	cutoff := now.Add(-cb.config.MonitoringWindow)
	validFailures := make([]time.Time, 0, len(cb.failures))

	for _, failure := range cb.failures {
		if failure.After(cutoff) {
			validFailures = append(validFailures, failure)
		}
	}

	cb.failures = validFailures
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return CircuitBreakerStats{
		Name:          cb.name,
		State:         cb.state,
		FailureCount:  len(cb.failures),
		SuccessCount:  cb.successes,
		LastFailureAt: cb.lastFailureTime,
		HalfOpenCalls: cb.halfOpenCalls,
		Config:        cb.config,
	}
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = CircuitBreakerClosed
	cb.failures = make([]time.Time, 0)
	cb.successes = 0
	cb.halfOpenCalls = 0

	cb.logger.Info("Circuit breaker manually reset",
		logger.String("name", cb.name),
	)
}

// CircuitBreakerStats contains statistics about a circuit breaker
type CircuitBreakerStats struct {
	Name          string               `json:"name"`
	State         CircuitBreakerState  `json:"state"`
	FailureCount  int                  `json:"failure_count"`
	SuccessCount  int                  `json:"success_count"`
	LastFailureAt time.Time            `json:"last_failure_at"`
	HalfOpenCalls int                  `json:"half_open_calls"`
	Config        CircuitBreakerConfig `json:"config"`
}

// CircuitBreakerEventHandler wraps an event handler with circuit breaker protection
type CircuitBreakerEventHandler struct {
	handler        EventHandler
	circuitBreaker *CircuitBreaker
	logger         logger.Logger
	id             string
}

// NewCircuitBreakerEventHandler creates a new circuit breaker protected event handler
func NewCircuitBreakerEventHandler(name string, handler EventHandler, config CircuitBreakerConfig) *CircuitBreakerEventHandler {
	return &CircuitBreakerEventHandler{
		handler:        handler,
		circuitBreaker: NewCircuitBreaker(name, config),
		logger:         logger.GetGlobalLogger(),
		id:             generateEventID(),
	}
}

// Handle implements the EventHandler interface with circuit breaker protection
func (cbh *CircuitBreakerEventHandler) Handle(ctx context.Context, event Event) error {
	return cbh.circuitBreaker.Execute(ctx, func() error {
		return cbh.handler.Handle(ctx, event)
	})
}

// EventTypes returns the event types this handler processes
func (cbh *CircuitBreakerEventHandler) EventTypes() []string {
	return cbh.handler.EventTypes()
}

// ID returns the unique identifier for this handler
func (cbh *CircuitBreakerEventHandler) ID() string {
	return cbh.id
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (cbh *CircuitBreakerEventHandler) GetCircuitBreakerStats() CircuitBreakerStats {
	return cbh.circuitBreaker.GetStats()
}

// ResetCircuitBreaker manually resets the circuit breaker
func (cbh *CircuitBreakerEventHandler) ResetCircuitBreaker() {
	cbh.circuitBreaker.Reset()
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	circuitBreakers map[string]*CircuitBreaker
	mutex           sync.RWMutex
	logger          logger.Logger
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		circuitBreakers: make(map[string]*CircuitBreaker),
		logger:          logger.GetGlobalLogger(),
	}
}

// GetOrCreateCircuitBreaker gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreateCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	if cb, exists := cbm.circuitBreakers[name]; exists {
		return cb
	}

	cb := NewCircuitBreaker(name, config)
	cbm.circuitBreakers[name] = cb

	cbm.logger.Info("Circuit breaker created",
		logger.String("name", name),
		logger.Any("config", config),
	)

	return cb
}

// GetCircuitBreaker returns a circuit breaker by name
func (cbm *CircuitBreakerManager) GetCircuitBreaker(name string) (*CircuitBreaker, bool) {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	cb, exists := cbm.circuitBreakers[name]
	return cb, exists
}

// GetAllStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllStats() map[string]CircuitBreakerStats {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	stats := make(map[string]CircuitBreakerStats)
	for name, cb := range cbm.circuitBreakers {
		stats[name] = cb.GetStats()
	}

	return stats
}

// ResetAll resets all circuit breakers
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	for _, cb := range cbm.circuitBreakers {
		cb.Reset()
	}

	cbm.logger.Info("All circuit breakers reset")
}

// RemoveCircuitBreaker removes a circuit breaker
func (cbm *CircuitBreakerManager) RemoveCircuitBreaker(name string) {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	delete(cbm.circuitBreakers, name)

	cbm.logger.Info("Circuit breaker removed",
		logger.String("name", name),
	)
}

// HealthCheck performs a health check on all circuit breakers
func (cbm *CircuitBreakerManager) HealthCheck() map[string]bool {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	health := make(map[string]bool)
	for name, cb := range cbm.circuitBreakers {
		// Consider a circuit breaker healthy if it's not open
		health[name] = cb.GetState() != CircuitBreakerOpen
	}

	return health
}

// Global circuit breaker manager
var globalCircuitBreakerManager *CircuitBreakerManager

// InitCircuitBreakerManager initializes the global circuit breaker manager
func InitCircuitBreakerManager() {
	globalCircuitBreakerManager = NewCircuitBreakerManager()
}

// GetCircuitBreakerManager returns the global circuit breaker manager
func GetCircuitBreakerManager() *CircuitBreakerManager {
	if globalCircuitBreakerManager == nil {
		InitCircuitBreakerManager()
	}
	return globalCircuitBreakerManager
}

// WrapHandlerWithCircuitBreaker wraps an event handler with circuit breaker protection
func WrapHandlerWithCircuitBreaker(name string, handler EventHandler, config ...CircuitBreakerConfig) EventHandler {
	var cbConfig CircuitBreakerConfig
	if len(config) > 0 {
		cbConfig = config[0]
	} else {
		cbConfig = DefaultCircuitBreakerConfig()
	}

	return NewCircuitBreakerEventHandler(name, handler, cbConfig)
}
