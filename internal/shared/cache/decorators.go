package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// KeyStrategy defines how cache keys are generated
type KeyStrategy interface {
	GenerateKey(prefix string, identifier string, params ...interface{}) string
	ValidateKey(key string) error
	NormalizeKey(key string) string
}

// TTLStrategy defines how TTL is determined for cache entries
type TTLStrategy interface {
	GetTTL(ctx context.Context, key string, value interface{}) time.Duration
	GetRefreshTTL(ctx context.Context, key string) time.Duration
	ShouldRefresh(ctx context.Context, key string, age time.Duration) bool
}

// DefaultKeyStrategy implements a simple key generation strategy
type DefaultKeyStrategy struct {
	separator string
	maxLength int
}

// NewDefaultKeyStrategy creates a new default key strategy
func NewDefaultKeyStrategy() *DefaultKeyStrategy {
	return &DefaultKeyStrategy{
		separator: ":",
		maxLength: 250, // Redis max key length is 512MB, but keep reasonable
	}
}

func (dks *DefaultKeyStrategy) GenerateKey(prefix string, identifier string, params ...interface{}) string {
	parts := []string{prefix, identifier}
	
	for _, param := range params {
		parts = append(parts, fmt.Sprintf("%v", param))
	}
	
	key := strings.Join(parts, dks.separator)
	return dks.NormalizeKey(key)
}

func (dks *DefaultKeyStrategy) ValidateKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	if len(key) > dks.maxLength {
		return fmt.Errorf("key length %d exceeds maximum %d", len(key), dks.maxLength)
	}
	return nil
}

func (dks *DefaultKeyStrategy) NormalizeKey(key string) string {
	// Remove any problematic characters and normalize
	normalized := strings.ReplaceAll(key, " ", "_")
	normalized = strings.ToLower(normalized)
	
	// If key is too long, hash it
	if len(normalized) > dks.maxLength {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(normalized)))
		prefix := normalized[:dks.maxLength-len(hash)-1]
		return prefix + "_" + hash
	}
	
	return normalized
}

// HashKeyStrategy implements a hash-based key generation strategy
type HashKeyStrategy struct {
	*DefaultKeyStrategy
	hashLongKeys bool
}

func NewHashKeyStrategy() *HashKeyStrategy {
	return &HashKeyStrategy{
		DefaultKeyStrategy: NewDefaultKeyStrategy(),
		hashLongKeys:       true,
	}
}

func (hks *HashKeyStrategy) GenerateKey(prefix string, identifier string, params ...interface{}) string {
	if !hks.hashLongKeys {
		return hks.DefaultKeyStrategy.GenerateKey(prefix, identifier, params...)
	}
	
	// Always hash for consistent length
	allParams := append([]interface{}{prefix, identifier}, params...)
	data, _ := json.Marshal(allParams)
	hash := fmt.Sprintf("%x", md5.Sum(data))
	
	return fmt.Sprintf("%s:%s", prefix, hash)
}

// FixedTTLStrategy implements a simple fixed TTL strategy
type FixedTTLStrategy struct {
	defaultTTL time.Duration
	refreshTTL time.Duration
}

func NewFixedTTLStrategy(defaultTTL time.Duration) *FixedTTLStrategy {
	return &FixedTTLStrategy{
		defaultTTL: defaultTTL,
		refreshTTL: defaultTTL / 4, // Refresh when 75% of TTL has passed
	}
}

func (fts *FixedTTLStrategy) GetTTL(ctx context.Context, key string, value interface{}) time.Duration {
	return fts.defaultTTL
}

func (fts *FixedTTLStrategy) GetRefreshTTL(ctx context.Context, key string) time.Duration {
	return fts.refreshTTL
}

func (fts *FixedTTLStrategy) ShouldRefresh(ctx context.Context, key string, age time.Duration) bool {
	return age >= fts.refreshTTL
}

// AdaptiveTTLStrategy implements an adaptive TTL strategy based on data characteristics
type AdaptiveTTLStrategy struct {
	baseTTL    time.Duration
	maxTTL     time.Duration
	minTTL     time.Duration
	refreshTTL time.Duration
}

func NewAdaptiveTTLStrategy(baseTTL, minTTL, maxTTL time.Duration) *AdaptiveTTLStrategy {
	return &AdaptiveTTLStrategy{
		baseTTL:    baseTTL,
		maxTTL:     maxTTL,
		minTTL:     minTTL,
		refreshTTL: baseTTL / 3,
	}
}

func (ats *AdaptiveTTLStrategy) GetTTL(ctx context.Context, key string, value interface{}) time.Duration {
	// Adjust TTL based on value characteristics
	ttl := ats.baseTTL
	
	if value != nil {
		// Larger objects get shorter TTL
		if data, err := json.Marshal(value); err == nil {
			size := len(data)
			if size > 10240 { // 10KB
				ttl = ats.minTTL
			} else if size < 1024 { // 1KB
				ttl = ats.maxTTL
			}
		}
	}
	
	// Ensure TTL is within bounds
	if ttl < ats.minTTL {
		ttl = ats.minTTL
	}
	if ttl > ats.maxTTL {
		ttl = ats.maxTTL
	}
	
	return ttl
}

func (ats *AdaptiveTTLStrategy) GetRefreshTTL(ctx context.Context, key string) time.Duration {
	return ats.refreshTTL
}

func (ats *AdaptiveTTLStrategy) ShouldRefresh(ctx context.Context, key string, age time.Duration) bool {
	return age >= ats.refreshTTL
}

// EnhancedCacheDecorator extends the basic cache decorator with strategies
type EnhancedCacheDecorator struct {
	cache       Cache
	keyStrategy KeyStrategy
	ttlStrategy TTLStrategy
	prefix      string
}

// NewEnhancedCacheDecorator creates a new enhanced cache decorator
func NewEnhancedCacheDecorator(cache Cache, prefix string, keyStrategy KeyStrategy, ttlStrategy TTLStrategy) *EnhancedCacheDecorator {
	if keyStrategy == nil {
		keyStrategy = NewDefaultKeyStrategy()
	}
	if ttlStrategy == nil {
		ttlStrategy = NewFixedTTLStrategy(5 * time.Minute)
	}
	
	return &EnhancedCacheDecorator{
		cache:       cache,
		keyStrategy: keyStrategy,
		ttlStrategy: ttlStrategy,
		prefix:      prefix,
	}
}

// GetWithStrategy gets a value using configured strategies
func (ecd *EnhancedCacheDecorator) GetWithStrategy(ctx context.Context, identifier string, fetch func() (interface{}, error), params ...interface{}) (interface{}, error) {
	cacheKey := ecd.keyStrategy.GenerateKey(ecd.prefix, identifier, params...)
	
	if err := ecd.keyStrategy.ValidateKey(cacheKey); err != nil {
		return fetch() // Fall back to direct fetch if key is invalid
	}
	
	cached, err := ecd.cache.Get(ctx, cacheKey)
	if err != nil {
		return ecd.fetchAndStore(ctx, cacheKey, fetch)
	}
	
	if cached != nil {
		var result interface{}
		if err := json.Unmarshal(cached, &result); err == nil {
			return result, nil
		}
	}
	
	return ecd.fetchAndStore(ctx, cacheKey, fetch)
}

// SetWithStrategy sets a value using configured strategies
func (ecd *EnhancedCacheDecorator) SetWithStrategy(ctx context.Context, identifier string, value interface{}, params ...interface{}) error {
	cacheKey := ecd.keyStrategy.GenerateKey(ecd.prefix, identifier, params...)
	
	if err := ecd.keyStrategy.ValidateKey(cacheKey); err != nil {
		return err
	}
	
	ttl := ecd.ttlStrategy.GetTTL(ctx, cacheKey, value)
	
	if value != nil {
		if data, err := json.Marshal(value); err == nil {
			return ecd.cache.Set(ctx, cacheKey, data, ttl)
		}
	}
	
	return nil
}

// InvalidateWithStrategy invalidates a cache entry using configured strategies
func (ecd *EnhancedCacheDecorator) InvalidateWithStrategy(ctx context.Context, identifier string, params ...interface{}) error {
	cacheKey := ecd.keyStrategy.GenerateKey(ecd.prefix, identifier, params...)
	return ecd.cache.Delete(ctx, cacheKey)
}

func (ecd *EnhancedCacheDecorator) fetchAndStore(ctx context.Context, cacheKey string, fetch func() (interface{}, error)) (interface{}, error) {
	result, err := fetch()
	if err != nil {
		return nil, err
	}
	
	if result != nil {
		ttl := ecd.ttlStrategy.GetTTL(ctx, cacheKey, result)
		if data, err := json.Marshal(result); err == nil {
			_ = ecd.cache.Set(ctx, cacheKey, data, ttl)
		}
	}
	
	return result, nil
}

// GetKeyStrategy returns the current key strategy
func (ecd *EnhancedCacheDecorator) GetKeyStrategy() KeyStrategy {
	return ecd.keyStrategy
}

// GetTTLStrategy returns the current TTL strategy
func (ecd *EnhancedCacheDecorator) GetTTLStrategy() TTLStrategy {
	return ecd.ttlStrategy
}

// SetKeyStrategy updates the key strategy
func (ecd *EnhancedCacheDecorator) SetKeyStrategy(strategy KeyStrategy) {
	if strategy != nil {
		ecd.keyStrategy = strategy
	}
}

// SetTTLStrategy updates the TTL strategy
func (ecd *EnhancedCacheDecorator) SetTTLStrategy(strategy TTLStrategy) {
	if strategy != nil {
		ecd.ttlStrategy = strategy
	}
}

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
