package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Flush(ctx context.Context) error
}

type TypedCache[T any] interface {
	Get(ctx context.Context, key string) (*T, error)
	Set(ctx context.Context, key string, value *T, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
}

type CacheManager interface {
	GetCache() Cache
	GetTypedCache(prefix string) TypedCache[any]
	InvalidateCache(ctx context.Context, patterns ...string) error
	GetStats(ctx context.Context) (*CacheStats, error)
}

// MultiLevelCacheManager extends CacheManager with multi-level capabilities
type MultiLevelCacheManager interface {
	CacheManager
	GetMultiLevelCache() *MultiLevelCache
	GetL1Cache() *MemoryCache
	GetL2Cache() Cache
	GetWarmer() *CacheWarmer
	GetInvalidator() *EventDrivenInvalidator
	GetMonitor() *MultiLevelCacheMonitor
	SwitchStrategy(writeStrategy WriteStrategy, readStrategy ReadStrategy) error
}

type CacheStats struct {
	Hits       int64
	Misses     int64
	Sets       int64
	Deletes    int64
	Evictions  int64
	HitRate    float64
	TotalKeys  int64
	MemoryUsed int64
}

type CacheConfig struct {
	DefaultTTL      time.Duration
	MaxRetries      int
	RetryDelay      time.Duration
	EnableMetrics   bool
	MetricsPrefix   string
	CompressionType string
}

type CacheKeyBuilder interface {
	Build(parts ...string) string
	BuildWithPrefix(prefix string, parts ...string) string
	ExtractPattern(key string) string
}

type CacheTags interface {
	AddToTag(ctx context.Context, tag string, keys ...string) error
	GetByTag(ctx context.Context, tag string) ([]string, error)
	InvalidateTag(ctx context.Context, tag string) error
	RemoveFromTag(ctx context.Context, tag string, keys ...string) error
}

type CacheOptions struct {
	TTL                  time.Duration
	Tags                 []string
	SkipCompression      bool
	RefreshOnMiss        bool
	StaleWhileRevalidate bool
}

type CacheEntry struct {
	Value      []byte
	CreatedAt  time.Time
	ExpiresAt  time.Time
	Tags       []string
	Version    int
	Compressed bool
}

type CacheError struct {
	Op      string
	Key     string
	Pattern string
	Err     error
}

func (e *CacheError) Error() string {
	if e.Pattern != "" {
		return e.Op + " cache pattern " + e.Pattern + ": " + e.Err.Error()
	}
	return e.Op + " cache key " + e.Key + ": " + e.Err.Error()
}

func (e *CacheError) Unwrap() error {
	return e.Err
}

const (
	CachePrefixUser         = "user:"
	CachePrefixSubscription = "subscription:"
	CachePrefixPayment      = "payment:"
	CachePrefixAuth         = "auth:"
	CachePrefixPlan         = "plan:"
	CachePrefixInvoice      = "invoice:"
	CachePrefixServer       = "server:"
	CachePrefixCoupon       = "coupon:"
	CachePrefixReferral     = "referral:"
	CachePrefixTicket       = "ticket:"
	CachePrefixConfig       = "config:"
	CachePrefixSession      = "session:"
	CachePrefixRateLimit    = "rate_limit:"
)

const (
	CacheTagUser         = "tag:user"
	CacheTagSubscription = "tag:subscription"
	CacheTagPayment      = "tag:payment"
	CacheTagConfig       = "tag:config"
	CacheTagGlobal       = "tag:global"
)

const (
	DefaultCacheTTL = 5 * time.Minute
	ShortCacheTTL   = 1 * time.Minute
	MediumCacheTTL  = 15 * time.Minute
	LongCacheTTL    = 1 * time.Hour
	SessionCacheTTL = 24 * time.Hour
	ConfigCacheTTL  = 12 * time.Hour
)
