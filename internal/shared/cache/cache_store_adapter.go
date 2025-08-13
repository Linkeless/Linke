package cache

import (
	"context"
	"encoding/json"
	"time"
)

// CacheStoreAdapter implements CacheStore interface using the existing Cache interface
// This adapter bridges between the event system and the cache system to avoid circular dependencies
type CacheStoreAdapter struct {
	cache Cache
}

// NewCacheStoreAdapter creates a new CacheStore adapter from a Cache implementation
func NewCacheStoreAdapter(cache Cache) CacheStore {
	return &CacheStoreAdapter{
		cache: cache,
	}
}

// Set stores a string value in the cache
func (a *CacheStoreAdapter) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	return a.cache.Set(ctx, key, []byte(value), expiration)
}

// Get retrieves a string value from the cache
func (a *CacheStoreAdapter) Get(ctx context.Context, key string) (string, error) {
	data, err := a.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", nil
	}
	return string(data), nil
}

// Delete removes a key from the cache
func (a *CacheStoreAdapter) Delete(ctx context.Context, key string) error {
	return a.cache.Delete(ctx, key)
}

// Exists checks if a key exists in the cache
func (a *CacheStoreAdapter) Exists(ctx context.Context, key string) (bool, error) {
	return a.cache.Exists(ctx, key)
}

// SetJSON stores a JSON-serializable value in the cache
func (a *CacheStoreAdapter) SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Op: "marshal", Key: key, Err: err}
	}
	return a.cache.Set(ctx, key, data, expiration)
}

// GetJSON retrieves and unmarshals a JSON value from the cache
func (a *CacheStoreAdapter) GetJSON(ctx context.Context, key string, dest any) error {
	data, err := a.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == nil {
		return nil
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return &CacheError{Op: "unmarshal", Key: key, Err: err}
	}
	return nil
}

// DeletePattern removes all keys matching a pattern
func (a *CacheStoreAdapter) DeletePattern(ctx context.Context, pattern string) error {
	return a.cache.DeleteByPattern(ctx, pattern)
}
