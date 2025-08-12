package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"linke/internal/shared/logger"
)

// WarmingStrategy defines different cache warming approaches
type WarmingStrategy string

const (
	WarmingStrategyEager      WarmingStrategy = "eager"      // Load all data at startup
	WarmingStrategyLazy       WarmingStrategy = "lazy"       // Load data on first access
	WarmingStrategyScheduled  WarmingStrategy = "scheduled"  // Load data on schedule
	WarmingStrategyPredictive WarmingStrategy = "predictive" // Load data based on usage patterns
)

// WarmingConfig configures cache warming behavior
type WarmingConfig struct {
	Strategy       WarmingStrategy `json:"strategy"`
	BatchSize      int             `json:"batch_size"`
	ConcurrentJobs int             `json:"concurrent_jobs"`
	WarmingTTL     time.Duration   `json:"warming_ttl" swaggertype:"string" example:"1h"`
	Schedule       string          `json:"schedule"`  // Cron expression for scheduled warming
	Prefixes       []string        `json:"prefixes"`  // Prefixes to warm
	MaxItems       int             `json:"max_items"` // Maximum items to warm per prefix
	Enabled        bool            `json:"enabled"`
}

// WarmingDataProvider defines interface for providing data to warm
type WarmingDataProvider interface {
	GetWarmingData(ctx context.Context, prefix string, limit int) (map[string][]byte, error)
	GetCriticalKeys(ctx context.Context) ([]string, error)
	GetPopularKeys(ctx context.Context, limit int) ([]string, error)
}

// WarmingMetrics tracks warming performance
type WarmingMetrics struct {
	TotalWarmed     int64         `json:"total_warmed"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	LastWarmTime    time.Time     `json:"last_warm_time" swaggertype:"string" format:"date-time" example:"2023-01-01T00:00:00Z"`
	WarmingDuration time.Duration `json:"warming_duration" swaggertype:"string" example:"1h30m"`
	ItemsPerSecond  float64       `json:"items_per_second"`
}

// CacheWarmer handles cache warming operations
type CacheWarmer struct {
	cache    Cache
	config   *WarmingConfig
	provider WarmingDataProvider
	logger   logger.Logger

	metrics   *WarmingMetrics
	metricsMu sync.RWMutex

	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewCacheWarmer creates a new cache warmer instance
func NewCacheWarmer(
	cache Cache,
	config *WarmingConfig,
	provider WarmingDataProvider,
	logger logger.Logger,
) *CacheWarmer {
	if config == nil {
		config = &WarmingConfig{
			Strategy:       WarmingStrategyLazy,
			BatchSize:      100,
			ConcurrentJobs: 5,
			WarmingTTL:     1 * time.Hour,
			MaxItems:       1000,
			Enabled:        true,
		}
	}

	return &CacheWarmer{
		cache:    cache,
		config:   config,
		provider: provider,
		logger:   logger,
		metrics:  &WarmingMetrics{},
		stopChan: make(chan struct{}),
	}
}

// Start begins the cache warming process based on configured strategy
func (cw *CacheWarmer) Start(ctx context.Context) error {
	if !cw.config.Enabled {
		cw.logger.Info("Cache warming is disabled")
		return nil
	}

	switch cw.config.Strategy {
	case WarmingStrategyEager:
		return cw.warmEager(ctx)
	case WarmingStrategyScheduled:
		return cw.startScheduledWarming(ctx)
	case WarmingStrategyPredictive:
		return cw.startPredictiveWarming(ctx)
	case WarmingStrategyLazy:
		cw.logger.Info("Lazy warming strategy - cache will be warmed on demand")
		return nil
	default:
		return fmt.Errorf("unknown warming strategy: %s", cw.config.Strategy)
	}
}

// Stop halts all warming operations
func (cw *CacheWarmer) Stop() {
	close(cw.stopChan)
	cw.wg.Wait()
}

// WarmPrefix warms cache for a specific prefix
func (cw *CacheWarmer) WarmPrefix(ctx context.Context, prefix string) error {
	startTime := time.Now()

	cw.logger.Info("Starting cache warming for prefix",
		logger.String("prefix", prefix),
		logger.Int("max_items", cw.config.MaxItems))

	data, err := cw.provider.GetWarmingData(ctx, prefix, cw.config.MaxItems)
	if err != nil {
		cw.updateMetrics(func(m *WarmingMetrics) {
			m.ErrorCount++
		})
		return fmt.Errorf("failed to get warming data for prefix %s: %w", prefix, err)
	}

	successCount, errorCount := cw.warmDataBatch(ctx, data)

	duration := time.Since(startTime)
	itemsPerSecond := float64(len(data)) / duration.Seconds()

	cw.updateMetrics(func(m *WarmingMetrics) {
		m.TotalWarmed += int64(len(data))
		m.SuccessCount += successCount
		m.ErrorCount += errorCount
		m.LastWarmTime = time.Now()
		m.WarmingDuration = duration
		m.ItemsPerSecond = itemsPerSecond
	})

	cw.logger.Info("Completed cache warming for prefix",
		logger.String("prefix", prefix),
		logger.Int("total_items", len(data)),
		logger.Int64("success_count", successCount),
		logger.Int64("error_count", errorCount),
		logger.Duration("duration", duration),
		logger.String("items_per_second", fmt.Sprintf("%.2f", itemsPerSecond)))

	return nil
}

// WarmCriticalData warms cache with critical data
func (cw *CacheWarmer) WarmCriticalData(ctx context.Context) error {
	keys, err := cw.provider.GetCriticalKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to get critical keys: %w", err)
	}

	cw.logger.Info("Warming critical data", logger.Int("key_count", len(keys)))

	for _, key := range keys {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cw.stopChan:
			return nil
		default:
			// The provider should supply the actual data for each key
			// This is a simplified implementation
			if exists, err := cw.cache.Exists(ctx, key); err == nil && !exists {
				cw.logger.Debug("Critical key not in cache, should be loaded",
					logger.String("key", key))
			}
		}
	}

	return nil
}

// WarmPopularData warms cache with popular data
func (cw *CacheWarmer) WarmPopularData(ctx context.Context, limit int) error {
	keys, err := cw.provider.GetPopularKeys(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to get popular keys: %w", err)
	}

	cw.logger.Info("Warming popular data",
		logger.Int("key_count", len(keys)),
		logger.Int("limit", limit))

	// Similar to critical data, the actual warming would depend on the provider
	// providing the actual data for these keys
	for _, key := range keys {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cw.stopChan:
			return nil
		default:
			if exists, err := cw.cache.Exists(ctx, key); err == nil && !exists {
				cw.logger.Debug("Popular key not in cache, should be loaded",
					logger.String("key", key))
			}
		}
	}

	return nil
}

// GetMetrics returns current warming metrics
func (cw *CacheWarmer) GetMetrics() *WarmingMetrics {
	cw.metricsMu.RLock()
	defer cw.metricsMu.RUnlock()

	return &WarmingMetrics{
		TotalWarmed:     cw.metrics.TotalWarmed,
		SuccessCount:    cw.metrics.SuccessCount,
		ErrorCount:      cw.metrics.ErrorCount,
		LastWarmTime:    cw.metrics.LastWarmTime,
		WarmingDuration: cw.metrics.WarmingDuration,
		ItemsPerSecond:  cw.metrics.ItemsPerSecond,
	}
}

// Private methods

func (cw *CacheWarmer) warmEager(ctx context.Context) error {
	cw.logger.Info("Starting eager cache warming")

	for _, prefix := range cw.config.Prefixes {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-cw.stopChan:
			return nil
		default:
			if err := cw.WarmPrefix(ctx, prefix); err != nil {
				cw.logger.Error("Failed to warm prefix",
					logger.String("prefix", prefix),
					logger.ErrorField(err))
			}
		}
	}

	return nil
}

func (cw *CacheWarmer) startScheduledWarming(ctx context.Context) error {
	if cw.config.Schedule == "" {
		return fmt.Errorf("schedule not configured for scheduled warming")
	}

	cw.logger.Info("Starting scheduled cache warming",
		logger.String("schedule", cw.config.Schedule))

	cw.wg.Add(1)
	go func() {
		defer cw.wg.Done()

		// Simple implementation - in production, use a proper cron library
		ticker := time.NewTicker(1 * time.Hour) // Default to hourly
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := cw.warmEager(ctx); err != nil {
					cw.logger.Error("Scheduled warming failed", logger.ErrorField(err))
				}
			case <-cw.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (cw *CacheWarmer) startPredictiveWarming(ctx context.Context) error {
	cw.logger.Info("Starting predictive cache warming")

	cw.wg.Add(1)
	go func() {
		defer cw.wg.Done()

		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Warm popular data based on recent usage patterns
				if err := cw.WarmPopularData(ctx, cw.config.MaxItems/2); err != nil {
					cw.logger.Error("Predictive warming failed", logger.ErrorField(err))
				}
			case <-cw.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (cw *CacheWarmer) warmDataBatch(ctx context.Context, data map[string][]byte) (int64, int64) {
	var successCount, errorCount int64

	// Create worker pool for concurrent warming
	workChan := make(chan warmingWork, len(data))
	resultChan := make(chan warmingResult, len(data))

	// Start workers
	for i := 0; i < cw.config.ConcurrentJobs; i++ {
		cw.wg.Add(1)
		go cw.warmingWorker(ctx, workChan, resultChan)
	}

	// Send work
	for key, value := range data {
		select {
		case workChan <- warmingWork{key: key, value: value}:
		case <-ctx.Done():
			close(workChan)
			return successCount, errorCount
		case <-cw.stopChan:
			close(workChan)
			return successCount, errorCount
		}
	}
	close(workChan)

	// Collect results
	for i := 0; i < len(data); i++ {
		select {
		case result := <-resultChan:
			if result.err != nil {
				errorCount++
				cw.logger.Debug("Failed to warm cache key",
					logger.String("key", result.key),
					logger.ErrorField(result.err))
			} else {
				successCount++
			}
		case <-ctx.Done():
			return successCount, errorCount
		case <-cw.stopChan:
			return successCount, errorCount
		}
	}

	return successCount, errorCount
}

type warmingWork struct {
	key   string
	value []byte
}

type warmingResult struct {
	key string
	err error
}

func (cw *CacheWarmer) warmingWorker(ctx context.Context, workChan <-chan warmingWork, resultChan chan<- warmingResult) {
	defer cw.wg.Done()

	for work := range workChan {
		select {
		case <-ctx.Done():
			return
		case <-cw.stopChan:
			return
		default:
			err := cw.cache.Set(ctx, work.key, work.value, cw.config.WarmingTTL)
			resultChan <- warmingResult{key: work.key, err: err}
		}
	}
}

func (cw *CacheWarmer) updateMetrics(updateFunc func(*WarmingMetrics)) {
	cw.metricsMu.Lock()
	updateFunc(cw.metrics)
	cw.metricsMu.Unlock()
}

// DefaultWarmingDataProvider provides a basic implementation of WarmingDataProvider
type DefaultWarmingDataProvider struct {
	cache       Cache
	keyPatterns map[string]string // prefix -> pattern mapping
}

// NewDefaultWarmingDataProvider creates a default warming data provider
func NewDefaultWarmingDataProvider(cache Cache) *DefaultWarmingDataProvider {
	return &DefaultWarmingDataProvider{
		cache: cache,
		keyPatterns: map[string]string{
			CachePrefixUser:         CachePrefixUser + "*",
			CachePrefixSubscription: CachePrefixSubscription + "*",
			CachePrefixPayment:      CachePrefixPayment + "*",
			CachePrefixPlan:         CachePrefixPlan + "*",
		},
	}
}

// GetWarmingData returns data to warm for a given prefix
func (dwdp *DefaultWarmingDataProvider) GetWarmingData(ctx context.Context, prefix string, limit int) (map[string][]byte, error) {
	// This is a placeholder implementation
	// In a real scenario, this would query the database or other data sources
	// to get the most relevant data to cache

	data := make(map[string][]byte)

	// For now, return empty data - this should be implemented based on specific business logic
	return data, nil
}

// GetCriticalKeys returns keys that are critical for application performance
func (dwdp *DefaultWarmingDataProvider) GetCriticalKeys(ctx context.Context) ([]string, error) {
	// Return a list of critical keys that should always be cached
	// This could include frequently accessed configuration, user permissions, etc.

	keys := []string{
		CachePrefixConfig + "app_settings",
		CachePrefixPlan + "active:all",
		CachePrefixServer + "active:all",
	}

	return keys, nil
}

// GetPopularKeys returns keys that are frequently accessed
func (dwdp *DefaultWarmingDataProvider) GetPopularKeys(ctx context.Context, limit int) ([]string, error) {
	// This should analyze access patterns and return the most popular keys
	// For now, return empty list - should be implemented based on metrics/analytics

	return []string{}, nil
}
