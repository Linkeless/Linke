package resilience

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryableError defines an error that can be retried
type RetryableError interface {
	error
	IsRetryable() bool
}

// retryableError implements RetryableError
type retryableError struct {
	err        error
	retryable  bool
}

func (re *retryableError) Error() string {
	return re.err.Error()
}

func (re *retryableError) IsRetryable() bool {
	return re.retryable
}

func (re *retryableError) Unwrap() error {
	return re.err
}

// NewRetryableError creates a new retryable error
func NewRetryableError(err error) RetryableError {
	return &retryableError{err: err, retryable: true}
}

// NewNonRetryableError creates a new non-retryable error
func NewNonRetryableError(err error) RetryableError {
	return &retryableError{err: err, retryable: false}
}

// BackoffStrategy defines how delays between retries are calculated
type BackoffStrategy interface {
	NextDelay(attempt int, baseDelay time.Duration) time.Duration
	Reset()
}

// ExponentialBackoff implements exponential backoff with jitter
type ExponentialBackoff struct {
	multiplier    float64
	maxDelay      time.Duration
	jitter        bool
	jitterPercent float64
}

// NewExponentialBackoff creates a new exponential backoff strategy
func NewExponentialBackoff(multiplier float64, maxDelay time.Duration, jitter bool) *ExponentialBackoff {
	return &ExponentialBackoff{
		multiplier:    multiplier,
		maxDelay:      maxDelay,
		jitter:        jitter,
		jitterPercent: 0.1, // 10% jitter by default
	}
}

func (eb *ExponentialBackoff) NextDelay(attempt int, baseDelay time.Duration) time.Duration {
	if attempt <= 0 {
		return baseDelay
	}
	
	// Calculate exponential delay
	delay := time.Duration(float64(baseDelay) * math.Pow(eb.multiplier, float64(attempt-1)))
	
	// Apply maximum delay limit
	if delay > eb.maxDelay {
		delay = eb.maxDelay
	}
	
	// Add jitter if enabled
	if eb.jitter {
		jitterAmount := float64(delay) * eb.jitterPercent
		jitterOffset := (rand.Float64() - 0.5) * 2 * jitterAmount
		delay = time.Duration(float64(delay) + jitterOffset)
		
		// Ensure delay is not negative
		if delay < 0 {
			delay = baseDelay
		}
	}
	
	return delay
}

func (eb *ExponentialBackoff) Reset() {
	// No state to reset for exponential backoff
}

// LinearBackoff implements linear backoff
type LinearBackoff struct {
	increment time.Duration
	maxDelay  time.Duration
}

// NewLinearBackoff creates a new linear backoff strategy
func NewLinearBackoff(increment, maxDelay time.Duration) *LinearBackoff {
	return &LinearBackoff{
		increment: increment,
		maxDelay:  maxDelay,
	}
}

func (lb *LinearBackoff) NextDelay(attempt int, baseDelay time.Duration) time.Duration {
	delay := baseDelay + time.Duration(attempt) * lb.increment
	if delay > lb.maxDelay {
		delay = lb.maxDelay
	}
	return delay
}

func (lb *LinearBackoff) Reset() {
	// No state to reset for linear backoff
}

// FixedBackoff implements fixed delay backoff
type FixedBackoff struct {
	delay time.Duration
}

// NewFixedBackoff creates a new fixed backoff strategy
func NewFixedBackoff(delay time.Duration) *FixedBackoff {
	return &FixedBackoff{delay: delay}
}

func (fb *FixedBackoff) NextDelay(attempt int, baseDelay time.Duration) time.Duration {
	return fb.delay
}

func (fb *FixedBackoff) Reset() {
	// No state to reset for fixed backoff
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts     int               `json:"max_attempts"`
	BaseDelay       time.Duration     `json:"base_delay"`
	MaxDelay        time.Duration     `json:"max_delay"`
	BackoffStrategy BackoffStrategy   `json:"-"`
	RetryIf         func(error) bool  `json:"-"`
	OnRetry         func(int, error)  `json:"-"`
	Context         context.Context   `json:"-"`
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:     3,
		BaseDelay:       100 * time.Millisecond,
		MaxDelay:        5 * time.Second,
		BackoffStrategy: NewExponentialBackoff(2.0, 5*time.Second, true),
		RetryIf:         DefaultRetryIf,
		OnRetry:         nil,
		Context:         context.Background(),
	}
}

// DefaultRetryIf is the default retry condition function
func DefaultRetryIf(err error) bool {
	if err == nil {
		return false
	}
	
	// Check if error explicitly implements RetryableError
	if retryableErr, ok := err.(RetryableError); ok {
		return retryableErr.IsRetryable()
	}
	
	// Default heuristics for common error types
	if isTemporaryError(err) {
		return true
	}
	
	if isTimeoutError(err) {
		return true
	}
	
	if isConnectionError(err) {
		return true
	}
	
	return false
}

// isTemporaryError checks if error is temporary
func isTemporaryError(err error) bool {
	type temporary interface {
		Temporary() bool
	}
	if temp, ok := err.(temporary); ok {
		return temp.Temporary()
	}
	return false
}

// isTimeoutError checks if error is a timeout
func isTimeoutError(err error) bool {
	type timeout interface {
		Timeout() bool
	}
	if t, ok := err.(timeout); ok {
		return t.Timeout()
	}
	return false
}

// isConnectionError checks if error is a connection error
func isConnectionError(err error) bool {
	// Simple heuristic - check error message for connection-related keywords
	errMsg := err.Error()
	connectionKeywords := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"network unreachable",
		"host unreachable",
	}
	
	for _, keyword := range connectionKeywords {
		if contains(errMsg, keyword) {
			return true
		}
	}
	
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   s[len(s)-len(substr):] == substr || 
		   findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Retrier handles retry logic
type Retrier struct {
	config *RetryConfig
}

// NewRetrier creates a new retrier with the given configuration
func NewRetrier(config *RetryConfig) *Retrier {
	if config == nil {
		config = DefaultRetryConfig()
	}
	
	return &Retrier{config: config}
}

// Do executes the function with retry logic
func (r *Retrier) Do(fn func() error) error {
	return r.DoWithResult(func() (interface{}, error) {
		return nil, fn()
	})
}

// DoWithResult executes the function with retry logic and returns a result
func (r *Retrier) DoWithResult(fn func() (interface{}, error)) error {
	_, err := r.DoWithReturn(fn)
	return err
}

// DoWithReturn executes the function with retry logic and returns both result and error
func (r *Retrier) DoWithReturn(fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	var result interface{}
	
	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// Check context cancellation
		if r.config.Context != nil {
			select {
			case <-r.config.Context.Done():
				return nil, r.config.Context.Err()
			default:
			}
		}
		
		// Execute the function
		result, lastErr = fn()
		
		// If no error, return success
		if lastErr == nil {
			return result, nil
		}
		
		// Check if we should retry this error
		if r.config.RetryIf != nil && !r.config.RetryIf(lastErr) {
			return result, lastErr
		}
		
		// If this was the last attempt, don't wait
		if attempt == r.config.MaxAttempts {
			break
		}
		
		// Call retry callback if provided
		if r.config.OnRetry != nil {
			r.config.OnRetry(attempt, lastErr)
		}
		
		// Calculate delay for next attempt
		delay := r.config.BackoffStrategy.NextDelay(attempt, r.config.BaseDelay)
		
		// Wait before next attempt (with context cancellation support)
		if r.config.Context != nil {
			timer := time.NewTimer(delay)
			select {
			case <-r.config.Context.Done():
				timer.Stop()
				return result, r.config.Context.Err()
			case <-timer.C:
			}
		} else {
			time.Sleep(delay)
		}
	}
	
	return result, fmt.Errorf("retry failed after %d attempts, last error: %w", r.config.MaxAttempts, lastErr)
}

// DoWithContext executes the function with retry logic using the provided context
func (r *Retrier) DoWithContext(ctx context.Context, fn func() error) error {
	originalCtx := r.config.Context
	r.config.Context = ctx
	defer func() {
		r.config.Context = originalCtx
	}()
	
	return r.Do(fn)
}

// Reset resets the retrier state
func (r *Retrier) Reset() {
	if r.config.BackoffStrategy != nil {
		r.config.BackoffStrategy.Reset()
	}
}

// GetConfig returns a copy of the current configuration
func (r *Retrier) GetConfig() RetryConfig {
	return *r.config
}

// UpdateConfig updates the retrier configuration
func (r *Retrier) UpdateConfig(config *RetryConfig) {
	if config != nil {
		r.config = config
	}
}

// Retry is a convenience function for simple retry scenarios
func Retry(fn func() error, maxAttempts int, baseDelay time.Duration) error {
	config := &RetryConfig{
		MaxAttempts:     maxAttempts,
		BaseDelay:       baseDelay,
		MaxDelay:        baseDelay * 10,
		BackoffStrategy: NewExponentialBackoff(2.0, baseDelay*10, true),
		RetryIf:         DefaultRetryIf,
	}
	
	retrier := NewRetrier(config)
	return retrier.Do(fn)
}

// RetryWithContext is a convenience function for simple retry scenarios with context
func RetryWithContext(ctx context.Context, fn func() error, maxAttempts int, baseDelay time.Duration) error {
	config := &RetryConfig{
		MaxAttempts:     maxAttempts,
		BaseDelay:       baseDelay,
		MaxDelay:        baseDelay * 10,
		BackoffStrategy: NewExponentialBackoff(2.0, baseDelay*10, true),
		RetryIf:         DefaultRetryIf,
		Context:         ctx,
	}
	
	retrier := NewRetrier(config)
	return retrier.Do(fn)
}

// RetryWithResult is a convenience function for retry scenarios that return a result
func RetryWithResult[T any](fn func() (T, error), maxAttempts int, baseDelay time.Duration) (T, error) {
	var result T
	
	config := &RetryConfig{
		MaxAttempts:     maxAttempts,
		BaseDelay:       baseDelay,
		MaxDelay:        baseDelay * 10,
		BackoffStrategy: NewExponentialBackoff(2.0, baseDelay*10, true),
		RetryIf:         DefaultRetryIf,
	}
	
	retrier := NewRetrier(config)
	
	value, err := retrier.DoWithReturn(func() (interface{}, error) {
		return fn()
	})
	
	if err != nil {
		return result, err
	}
	
	if value != nil {
		result = value.(T)
	}
	
	return result, nil
}

// RetryStats tracks retry statistics
type RetryStats struct {
	TotalAttempts    int           `json:"total_attempts"`
	SuccessfulRetries int          `json:"successful_retries"`
	FailedRetries    int           `json:"failed_retries"`
	TotalDelay       time.Duration `json:"total_delay"`
	LastError        string        `json:"last_error,omitempty"`
	LastAttemptAt    time.Time     `json:"last_attempt_at"`
}

// StatsTrackingRetrier wraps a retrier to track statistics
type StatsTrackingRetrier struct {
	*Retrier
	stats RetryStats
}

// NewStatsTrackingRetrier creates a new stats-tracking retrier
func NewStatsTrackingRetrier(config *RetryConfig) *StatsTrackingRetrier {
	return &StatsTrackingRetrier{
		Retrier: NewRetrier(config),
		stats:   RetryStats{},
	}
}

// Do executes the function with retry logic and tracks statistics
func (str *StatsTrackingRetrier) Do(fn func() error) error {
	startTime := time.Now()
	str.stats.LastAttemptAt = startTime
	
	err := str.Retrier.Do(fn)
	
	str.stats.TotalAttempts++
	str.stats.TotalDelay += time.Since(startTime)
	
	if err != nil {
		str.stats.FailedRetries++
		str.stats.LastError = err.Error()
	} else {
		str.stats.SuccessfulRetries++
		str.stats.LastError = ""
	}
	
	return err
}

// GetStats returns current retry statistics
func (str *StatsTrackingRetrier) GetStats() RetryStats {
	return str.stats
}

// ResetStats resets retry statistics
func (str *StatsTrackingRetrier) ResetStats() {
	str.stats = RetryStats{}
}