package cache

import (
	"linke/internal/shared/config"
	"linke/internal/shared/database"
	"linke/internal/shared/logger"
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
		// Enhanced cache monitoring handler with optional multi-level support
		func(
			manager CacheManager,
			collector MetricsCollector,
			logger logger.Logger,
		) *CacheMonitoringHandler {
			// For basic cache module, no multi-level manager available
			return NewCacheMonitoringHandler(manager, collector, nil, logger)
		},
		NewCacheHealthCheck,
	),
)

var MultiLevelModule = fx.Module("multilevel-cache",
	fx.Provide(
		NewMultiLevelCacheConfig,
		NewMemoryCacheConfig,
		NewMultiLevelCacheFromConfig,
		fx.Annotate(
			NewMultiLevelCacheManagerFromConfig,
			fx.As(new(MultiLevelCacheManager)),
		),
		NewAllCacheKeys,
		// Enhanced cache monitoring handler WITH multi-level support
		func(
			manager CacheManager,
			collector MetricsCollector,
			multiLevelMgr MultiLevelCacheManager,
			logger logger.Logger,
		) *CacheMonitoringHandler {
			// For multi-level cache module, provide the multi-level manager
			return NewCacheMonitoringHandler(manager, collector, multiLevelMgr, logger)
		},
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
	UserCache *CacheAside[any] `name:"userCache"`
	PlanCache *CacheAside[any] `name:"planCache"`
	AuthCache *CacheDecorator  `name:"authCache"`
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

// Multi-level cache configuration functions

func NewMultiLevelCacheConfig(cfg *config.Config) *MultiLevelCacheConfig {
	return &MultiLevelCacheConfig{
		EnableL1:         true,
		EnableL2:         true,
		WriteStrategy:    WriteStrategyThrough,
		ReadStrategy:     ReadStrategyPromotion,
		PromotionRatio:   0.8,
		ReplicationDelay: 100 * time.Millisecond,
		L1Config:         NewMemoryCacheConfig(cfg),
		L2Config:         NewCacheConfig(cfg),
	}
}

func NewMemoryCacheConfig(cfg *config.Config) *MemoryCacheConfig {
	return &MemoryCacheConfig{
		MaxSize:         1000,
		DefaultTTL:      DefaultCacheTTL,
		EvictionPolicy:  EvictionPolicyLRU,
		CleanupInterval: 1 * time.Minute,
	}
}

func NewMultiLevelCacheFromConfig(
	config *MultiLevelCacheConfig,
	l2Cache Cache,
	collector MetricsCollector,
	logger logger.Logger,
) *MultiLevelCache {
	return NewMultiLevelCache(config, l2Cache, collector, logger)
}

func NewMultiLevelCacheManagerFromConfig(
	config *MultiLevelCacheConfig,
	l2Cache Cache,
	cacheKeys *AllCacheKeys,
	collector MetricsCollector,
	logger logger.Logger,
) MultiLevelCacheManager {
	return NewMultiLevelCacheManager(config, l2Cache, cacheKeys, collector, logger)
}
