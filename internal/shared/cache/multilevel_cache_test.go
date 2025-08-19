package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"linke/internal/shared/logger"
)

// MockCache is a mock implementation of Cache interface
type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCache) DeleteByPattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCache) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	args := m.Called(ctx, key, ttl)
	return args.Error(0)
}

func (m *MockCache) Flush(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockLogger is a mock implementation for testing
type MockLogger struct{}

func (ml *MockLogger) Debug(msg string, fields ...zap.Field)  {}
func (ml *MockLogger) Info(msg string, fields ...zap.Field)   {}
func (ml *MockLogger) Warn(msg string, fields ...zap.Field)   {}
func (ml *MockLogger) Error(msg string, fields ...zap.Field)  {}
func (ml *MockLogger) Fatal(msg string, fields ...zap.Field)  {}
func (ml *MockLogger) With(fields ...zap.Field) logger.Logger { return ml }
func (ml *MockLogger) Sync() error                            { return nil }

// MockMetricsCollector is a mock implementation of MetricsCollector
type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) RecordHit(ctx context.Context, key string) {
	m.Called(ctx, key)
}

func (m *MockMetricsCollector) RecordMiss(ctx context.Context, key string) {
	m.Called(ctx, key)
}

func (m *MockMetricsCollector) RecordSet(ctx context.Context, key string, ttl time.Duration) {
	m.Called(ctx, key, ttl)
}

func (m *MockMetricsCollector) RecordDelete(ctx context.Context, key string) {
	m.Called(ctx, key)
}

func (m *MockMetricsCollector) RecordError(ctx context.Context, key, operation string) {
	m.Called(ctx, key, operation)
}

func (m *MockMetricsCollector) RecordEviction(ctx context.Context, key string) {
	m.Called(ctx, key)
}

func (m *MockMetricsCollector) GetMetrics() *Metrics {
	args := m.Called()
	return args.Get(0).(*Metrics)
}

func (m *MockMetricsCollector) GetMetricsByPrefix(prefix string) *Metrics {
	args := m.Called(prefix)
	return args.Get(0).(*Metrics)
}

func (m *MockMetricsCollector) Reset() {
	m.Called()
}

func TestMemoryCache(t *testing.T) {
	t.Run("Basic Operations", func(t *testing.T) {
		config := &MemoryCacheConfig{
			MaxSize:         100,
			DefaultTTL:      1 * time.Minute,
			EvictionPolicy:  EvictionPolicyLRU,
			CleanupInterval: 10 * time.Second,
		}

		cache := NewMemoryCache(config)
		defer cache.Close()

		ctx := context.Background()
		key := "test_key"
		value := []byte("test_value")

		// Test Set and Get
		err := cache.Set(ctx, key, value, 1*time.Minute)
		assert.NoError(t, err)

		retrieved, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrieved)

		// Test Exists
		exists, err := cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Test Delete
		err = cache.Delete(ctx, key)
		assert.NoError(t, err)

		retrieved, err = cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("TTL Expiration", func(t *testing.T) {
		config := &MemoryCacheConfig{
			MaxSize:         100,
			DefaultTTL:      100 * time.Millisecond,
			EvictionPolicy:  EvictionPolicyTTL,
			CleanupInterval: 50 * time.Millisecond,
		}

		cache := NewMemoryCache(config)
		defer cache.Close()

		ctx := context.Background()
		key := "expire_test"
		value := []byte("expire_value")

		// Set with short TTL
		err := cache.Set(ctx, key, value, 100*time.Millisecond)
		assert.NoError(t, err)

		// Should exist immediately
		retrieved, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, retrieved)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired
		retrieved, err = cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("LRU Eviction", func(t *testing.T) {
		// Use very small memory limit to force eviction based on size
		// Since hasSpace() converts MaxSize to MB, we need MaxSize=1 to get 1MB limit
		config := &MemoryCacheConfig{
			MaxSize:         1, // 1MB memory limit
			DefaultTTL:      1 * time.Minute,
			EvictionPolicy:  EvictionPolicyLRU,
			CleanupInterval: 10 * time.Second,
		}

		cache := NewMemoryCache(config)
		defer cache.Close()

		ctx := context.Background()

		// Create large values to trigger memory-based eviction
		// Each value is ~300KB, so 3 values = 900KB < 1MB, but 4 values = 1200KB > 1MB
		largeValue := make([]byte, 300*1024) // 300KB
		for i := range largeValue {
			largeValue[i] = byte(i % 256)
		}

		// Add 3 items, approaching the 1MB limit
		for i := 0; i < 3; i++ {
			key := fmt.Sprintf("key_%d", i)
			value := make([]byte, len(largeValue))
			copy(value, largeValue)
			// Make each value slightly different
			value[0] = byte(i)
			err := cache.Set(ctx, key, value, 1*time.Minute)
			assert.NoError(t, err)
		}

		// Access key_1 to make it recently used (move to front of LRU)
		_, err := cache.Get(ctx, "key_1")
		assert.NoError(t, err)

		// Access key_2 to make it recently used (move to front of LRU)
		_, err = cache.Get(ctx, "key_2")
		assert.NoError(t, err)

		// At this point, LRU order should be: key_0 (least recently used), key_1, key_2 (most recently used)

		// Add one more large item to trigger eviction of key_0
		value3 := make([]byte, len(largeValue))
		copy(value3, largeValue)
		value3[0] = 3
		err = cache.Set(ctx, "key_3", value3, 1*time.Minute)
		assert.NoError(t, err)

		// key_0 should be evicted (least recently used)
		retrieved, err := cache.Get(ctx, "key_0")
		assert.NoError(t, err)
		assert.Nil(t, retrieved)

		// key_1 should still exist
		retrieved, err = cache.Get(ctx, "key_1")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, byte(1), retrieved[0]) // Verify it's the right value

		// key_2 should still exist
		retrieved, err = cache.Get(ctx, "key_2")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, byte(2), retrieved[0]) // Verify it's the right value

		// key_3 should exist (newly added)
		retrieved, err = cache.Get(ctx, "key_3")
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, byte(3), retrieved[0]) // Verify it's the right value
	})
}

func TestMultiLevelCache(t *testing.T) {
	t.Run("Write-Through Strategy", func(t *testing.T) {
		mockL2 := new(MockCache)
		mockCollector := new(MockMetricsCollector)
		logger := &MockLogger{}

		config := &MultiLevelCacheConfig{
			EnableL1:         true,
			EnableL2:         true,
			WriteStrategy:    WriteStrategyThrough,
			ReadStrategy:     ReadStrategyFailover,
			PromotionRatio:   0.8,
			ReplicationDelay: 100 * time.Millisecond,
			L1Config: &MemoryCacheConfig{
				MaxSize:         100,
				DefaultTTL:      1 * time.Minute,
				EvictionPolicy:  EvictionPolicyLRU,
				CleanupInterval: 10 * time.Second,
			},
		}

		mlc := NewMultiLevelCache(config, mockL2, mockCollector, logger)
		defer mlc.Close()

		ctx := context.Background()
		key := "test_key"
		value := []byte("test_value")
		ttl := 1 * time.Minute

		// Mock L2 cache Set operation
		mockL2.On("Set", ctx, key, value, ttl).Return(nil)
		mockCollector.On("RecordSet", ctx, key, ttl).Return()

		// Test write-through
		err := mlc.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// Verify that both L1 and L2 were called
		mockL2.AssertExpectations(t)

		// Test that value exists in L1
		l1Value, err := mlc.l1Cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, l1Value)
	})

	t.Run("Read Failover Strategy", func(t *testing.T) {
		mockL2 := new(MockCache)
		mockCollector := new(MockMetricsCollector)
		logger := &MockLogger{}

		config := &MultiLevelCacheConfig{
			EnableL1:         true,
			EnableL2:         true,
			WriteStrategy:    WriteStrategyThrough,
			ReadStrategy:     ReadStrategyFailover,
			PromotionRatio:   0.8,
			ReplicationDelay: 100 * time.Millisecond,
			L1Config: &MemoryCacheConfig{
				MaxSize:         100,
				DefaultTTL:      1 * time.Minute,
				EvictionPolicy:  EvictionPolicyLRU,
				CleanupInterval: 10 * time.Second,
			},
		}

		mlc := NewMultiLevelCache(config, mockL2, mockCollector, logger)
		defer mlc.Close()

		ctx := context.Background()
		key := "test_key"
		value := []byte("test_value")

		// First, test L1 hit
		mlc.l1Cache.Set(ctx, key, value, 1*time.Minute)
		mockCollector.On("RecordHit", ctx, "L1:"+key).Return()

		result, err := mlc.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// Test L1 miss, L2 hit
		mlc.l1Cache.Delete(ctx, key)
		mockL2.On("Get", ctx, key).Return(value, nil)
		mockCollector.On("RecordHit", ctx, "L2:"+key).Return()

		result, err = mlc.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, result)

		// Test complete miss
		mockL2.On("Get", ctx, "missing_key").Return([]byte(nil), nil)
		mockCollector.On("RecordMiss", ctx, "missing_key").Return()

		result, err = mlc.Get(ctx, "missing_key")
		assert.NoError(t, err)
		assert.Nil(t, result)

		mockL2.AssertExpectations(t)
		mockCollector.AssertExpectations(t)
	})

	t.Run("Write-Behind Strategy", func(t *testing.T) {
		mockL2 := new(MockCache)
		mockCollector := new(MockMetricsCollector)
		logger := &MockLogger{}

		config := &MultiLevelCacheConfig{
			EnableL1:         true,
			EnableL2:         true,
			WriteStrategy:    WriteStrategyBehind,
			ReadStrategy:     ReadStrategyFailover,
			PromotionRatio:   0.8,
			ReplicationDelay: 10 * time.Millisecond, // Short delay for testing
			L1Config: &MemoryCacheConfig{
				MaxSize:         100,
				DefaultTTL:      1 * time.Minute,
				EvictionPolicy:  EvictionPolicyLRU,
				CleanupInterval: 10 * time.Second,
			},
		}

		mlc := NewMultiLevelCache(config, mockL2, mockCollector, logger)
		defer mlc.Close()

		ctx := context.Background()
		key := "test_key"
		value := []byte("test_value")
		ttl := 1 * time.Minute

		// Mock L2 cache Set operation (will be called asynchronously)
		mockL2.On("Set", ctx, key, value, ttl).Return(nil)
		mockCollector.On("RecordSet", ctx, key, ttl).Return()

		// Test write-behind
		err := mlc.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// Value should be immediately available in L1
		l1Value, err := mlc.l1Cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, value, l1Value)

		// Wait for write-behind to complete
		time.Sleep(50 * time.Millisecond)

		// L2 should have been called
		mockL2.AssertExpectations(t)
	})
}

func TestCacheWarming(t *testing.T) {
	t.Run("Eager Warming", func(t *testing.T) {
		mockCache := new(MockCache)
		logger := &MockLogger{}

		config := &WarmingConfig{
			Strategy:       WarmingStrategyEager,
			BatchSize:      10,
			ConcurrentJobs: 2,
			WarmingTTL:     1 * time.Hour,
			Prefixes:       []string{"test:"},
			MaxItems:       100,
			Enabled:        true,
		}

		provider := &DefaultWarmingDataProvider{cache: mockCache}
		warmer := NewCacheWarmer(mockCache, config, provider, logger)

		ctx := context.Background()

		// Test warming (no actual data to warm with default provider)
		err := warmer.Start(ctx)
		assert.NoError(t, err)

		warmer.Stop()
	})

	t.Run("Critical Data Warming", func(t *testing.T) {
		mockCache := new(MockCache)
		logger := &MockLogger{}

		config := &WarmingConfig{
			Strategy:       WarmingStrategyEager,
			BatchSize:      10,
			ConcurrentJobs: 2,
			WarmingTTL:     1 * time.Hour,
			MaxItems:       100,
			Enabled:        true,
		}

		provider := &DefaultWarmingDataProvider{cache: mockCache}
		warmer := NewCacheWarmer(mockCache, config, provider, logger)

		ctx := context.Background()

		// Set up mock expectations for critical keys
		// The DefaultWarmingDataProvider returns these critical keys:
		// - "config:app_settings"
		// - "plan:active:all"
		// - "server:active:all"
		mockCache.On("Exists", ctx, "config:app_settings").Return(false, nil)
		mockCache.On("Exists", ctx, "plan:active:all").Return(false, nil)
		mockCache.On("Exists", ctx, "server:active:all").Return(false, nil)

		// Test critical data warming
		err := warmer.WarmCriticalData(ctx)
		assert.NoError(t, err)

		metrics := warmer.GetMetrics()
		assert.NotNil(t, metrics)

		// Verify that all expected Exists calls were made
		mockCache.AssertExpectations(t)
	})
}

func TestEventDrivenInvalidation(t *testing.T) {
	t.Run("User Event Invalidation", func(t *testing.T) {
		mockCache := new(MockCache)
		cacheKeys := NewAllCacheKeys()
		logger := &MockLogger{}

		config := &CacheInvalidationConfig{
			Enabled:       true,
			AsyncMode:     false, // Synchronous for testing
			BatchSize:     100,
			BufferTimeout: 1 * time.Second,
		}

		invalidator := NewEventDrivenInvalidator(mockCache, cacheKeys, config, logger)

		// Create a user event using cache package's own event interface
		userEvent := &UserCacheEvent{
			ID:     "test_event_123",
			Type:   "user.updated",
			UserID: 123,
			Time:   time.Now(),
			Data:   map[string]any{"updated_field": "value"},
		}

		ctx := context.Background()

		// Mock cache delete operations
		expectedPattern := cacheKeys.User.UserPattern(123)
		mockCache.On("DeleteByPattern", ctx, expectedPattern).Return(nil)

		// Test event handling
		err := invalidator.Handle(ctx, userEvent)
		assert.NoError(t, err)

		// Wait a bit for async processing if any
		time.Sleep(10 * time.Millisecond)

		mockCache.AssertExpectations(t)
		invalidator.Stop()
	})
}

func TestMultiLevelCacheManager(t *testing.T) {
	t.Run("Manager Creation and Basic Operations", func(t *testing.T) {
		mockL2 := new(MockCache)
		mockCollector := new(MockMetricsCollector)
		cacheKeys := NewAllCacheKeys()
		logger := &MockLogger{}

		config := &MultiLevelCacheConfig{
			EnableL1:         true,
			EnableL2:         true,
			WriteStrategy:    WriteStrategyThrough,
			ReadStrategy:     ReadStrategyFailover,
			PromotionRatio:   0.8,
			ReplicationDelay: 100 * time.Millisecond,
			L1Config: &MemoryCacheConfig{
				MaxSize:         100,
				DefaultTTL:      1 * time.Minute,
				EvictionPolicy:  EvictionPolicyLRU,
				CleanupInterval: 10 * time.Second,
			},
			L2Config: &CacheConfig{
				DefaultTTL:      10 * time.Minute,
				MaxRetries:      3,
				RetryDelay:      100 * time.Millisecond,
				EnableMetrics:   true,
				MetricsPrefix:   "l2",
				CompressionType: "gzip",
			},
		}

		manager := NewMultiLevelCacheManager(config, mockL2, cacheKeys, mockCollector, logger)

		// Test manager methods
		assert.NotNil(t, manager.GetMultiLevelCache())
		assert.NotNil(t, manager.GetL1Cache())
		assert.NotNil(t, manager.GetL2Cache())
		assert.NotNil(t, manager.GetWarmer())
		assert.NotNil(t, manager.GetInvalidator())
		assert.NotNil(t, manager.GetMonitor())

		// Test strategy switching
		err := manager.SwitchStrategy(WriteStrategyBehind, ReadStrategyPromotion)
		assert.NoError(t, err)

		// Test stats
		mockCollector.On("GetMetrics").Return(&Metrics{
			Hits:     100,
			Misses:   20,
			Sets:     50,
			HitRate:  83.3,
			TotalOps: 120,
		})

		stats, err := manager.GetStats(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		if defaultManager, ok := manager.(*DefaultMultiLevelCacheManager); ok {
			defaultManager.Stop()
		}
	})
}

func TestCacheConfiguration(t *testing.T) {
	t.Run("Configuration Manager", func(t *testing.T) {
		manager := NewCacheConfigManager()

		// Test getting predefined config
		config, err := manager.GetConfig("hot_data")
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "hot_data", config.Name)

		// Test domain config
		domainConfig, err := manager.GetDomainConfig(CachePrefixUser)
		assert.NoError(t, err)
		assert.NotNil(t, domainConfig)

		// Test adding custom config
		customConfig := &CacheLayerConfig{
			Name:            "custom",
			Enabled:         true,
			TTL:             30 * time.Minute,
			MaxSize:         500,
			EvictionPolicy:  EvictionPolicyLFU,
			CompressionType: "gzip",
			Strategies: &CacheStrategies{
				ReadStrategy:       ReadStrategyFailover,
				WriteStrategy:      WriteStrategyThrough,
				WarmingStrategy:    WarmingStrategyLazy,
				PromotionRatio:     0.5,
				WarmingBatchSize:   50,
				WarmingConcurrency: 3,
			},
		}

		manager.AddConfig("custom", customConfig)

		retrieved, err := manager.GetConfig("custom")
		assert.NoError(t, err)
		assert.Equal(t, customConfig, retrieved)
	})

	t.Run("Configuration Validation", func(t *testing.T) {
		validator := NewConfigurationValidator()

		// Valid configuration
		validConfig := &CacheLayerConfig{
			Name:            "valid",
			Enabled:         true,
			TTL:             1 * time.Hour,
			MaxSize:         1000,
			EvictionPolicy:  EvictionPolicyLRU,
			CompressionType: "gzip",
			Strategies: &CacheStrategies{
				ReadStrategy:       ReadStrategyFailover,
				WriteStrategy:      WriteStrategyThrough,
				WarmingStrategy:    WarmingStrategyLazy,
				PromotionRatio:     0.7,
				WarmingBatchSize:   100,
				WarmingConcurrency: 5,
			},
		}

		err := validator.ValidateConfig(validConfig)
		assert.NoError(t, err)

		// Invalid configuration - empty name
		invalidConfig := *validConfig
		invalidConfig.Name = ""
		err = validator.ValidateConfig(&invalidConfig)
		assert.Error(t, err)

		// Invalid configuration - negative TTL
		invalidConfig = *validConfig
		invalidConfig.TTL = -1 * time.Hour
		err = validator.ValidateConfig(&invalidConfig)
		assert.Error(t, err)
	})
}

// Benchmark tests
func BenchmarkMemoryCache(b *testing.B) {
	config := &MemoryCacheConfig{
		MaxSize:         10000,
		DefaultTTL:      1 * time.Hour,
		EvictionPolicy:  EvictionPolicyLRU,
		CleanupInterval: 1 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	ctx := context.Background()
	value := []byte("benchmark_value")

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key_%d", i)
			cache.Set(ctx, key, value, 1*time.Hour)
		}
	})

	// Pre-populate for Get benchmark
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("get_key_%d", i)
		cache.Set(ctx, key, value, 1*time.Hour)
	}

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("get_key_%d", i%1000)
			cache.Get(ctx, key)
		}
	})
}

func BenchmarkMultiLevelCache(b *testing.B) {
	mockL2 := new(MockCache)
	mockCollector := new(MockMetricsCollector)
	logger := &MockLogger{}

	config := &MultiLevelCacheConfig{
		EnableL1:         true,
		EnableL2:         true,
		WriteStrategy:    WriteStrategyThrough,
		ReadStrategy:     ReadStrategyFailover,
		PromotionRatio:   0.8,
		ReplicationDelay: 100 * time.Millisecond,
		L1Config: &MemoryCacheConfig{
			MaxSize:         10000,
			DefaultTTL:      1 * time.Hour,
			EvictionPolicy:  EvictionPolicyLRU,
			CleanupInterval: 1 * time.Minute,
		},
	}

	mlc := NewMultiLevelCache(config, mockL2, mockCollector, logger)
	defer mlc.Close()

	ctx := context.Background()
	value := []byte("benchmark_value")

	// Setup mocks for benchmark
	mockL2.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockCollector.On("RecordSet", mock.Anything, mock.Anything, mock.Anything).Return()
	mockCollector.On("RecordHit", mock.Anything, mock.Anything).Return()

	b.Run("MultiLevel_Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("ml_key_%d", i)
			mlc.Set(ctx, key, value, 1*time.Hour)
		}
	})

	// Pre-populate for Get benchmark
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("ml_get_key_%d", i)
		mlc.Set(ctx, key, value, 1*time.Hour)
	}

	b.Run("MultiLevel_Get_L1_Hit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("ml_get_key_%d", i%1000)
			mlc.Get(ctx, key)
		}
	})
}
