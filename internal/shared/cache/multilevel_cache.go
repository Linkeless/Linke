package cache

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"linke/internal/shared/logger"
)

// CacheLevel represents different cache levels
type CacheLevel int

const (
	CacheLevelL1 CacheLevel = iota // Memory cache
	CacheLevelL2                   // Redis cache
)

// MultiLevelCacheConfig configures the multi-level cache
type MultiLevelCacheConfig struct {
	L1Config         *MemoryCacheConfig `json:"l1_config"`
	L2Config         *CacheConfig       `json:"l2_config"`
	EnableL1         bool               `json:"enable_l1"`
	EnableL2         bool               `json:"enable_l2"`
	WriteStrategy    WriteStrategy      `json:"write_strategy"`
	ReadStrategy     ReadStrategy       `json:"read_strategy"`
	PromotionRatio   float64            `json:"promotion_ratio"`   // Ratio of L2 hits that get promoted to L1
	ReplicationDelay time.Duration      `json:"replication_delay"` // Delay for write-behind strategy
	
	// Cache avalanche protection
	EnableTTLJitter  bool    `json:"enable_ttl_jitter"`  // Enable TTL randomization to prevent cache avalanche
	JitterPercent    float64 `json:"jitter_percent"`     // Percentage of jitter to add to TTL (0.0-1.0)
}

// WriteStrategy defines how writes are handled across cache levels
type WriteStrategy string

const (
	WriteStrategyThrough WriteStrategy = "write_through" // Write to all levels synchronously
	WriteStrategyBehind  WriteStrategy = "write_behind"  // Write to L1 immediately, L2 asynchronously
	WriteStrategyAround  WriteStrategy = "write_around"  // Write only to L2, bypass L1
)

// ReadStrategy defines how reads are handled across cache levels
type ReadStrategy string

const (
	ReadStrategyFailover    ReadStrategy = "failover"    // Try L1, then L2 on miss
	ReadStrategyPromotion   ReadStrategy = "promotion"   // Read from L1/L2, promote hot data to L1
	ReadStrategyReplication ReadStrategy = "replication" // Replicate all L2 hits to L1
)

// MultiLevelCacheMetrics tracks performance across cache levels
type MultiLevelCacheMetrics struct {
	L1Metrics *MemoryCacheMetrics `json:"l1_metrics"`
	L2Metrics *Metrics            `json:"l2_metrics"`

	// Multi-level specific metrics
	L1Hits       int64 `json:"l1_hits"`
	L2Hits       int64 `json:"l2_hits"`
	TotalMisses  int64 `json:"total_misses"`
	Promotions   int64 `json:"promotions"`
	Demotions    int64 `json:"demotions"`
	WriteBehinds int64 `json:"write_behinds"`

	// Computed metrics
	OverallHitRate float64 `json:"overall_hit_rate"`
	L1HitRate      float64 `json:"l1_hit_rate"`
	L2HitRate      float64 `json:"l2_hit_rate"`
}

// MultiLevelCache implements a multi-level caching strategy
type MultiLevelCache struct {
	config    *MultiLevelCacheConfig
	l1Cache   *MemoryCache
	l2Cache   Cache
	collector MetricsCollector
	logger    logger.Logger

	// Write-behind support
	writeQueue chan *writeOperation
	stopWriter chan struct{}

	// Metrics
	metrics struct {
		l1Hits      int64
		l2Hits      int64
		totalMisses int64
		promotions  int64
		demotions   int64
		writeBehind int64
	}
}

type writeOperation struct {
	key       string
	value     []byte
	ttl       time.Duration
	timestamp time.Time
}

// NewMultiLevelCache creates a new multi-level cache instance
func NewMultiLevelCache(
	config *MultiLevelCacheConfig,
	l2Cache Cache,
	collector MetricsCollector,
	logger logger.Logger,
) *MultiLevelCache {
	if config == nil {
		config = &MultiLevelCacheConfig{
			EnableL1:         true,
			EnableL2:         true,
			WriteStrategy:    WriteStrategyThrough,
			ReadStrategy:     ReadStrategyPromotion,
			PromotionRatio:   0.8,
			ReplicationDelay: 100 * time.Millisecond,
			EnableTTLJitter:  true,   // Enable jitter by default
			JitterPercent:    0.2,    // 20% jitter by default
		}
	}

	// Validate jitter configuration
	if config.EnableTTLJitter && (config.JitterPercent < 0 || config.JitterPercent > 1.0) {
		config.JitterPercent = 0.2 // Default to 20% if invalid
	}

	mlc := &MultiLevelCache{
		config:     config,
		l2Cache:    l2Cache,
		collector:  collector,
		logger:     logger,
		writeQueue: make(chan *writeOperation, 1000),
		stopWriter: make(chan struct{}),
	}

	// Initialize L1 cache if enabled
	if config.EnableL1 {
		mlc.l1Cache = NewMemoryCache(config.L1Config)
	}

	// Start write-behind worker if needed
	if config.WriteStrategy == WriteStrategyBehind {
		mlc.startWriteBehindWorker()
	}

	return mlc
}

// Get retrieves a value using the configured read strategy
func (mlc *MultiLevelCache) Get(ctx context.Context, key string) ([]byte, error) {
	switch mlc.config.ReadStrategy {
	case ReadStrategyFailover:
		return mlc.getFailover(ctx, key)
	case ReadStrategyPromotion:
		return mlc.getWithPromotion(ctx, key)
	case ReadStrategyReplication:
		return mlc.getWithReplication(ctx, key)
	default:
		return mlc.getFailover(ctx, key)
	}
}

// Set stores a value using the configured write strategy
func (mlc *MultiLevelCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	switch mlc.config.WriteStrategy {
	case WriteStrategyThrough:
		return mlc.setWriteThrough(ctx, key, value, ttl)
	case WriteStrategyBehind:
		return mlc.setWriteBehind(ctx, key, value, ttl)
	case WriteStrategyAround:
		return mlc.setWriteAround(ctx, key, value, ttl)
	default:
		return mlc.setWriteThrough(ctx, key, value, ttl)
	}
}

// Delete removes a value from all cache levels
func (mlc *MultiLevelCache) Delete(ctx context.Context, key string) error {
	var errs []error

	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if err := mlc.l1Cache.Delete(ctx, key); err != nil {
			errs = append(errs, fmt.Errorf("L1 delete error: %w", err))
		}
	}

	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if err := mlc.l2Cache.Delete(ctx, key); err != nil {
			errs = append(errs, fmt.Errorf("L2 delete error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multi-level delete errors: %v", errs)
	}

	return nil
}

// DeleteByPattern removes entries matching a pattern from all cache levels
func (mlc *MultiLevelCache) DeleteByPattern(ctx context.Context, pattern string) error {
	var errs []error

	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if err := mlc.l1Cache.DeleteByPattern(ctx, pattern); err != nil {
			errs = append(errs, fmt.Errorf("L1 delete pattern error: %w", err))
		}
	}

	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if err := mlc.l2Cache.DeleteByPattern(ctx, pattern); err != nil {
			errs = append(errs, fmt.Errorf("L2 delete pattern error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multi-level delete pattern errors: %v", errs)
	}

	return nil
}

// Exists checks if a key exists in any cache level
func (mlc *MultiLevelCache) Exists(ctx context.Context, key string) (bool, error) {
	// Check L1 first
	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if exists, err := mlc.l1Cache.Exists(ctx, key); err == nil && exists {
			return true, nil
		}
	}

	// Check L2
	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		return mlc.l2Cache.Exists(ctx, key)
	}

	return false, nil
}

// Expire sets TTL for a key in all cache levels
func (mlc *MultiLevelCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	var errs []error

	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if err := mlc.l1Cache.Expire(ctx, key, ttl); err != nil {
			errs = append(errs, fmt.Errorf("L1 expire error: %w", err))
		}
	}

	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if err := mlc.l2Cache.Expire(ctx, key, ttl); err != nil {
			errs = append(errs, fmt.Errorf("L2 expire error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multi-level expire errors: %v", errs)
	}

	return nil
}

// Flush clears all cache levels
func (mlc *MultiLevelCache) Flush(ctx context.Context) error {
	var errs []error

	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if err := mlc.l1Cache.Flush(ctx); err != nil {
			errs = append(errs, fmt.Errorf("L1 flush error: %w", err))
		}
	}

	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if err := mlc.l2Cache.Flush(ctx); err != nil {
			errs = append(errs, fmt.Errorf("L2 flush error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multi-level flush errors: %v", errs)
	}

	return nil
}

// GetMetrics returns comprehensive multi-level cache metrics
func (mlc *MultiLevelCache) GetMetrics() *MultiLevelCacheMetrics {
	metrics := &MultiLevelCacheMetrics{
		L1Hits:       atomic.LoadInt64(&mlc.metrics.l1Hits),
		L2Hits:       atomic.LoadInt64(&mlc.metrics.l2Hits),
		TotalMisses:  atomic.LoadInt64(&mlc.metrics.totalMisses),
		Promotions:   atomic.LoadInt64(&mlc.metrics.promotions),
		Demotions:    atomic.LoadInt64(&mlc.metrics.demotions),
		WriteBehinds: atomic.LoadInt64(&mlc.metrics.writeBehind),
	}

	// Get L1 metrics
	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		metrics.L1Metrics = mlc.l1Cache.GetMetrics()
	}

	// Get L2 metrics from collector
	if mlc.collector != nil {
		metrics.L2Metrics = mlc.collector.GetMetrics()
	}

	// Calculate computed metrics
	totalOperations := metrics.L1Hits + metrics.L2Hits + metrics.TotalMisses
	if totalOperations > 0 {
		metrics.OverallHitRate = float64(metrics.L1Hits+metrics.L2Hits) / float64(totalOperations) * 100
		metrics.L1HitRate = float64(metrics.L1Hits) / float64(totalOperations) * 100
		metrics.L2HitRate = float64(metrics.L2Hits) / float64(totalOperations) * 100
	}

	return metrics
}

// Close gracefully shuts down the multi-level cache
func (mlc *MultiLevelCache) Close() {
	if mlc.config.WriteStrategy == WriteStrategyBehind {
		close(mlc.stopWriter)
	}

	if mlc.l1Cache != nil {
		mlc.l1Cache.Close()
	}
}

// Read strategies implementation

func (mlc *MultiLevelCache) getFailover(ctx context.Context, key string) ([]byte, error) {
	// Try L1 first
	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if value, err := mlc.l1Cache.Get(ctx, key); err == nil && value != nil {
			atomic.AddInt64(&mlc.metrics.l1Hits, 1)
			mlc.recordHit("L1", key)
			return value, nil
		}
	}

	// Try L2 on L1 miss
	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if value, err := mlc.l2Cache.Get(ctx, key); err == nil && value != nil {
			atomic.AddInt64(&mlc.metrics.l2Hits, 1)
			mlc.recordHit("L2", key)
			return value, nil
		}
	}

	atomic.AddInt64(&mlc.metrics.totalMisses, 1)
	mlc.recordMiss(key)
	return nil, nil
}

func (mlc *MultiLevelCache) getWithPromotion(ctx context.Context, key string) ([]byte, error) {
	// Try L1 first
	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if value, err := mlc.l1Cache.Get(ctx, key); err == nil && value != nil {
			atomic.AddInt64(&mlc.metrics.l1Hits, 1)
			mlc.recordHit("L1", key)
			return value, nil
		}
	}

	// Try L2 and promote to L1 based on promotion ratio
	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if value, err := mlc.l2Cache.Get(ctx, key); err == nil && value != nil {
			atomic.AddInt64(&mlc.metrics.l2Hits, 1)
			mlc.recordHit("L2", key)

			// Promote to L1 based on promotion ratio
			if mlc.shouldPromote() && mlc.config.EnableL1 && mlc.l1Cache != nil {
				go mlc.promoteToL1(ctx, key, value)
			}

			return value, nil
		}
	}

	atomic.AddInt64(&mlc.metrics.totalMisses, 1)
	mlc.recordMiss(key)
	return nil, nil
}

func (mlc *MultiLevelCache) getWithReplication(ctx context.Context, key string) ([]byte, error) {
	// Try L1 first
	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if value, err := mlc.l1Cache.Get(ctx, key); err == nil && value != nil {
			atomic.AddInt64(&mlc.metrics.l1Hits, 1)
			mlc.recordHit("L1", key)
			return value, nil
		}
	}

	// Try L2 and always replicate to L1
	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if value, err := mlc.l2Cache.Get(ctx, key); err == nil && value != nil {
			atomic.AddInt64(&mlc.metrics.l2Hits, 1)
			mlc.recordHit("L2", key)

			// Always replicate to L1
			if mlc.config.EnableL1 && mlc.l1Cache != nil {
				go mlc.replicateToL1(ctx, key, value)
			}

			return value, nil
		}
	}

	atomic.AddInt64(&mlc.metrics.totalMisses, 1)
	mlc.recordMiss(key)
	return nil, nil
}

// Write strategies implementation

func (mlc *MultiLevelCache) setWriteThrough(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	var errs []error

	// Apply TTL jitter for cache avalanche protection
	l1TTL := mlc.addTTLJitter(ttl)
	l2TTL := mlc.addTTLJitter(ttl)

	// Write to L1
	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if err := mlc.l1Cache.Set(ctx, key, value, l1TTL); err != nil {
			errs = append(errs, fmt.Errorf("L1 set error: %w", err))
		}
	}

	// Write to L2
	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if err := mlc.l2Cache.Set(ctx, key, value, l2TTL); err != nil {
			errs = append(errs, fmt.Errorf("L2 set error: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("write-through errors: %v", errs)
	}

	mlc.recordSet(key, ttl)
	return nil
}

func (mlc *MultiLevelCache) setWriteBehind(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// Apply TTL jitter for cache avalanche protection
	l1TTL := mlc.addTTLJitter(ttl)
	l2TTL := mlc.addTTLJitter(ttl)

	// Write to L1 immediately
	if mlc.config.EnableL1 && mlc.l1Cache != nil {
		if err := mlc.l1Cache.Set(ctx, key, value, l1TTL); err != nil {
			return fmt.Errorf("L1 set error: %w", err)
		}
	}

	// Queue write to L2
	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		select {
		case mlc.writeQueue <- &writeOperation{
			key:       key,
			value:     value,
			ttl:       l2TTL,
			timestamp: time.Now(),
		}:
			atomic.AddInt64(&mlc.metrics.writeBehind, 1)
		default:
			// Queue is full, fallback to synchronous write with jitter
			if err := mlc.l2Cache.Set(ctx, key, value, l2TTL); err != nil {
				mlc.logger.Warn("Write-behind queue full, synchronous L2 write failed",
					logger.String("key", key),
					logger.ErrorField(err))
			}
		}
	}

	mlc.recordSet(key, ttl)
	return nil
}

func (mlc *MultiLevelCache) setWriteAround(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// Apply TTL jitter for cache avalanche protection
	l2TTL := mlc.addTTLJitter(ttl)

	// Write only to L2, bypass L1
	if mlc.config.EnableL2 && mlc.l2Cache != nil {
		if err := mlc.l2Cache.Set(ctx, key, value, l2TTL); err != nil {
			return fmt.Errorf("L2 set error: %w", err)
		}
	}

	mlc.recordSet(key, ttl)
	return nil
}

// Helper methods

// addTTLJitter adds random jitter to TTL to prevent cache avalanche
func (mlc *MultiLevelCache) addTTLJitter(ttl time.Duration) time.Duration {
	if !mlc.config.EnableTTLJitter || mlc.config.JitterPercent <= 0 {
		return ttl
	}

	// Calculate jitter range: ±(jitterPercent * ttl)
	jitterRange := time.Duration(float64(ttl) * mlc.config.JitterPercent)
	
	// Generate random jitter between -jitterRange and +jitterRange
	jitter := time.Duration(rand.Int63n(int64(2*jitterRange))) - jitterRange
	
	adjustedTTL := ttl + jitter
	
	// Ensure TTL doesn't go negative or below 1 second minimum
	if adjustedTTL < time.Second {
		adjustedTTL = time.Second
	}
	
	return adjustedTTL
}

func (mlc *MultiLevelCache) shouldPromote() bool {
	// Simple random-based promotion decision
	// In production, this could be more sophisticated (frequency-based, etc.)
	return mlc.config.PromotionRatio > 0 &&
		(mlc.config.PromotionRatio >= 1.0 ||
			time.Now().UnixNano()%100 < int64(mlc.config.PromotionRatio*100))
}

func (mlc *MultiLevelCache) promoteToL1(ctx context.Context, key string, value []byte) {
	// Apply TTL jitter to promotion as well to prevent avalanche on promoted data
	l1TTL := mlc.addTTLJitter(mlc.config.L1Config.DefaultTTL)
	
	if err := mlc.l1Cache.Set(ctx, key, value, l1TTL); err != nil {
		mlc.logger.Debug("Failed to promote key to L1",
			logger.String("key", key),
			logger.ErrorField(err))
	} else {
		atomic.AddInt64(&mlc.metrics.promotions, 1)
		mlc.logger.Debug("Promoted key to L1", logger.String("key", key))
	}
}

func (mlc *MultiLevelCache) replicateToL1(ctx context.Context, key string, value []byte) {
	// Apply TTL jitter to replication as well to prevent avalanche on replicated data
	l1TTL := mlc.addTTLJitter(mlc.config.L1Config.DefaultTTL)
	
	if err := mlc.l1Cache.Set(ctx, key, value, l1TTL); err != nil {
		mlc.logger.Debug("Failed to replicate key to L1",
			logger.String("key", key),
			logger.ErrorField(err))
	} else {
		mlc.logger.Debug("Replicated key to L1", logger.String("key", key))
	}
}

func (mlc *MultiLevelCache) startWriteBehindWorker() {
	go func() {
		for {
			select {
			case op := <-mlc.writeQueue:
				ctx := context.Background()
				if err := mlc.l2Cache.Set(ctx, op.key, op.value, op.ttl); err != nil {
					mlc.logger.Error("Write-behind operation failed",
						logger.String("key", op.key),
						logger.ErrorField(err))
				}

			case <-mlc.stopWriter:
				// Drain remaining operations
				for {
					select {
					case op := <-mlc.writeQueue:
						ctx := context.Background()
						_ = mlc.l2Cache.Set(ctx, op.key, op.value, op.ttl)
					default:
						return
					}
				}
			}
		}
	}()
}

func (mlc *MultiLevelCache) recordHit(level, key string) {
	if mlc.collector != nil {
		mlc.collector.RecordHit(context.Background(), fmt.Sprintf("%s:%s", level, key))
	}
}

func (mlc *MultiLevelCache) recordMiss(key string) {
	if mlc.collector != nil {
		mlc.collector.RecordMiss(context.Background(), key)
	}
}

func (mlc *MultiLevelCache) recordSet(key string, ttl time.Duration) {
	if mlc.collector != nil {
		mlc.collector.RecordSet(context.Background(), key, ttl)
	}
}
