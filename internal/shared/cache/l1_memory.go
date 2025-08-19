package cache

import (
	"context"
	"time"
)

// L1MemoryCacheConfig configures the L1 memory cache specifically for multi-level cache
type L1MemoryCacheConfig struct {
	MaxSize         int           `json:"max_size"`         // Maximum number of entries
	DefaultTTL      time.Duration `json:"default_ttl"`      // Default TTL set to 1 minute
	CleanupInterval time.Duration `json:"cleanup_interval"` // How often to clean expired entries
	EnableMetrics   bool          `json:"enable_metrics"`   // Enable detailed metrics collection
}

// L1MemoryCache represents the L1 (memory) cache in a multi-level cache system
type L1MemoryCache struct {
	*MemoryCache
	config *L1MemoryCacheConfig
}

// NewL1MemoryCache creates a new L1 memory cache optimized for multi-level caching
func NewL1MemoryCache(config *L1MemoryCacheConfig) *L1MemoryCache {
	if config == nil {
		config = &L1MemoryCacheConfig{
			MaxSize:         1000,      // Reasonable default for L1
			DefaultTTL:      time.Minute, // 1 minute as specified
			CleanupInterval: 30 * time.Second, // More frequent cleanup for L1
			EnableMetrics:   true,
		}
	}

	// Convert to MemoryCacheConfig
	memoryCacheConfig := &MemoryCacheConfig{
		MaxSize:         config.MaxSize,
		DefaultTTL:      config.DefaultTTL,
		EvictionPolicy:  EvictionPolicyLRU, // Always use LRU for L1
		CleanupInterval: config.CleanupInterval,
	}

	memoryCache := NewMemoryCache(memoryCacheConfig)

	return &L1MemoryCache{
		MemoryCache: memoryCache,
		config:      config,
	}
}

// GetL1Config returns the L1-specific configuration
func (l1 *L1MemoryCache) GetL1Config() *L1MemoryCacheConfig {
	return l1.config
}

// IsL1Suitable checks if an entry is suitable for L1 caching based on size and access patterns
func (l1 *L1MemoryCache) IsL1Suitable(ctx context.Context, key string, value []byte) bool {
	// Small entries are preferred in L1
	if len(value) > 1024*10 { // 10KB threshold
		return false
	}
	
	// Check if we have room
	l1.MemoryCache.mu.RLock()
	hasRoom := len(l1.MemoryCache.entries) < l1.config.MaxSize
	l1.MemoryCache.mu.RUnlock()
	
	return hasRoom
}

// PromoteFromL2 promotes an entry from L2 to L1 cache
func (l1 *L1MemoryCache) PromoteFromL2(ctx context.Context, key string, value []byte) error {
	if !l1.IsL1Suitable(ctx, key, value) {
		return nil // Don't promote if not suitable
	}
	
	return l1.Set(ctx, key, value, l1.config.DefaultTTL)
}

// GetWithPromotion gets a value and tracks it for potential promotion
func (l1 *L1MemoryCache) GetWithPromotion(ctx context.Context, key string) ([]byte, bool, error) {
	value, err := l1.Get(ctx, key)
	if err != nil {
		return nil, false, err
	}
	
	found := value != nil
	return value, found, nil
}

// SetWithL1TTL sets a value with L1-optimized TTL
func (l1 *L1MemoryCache) SetWithL1TTL(ctx context.Context, key string, value []byte) error {
	return l1.Set(ctx, key, value, l1.config.DefaultTTL)
}

// GetL1Stats returns L1-specific statistics
func (l1 *L1MemoryCache) GetL1Stats() L1Stats {
	metrics := l1.MemoryCache.GetMetrics()
	
	// Access private fields via getter pattern
	metrics.mu.RLock()
	hits := metrics.hits
	misses := metrics.misses
	evictions := metrics.evictions
	entryCount := metrics.entryCount
	currentSize := metrics.currentSize
	maxSize := metrics.maxSize
	metrics.mu.RUnlock()
	
	// Calculate hit rate
	var hitRate float64
	if hits+misses > 0 {
		hitRate = float64(hits) / float64(hits+misses) * 100
	}
	
	return L1Stats{
		Hits:        hits,
		Misses:      misses,
		Evictions:   evictions,
		EntryCount:  entryCount,
		CurrentSize: currentSize,
		MaxSize:     maxSize,
		HitRate:     hitRate,
		Capacity:    float64(entryCount) / float64(maxSize) * 100,
	}
}

// L1Stats represents L1 cache statistics
type L1Stats struct {
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Evictions   int64   `json:"evictions"`
	EntryCount  int64   `json:"entry_count"`
	CurrentSize int64   `json:"current_size"`
	MaxSize     int64   `json:"max_size"`
	HitRate     float64 `json:"hit_rate"`
	Capacity    float64 `json:"capacity_percentage"`
}

// WarmupL1 preloads frequently accessed keys into L1 cache
func (l1 *L1MemoryCache) WarmupL1(ctx context.Context, keys []string, fetcher func(string) ([]byte, error)) error {
	for _, key := range keys {
		if value, err := fetcher(key); err == nil && value != nil {
			if l1.IsL1Suitable(ctx, key, value) {
				_ = l1.SetWithL1TTL(ctx, key, value)
			}
		}
	}
	return nil
}

// ClearL1 clears the L1 cache while preserving configuration
func (l1 *L1MemoryCache) ClearL1(ctx context.Context) error {
	return l1.Flush(ctx)
}

// GetL1Capacity returns current capacity utilization
func (l1 *L1MemoryCache) GetL1Capacity() float64 {
	l1.MemoryCache.mu.RLock()
	defer l1.MemoryCache.mu.RUnlock()
	
	if l1.config.MaxSize == 0 {
		return 0
	}
	
	return float64(len(l1.MemoryCache.entries)) / float64(l1.config.MaxSize) * 100
}

// DefaultL1MemoryCacheConfig returns the default L1 memory cache configuration
func DefaultL1MemoryCacheConfig() *L1MemoryCacheConfig {
	return &L1MemoryCacheConfig{
		MaxSize:         1000,
		DefaultTTL:      time.Minute,
		CleanupInterval: 30 * time.Second,
		EnableMetrics:   true,
	}
}