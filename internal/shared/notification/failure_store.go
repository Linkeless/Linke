package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/cache"
	"linke/internal/shared/logger"
)

// MemoryFailureStore provides an in-memory implementation of FailureStore
type MemoryFailureStore struct {
	failures    map[string]*NotificationFailure
	deadLetters map[string]*DeadLetterEntry
	mutex       sync.RWMutex
	logger      logger.Logger
}

// DeadLetterEntry represents a failed notification that has exceeded max retries
type DeadLetterEntry struct {
	Failure     *NotificationFailure `json:"failure"`
	Reason      string               `json:"reason"`
	MovedAt     time.Time            `json:"moved_at"`
	ProcessedBy string               `json:"processed_by,omitempty"`
}

// NewMemoryFailureStore creates a new in-memory failure store
func NewMemoryFailureStore(logger logger.Logger) *MemoryFailureStore {
	return &MemoryFailureStore{
		failures:    make(map[string]*NotificationFailure),
		deadLetters: make(map[string]*DeadLetterEntry),
		logger:      logger,
	}
}

// StoreFailure stores a notification failure
func (s *MemoryFailureStore) StoreFailure(ctx context.Context, failure *NotificationFailure) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.failures[failure.RequestID] = failure

	s.logger.Debug("Stored notification failure",
		logger.String("request_id", failure.RequestID),
		logger.String("channel", string(failure.Channel)),
		logger.Int("attempt_count", failure.AttemptCount))

	return nil
}

// GetFailure retrieves a notification failure by request ID
func (s *MemoryFailureStore) GetFailure(ctx context.Context, requestID string) (*NotificationFailure, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	failure, exists := s.failures[requestID]
	if !exists {
		return nil, fmt.Errorf("failure not found: %s", requestID)
	}

	return failure, nil
}

// UpdateFailure updates an existing notification failure
func (s *MemoryFailureStore) UpdateFailure(ctx context.Context, failure *NotificationFailure) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.failures[failure.RequestID]; !exists {
		return fmt.Errorf("failure not found for update: %s", failure.RequestID)
	}

	s.failures[failure.RequestID] = failure

	s.logger.Debug("Updated notification failure",
		logger.String("request_id", failure.RequestID),
		logger.Int("attempt_count", failure.AttemptCount))

	return nil
}

// GetPendingRetries retrieves pending notification retries
func (s *MemoryFailureStore) GetPendingRetries(ctx context.Context, limit int) ([]*NotificationFailure, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var pending []*NotificationFailure
	now := time.Now()

	for _, failure := range s.failures {
		// Check if this failure is ready for retry
		if failure.NextRetryAt == nil || failure.NextRetryAt.After(now) {
			continue
		}

		pending = append(pending, failure)

		if len(pending) >= limit {
			break
		}
	}

	return pending, nil
}

// MarkAsDeadLetter moves a failure to the dead letter queue
func (s *MemoryFailureStore) MarkAsDeadLetter(ctx context.Context, requestID, reason string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	failure, exists := s.failures[requestID]
	if !exists {
		// If it's a success case (retry succeeded), just log it
		if reason == "retry_succeeded" {
			s.logger.Info("Notification retry succeeded, removing from failure tracking",
				logger.String("request_id", requestID))
			return nil
		}
		return fmt.Errorf("failure not found for dead letter: %s", requestID)
	}

	// Move to dead letters
	deadLetter := &DeadLetterEntry{
		Failure: failure,
		Reason:  reason,
		MovedAt: time.Now(),
	}

	s.deadLetters[requestID] = deadLetter
	delete(s.failures, requestID)

	s.logger.Warn("Moved notification failure to dead letter queue",
		logger.String("request_id", requestID),
		logger.String("reason", reason),
		logger.String("channel", string(failure.Channel)),
		logger.Int("attempts", failure.AttemptCount))

	return nil
}

// GetDeadLetters retrieves dead letter entries for analysis
func (s *MemoryFailureStore) GetDeadLetters(ctx context.Context, limit int) ([]*DeadLetterEntry, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var deadLetters []*DeadLetterEntry
	for _, entry := range s.deadLetters {
		deadLetters = append(deadLetters, entry)
		if len(deadLetters) >= limit {
			break
		}
	}

	return deadLetters, nil
}

// GetStats returns statistics about the failure store
func (s *MemoryFailureStore) GetStats() map[string]any {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return map[string]any{
		"pending_failures": len(s.failures),
		"dead_letters":     len(s.deadLetters),
		"store_type":       "memory",
		"last_updated":     time.Now().Format(time.RFC3339),
	}
}

// RedisFailureStore provides a Redis-based implementation of FailureStore
type RedisFailureStore struct {
	cache  cache.CacheStore
	logger logger.Logger
	prefix string
}

// NewRedisFailureStore creates a new Redis-based failure store
func NewRedisFailureStore(cache cache.CacheStore, logger logger.Logger) *RedisFailureStore {
	return &RedisFailureStore{
		cache:  cache,
		logger: logger,
		prefix: "notification_failures:",
	}
}

// StoreFailure stores a notification failure in Redis
func (s *RedisFailureStore) StoreFailure(ctx context.Context, failure *NotificationFailure) error {
	key := s.prefix + failure.RequestID

	// Store with TTL of 7 days to prevent indefinite growth
	ttl := 7 * 24 * time.Hour

	if err := s.cache.SetJSON(ctx, key, failure, ttl); err != nil {
		return fmt.Errorf("failed to store failure in Redis: %w", err)
	}

	// Store in pending retry index for easier querying
	if failure.NextRetryAt != nil {
		retryIndexKey := fmt.Sprintf("%spending_index:%d:%s", s.prefix, failure.NextRetryAt.Unix(), failure.RequestID)
		if err := s.cache.Set(ctx, retryIndexKey, "1", ttl); err != nil {
			s.logger.Warn("Failed to add to pending retry index",
				logger.String("request_id", failure.RequestID),
				logger.ErrorField(err))
		}
	}

	s.logger.Debug("Stored notification failure in Redis",
		logger.String("request_id", failure.RequestID),
		logger.String("channel", string(failure.Channel)),
		logger.Int("attempt_count", failure.AttemptCount))

	return nil
}

// GetFailure retrieves a notification failure from Redis
func (s *RedisFailureStore) GetFailure(ctx context.Context, requestID string) (*NotificationFailure, error) {
	key := s.prefix + requestID

	var failure NotificationFailure
	if err := s.cache.GetJSON(ctx, key, &failure); err != nil {
		return nil, fmt.Errorf("failed to get failure from Redis: %w", err)
	}

	return &failure, nil
}

// UpdateFailure updates an existing notification failure in Redis
func (s *RedisFailureStore) UpdateFailure(ctx context.Context, failure *NotificationFailure) error {
	key := s.prefix + failure.RequestID

	// Check if failure exists
	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check failure existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("failure not found for update: %s", failure.RequestID)
	}

	// Update with extended TTL
	ttl := 7 * 24 * time.Hour
	if err := s.cache.SetJSON(ctx, key, failure, ttl); err != nil {
		return fmt.Errorf("failed to update failure in Redis: %w", err)
	}

	// Update pending retry index
	if failure.NextRetryAt != nil {
		retryIndexKey := fmt.Sprintf("%spending_index:%d:%s", s.prefix, failure.NextRetryAt.Unix(), failure.RequestID)
		if err := s.cache.Set(ctx, retryIndexKey, "1", ttl); err != nil {
			s.logger.Warn("Failed to update pending retry index",
				logger.String("request_id", failure.RequestID),
				logger.ErrorField(err))
		}
	}

	s.logger.Debug("Updated notification failure in Redis",
		logger.String("request_id", failure.RequestID),
		logger.Int("attempt_count", failure.AttemptCount))

	return nil
}

// GetPendingRetries retrieves pending notification retries from Redis
func (s *RedisFailureStore) GetPendingRetries(ctx context.Context, limit int) ([]*NotificationFailure, error) {
	// Since we can't use sorted sets or pattern matching with the current cache interface,
	// we'll implement a simplified approach that would work in a real production system
	// with proper Redis operations or a background indexing system

	s.logger.Debug("Getting pending retries - simplified implementation",
		logger.Int("requested_limit", limit))

	var failures []*NotificationFailure

	// TODO: In a production system, implement one of these approaches:
	// 1. Use a background task to maintain a list of ready-for-retry failures
	// 2. Use Redis sorted sets with the full Redis client (not just CacheStore)
	// 3. Store retry timestamps in a queryable format
	//
	// For now, returning empty list to satisfy the interface
	// The memory-based failure store will handle retries in development

	return failures, nil
}

// MarkAsDeadLetter moves a failure to the dead letter queue in Redis
func (s *RedisFailureStore) MarkAsDeadLetter(ctx context.Context, requestID, reason string) error {
	// Special handling for success case
	if reason == "retry_succeeded" {
		return s.removeFailure(ctx, requestID, "retry succeeded")
	}

	// Get the failure first
	failure, err := s.GetFailure(ctx, requestID)
	if err != nil {
		return fmt.Errorf("failed to get failure for dead letter: %w", err)
	}

	// Create dead letter entry
	deadLetter := &DeadLetterEntry{
		Failure: failure,
		Reason:  reason,
		MovedAt: time.Now(),
	}

	// Store in dead letters with 30-day TTL
	deadLetterKey := s.prefix + "dead_letter:" + requestID
	deadLetterTTL := 30 * 24 * time.Hour

	if err := s.cache.SetJSON(ctx, deadLetterKey, deadLetter, deadLetterTTL); err != nil {
		return fmt.Errorf("failed to store dead letter: %w", err)
	}

	// Remove from active failures
	if err := s.removeFailure(ctx, requestID, reason); err != nil {
		s.logger.Warn("Failed to remove failure after moving to dead letter",
			logger.String("request_id", requestID),
			logger.ErrorField(err))
	}

	s.logger.Warn("Moved notification failure to dead letter queue",
		logger.String("request_id", requestID),
		logger.String("reason", reason),
		logger.String("channel", string(failure.Channel)),
		logger.Int("attempts", failure.AttemptCount))

	return nil
}

// removeFailure removes a failure from active tracking
func (s *RedisFailureStore) removeFailure(ctx context.Context, requestID, reason string) error {
	key := s.prefix + requestID

	// Remove from failure store
	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Warn("Failed to delete failure from Redis",
			logger.String("request_id", requestID),
			logger.ErrorField(err))
	}

	// Remove any associated retry index entries (simplified cleanup)
	// In a production system, you'd want a more efficient cleanup mechanism
	s.logger.Info("Removed notification failure from tracking",
		logger.String("request_id", requestID),
		logger.String("reason", reason))

	return nil
}

// GetDeadLetters retrieves dead letter entries from Redis
func (s *RedisFailureStore) GetDeadLetters(ctx context.Context, limit int) ([]*DeadLetterEntry, error) {
	// Since we don't have Keys pattern matching available, we'll return a simplified implementation
	// In a production system, you'd maintain a separate index of dead letter entries
	s.logger.Debug("Getting dead letters - simplified implementation",
		logger.Int("requested_limit", limit))

	var deadLetters []*DeadLetterEntry
	// TODO: Implement proper dead letter retrieval when pattern matching is available
	return deadLetters, nil
}

// GetStats returns statistics about the Redis failure store
func (s *RedisFailureStore) GetStats(ctx context.Context) (map[string]any, error) {
	// Since we don't have advanced Redis operations available, return basic stats
	s.logger.Debug("Getting failure store stats - simplified implementation")

	return map[string]any{
		"pending_failures": "unavailable", // Would need pattern matching to count
		"dead_letters":     "unavailable", // Would need pattern matching to count
		"store_type":       "redis_simple",
		"last_updated":     time.Now().Format(time.RFC3339),
		"note":             "Advanced statistics require Redis pattern matching operations",
	}, nil
}
