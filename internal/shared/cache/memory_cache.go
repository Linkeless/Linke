package cache

import (
	"context"
	"sync"
	"time"
)

// EvictionPolicy defines the eviction strategy for memory cache
type EvictionPolicy string

const (
	EvictionPolicyLRU EvictionPolicy = "lru"
	EvictionPolicyLFU EvictionPolicy = "lfu"
	EvictionPolicyTTL EvictionPolicy = "ttl"
)

// MemoryCacheConfig configures the memory cache
type MemoryCacheConfig struct {
	MaxSize         int            `json:"max_size"`
	DefaultTTL      time.Duration  `json:"default_ttl"`
	EvictionPolicy  EvictionPolicy `json:"eviction_policy"`
	CleanupInterval time.Duration  `json:"cleanup_interval"`
}

// MemoryCacheEntry represents a cache entry with metadata
type MemoryCacheEntry struct {
	Value       []byte
	CreatedAt   time.Time
	ExpiresAt   time.Time
	AccessedAt  time.Time
	AccessCount int64
	Size        int64
}

// IsExpired checks if the entry has expired
func (e *MemoryCacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// Touch updates the access time and count
func (e *MemoryCacheEntry) Touch() {
	e.AccessedAt = time.Now()
	e.AccessCount++
}

// MemoryCache implements an in-memory cache with configurable eviction policies
type MemoryCache struct {
	mu      sync.RWMutex
	config  *MemoryCacheConfig
	entries map[string]*MemoryCacheEntry

	// LRU tracking
	lruHead   *lruNode
	lruTail   *lruNode
	keyToNode map[string]*lruNode

	// Stats
	currentSize int64
	stopCleanup chan struct{}
	metrics     *MemoryCacheMetrics
}

type lruNode struct {
	key  string
	prev *lruNode
	next *lruNode
}

// MemoryCacheMetrics tracks memory cache performance
type MemoryCacheMetrics struct {
	mu          sync.RWMutex
	hits        int64
	misses      int64
	evictions   int64
	expired     int64
	currentSize int64
	maxSize     int64
	entryCount  int64
}

// NewMemoryCache creates a new memory cache instance
func NewMemoryCache(config *MemoryCacheConfig) *MemoryCache {
	if config == nil {
		config = &MemoryCacheConfig{
			MaxSize:         1000,
			DefaultTTL:      5 * time.Minute,
			EvictionPolicy:  EvictionPolicyLRU,
			CleanupInterval: 1 * time.Minute,
		}
	}

	mc := &MemoryCache{
		config:      config,
		entries:     make(map[string]*MemoryCacheEntry),
		keyToNode:   make(map[string]*lruNode),
		stopCleanup: make(chan struct{}),
		metrics:     &MemoryCacheMetrics{maxSize: int64(config.MaxSize)},
	}

	// Initialize LRU list
	mc.lruHead = &lruNode{}
	mc.lruTail = &lruNode{}
	mc.lruHead.next = mc.lruTail
	mc.lruTail.prev = mc.lruHead

	// Start background cleanup
	mc.startCleanup()

	return mc
}

// Get retrieves a value from the memory cache
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	mc.mu.RLock()
	entry, exists := mc.entries[key]
	mc.mu.RUnlock()

	if !exists {
		mc.updateMetrics(func(m *MemoryCacheMetrics) {
			m.misses++
		})
		return nil, nil
	}

	if entry.IsExpired() {
		mc.mu.Lock()
		mc.removeEntry(key)
		mc.mu.Unlock()
		mc.updateMetrics(func(m *MemoryCacheMetrics) {
			m.misses++
			m.expired++
		})
		return nil, nil
	}

	// Update access information
	mc.mu.Lock()
	entry.Touch()
	if mc.config.EvictionPolicy == EvictionPolicyLRU {
		mc.moveToFront(key)
	}
	value := make([]byte, len(entry.Value))
	copy(value, entry.Value)
	mc.mu.Unlock()

	mc.updateMetrics(func(m *MemoryCacheMetrics) {
		m.hits++
	})

	return value, nil
}

// Set stores a value in the memory cache
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = mc.config.DefaultTTL
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if we need to evict entries
	entrySize := int64(len(value))
	for len(mc.entries) >= mc.config.MaxSize && !mc.hasSpace(entrySize) {
		mc.evictEntry()
	}

	entry := &MemoryCacheEntry{
		Value:       make([]byte, len(value)),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		AccessedAt:  time.Now(),
		AccessCount: 1,
		Size:        entrySize,
	}
	copy(entry.Value, value)

	// Remove existing entry if it exists
	if _, exists := mc.entries[key]; exists {
		mc.removeEntry(key)
	}

	mc.entries[key] = entry
	mc.currentSize += entrySize

	// Update LRU tracking
	if mc.config.EvictionPolicy == EvictionPolicyLRU {
		mc.addToFront(key)
	}

	mc.updateMetrics(func(m *MemoryCacheMetrics) {
		m.currentSize = mc.currentSize
		m.entryCount = int64(len(mc.entries))
	})

	return nil
}

// Delete removes a value from the memory cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.removeEntry(key)
	return nil
}

// DeleteByPattern removes entries matching a pattern
func (mc *MemoryCache) DeleteByPattern(ctx context.Context, pattern string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	keysToDelete := make([]string, 0)
	for key := range mc.entries {
		if matchPattern(key, pattern) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		mc.removeEntry(key)
	}

	return nil
}

// Exists checks if a key exists in the cache
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	entry, exists := mc.entries[key]
	mc.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if entry.IsExpired() {
		mc.mu.Lock()
		mc.removeEntry(key)
		mc.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// Expire is not supported in memory cache (TTL is set on creation)
func (mc *MemoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if entry, exists := mc.entries[key]; exists {
		entry.ExpiresAt = time.Now().Add(ttl)
	}

	return nil
}

// Flush clears all entries from the memory cache
func (mc *MemoryCache) Flush(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.entries = make(map[string]*MemoryCacheEntry)
	mc.keyToNode = make(map[string]*lruNode)
	mc.lruHead.next = mc.lruTail
	mc.lruTail.prev = mc.lruHead
	mc.currentSize = 0

	mc.updateMetrics(func(m *MemoryCacheMetrics) {
		m.currentSize = 0
		m.entryCount = 0
	})

	return nil
}

// GetMetrics returns current cache metrics
func (mc *MemoryCache) GetMetrics() *MemoryCacheMetrics {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	return &MemoryCacheMetrics{
		hits:        mc.metrics.hits,
		misses:      mc.metrics.misses,
		evictions:   mc.metrics.evictions,
		expired:     mc.metrics.expired,
		currentSize: mc.metrics.currentSize,
		maxSize:     mc.metrics.maxSize,
		entryCount:  mc.metrics.entryCount,
	}
}

// Close stops the background cleanup goroutine
func (mc *MemoryCache) Close() {
	close(mc.stopCleanup)
}

// Private methods

func (mc *MemoryCache) removeEntry(key string) {
	if entry, exists := mc.entries[key]; exists {
		delete(mc.entries, key)
		mc.currentSize -= entry.Size

		// Update LRU tracking
		if mc.config.EvictionPolicy == EvictionPolicyLRU {
			mc.removeFromLRU(key)
		}
	}
}

func (mc *MemoryCache) evictEntry() {
	var keyToEvict string

	switch mc.config.EvictionPolicy {
	case EvictionPolicyLRU:
		keyToEvict = mc.getLRUKey()
	case EvictionPolicyLFU:
		keyToEvict = mc.getLFUKey()
	case EvictionPolicyTTL:
		keyToEvict = mc.getTTLKey()
	default:
		keyToEvict = mc.getLRUKey()
	}

	if keyToEvict != "" {
		mc.removeEntry(keyToEvict)
		mc.updateMetrics(func(m *MemoryCacheMetrics) {
			m.evictions++
		})
	}
}

func (mc *MemoryCache) hasSpace(requiredSize int64) bool {
	return mc.currentSize+requiredSize <= int64(mc.config.MaxSize)*1024*1024 // Assume MaxSize is in MB
}

func (mc *MemoryCache) getLRUKey() string {
	if mc.lruTail.prev != mc.lruHead {
		return mc.lruTail.prev.key
	}
	return ""
}

func (mc *MemoryCache) getLFUKey() string {
	var minKey string
	var minCount int64 = -1

	for key, entry := range mc.entries {
		if minCount == -1 || entry.AccessCount < minCount {
			minCount = entry.AccessCount
			minKey = key
		}
	}

	return minKey
}

func (mc *MemoryCache) getTTLKey() string {
	var earliestKey string
	var earliestTime time.Time

	for key, entry := range mc.entries {
		if earliestTime.IsZero() || entry.ExpiresAt.Before(earliestTime) {
			earliestTime = entry.ExpiresAt
			earliestKey = key
		}
	}

	return earliestKey
}

// LRU list operations

func (mc *MemoryCache) addToFront(key string) {
	node := &lruNode{key: key}
	node.next = mc.lruHead.next
	node.prev = mc.lruHead
	mc.lruHead.next.prev = node
	mc.lruHead.next = node
	mc.keyToNode[key] = node
}

func (mc *MemoryCache) moveToFront(key string) {
	if node, exists := mc.keyToNode[key]; exists {
		// Remove from current position
		node.prev.next = node.next
		node.next.prev = node.prev

		// Add to front
		node.next = mc.lruHead.next
		node.prev = mc.lruHead
		mc.lruHead.next.prev = node
		mc.lruHead.next = node
	}
}

func (mc *MemoryCache) removeFromLRU(key string) {
	if node, exists := mc.keyToNode[key]; exists {
		node.prev.next = node.next
		node.next.prev = node.prev
		delete(mc.keyToNode, key)
	}
}

func (mc *MemoryCache) updateMetrics(updateFunc func(*MemoryCacheMetrics)) {
	mc.metrics.mu.Lock()
	updateFunc(mc.metrics)
	mc.metrics.mu.Unlock()
}

func (mc *MemoryCache) startCleanup() {
	go func() {
		ticker := time.NewTicker(mc.config.CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mc.cleanupExpired()
			case <-mc.stopCleanup:
				return
			}
		}
	}()
}

func (mc *MemoryCache) cleanupExpired() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	keysToDelete := make([]string, 0)
	for key, entry := range mc.entries {
		if entry.IsExpired() {
			keysToDelete = append(keysToDelete, key)
		}
	}

	expiredCount := int64(len(keysToDelete))
	for _, key := range keysToDelete {
		mc.removeEntry(key)
	}

	if expiredCount > 0 {
		mc.updateMetrics(func(m *MemoryCacheMetrics) {
			m.expired += expiredCount
			m.currentSize = mc.currentSize
			m.entryCount = int64(len(mc.entries))
		})
	}
}

// Helper function for pattern matching
func matchPattern(key, pattern string) bool {
	// Simple wildcard pattern matching
	// Support for * at the end of pattern
	if pattern == "*" {
		return true
	}

	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}

	return key == pattern
}
