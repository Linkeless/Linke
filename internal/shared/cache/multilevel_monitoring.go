package cache

// DEPRECATED: This file has been merged into monitoring.go
// The MultiLevelCacheMonitor functionality is now part of the unified CacheMonitoringHandler
// This file is kept for backwards compatibility but should not be used directly

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"
)

// MultiLevelCacheMonitor provides comprehensive monitoring for multi-level cache
type MultiLevelCacheMonitor struct {
	multilevelCache *MultiLevelCache
	warmer          *CacheWarmer
	invalidator     *EventDrivenInvalidator
	logger          logger.Logger

	// Performance tracking
	performanceMetrics *DeprecatedPerformanceMetrics
	alertThresholds    *DeprecatedAlertThresholds

	mu sync.RWMutex
}

// DeprecatedPerformanceMetrics tracks detailed performance metrics (DEPRECATED - use monitoring.go)
type DeprecatedPerformanceMetrics struct {
	StartTime time.Time `json:"start_time" swaggertype:"string" format:"date-time" example:"2023-01-01T00:00:00Z"`

	// Response time metrics (in milliseconds)
	L1AvgResponseTime float64 `json:"l1_avg_response_time"`
	L2AvgResponseTime float64 `json:"l2_avg_response_time"`

	// Cache efficiency
	MemoryEfficiency  float64 `json:"memory_efficiency"`  // (Hits / (Hits + Misses)) * 100
	StorageEfficiency float64 `json:"storage_efficiency"` // Useful data / Total storage * 100

	// Cost metrics
	L1CostPerHit float64 `json:"l1_cost_per_hit"`
	L2CostPerHit float64 `json:"l2_cost_per_hit"`

	// Trend data (last 24h)
	HourlyHitRates  []float64 `json:"hourly_hit_rates"`
	HourlyMissRates []float64 `json:"hourly_miss_rates"`
	HourlyEvictions []int64   `json:"hourly_evictions"`

	// Quality metrics
	DataFreshness     float64 `json:"data_freshness"`     // Average age of cached data
	CacheConsistency  float64 `json:"cache_consistency"`  // L1-L2 consistency score
	PromotionAccuracy float64 `json:"promotion_accuracy"` // Accuracy of promotion decisions
}

// DeprecatedAlertThresholds defines when to trigger alerts (DEPRECATED - use monitoring.go)
type DeprecatedAlertThresholds struct {
	MaxL1HitRateDropPercent     float64 `json:"max_l1_hit_rate_drop_percent"`
	MaxL2ResponseTimeMs         float64 `json:"max_l2_response_time_ms"`
	MaxMemoryUsagePercent       float64 `json:"max_memory_usage_percent"`
	MaxErrorRatePercent         float64 `json:"max_error_rate_percent"`
	MinPromotionAccuracyPercent float64 `json:"min_promotion_accuracy_percent"`
}

// DeprecatedCacheHealthStatus represents the overall health of the cache system (DEPRECATED - use monitoring.go)
type DeprecatedCacheHealthStatus struct {
	Overall     string                        `json:"overall"`
	Components  map[string]string             `json:"components"`
	Issues      []string                      `json:"issues"`
	Metrics     *MultiLevelCacheMetrics       `json:"metrics"`
	Performance *DeprecatedPerformanceMetrics `json:"performance"`
	Timestamp   time.Time                     `json:"timestamp" swaggertype:"string" format:"date-time" example:"2023-01-01T00:00:00Z"`
}

// NewMultiLevelCacheMonitor creates a new multi-level cache monitor
func NewMultiLevelCacheMonitor(
	multilevelCache *MultiLevelCache,
	warmer *CacheWarmer,
	invalidator *EventDrivenInvalidator,
	logger logger.Logger,
) *MultiLevelCacheMonitor {
	return &MultiLevelCacheMonitor{
		multilevelCache: multilevelCache,
		warmer:          warmer,
		invalidator:     invalidator,
		logger:          logger,
		performanceMetrics: &DeprecatedPerformanceMetrics{
			StartTime:       time.Now(),
			HourlyHitRates:  make([]float64, 24),
			HourlyMissRates: make([]float64, 24),
			HourlyEvictions: make([]int64, 24),
		},
		alertThresholds: &DeprecatedAlertThresholds{
			MaxL1HitRateDropPercent:     20.0,
			MaxL2ResponseTimeMs:         100.0,
			MaxMemoryUsagePercent:       85.0,
			MaxErrorRatePercent:         5.0,
			MinPromotionAccuracyPercent: 80.0,
		},
	}
}

// RegisterRoutes registers monitoring endpoints (DEPRECATED - functionality moved to CacheMonitoringHandler)
func (m *MultiLevelCacheMonitor) RegisterRoutes(router *gin.RouterGroup) {
	// DEPRECATED: This functionality has been merged into CacheMonitoringHandler
	// These routes are no longer registered to avoid conflicts
	// The unified CacheMonitoringHandler now provides all cache monitoring endpoints
}

// GetHealth returns comprehensive health status (DEPRECATED - use CacheMonitoringHandler.GetMultiLevelHealth)
func (m *MultiLevelCacheMonitor) GetHealth(c *gin.Context) {
	health := m.checkHealth(c.Request.Context())
	response.Success(c, health)
}

// GetMetrics returns detailed cache metrics (DEPRECATED - use CacheMonitoringHandler.GetMultiLevelMetrics)
func (m *MultiLevelCacheMonitor) GetMetrics(c *gin.Context) {
	metrics := m.multilevelCache.GetMetrics()
	response.Success(c, metrics)
}

// GetPerformanceMetrics returns performance analysis (DEPRECATED - use CacheMonitoringHandler.GetPerformanceMetrics)
func (m *MultiLevelCacheMonitor) GetPerformanceMetrics(c *gin.Context) {
	m.updatePerformanceMetrics()

	m.mu.RLock()
	performance := *m.performanceMetrics
	m.mu.RUnlock()

	response.Success(c, performance)
}

// GetDashboard returns dashboard data (DEPRECATED - use CacheMonitoringHandler.GetDashboard)
func (m *MultiLevelCacheMonitor) GetDashboard(c *gin.Context) {
	dashboard := map[string]any{
		"health":       m.checkHealth(c.Request.Context()),
		"metrics":      m.multilevelCache.GetMetrics(),
		"performance":  m.getPerformanceSnapshot(),
		"warming":      m.getWarmingSnapshot(),
		"invalidation": m.getInvalidationSnapshot(),
		"alerts":       m.checkAlerts(),
	}

	response.Success(c, dashboard)
}

// GetAlerts checks and returns active alerts (DEPRECATED - use CacheMonitoringHandler.GetAlerts)
func (m *MultiLevelCacheMonitor) GetAlerts(c *gin.Context) {
	alerts := m.checkAlerts()
	response.Success(c, alerts)
}

// RunBenchmark runs a performance benchmark (DEPRECATED - use CacheMonitoringHandler.RunBenchmark)
func (m *MultiLevelCacheMonitor) RunBenchmark(c *gin.Context) {
	result := m.runBenchmark(c.Request.Context())
	response.Success(c, result)
}

// GetWarmingStatus returns cache warming status (DEPRECATED - use CacheMonitoringHandler.GetWarmingStatus)
func (m *MultiLevelCacheMonitor) GetWarmingStatus(c *gin.Context) {
	if m.warmer != nil {
		metrics := m.warmer.GetMetrics()
		response.Success(c, metrics)
	} else {
		response.Success(c, map[string]string{"status": "warming not configured"})
	}
}

// TriggerWarming manually triggers cache warming (DEPRECATED - use CacheMonitoringHandler.TriggerWarming)
func (m *MultiLevelCacheMonitor) TriggerWarming(c *gin.Context) {
	var req struct {
		Prefixes []string `json:"prefixes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	if m.warmer == nil {
		response.Error(c, http.StatusServiceUnavailable, 503, "Cache warming not configured")
		return
	}

	results := make(map[string]any)
	for _, prefix := range req.Prefixes {
		err := m.warmer.WarmPrefix(c.Request.Context(), prefix)
		if err != nil {
			results[prefix] = map[string]string{"status": "failed", "error": err.Error()}
		} else {
			results[prefix] = map[string]string{"status": "success"}
		}
	}

	response.Success(c, map[string]any{
		"triggered_at": time.Now().Format(time.RFC3339),
		"results":      results,
	})
}

// GetInvalidationMetrics returns cache invalidation metrics (DEPRECATED - use CacheMonitoringHandler.GetInvalidationMetrics)
func (m *MultiLevelCacheMonitor) GetInvalidationMetrics(c *gin.Context) {
	if m.invalidator != nil {
		metrics := m.invalidator.GetMetrics()
		response.Success(c, metrics)
	} else {
		response.Success(c, map[string]string{"status": "invalidation not configured"})
	}
}

// Private methods

func (m *MultiLevelCacheMonitor) checkHealth(ctx context.Context) *DeprecatedCacheHealthStatus {
	health := &DeprecatedCacheHealthStatus{
		Components: make(map[string]string),
		Issues:     make([]string, 0),
		Timestamp:  time.Now(),
	}

	// Check L1 cache health
	if m.multilevelCache.config.EnableL1 && m.multilevelCache.l1Cache != nil {
		l1Metrics := m.multilevelCache.l1Cache.GetMetrics()
		if l1Metrics.entryCount > 0 {
			health.Components["L1_cache"] = "healthy"
		} else {
			health.Components["L1_cache"] = "empty"
		}

		// Check memory usage
		memoryUsage := float64(l1Metrics.currentSize) / float64(l1Metrics.maxSize) * 100
		if memoryUsage > m.alertThresholds.MaxMemoryUsagePercent {
			health.Issues = append(health.Issues, fmt.Sprintf("L1 memory usage high: %.1f%%", memoryUsage))
			health.Components["L1_cache"] = "warning"
		}
	} else {
		health.Components["L1_cache"] = "disabled"
	}

	// Check L2 cache health
	if m.multilevelCache.config.EnableL2 && m.multilevelCache.l2Cache != nil {
		// Try a test operation
		testKey := "health_check_" + fmt.Sprintf("%d", time.Now().Unix())
		err := m.multilevelCache.l2Cache.Set(ctx, testKey, []byte("test"), 1*time.Second)
		if err == nil {
			health.Components["L2_cache"] = "healthy"
			_ = m.multilevelCache.l2Cache.Delete(ctx, testKey)
		} else {
			health.Components["L2_cache"] = "unhealthy"
			health.Issues = append(health.Issues, fmt.Sprintf("L2 cache error: %v", err))
		}
	} else {
		health.Components["L2_cache"] = "disabled"
	}

	// Check overall metrics
	metrics := m.multilevelCache.GetMetrics()
	health.Metrics = metrics

	// Check performance
	m.updatePerformanceMetrics()
	m.mu.RLock()
	health.Performance = m.performanceMetrics
	m.mu.RUnlock()

	// Check alerts
	alerts := m.checkAlerts()
	health.Issues = append(health.Issues, alerts...)

	// Determine overall health
	if len(health.Issues) == 0 {
		health.Overall = "healthy"
	} else if len(health.Issues) <= 2 {
		health.Overall = "warning"
	} else {
		health.Overall = "unhealthy"
	}

	return health
}

func (m *MultiLevelCacheMonitor) updatePerformanceMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	metrics := m.multilevelCache.GetMetrics()

	// Calculate efficiency metrics
	totalOps := metrics.L1Hits + metrics.L2Hits + metrics.TotalMisses
	if totalOps > 0 {
		m.performanceMetrics.MemoryEfficiency = float64(metrics.L1Hits+metrics.L2Hits) / float64(totalOps) * 100
	}

	// Update hourly trends (simplified - should use proper time windows)
	currentHour := time.Now().Hour()
	if totalOps > 0 {
		m.performanceMetrics.HourlyHitRates[currentHour] = metrics.OverallHitRate
		m.performanceMetrics.HourlyMissRates[currentHour] = 100 - metrics.OverallHitRate
	}

	// Update response times (placeholder - would need actual measurement)
	m.performanceMetrics.L1AvgResponseTime = 0.5 // 0.5ms avg for memory cache
	m.performanceMetrics.L2AvgResponseTime = 5.0 // 5ms avg for Redis cache

	// Calculate data freshness (placeholder)
	m.performanceMetrics.DataFreshness = 85.0 // 85% fresh data

	// Calculate cache consistency
	if metrics.L1Metrics != nil && metrics.L2Metrics != nil {
		// Simple consistency score based on similar hit rates
		hitRateDiff := metrics.L1HitRate - metrics.L2HitRate
		if hitRateDiff < 0 {
			hitRateDiff = -hitRateDiff
		}
		m.performanceMetrics.CacheConsistency = 100 - hitRateDiff
	}

	// Update promotion accuracy (placeholder)
	if metrics.Promotions > 0 {
		m.performanceMetrics.PromotionAccuracy = 88.5 // 88.5% accuracy
	}
}

func (m *MultiLevelCacheMonitor) checkAlerts() []string {
	alerts := make([]string, 0)

	metrics := m.multilevelCache.GetMetrics()

	// Check hit rate drop
	if metrics.OverallHitRate < 50.0 {
		alerts = append(alerts, fmt.Sprintf("Low overall hit rate: %.1f%%", metrics.OverallHitRate))
	}

	// Check error rates
	if metrics.L2Metrics != nil && metrics.L2Metrics.ErrorRate > m.alertThresholds.MaxErrorRatePercent {
		alerts = append(alerts, fmt.Sprintf("High L2 error rate: %.1f%%", metrics.L2Metrics.ErrorRate))
	}

	// Check memory usage
	if metrics.L1Metrics != nil {
		memoryUsage := float64(metrics.L1Metrics.currentSize) / float64(metrics.L1Metrics.maxSize) * 100
		if memoryUsage > m.alertThresholds.MaxMemoryUsagePercent {
			alerts = append(alerts, fmt.Sprintf("High L1 memory usage: %.1f%%", memoryUsage))
		}
	}

	return alerts
}

func (m *MultiLevelCacheMonitor) getPerformanceSnapshot() map[string]any {
	m.updatePerformanceMetrics()

	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]any{
		"memory_efficiency":  m.performanceMetrics.MemoryEfficiency,
		"l1_response_time":   m.performanceMetrics.L1AvgResponseTime,
		"l2_response_time":   m.performanceMetrics.L2AvgResponseTime,
		"data_freshness":     m.performanceMetrics.DataFreshness,
		"cache_consistency":  m.performanceMetrics.CacheConsistency,
		"promotion_accuracy": m.performanceMetrics.PromotionAccuracy,
	}
}

func (m *MultiLevelCacheMonitor) getWarmingSnapshot() map[string]any {
	if m.warmer == nil {
		return map[string]any{"status": "disabled"}
	}

	metrics := m.warmer.GetMetrics()
	return map[string]any{
		"total_warmed":     metrics.TotalWarmed,
		"success_count":    metrics.SuccessCount,
		"error_count":      metrics.ErrorCount,
		"last_warm_time":   metrics.LastWarmTime,
		"warming_duration": metrics.WarmingDuration,
		"items_per_second": metrics.ItemsPerSecond,
	}
}

func (m *MultiLevelCacheMonitor) getInvalidationSnapshot() map[string]any {
	if m.invalidator == nil {
		return map[string]any{"status": "disabled"}
	}

	metrics := m.invalidator.GetMetrics()
	result := make(map[string]any)
	for k, v := range metrics {
		result[k] = v
	}
	return result
}

func (m *MultiLevelCacheMonitor) runBenchmark(ctx context.Context) map[string]any {
	startTime := time.Now()

	// Run simple benchmark
	numOps := 1000
	var getTime, setTime time.Duration

	// Benchmark sets
	setStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		value := []byte(fmt.Sprintf("benchmark_value_%d", i))
		_ = m.multilevelCache.Set(ctx, key, value, 1*time.Minute)
	}
	setTime = time.Since(setStart)

	// Benchmark gets
	getStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		_, _ = m.multilevelCache.Get(ctx, key)
	}
	getTime = time.Since(getStart)

	// Cleanup
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		_ = m.multilevelCache.Delete(ctx, key)
	}

	totalTime := time.Since(startTime)

	return map[string]any{
		"total_operations":    numOps * 2, // sets + gets
		"total_duration_ms":   totalTime.Milliseconds(),
		"set_duration_ms":     setTime.Milliseconds(),
		"get_duration_ms":     getTime.Milliseconds(),
		"ops_per_second":      float64(numOps*2) / totalTime.Seconds(),
		"avg_set_time_us":     setTime.Microseconds() / int64(numOps),
		"avg_get_time_us":     getTime.Microseconds() / int64(numOps),
		"benchmark_timestamp": time.Now().Format(time.RFC3339),
	}
}
