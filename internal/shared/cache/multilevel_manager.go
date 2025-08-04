package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"linke/internal/shared/logger"
)

// DefaultMultiLevelCacheManager implements MultiLevelCacheManager
type DefaultMultiLevelCacheManager struct {
	multilevelCache *MultiLevelCache
	l1Cache         *MemoryCache
	l2Cache         Cache
	warmer          *CacheWarmer
	invalidator     *EventDrivenInvalidator
	monitor         *MultiLevelCacheMonitor
	cacheKeys       *AllCacheKeys
	collector       MetricsCollector
	logger          logger.Logger
}

// NewMultiLevelCacheManager creates a new multi-level cache manager
func NewMultiLevelCacheManager(
	config *MultiLevelCacheConfig,
	l2Cache Cache,
	cacheKeys *AllCacheKeys,
	collector MetricsCollector,
	logger logger.Logger,
) MultiLevelCacheManager {
	// Create multi-level cache
	multilevelCache := NewMultiLevelCache(config, l2Cache, collector, logger)

	// Create cache warmer
	warmingConfig := &WarmingConfig{
		Strategy:       WarmingStrategyLazy,
		BatchSize:      100,
		ConcurrentJobs: 5,
		WarmingTTL:     config.L2Config.DefaultTTL,
		Prefixes: []string{
			CachePrefixUser,
			CachePrefixSubscription,
			CachePrefixPayment,
			CachePrefixPlan,
		},
		MaxItems: 1000,
		Enabled:  true,
	}

	warmingProvider := NewDefaultWarmingDataProvider(l2Cache)
	warmer := NewCacheWarmer(multilevelCache, warmingConfig, warmingProvider, logger)

	// Create event-driven invalidator
	invalidationConfig := &CacheInvalidationConfig{
		Enabled:       true,
		AsyncMode:     true,
		BatchSize:     100,
		BufferTimeout: 1 * time.Second,
	}
	invalidator := NewEventDrivenInvalidator(multilevelCache, cacheKeys, invalidationConfig, logger)

	// Create monitor
	monitor := NewMultiLevelCacheMonitor(multilevelCache, warmer, invalidator, logger)

	return &DefaultMultiLevelCacheManager{
		multilevelCache: multilevelCache,
		l1Cache:         multilevelCache.l1Cache,
		l2Cache:         l2Cache,
		warmer:          warmer,
		invalidator:     invalidator,
		monitor:         monitor,
		cacheKeys:       cacheKeys,
		collector:       collector,
		logger:          logger,
	}
}

// GetCache returns the multi-level cache as a standard Cache interface
func (mlcm *DefaultMultiLevelCacheManager) GetCache() Cache {
	return mlcm.multilevelCache
}

// GetTypedCache returns a typed cache with the specified prefix
func (mlcm *DefaultMultiLevelCacheManager) GetTypedCache(prefix string) TypedCache[any] {
	return NewTypedCache[any](mlcm.multilevelCache, prefix)
}

// InvalidateCache invalidates cache entries matching the given patterns
func (mlcm *DefaultMultiLevelCacheManager) InvalidateCache(ctx context.Context, patterns ...string) error {
	for _, pattern := range patterns {
		if err := mlcm.multilevelCache.DeleteByPattern(ctx, pattern); err != nil {
			return fmt.Errorf("failed to invalidate pattern %s: %w", pattern, err)
		}
	}
	return nil
}

// GetStats returns cache statistics
func (mlcm *DefaultMultiLevelCacheManager) GetStats(ctx context.Context) (*CacheStats, error) {
	metrics := mlcm.multilevelCache.GetMetrics()

	stats := &CacheStats{
		Hits:      metrics.L1Hits + metrics.L2Hits,
		Misses:    metrics.TotalMisses,
		HitRate:   metrics.OverallHitRate,
		TotalKeys: 0, // Would need to be calculated from both levels
	}

	if metrics.L1Metrics != nil {
		stats.TotalKeys += metrics.L1Metrics.entryCount
		stats.MemoryUsed = metrics.L1Metrics.currentSize
		stats.Evictions = metrics.L1Metrics.evictions
	}

	if metrics.L2Metrics != nil {
		stats.Sets = metrics.L2Metrics.Sets
		stats.Deletes = metrics.L2Metrics.Deletes
		stats.TotalKeys += metrics.L2Metrics.TotalOps
	}

	return stats, nil
}

// Multi-level specific methods

// GetMultiLevelCache returns the multi-level cache instance
func (mlcm *DefaultMultiLevelCacheManager) GetMultiLevelCache() *MultiLevelCache {
	return mlcm.multilevelCache
}

// GetL1Cache returns the L1 (memory) cache
func (mlcm *DefaultMultiLevelCacheManager) GetL1Cache() *MemoryCache {
	return mlcm.l1Cache
}

// GetL2Cache returns the L2 (Redis) cache
func (mlcm *DefaultMultiLevelCacheManager) GetL2Cache() Cache {
	return mlcm.l2Cache
}

// GetWarmer returns the cache warmer
func (mlcm *DefaultMultiLevelCacheManager) GetWarmer() *CacheWarmer {
	return mlcm.warmer
}

// GetInvalidator returns the event-driven invalidator
func (mlcm *DefaultMultiLevelCacheManager) GetInvalidator() *EventDrivenInvalidator {
	return mlcm.invalidator
}

// GetMonitor returns the cache monitor
func (mlcm *DefaultMultiLevelCacheManager) GetMonitor() *MultiLevelCacheMonitor {
	return mlcm.monitor
}

// SwitchStrategy dynamically switches caching strategies
func (mlcm *DefaultMultiLevelCacheManager) SwitchStrategy(writeStrategy WriteStrategy, readStrategy ReadStrategy) error {
	mlcm.logger.Info("Switching cache strategies",
		logger.String("write_strategy", string(writeStrategy)),
		logger.String("read_strategy", string(readStrategy)))

	// Update configuration
	mlcm.multilevelCache.config.WriteStrategy = writeStrategy
	mlcm.multilevelCache.config.ReadStrategy = readStrategy

	// If switching to write-behind and worker isn't running, start it
	if writeStrategy == WriteStrategyBehind && mlcm.multilevelCache.config.WriteStrategy != WriteStrategyBehind {
		mlcm.multilevelCache.startWriteBehindWorker()
	}

	return nil
}

// Start initializes all components
func (mlcm *DefaultMultiLevelCacheManager) Start(ctx context.Context) error {
	mlcm.logger.Info("Starting multi-level cache manager")

	// Start cache warming
	if err := mlcm.warmer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start cache warmer: %w", err)
	}

	mlcm.logger.Info("Multi-level cache manager started successfully")
	return nil
}

// Stop gracefully shuts down all components
func (mlcm *DefaultMultiLevelCacheManager) Stop() {
	mlcm.logger.Info("Stopping multi-level cache manager")

	// Stop components in reverse order
	if mlcm.warmer != nil {
		mlcm.warmer.Stop()
	}

	if mlcm.invalidator != nil {
		mlcm.invalidator.Stop()
	}

	if mlcm.multilevelCache != nil {
		mlcm.multilevelCache.Close()
	}

	mlcm.logger.Info("Multi-level cache manager stopped")
}

// RegisterEventHandler registers the invalidator as an event handler
func (mlcm *DefaultMultiLevelCacheManager) RegisterEventHandler(eventBus interface{}) error {
	// This would integrate with the actual event bus implementation
	// For now, we'll just log that it should be registered
	mlcm.logger.Info("Cache invalidator should be registered with event bus",
		logger.String("event_types", strings.Join(mlcm.invalidator.EventTypes(), ",")))

	return nil
}

// GetCacheKeyBuilder returns the cache key builder
func (mlcm *DefaultMultiLevelCacheManager) GetCacheKeyBuilder() *AllCacheKeys {
	return mlcm.cacheKeys
}

// TypedCache implementation for multi-level cache
type TypedMultiLevelCache[T any] struct {
	cache  Cache
	prefix string
}

// NewTypedCache creates a new typed cache wrapper
func NewTypedCache[T any](cache Cache, prefix string) TypedCache[T] {
	return &TypedMultiLevelCache[T]{
		cache:  cache,
		prefix: prefix,
	}
}

// Get retrieves a typed value from cache
func (tc *TypedMultiLevelCache[T]) Get(ctx context.Context, key string) (*T, error) {
	fullKey := tc.prefix + key
	data, err := tc.cache.Get(ctx, fullKey)
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, &CacheError{Op: "unmarshal", Key: fullKey, Err: err}
	}

	return &result, nil
}

// Set stores a typed value in cache
func (tc *TypedMultiLevelCache[T]) Set(ctx context.Context, key string, value *T, ttl time.Duration) error {
	fullKey := tc.prefix + key
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Op: "marshal", Key: fullKey, Err: err}
	}

	return tc.cache.Set(ctx, fullKey, data, ttl)
}

// Delete removes a value from cache
func (tc *TypedMultiLevelCache[T]) Delete(ctx context.Context, key string) error {
	fullKey := tc.prefix + key
	return tc.cache.Delete(ctx, fullKey)
}

// DeleteByPattern removes values matching a pattern
func (tc *TypedMultiLevelCache[T]) DeleteByPattern(ctx context.Context, pattern string) error {
	fullPattern := tc.prefix + pattern
	return tc.cache.DeleteByPattern(ctx, fullPattern)
}

// Exists checks if a key exists
func (tc *TypedMultiLevelCache[T]) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := tc.prefix + key
	return tc.cache.Exists(ctx, fullKey)
}
