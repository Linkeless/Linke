package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// L2RedisCacheConfig configures the L2 Redis cache specifically for multi-level cache
type L2RedisCacheConfig struct {
	DefaultTTL       time.Duration `json:"default_ttl"`       // Default TTL set to 15 minutes
	CompressionType  string        `json:"compression_type"`  // Compression algorithm (gzip, lz4, none)
	EnableTagging    bool          `json:"enable_tagging"`    // Enable tag-based invalidation
	TagPrefix        string        `json:"tag_prefix"`        // Prefix for tag keys
	MaxRetries       int           `json:"max_retries"`       // Max retry attempts
	RetryDelay       time.Duration `json:"retry_delay"`       // Delay between retries
	EnableMetrics    bool          `json:"enable_metrics"`    // Enable detailed metrics
	KeyPrefix        string        `json:"key_prefix"`        // Prefix for all keys
	EnablePipeline   bool          `json:"enable_pipeline"`   // Enable pipeline for batch operations
}

// L2RedisCache represents the L2 (Redis) cache in a multi-level cache system
type L2RedisCache struct {
	*RedisCache
	config *L2RedisCacheConfig
	client *redis.Client
}

// NewL2RedisCache creates a new L2 Redis cache optimized for multi-level caching
func NewL2RedisCache(client *redis.Client, config *L2RedisCacheConfig) *L2RedisCache {
	if config == nil {
		config = &L2RedisCacheConfig{
			DefaultTTL:       15 * time.Minute, // 15 minutes as specified
			CompressionType:  "gzip",
			EnableTagging:    true,
			TagPrefix:        "tag:",
			MaxRetries:       3,
			RetryDelay:       100 * time.Millisecond,
			EnableMetrics:    true,
			KeyPrefix:        "l2:",
			EnablePipeline:   true,
		}
	}

	// Convert to base CacheConfig
	baseCacheConfig := &CacheConfig{
		DefaultTTL:      config.DefaultTTL,
		MaxRetries:      config.MaxRetries,
		RetryDelay:      config.RetryDelay,
		EnableMetrics:   config.EnableMetrics,
		CompressionType: config.CompressionType,
	}

	redisCache := NewRedisCache(client, baseCacheConfig)

	return &L2RedisCache{
		RedisCache: redisCache,
		config:     config,
		client:     client,
	}
}

// GetL2Config returns the L2-specific configuration
func (l2 *L2RedisCache) GetL2Config() *L2RedisCacheConfig {
	return l2.config
}

// SetWithTags sets a value with associated tags for later invalidation
func (l2 *L2RedisCache) SetWithTags(ctx context.Context, key string, value []byte, tags []string) error {
	prefixedKey := l2.prefixKey(key)
	
	// Set the main value with L2 TTL
	if err := l2.Set(ctx, prefixedKey, value, l2.config.DefaultTTL); err != nil {
		return err
	}
	
	// If tagging is disabled, we're done
	if !l2.config.EnableTagging || len(tags) == 0 {
		return nil
	}
	
	// Add key to each tag set
	if l2.config.EnablePipeline {
		return l2.setTagsPipeline(ctx, prefixedKey, tags)
	}
	
	return l2.setTagsSequential(ctx, prefixedKey, tags)
}

// setTagsPipeline uses Redis pipeline for efficient tag operations
func (l2 *L2RedisCache) setTagsPipeline(ctx context.Context, key string, tags []string) error {
	pipe := l2.client.Pipeline()
	
	for _, tag := range tags {
		tagKey := l2.tagKey(tag)
		pipe.SAdd(ctx, tagKey, key)
		pipe.Expire(ctx, tagKey, l2.config.DefaultTTL+time.Hour) // Tag TTL slightly longer
	}
	
	_, err := pipe.Exec(ctx)
	return err
}

// setTagsSequential sets tags sequentially (fallback)
func (l2 *L2RedisCache) setTagsSequential(ctx context.Context, key string, tags []string) error {
	for _, tag := range tags {
		tagKey := l2.tagKey(tag)
		if err := l2.client.SAdd(ctx, tagKey, key).Err(); err != nil {
			return err
		}
		if err := l2.client.Expire(ctx, tagKey, l2.config.DefaultTTL+time.Hour).Err(); err != nil {
			return err
		}
	}
	return nil
}

// InvalidateByTags removes all keys associated with the given tags
func (l2 *L2RedisCache) InvalidateByTags(ctx context.Context, tags []string) error {
	if !l2.config.EnableTagging {
		return fmt.Errorf("tagging is not enabled")
	}
	
	var allKeys []string
	
	// Collect all keys for each tag
	for _, tag := range tags {
		tagKey := l2.tagKey(tag)
		keys, err := l2.client.SMembers(ctx, tagKey).Result()
		if err != nil {
			continue // Skip if tag doesn't exist
		}
		allKeys = append(allKeys, keys...)
	}
	
	if len(allKeys) == 0 {
		return nil
	}
	
	// Remove duplicates
	keySet := make(map[string]struct{})
	uniqueKeys := make([]string, 0, len(allKeys))
	for _, key := range allKeys {
		if _, exists := keySet[key]; !exists {
			keySet[key] = struct{}{}
			uniqueKeys = append(uniqueKeys, key)
		}
	}
	
	// Delete keys in batches
	return l2.deleteKeysBatch(ctx, uniqueKeys, tags)
}

// deleteKeysBatch deletes keys and cleans up tags in batches
func (l2 *L2RedisCache) deleteKeysBatch(ctx context.Context, keys []string, tags []string) error {
	const batchSize = 100
	
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		
		batch := keys[i:end]
		
		pipe := l2.client.Pipeline()
		
		// Delete the actual cache keys
		for _, key := range batch {
			pipe.Del(ctx, key)
		}
		
		// Remove keys from tag sets
		for _, tag := range tags {
			tagKey := l2.tagKey(tag)
			for _, key := range batch {
				pipe.SRem(ctx, tagKey, key)
			}
		}
		
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}
	
	return nil
}

// GetWithL2TTL gets a value and resets TTL to L2 default
func (l2 *L2RedisCache) GetWithL2TTL(ctx context.Context, key string) ([]byte, error) {
	prefixedKey := l2.prefixKey(key)
	
	value, err := l2.Get(ctx, prefixedKey)
	if err != nil || value == nil {
		return value, err
	}
	
	// Reset TTL to extend cache life
	_ = l2.Expire(ctx, prefixedKey, l2.config.DefaultTTL)
	
	return value, nil
}

// SetWithL2TTL sets a value with L2-optimized TTL
func (l2 *L2RedisCache) SetWithL2TTL(ctx context.Context, key string, value []byte) error {
	prefixedKey := l2.prefixKey(key)
	return l2.Set(ctx, prefixedKey, value, l2.config.DefaultTTL)
}

// GetL2Stats returns L2-specific statistics
func (l2 *L2RedisCache) GetL2Stats(ctx context.Context) (*L2Stats, error) {
	info, err := l2.client.Info(ctx, "memory", "stats").Result()
	if err != nil {
		return nil, err
	}
	
	stats := &L2Stats{
		ConnectedClients: l2.parseRedisInfo(info, "connected_clients"),
		UsedMemory:       l2.parseRedisInfo(info, "used_memory"),
		TotalCommands:    l2.parseRedisInfo(info, "total_commands_processed"),
		KeyspaceHits:     l2.parseRedisInfo(info, "keyspace_hits"),
		KeyspaceMisses:   l2.parseRedisInfo(info, "keyspace_misses"),
	}
	
	if stats.KeyspaceHits+stats.KeyspaceMisses > 0 {
		stats.HitRate = float64(stats.KeyspaceHits) / float64(stats.KeyspaceHits+stats.KeyspaceMisses) * 100
	}
	
	return stats, nil
}

// L2Stats represents L2 cache statistics
type L2Stats struct {
	ConnectedClients int64   `json:"connected_clients"`
	UsedMemory       int64   `json:"used_memory"`
	TotalCommands    int64   `json:"total_commands"`
	KeyspaceHits     int64   `json:"keyspace_hits"`
	KeyspaceMisses   int64   `json:"keyspace_misses"`
	HitRate          float64 `json:"hit_rate"`
}

// ListTags returns all available tags
func (l2 *L2RedisCache) ListTags(ctx context.Context) ([]string, error) {
	if !l2.config.EnableTagging {
		return nil, fmt.Errorf("tagging is not enabled")
	}
	
	pattern := l2.config.TagPrefix + "*"
	keys, err := l2.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	
	tags := make([]string, len(keys))
	for i, key := range keys {
		tags[i] = strings.TrimPrefix(key, l2.config.TagPrefix)
	}
	
	return tags, nil
}

// GetTagKeys returns all keys associated with a tag
func (l2 *L2RedisCache) GetTagKeys(ctx context.Context, tag string) ([]string, error) {
	if !l2.config.EnableTagging {
		return nil, fmt.Errorf("tagging is not enabled")
	}
	
	tagKey := l2.tagKey(tag)
	return l2.client.SMembers(ctx, tagKey).Result()
}

// Helper methods

func (l2 *L2RedisCache) prefixKey(key string) string {
	if l2.config.KeyPrefix == "" {
		return key
	}
	return l2.config.KeyPrefix + key
}

func (l2 *L2RedisCache) tagKey(tag string) string {
	return l2.config.TagPrefix + tag
}

func (l2 *L2RedisCache) parseRedisInfo(info, key string) int64 {
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, key+":") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				var value int64
				if _, err := fmt.Sscanf(parts[1], "%d", &value); err == nil {
					return value
				}
			}
		}
	}
	return 0
}

// DefaultL2RedisCacheConfig returns the default L2 Redis cache configuration
func DefaultL2RedisCacheConfig() *L2RedisCacheConfig {
	return &L2RedisCacheConfig{
		DefaultTTL:       15 * time.Minute,
		CompressionType:  "gzip",
		EnableTagging:    true,
		TagPrefix:        "tag:",
		MaxRetries:       3,
		RetryDelay:       100 * time.Millisecond,
		EnableMetrics:    true,
		KeyPrefix:        "l2:",
		EnablePipeline:   true,
	}
}

// BulkSet sets multiple key-value pairs efficiently
func (l2 *L2RedisCache) BulkSet(ctx context.Context, items map[string][]byte, tags map[string][]string) error {
	if !l2.config.EnablePipeline {
		return l2.bulkSetSequential(ctx, items, tags)
	}
	
	pipe := l2.client.Pipeline()
	
	// Set all values
	for key, value := range items {
		prefixedKey := l2.prefixKey(key)
		pipe.Set(ctx, prefixedKey, value, l2.config.DefaultTTL)
		
		// Add tags if tagging is enabled
		if l2.config.EnableTagging {
			if keyTags, exists := tags[key]; exists {
				for _, tag := range keyTags {
					tagKey := l2.tagKey(tag)
					pipe.SAdd(ctx, tagKey, prefixedKey)
					pipe.Expire(ctx, tagKey, l2.config.DefaultTTL+time.Hour)
				}
			}
		}
	}
	
	_, err := pipe.Exec(ctx)
	return err
}

func (l2 *L2RedisCache) bulkSetSequential(ctx context.Context, items map[string][]byte, tags map[string][]string) error {
	for key, value := range items {
		keyTags := tags[key]
		if err := l2.SetWithTags(ctx, key, value, keyTags); err != nil {
			return err
		}
	}
	return nil
}