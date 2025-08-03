package cache

import (
	"linke/internal/shared/config"
	"linke/internal/shared/database"
	"time"

	"go.uber.org/fx"
)

var Module = fx.Module("cache",
	fx.Provide(
		NewCacheConfig,
		NewRedisCacheFromDB,
		NewMetricsCollector,
		NewMetricsCacheFromRedis,
		fx.Annotate(
			NewRedisCacheManagerWithMetrics,
			fx.As(new(CacheManager)),
		),
		NewAllCacheKeys,
		NewCacheMonitoringHandler,
		NewCacheHealthCheck,
	),
)

func NewCacheConfig(cfg *config.Config) *CacheConfig {
	defaultTTL := DefaultCacheTTL
	if cfg.Cache.DefaultTTL > 0 {
		defaultTTL = time.Duration(cfg.Cache.DefaultTTL) * time.Second
	}

	return &CacheConfig{
		DefaultTTL:      defaultTTL,
		MaxRetries:      3,
		RetryDelay:      100 * time.Millisecond,
		EnableMetrics:   cfg.Cache.EnableMetrics,
		MetricsPrefix:   "linke_cache_",
		CompressionType: "gzip",
	}
}

func NewRedisCacheFromDB(db *database.Database, config *CacheConfig) *RedisCache {
	return NewRedisCache(db.Redis, config)
}

func NewMetricsCacheFromRedis(cache *RedisCache, collector MetricsCollector) Cache {
	if collector != nil {
		return NewMetricsCacheWrapper(cache, collector)
	}
	return cache
}

func NewRedisCacheManagerWithMetrics(cache Cache, db *database.Database, config *CacheConfig) *RedisCacheManager {
	return NewRedisCacheManager(db.Redis, config)
}

type CacheMetrics struct {
	manager CacheManager
}

func NewCacheMetrics(manager CacheManager) *CacheMetrics {
	return &CacheMetrics{manager: manager}
}

type CacheProviders struct {
	fx.Out

	Manager   CacheManager
	Keys      *AllCacheKeys
	Metrics   *CacheMetrics
	UserCache *CacheAside[any]      `name:"userCache"`
	PlanCache *CacheAside[any]      `name:"planCache"`
	AuthCache *CacheDecorator       `name:"authCache"`
}

func ProvideDomainCaches(manager CacheManager, keys *AllCacheKeys) CacheProviders {
	return CacheProviders{
		Manager: manager,
		Keys:    keys,
		UserCache: NewCacheAside[any](
			manager.GetCache(),
			CachePrefixUser,
			nil,
			MediumCacheTTL,
		),
		PlanCache: NewCacheAside[any](
			manager.GetCache(),
			CachePrefixPlan,
			nil,
			LongCacheTTL,
		),
		AuthCache: NewCacheDecorator(
			manager.GetCache(),
			CachePrefixAuth,
			ShortCacheTTL,
		),
	}
}