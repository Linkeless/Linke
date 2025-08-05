package cache

import (
	"fmt"
	"time"
)

// CacheConfigManager manages different cache configurations for various use cases
type CacheConfigManager struct {
	configs map[string]*CacheLayerConfig
}

// CacheLayerConfig defines configuration for a specific cache layer or domain
type CacheLayerConfig struct {
	Name            string                 `json:"name"`
	Enabled         bool                   `json:"enabled"`
	TTL             time.Duration          `json:"ttl"`
	MaxSize         int64                  `json:"max_size"`
	EvictionPolicy  EvictionPolicy         `json:"eviction_policy"`
	CompressionType string                 `json:"compression_type"`
	Strategies      *CacheStrategies       `json:"strategies"`
	DomainSpecific  map[string]any `json:"domain_specific"`
}

// CacheStrategies defines caching strategies for different operations
type CacheStrategies struct {
	ReadStrategy    ReadStrategy    `json:"read_strategy"`
	WriteStrategy   WriteStrategy   `json:"write_strategy"`
	WarmingStrategy WarmingStrategy `json:"warming_strategy"`

	// Strategy-specific configurations
	PromotionRatio     float64       `json:"promotion_ratio"`
	ReplicationDelay   time.Duration `json:"replication_delay"`
	WriteBehindDelay   time.Duration `json:"write_behind_delay"`
	WarmingBatchSize   int           `json:"warming_batch_size"`
	WarmingConcurrency int           `json:"warming_concurrency"`
}

// PredefinedConfigs contains common cache configurations
var PredefinedConfigs = map[string]*CacheLayerConfig{
	"hot_data": {
		Name:            "hot_data",
		Enabled:         true,
		TTL:             1 * time.Hour,
		MaxSize:         10000,
		EvictionPolicy:  EvictionPolicyLRU,
		CompressionType: "none",
		Strategies: &CacheStrategies{
			ReadStrategy:       ReadStrategyPromotion,
			WriteStrategy:      WriteStrategyThrough,
			WarmingStrategy:    WarmingStrategyEager,
			PromotionRatio:     0.9,
			ReplicationDelay:   50 * time.Millisecond,
			WriteBehindDelay:   0,
			WarmingBatchSize:   100,
			WarmingConcurrency: 10,
		},
	},

	"warm_data": {
		Name:            "warm_data",
		Enabled:         true,
		TTL:             15 * time.Minute,
		MaxSize:         5000,
		EvictionPolicy:  EvictionPolicyLRU,
		CompressionType: "gzip",
		Strategies: &CacheStrategies{
			ReadStrategy:       ReadStrategyFailover,
			WriteStrategy:      WriteStrategyBehind,
			WarmingStrategy:    WarmingStrategyScheduled,
			PromotionRatio:     0.7,
			ReplicationDelay:   100 * time.Millisecond,
			WriteBehindDelay:   200 * time.Millisecond,
			WarmingBatchSize:   50,
			WarmingConcurrency: 5,
		},
	},

	"cold_data": {
		Name:            "cold_data",
		Enabled:         true,
		TTL:             5 * time.Minute,
		MaxSize:         1000,
		EvictionPolicy:  EvictionPolicyLFU,
		CompressionType: "gzip",
		Strategies: &CacheStrategies{
			ReadStrategy:       ReadStrategyFailover,
			WriteStrategy:      WriteStrategyAround,
			WarmingStrategy:    WarmingStrategyLazy,
			PromotionRatio:     0.3,
			ReplicationDelay:   500 * time.Millisecond,
			WriteBehindDelay:   1 * time.Second,
			WarmingBatchSize:   20,
			WarmingConcurrency: 2,
		},
	},

	"session_data": {
		Name:            "session_data",
		Enabled:         true,
		TTL:             24 * time.Hour,
		MaxSize:         50000,
		EvictionPolicy:  EvictionPolicyTTL,
		CompressionType: "none",
		Strategies: &CacheStrategies{
			ReadStrategy:       ReadStrategyReplication,
			WriteStrategy:      WriteStrategyThrough,
			WarmingStrategy:    WarmingStrategyLazy,
			PromotionRatio:     1.0,
			ReplicationDelay:   10 * time.Millisecond,
			WriteBehindDelay:   0,
			WarmingBatchSize:   200,
			WarmingConcurrency: 20,
		},
	},

	"configuration_data": {
		Name:            "configuration_data",
		Enabled:         true,
		TTL:             12 * time.Hour,
		MaxSize:         500,
		EvictionPolicy:  EvictionPolicyLRU,
		CompressionType: "gzip",
		Strategies: &CacheStrategies{
			ReadStrategy:       ReadStrategyReplication,
			WriteStrategy:      WriteStrategyThrough,
			WarmingStrategy:    WarmingStrategyEager,
			PromotionRatio:     1.0,
			ReplicationDelay:   0,
			WriteBehindDelay:   0,
			WarmingBatchSize:   100,
			WarmingConcurrency: 5,
		},
	},
}

// DomainCacheConfigs maps domain prefixes to appropriate cache configurations
var DomainCacheConfigs = map[string]string{
	CachePrefixUser:         "warm_data",
	CachePrefixSubscription: "hot_data",
	CachePrefixPayment:      "warm_data",
	CachePrefixAuth:         "session_data",
	CachePrefixPlan:         "configuration_data",
	CachePrefixInvoice:      "cold_data",
	CachePrefixServer:       "configuration_data",
	CachePrefixCoupon:       "warm_data",
	CachePrefixReferral:     "cold_data",
	CachePrefixTicket:       "cold_data",
	CachePrefixConfig:       "configuration_data",
	CachePrefixSession:      "session_data",
	CachePrefixRateLimit:    "hot_data",
}

// NewCacheConfigManager creates a new cache configuration manager
func NewCacheConfigManager() *CacheConfigManager {
	manager := &CacheConfigManager{
		configs: make(map[string]*CacheLayerConfig),
	}

	// Load predefined configurations
	for name, config := range PredefinedConfigs {
		manager.configs[name] = config
	}

	return manager
}

// GetConfig returns configuration for a specific cache layer
func (ccm *CacheConfigManager) GetConfig(name string) (*CacheLayerConfig, error) {
	config, exists := ccm.configs[name]
	if !exists {
		return nil, fmt.Errorf("cache configuration '%s' not found", name)
	}
	return config, nil
}

// GetDomainConfig returns appropriate configuration for a domain prefix
func (ccm *CacheConfigManager) GetDomainConfig(prefix string) (*CacheLayerConfig, error) {
	configName, exists := DomainCacheConfigs[prefix]
	if !exists {
		// Default to warm_data configuration
		configName = "warm_data"
	}

	return ccm.GetConfig(configName)
}

// AddConfig adds or updates a cache configuration
func (ccm *CacheConfigManager) AddConfig(name string, config *CacheLayerConfig) {
	ccm.configs[name] = config
}

// RemoveConfig removes a cache configuration
func (ccm *CacheConfigManager) RemoveConfig(name string) {
	delete(ccm.configs, name)
}

// ListConfigs returns all available configurations
func (ccm *CacheConfigManager) ListConfigs() map[string]*CacheLayerConfig {
	result := make(map[string]*CacheLayerConfig)
	for name, config := range ccm.configs {
		result[name] = config
	}
	return result
}

// ToMultiLevelConfig converts a cache layer config to multi-level cache config
func (clc *CacheLayerConfig) ToMultiLevelConfig() *MultiLevelCacheConfig {
	return &MultiLevelCacheConfig{
		EnableL1:         clc.Enabled,
		EnableL2:         true, // Always enable L2 (Redis)
		WriteStrategy:    clc.Strategies.WriteStrategy,
		ReadStrategy:     clc.Strategies.ReadStrategy,
		PromotionRatio:   clc.Strategies.PromotionRatio,
		ReplicationDelay: clc.Strategies.ReplicationDelay,
		L1Config: &MemoryCacheConfig{
			MaxSize:         int(clc.MaxSize),
			DefaultTTL:      clc.TTL,
			EvictionPolicy:  clc.EvictionPolicy,
			CleanupInterval: clc.TTL / 10, // Cleanup every 10% of TTL
		},
		L2Config: &CacheConfig{
			DefaultTTL:      clc.TTL,
			MaxRetries:      3,
			RetryDelay:      100 * time.Millisecond,
			EnableMetrics:   true,
			CompressionType: clc.CompressionType,
		},
	}
}

// ToWarmingConfig converts strategies to warming configuration
func (cs *CacheStrategies) ToWarmingConfig(prefixes []string) *WarmingConfig {
	return &WarmingConfig{
		Strategy:       cs.WarmingStrategy,
		BatchSize:      cs.WarmingBatchSize,
		ConcurrentJobs: cs.WarmingConcurrency,
		WarmingTTL:     1 * time.Hour, // Default warming TTL
		Prefixes:       prefixes,
		MaxItems:       cs.WarmingBatchSize * 10, // 10x batch size
		Enabled:        cs.WarmingStrategy != WarmingStrategyLazy,
	}
}

// CacheProfileOptimizer optimizes cache configurations based on usage patterns
type CacheProfileOptimizer struct {
	metrics map[string]*UsageMetrics
}

// UsageMetrics tracks usage patterns for cache optimization
type UsageMetrics struct {
	HitRate          float64       `json:"hit_rate"`
	MissRate         float64       `json:"miss_rate"`
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	AccessFrequency  int64         `json:"access_frequency"`
	DataSize         int64         `json:"data_size"`
	LastOptimization time.Time     `json:"last_optimization"`
}

// NewCacheProfileOptimizer creates a new cache profile optimizer
func NewCacheProfileOptimizer() *CacheProfileOptimizer {
	return &CacheProfileOptimizer{
		metrics: make(map[string]*UsageMetrics),
	}
}

// AnalyzeAndOptimize analyzes usage patterns and suggests optimizations
func (cpo *CacheProfileOptimizer) AnalyzeAndOptimize(prefix string, currentConfig *CacheLayerConfig) *CacheLayerConfig {
	metrics, exists := cpo.metrics[prefix]
	if !exists {
		// No metrics available, return current config
		return currentConfig
	}

	optimized := *currentConfig // Copy current config

	// Optimize based on hit rate
	if metrics.HitRate < 50.0 {
		// Low hit rate - reduce TTL and cache size
		optimized.TTL = optimized.TTL / 2
		optimized.MaxSize = optimized.MaxSize / 2
		optimized.Strategies.ReadStrategy = ReadStrategyFailover
	} else if metrics.HitRate > 90.0 {
		// High hit rate - increase TTL and consider more aggressive caching
		optimized.TTL = optimized.TTL * 2
		optimized.Strategies.ReadStrategy = ReadStrategyPromotion
		optimized.Strategies.PromotionRatio = 0.9
	}

	// Optimize based on response time
	if metrics.AvgResponseTime > 10*time.Millisecond {
		// Slow responses - prioritize L1 cache
		optimized.Strategies.ReadStrategy = ReadStrategyReplication
		optimized.Strategies.PromotionRatio = 1.0
	}

	// Optimize based on access frequency
	if metrics.AccessFrequency > 1000 {
		// High frequency - use hot data configuration
		hotConfig := PredefinedConfigs["hot_data"]
		optimized.Strategies = hotConfig.Strategies
	} else if metrics.AccessFrequency < 10 {
		// Low frequency - use cold data configuration
		coldConfig := PredefinedConfigs["cold_data"]
		optimized.Strategies = coldConfig.Strategies
	}

	return &optimized
}

// UpdateMetrics updates usage metrics for a cache prefix
func (cpo *CacheProfileOptimizer) UpdateMetrics(prefix string, metrics *UsageMetrics) {
	cpo.metrics[prefix] = metrics
}

// GetMetrics returns current usage metrics for a prefix
func (cpo *CacheProfileOptimizer) GetMetrics(prefix string) *UsageMetrics {
	return cpo.metrics[prefix]
}

// ConfigurationValidator validates cache configurations
type ConfigurationValidator struct{}

// NewConfigurationValidator creates a new configuration validator
func NewConfigurationValidator() *ConfigurationValidator {
	return &ConfigurationValidator{}
}

// ValidateConfig validates a cache layer configuration
func (cv *ConfigurationValidator) ValidateConfig(config *CacheLayerConfig) error {
	if config.Name == "" {
		return fmt.Errorf("cache configuration name cannot be empty")
	}

	if config.TTL <= 0 {
		return fmt.Errorf("cache TTL must be positive")
	}

	if config.MaxSize <= 0 {
		return fmt.Errorf("cache max size must be positive")
	}

	if config.Strategies == nil {
		return fmt.Errorf("cache strategies cannot be nil")
	}

	// Validate strategies
	if err := cv.validateStrategies(config.Strategies); err != nil {
		return fmt.Errorf("invalid cache strategies: %w", err)
	}

	return nil
}

// ValidateMultiLevelConfig validates a multi-level cache configuration
func (cv *ConfigurationValidator) ValidateMultiLevelConfig(config *MultiLevelCacheConfig) error {
	if !config.EnableL1 && !config.EnableL2 {
		return fmt.Errorf("at least one cache level must be enabled")
	}

	if config.PromotionRatio < 0 || config.PromotionRatio > 1 {
		return fmt.Errorf("promotion ratio must be between 0 and 1")
	}

	if config.ReplicationDelay < 0 {
		return fmt.Errorf("replication delay cannot be negative")
	}

	if config.EnableL1 && config.L1Config == nil {
		return fmt.Errorf("L1 configuration required when L1 is enabled")
	}

	if config.EnableL2 && config.L2Config == nil {
		return fmt.Errorf("L2 configuration required when L2 is enabled")
	}

	return nil
}

func (cv *ConfigurationValidator) validateStrategies(strategies *CacheStrategies) error {
	// Validate read strategy
	validReadStrategies := []ReadStrategy{ReadStrategyFailover, ReadStrategyPromotion, ReadStrategyReplication}
	if !cv.isValidReadStrategy(strategies.ReadStrategy, validReadStrategies) {
		return fmt.Errorf("invalid read strategy: %s", strategies.ReadStrategy)
	}

	// Validate write strategy
	validWriteStrategies := []WriteStrategy{WriteStrategyThrough, WriteStrategyBehind, WriteStrategyAround}
	if !cv.isValidWriteStrategy(strategies.WriteStrategy, validWriteStrategies) {
		return fmt.Errorf("invalid write strategy: %s", strategies.WriteStrategy)
	}

	// Validate warming strategy
	validWarmingStrategies := []WarmingStrategy{WarmingStrategyEager, WarmingStrategyLazy, WarmingStrategyScheduled, WarmingStrategyPredictive}
	if !cv.isValidWarmingStrategy(strategies.WarmingStrategy, validWarmingStrategies) {
		return fmt.Errorf("invalid warming strategy: %s", strategies.WarmingStrategy)
	}

	// Validate numeric values
	if strategies.PromotionRatio < 0 || strategies.PromotionRatio > 1 {
		return fmt.Errorf("promotion ratio must be between 0 and 1")
	}

	if strategies.WarmingBatchSize <= 0 {
		return fmt.Errorf("warming batch size must be positive")
	}

	if strategies.WarmingConcurrency <= 0 {
		return fmt.Errorf("warming concurrency must be positive")
	}

	return nil
}

func (cv *ConfigurationValidator) isValidReadStrategy(strategy ReadStrategy, validStrategies []ReadStrategy) bool {
	for _, valid := range validStrategies {
		if strategy == valid {
			return true
		}
	}
	return false
}

func (cv *ConfigurationValidator) isValidWriteStrategy(strategy WriteStrategy, validStrategies []WriteStrategy) bool {
	for _, valid := range validStrategies {
		if strategy == valid {
			return true
		}
	}
	return false
}

func (cv *ConfigurationValidator) isValidWarmingStrategy(strategy WarmingStrategy, validStrategies []WarmingStrategy) bool {
	for _, valid := range validStrategies {
		if strategy == valid {
			return true
		}
	}
	return false
}
