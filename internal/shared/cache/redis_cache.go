package cache

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	config *CacheConfig
}

func NewRedisCache(client *redis.Client, config *CacheConfig) *RedisCache {
	if config == nil {
		config = &CacheConfig{
			DefaultTTL:      DefaultCacheTTL,
			MaxRetries:      3,
			RetryDelay:      100 * time.Millisecond,
			EnableMetrics:   false,
			CompressionType: "gzip",
		}
	}

	return &RedisCache{
		client: client,
		config: config,
	}
}

func (rc *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, &CacheError{Op: "get", Key: key, Err: err}
	}

	if rc.config.CompressionType == "gzip" && len(val) > 0 {
		decompressed, err := rc.decompress(val)
		if err != nil {
			return val, nil
		}
		return decompressed, nil
	}

	return val, nil
}

func (rc *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = rc.config.DefaultTTL
	}

	data := value
	if rc.config.CompressionType == "gzip" && len(value) > 1024 {
		compressed, err := rc.compress(value)
		if err == nil && len(compressed) < len(value) {
			data = compressed
		}
	}

	err := rc.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return &CacheError{Op: "set", Key: key, Err: err}
	}

	return nil
}

func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	err := rc.client.Del(ctx, key).Err()
	if err != nil {
		return &CacheError{Op: "delete", Key: key, Err: err}
	}
	return nil
}

func (rc *RedisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var keys []string
	
	for {
		var batch []string
		var err error
		batch, cursor, err = rc.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return &CacheError{Op: "scan", Pattern: pattern, Err: err}
		}
		
		keys = append(keys, batch...)
		
		if cursor == 0 {
			break
		}
	}
	
	if len(keys) == 0 {
		return nil
	}
	
	pipe := rc.client.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return &CacheError{Op: "delete", Pattern: pattern, Err: err}
	}
	
	return nil
}

func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, &CacheError{Op: "exists", Key: key, Err: err}
	}
	return count > 0, nil
}

func (rc *RedisCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	err := rc.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return &CacheError{Op: "expire", Key: key, Err: err}
	}
	return nil
}

func (rc *RedisCache) Flush(ctx context.Context) error {
	err := rc.client.FlushDB(ctx).Err()
	if err != nil {
		return &CacheError{Op: "flush", Key: "*", Err: err}
	}
	return nil
}

func (rc *RedisCache) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	
	if err := gz.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (rc *RedisCache) decompress(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()
	
	return io.ReadAll(r)
}

type TypedRedisCache[T any] struct {
	cache  *RedisCache
	prefix string
}

func NewTypedRedisCache[T any](cache *RedisCache, prefix string) *TypedRedisCache[T] {
	return &TypedRedisCache[T]{
		cache:  cache,
		prefix: prefix,
	}
}

func (tc *TypedRedisCache[T]) Get(ctx context.Context, key string) (*T, error) {
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

func (tc *TypedRedisCache[T]) Set(ctx context.Context, key string, value *T, ttl time.Duration) error {
	fullKey := tc.prefix + key
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Op: "marshal", Key: fullKey, Err: err}
	}
	
	return tc.cache.Set(ctx, fullKey, data, ttl)
}

func (tc *TypedRedisCache[T]) Delete(ctx context.Context, key string) error {
	fullKey := tc.prefix + key
	return tc.cache.Delete(ctx, fullKey)
}

func (tc *TypedRedisCache[T]) DeleteByPattern(ctx context.Context, pattern string) error {
	fullPattern := tc.prefix + pattern
	return tc.cache.DeleteByPattern(ctx, fullPattern)
}

func (tc *TypedRedisCache[T]) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := tc.prefix + key
	return tc.cache.Exists(ctx, fullKey)
}

type RedisCacheManager struct {
	cache       *RedisCache
	typedCaches map[string]TypedCache[any]
	tags        *RedisCacheTags
}

func NewRedisCacheManager(client *redis.Client, config *CacheConfig) *RedisCacheManager {
	cache := NewRedisCache(client, config)
	return &RedisCacheManager{
		cache:       cache,
		typedCaches: make(map[string]TypedCache[any]),
		tags:        NewRedisCacheTags(client),
	}
}

func (rcm *RedisCacheManager) GetCache() Cache {
	return rcm.cache
}

func (rcm *RedisCacheManager) GetTypedCache(prefix string) TypedCache[any] {
	if tc, exists := rcm.typedCaches[prefix]; exists {
		return tc
	}
	
	tc := NewTypedRedisCache[any](rcm.cache, prefix)
	rcm.typedCaches[prefix] = tc
	return tc
}

func (rcm *RedisCacheManager) InvalidateCache(ctx context.Context, patterns ...string) error {
	for _, pattern := range patterns {
		if err := rcm.cache.DeleteByPattern(ctx, pattern); err != nil {
			return err
		}
	}
	return nil
}

func (rcm *RedisCacheManager) GetStats(ctx context.Context) (*CacheStats, error) {
	_, err := rcm.cache.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, &CacheError{Op: "stats", Key: "*", Err: err}
	}
	
	dbSize, err := rcm.cache.client.DBSize(ctx).Result()
	if err != nil {
		return nil, &CacheError{Op: "dbsize", Key: "*", Err: err}
	}
	
	stats := &CacheStats{
		TotalKeys: dbSize,
	}
	
	return stats, nil
}

type RedisCacheTags struct {
	client *redis.Client
}

func NewRedisCacheTags(client *redis.Client) *RedisCacheTags {
	return &RedisCacheTags{client: client}
}

func (rct *RedisCacheTags) AddToTag(ctx context.Context, tag string, keys ...string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)
	members := make([]any, len(keys))
	for i, key := range keys {
		members[i] = key
	}
	
	return rct.client.SAdd(ctx, tagKey, members...).Err()
}

func (rct *RedisCacheTags) GetByTag(ctx context.Context, tag string) ([]string, error) {
	tagKey := fmt.Sprintf("tag:%s", tag)
	return rct.client.SMembers(ctx, tagKey).Result()
}

func (rct *RedisCacheTags) InvalidateTag(ctx context.Context, tag string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)
	
	keys, err := rct.GetByTag(ctx, tag)
	if err != nil {
		return err
	}
	
	if len(keys) == 0 {
		return nil
	}
	
	pipe := rct.client.Pipeline()
	for _, key := range keys {
		pipe.Del(ctx, key)
	}
	pipe.Del(ctx, tagKey)
	
	_, err = pipe.Exec(ctx)
	return err
}

func (rct *RedisCacheTags) RemoveFromTag(ctx context.Context, tag string, keys ...string) error {
	tagKey := fmt.Sprintf("tag:%s", tag)
	members := make([]any, len(keys))
	for i, key := range keys {
		members[i] = key
	}
	
	return rct.client.SRem(ctx, tagKey, members...).Err()
}