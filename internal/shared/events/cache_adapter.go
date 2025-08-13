package events

import (
	"context"
	"time"
)

// CacheStoreAdapter adapts any implementation with the required methods to EventCacheStore
// This breaks the circular dependency by avoiding direct import of cache package
type CacheStoreAdapter struct {
	setFunc           func(ctx context.Context, key string, value string, expiration time.Duration) error
	getFunc           func(ctx context.Context, key string) (string, error)
	deleteFunc        func(ctx context.Context, key string) error
	existsFunc        func(ctx context.Context, key string) (bool, error)
	setJSONFunc       func(ctx context.Context, key string, value any, expiration time.Duration) error
	getJSONFunc       func(ctx context.Context, key string, dest any) error
	deletePatternFunc func(ctx context.Context, pattern string) error
}

// NewCacheStoreAdapter creates a new cache store adapter with function implementations
func NewCacheStoreAdapter(
	setFunc func(ctx context.Context, key string, value string, expiration time.Duration) error,
	getFunc func(ctx context.Context, key string) (string, error),
	deleteFunc func(ctx context.Context, key string) error,
	existsFunc func(ctx context.Context, key string) (bool, error),
	setJSONFunc func(ctx context.Context, key string, value any, expiration time.Duration) error,
	getJSONFunc func(ctx context.Context, key string, dest any) error,
	deletePatternFunc func(ctx context.Context, pattern string) error,
) EventCacheStore {
	return &CacheStoreAdapter{
		setFunc:           setFunc,
		getFunc:           getFunc,
		deleteFunc:        deleteFunc,
		existsFunc:        existsFunc,
		setJSONFunc:       setJSONFunc,
		getJSONFunc:       getJSONFunc,
		deletePatternFunc: deletePatternFunc,
	}
}

func (a *CacheStoreAdapter) Set(ctx context.Context, key string, value string, expiration time.Duration) error {
	if a.setFunc == nil {
		return nil // No-op if function not provided
	}
	return a.setFunc(ctx, key, value, expiration)
}

func (a *CacheStoreAdapter) Get(ctx context.Context, key string) (string, error) {
	if a.getFunc == nil {
		return "", nil // No-op if function not provided
	}
	return a.getFunc(ctx, key)
}

func (a *CacheStoreAdapter) Delete(ctx context.Context, key string) error {
	if a.deleteFunc == nil {
		return nil // No-op if function not provided
	}
	return a.deleteFunc(ctx, key)
}

func (a *CacheStoreAdapter) Exists(ctx context.Context, key string) (bool, error) {
	if a.existsFunc == nil {
		return false, nil // No-op if function not provided
	}
	return a.existsFunc(ctx, key)
}

func (a *CacheStoreAdapter) SetJSON(ctx context.Context, key string, value any, expiration time.Duration) error {
	if a.setJSONFunc == nil {
		return nil // No-op if function not provided
	}
	return a.setJSONFunc(ctx, key, value, expiration)
}

func (a *CacheStoreAdapter) GetJSON(ctx context.Context, key string, dest any) error {
	if a.getJSONFunc == nil {
		return nil // No-op if function not provided
	}
	return a.getJSONFunc(ctx, key, dest)
}

func (a *CacheStoreAdapter) DeletePattern(ctx context.Context, pattern string) error {
	if a.deletePatternFunc == nil {
		return nil // No-op if function not provided
	}
	return a.deletePatternFunc(ctx, pattern)
}
