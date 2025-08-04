package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

type CacheDecorator struct {
	cache     Cache
	keyPrefix string
	ttl       time.Duration
}

func NewCacheDecorator(cache Cache, keyPrefix string, ttl time.Duration) *CacheDecorator {
	return &CacheDecorator{
		cache:     cache,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}
}

func (cd *CacheDecorator) Get(ctx context.Context, key string, fetch func() (any, error)) (any, error) {
	cacheKey := cd.keyPrefix + key

	cached, err := cd.cache.Get(ctx, cacheKey)
	if err != nil {
		return fetch()
	}

	if cached != nil {
		var result any
		if err := json.Unmarshal(cached, &result); err == nil {
			return result, nil
		}
	}

	result, err := fetch()
	if err != nil {
		return nil, err
	}

	if result != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = cd.cache.Set(ctx, cacheKey, data, cd.ttl)
		}
	}

	return result, nil
}

func (cd *CacheDecorator) Invalidate(ctx context.Context, key string) error {
	cacheKey := cd.keyPrefix + key
	return cd.cache.Delete(ctx, cacheKey)
}

func (cd *CacheDecorator) InvalidatePattern(ctx context.Context, pattern string) error {
	fullPattern := cd.keyPrefix + pattern
	return cd.cache.DeleteByPattern(ctx, fullPattern)
}

type CacheAside[T any] struct {
	cache     Cache
	keyFunc   func(T) string
	ttl       time.Duration
	keyPrefix string
}

func NewCacheAside[T any](cache Cache, keyPrefix string, keyFunc func(T) string, ttl time.Duration) *CacheAside[T] {
	return &CacheAside[T]{
		cache:     cache,
		keyFunc:   keyFunc,
		ttl:       ttl,
		keyPrefix: keyPrefix,
	}
}

func (ca *CacheAside[T]) Get(ctx context.Context, key string, fetch func() (*T, error)) (*T, error) {
	cacheKey := ca.keyPrefix + key

	cached, err := ca.cache.Get(ctx, cacheKey)
	if err != nil {
		return ca.fetchAndStore(ctx, cacheKey, fetch)
	}

	if cached != nil {
		var result T
		if err := json.Unmarshal(cached, &result); err == nil {
			return &result, nil
		}
	}

	return ca.fetchAndStore(ctx, cacheKey, fetch)
}

func (ca *CacheAside[T]) GetMany(ctx context.Context, keys []string, fetch func([]string) (map[string]*T, error)) (map[string]*T, error) {
	result := make(map[string]*T)
	missingKeys := make([]string, 0)

	for _, key := range keys {
		cacheKey := ca.keyPrefix + key
		cached, err := ca.cache.Get(ctx, cacheKey)
		if err != nil || cached == nil {
			missingKeys = append(missingKeys, key)
			continue
		}

		var item T
		if err := json.Unmarshal(cached, &item); err == nil {
			result[key] = &item
		} else {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		fetched, err := fetch(missingKeys)
		if err != nil {
			return result, err
		}

		for key, value := range fetched {
			result[key] = value
			if value != nil {
				cacheKey := ca.keyPrefix + key
				if data, err := json.Marshal(value); err == nil {
					_ = ca.cache.Set(ctx, cacheKey, data, ca.ttl)
				}
			}
		}
	}

	return result, nil
}

func (ca *CacheAside[T]) Set(ctx context.Context, item *T) error {
	if item == nil || ca.keyFunc == nil {
		return nil
	}

	key := ca.keyFunc(*item)
	cacheKey := ca.keyPrefix + key

	data, err := json.Marshal(item)
	if err != nil {
		return err
	}

	return ca.cache.Set(ctx, cacheKey, data, ca.ttl)
}

func (ca *CacheAside[T]) Invalidate(ctx context.Context, key string) error {
	cacheKey := ca.keyPrefix + key
	return ca.cache.Delete(ctx, cacheKey)
}

func (ca *CacheAside[T]) InvalidateMany(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := ca.Invalidate(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (ca *CacheAside[T]) fetchAndStore(ctx context.Context, cacheKey string, fetch func() (*T, error)) (*T, error) {
	result, err := fetch()
	if err != nil {
		return nil, err
	}

	if result != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = ca.cache.Set(ctx, cacheKey, data, ca.ttl)
		}
	}

	return result, nil
}

type WriteThrough[T any] struct {
	cache     Cache
	keyFunc   func(T) string
	ttl       time.Duration
	keyPrefix string
}

func NewWriteThrough[T any](cache Cache, keyPrefix string, keyFunc func(T) string, ttl time.Duration) *WriteThrough[T] {
	return &WriteThrough[T]{
		cache:     cache,
		keyFunc:   keyFunc,
		ttl:       ttl,
		keyPrefix: keyPrefix,
	}
}

func (wt *WriteThrough[T]) Write(ctx context.Context, item *T, persist func(*T) error) error {
	if err := persist(item); err != nil {
		return err
	}

	if item != nil && wt.keyFunc != nil {
		key := wt.keyFunc(*item)
		cacheKey := wt.keyPrefix + key

		if data, err := json.Marshal(item); err == nil {
			_ = wt.cache.Set(ctx, cacheKey, data, wt.ttl)
		}
	}

	return nil
}

func (wt *WriteThrough[T]) Delete(ctx context.Context, key string, deleteFunc func(string) error) error {
	if err := deleteFunc(key); err != nil {
		return err
	}

	cacheKey := wt.keyPrefix + key
	_ = wt.cache.Delete(ctx, cacheKey)

	return nil
}

type RefreshAhead[T any] struct {
	cache          Cache
	keyFunc        func(T) string
	ttl            time.Duration
	refreshBefore  time.Duration
	keyPrefix      string
	refreshChannel chan string
}

func NewRefreshAhead[T any](cache Cache, keyPrefix string, keyFunc func(T) string, ttl, refreshBefore time.Duration) *RefreshAhead[T] {
	return &RefreshAhead[T]{
		cache:          cache,
		keyFunc:        keyFunc,
		ttl:            ttl,
		refreshBefore:  refreshBefore,
		keyPrefix:      keyPrefix,
		refreshChannel: make(chan string, 100),
	}
}

func (ra *RefreshAhead[T]) Get(ctx context.Context, key string, fetch func() (*T, error)) (*T, error) {
	cacheKey := ra.keyPrefix + key

	entry, err := ra.getWithMetadata(ctx, cacheKey)
	if err != nil || entry == nil {
		return ra.fetchAndStore(ctx, cacheKey, fetch)
	}

	if time.Until(entry.ExpiresAt) < ra.refreshBefore {
		select {
		case ra.refreshChannel <- key:
		default:
		}
	}

	var result T
	if err := json.Unmarshal(entry.Value, &result); err != nil {
		return ra.fetchAndStore(ctx, cacheKey, fetch)
	}

	return &result, nil
}

func (ra *RefreshAhead[T]) StartRefreshWorker(ctx context.Context, fetch func(string) (*T, error)) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case key := <-ra.refreshChannel:
				if item, err := fetch(key); err == nil && item != nil {
					cacheKey := ra.keyPrefix + key
					if data, err := json.Marshal(item); err == nil {
						_ = ra.storeWithMetadata(ctx, cacheKey, data)
					}
				}
			}
		}
	}()
}

func (ra *RefreshAhead[T]) getWithMetadata(ctx context.Context, key string) (*CacheEntry, error) {
	data, err := ra.cache.Get(ctx, key+"_meta")
	if err != nil || data == nil {
		return nil, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (ra *RefreshAhead[T]) storeWithMetadata(ctx context.Context, key string, data []byte) error {
	entry := &CacheEntry{
		Value:     data,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ra.ttl),
	}

	metaData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	if err := ra.cache.Set(ctx, key+"_meta", metaData, ra.ttl); err != nil {
		return err
	}

	return ra.cache.Set(ctx, key, data, ra.ttl)
}

func (ra *RefreshAhead[T]) fetchAndStore(ctx context.Context, cacheKey string, fetch func() (*T, error)) (*T, error) {
	result, err := fetch()
	if err != nil {
		return nil, err
	}

	if result != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = ra.storeWithMetadata(ctx, cacheKey, data)
		}
	}

	return result, nil
}

type CachingProxy struct {
	target any
	cache  Cache
	ttl    time.Duration
	prefix string
}

func NewCachingProxy(target any, cache Cache, prefix string, ttl time.Duration) *CachingProxy {
	return &CachingProxy{
		target: target,
		cache:  cache,
		ttl:    ttl,
		prefix: prefix,
	}
}

func (cp *CachingProxy) Call(ctx context.Context, method string, args ...any) (any, error) {
	cacheKey := cp.generateKey(method, args...)

	cached, err := cp.cache.Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var result any
		if err := json.Unmarshal(cached, &result); err == nil {
			return result, nil
		}
	}

	targetValue := reflect.ValueOf(cp.target)
	methodValue := targetValue.MethodByName(method)
	if !methodValue.IsValid() {
		return nil, fmt.Errorf("method %s not found", method)
	}

	callArgs := make([]reflect.Value, len(args)+1)
	callArgs[0] = reflect.ValueOf(ctx)
	for i, arg := range args {
		callArgs[i+1] = reflect.ValueOf(arg)
	}

	results := methodValue.Call(callArgs)
	if len(results) != 2 {
		return nil, fmt.Errorf("method %s must return (result, error)", method)
	}

	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	result := results[0].Interface()
	if result != nil {
		if data, err := json.Marshal(result); err == nil {
			_ = cp.cache.Set(ctx, cacheKey, data, cp.ttl)
		}
	}

	return result, nil
}

func (cp *CachingProxy) generateKey(method string, args ...any) string {
	key := cp.prefix + method
	for _, arg := range args {
		key += fmt.Sprintf(":%v", arg)
	}
	return key
}
