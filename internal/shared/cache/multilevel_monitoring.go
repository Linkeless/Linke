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
	Overall     string                  `json:"overall"`
	Components  map[string]string       `json:"components"`
	Issues      []string                `json:"issues"`
	Metrics     *MultiLevelCacheMetrics `json:"metrics"`
	Performance *DeprecatedPerformanceMetrics `json:"performance"`
	Timestamp   time.Time               `json:"timestamp" swaggertype:"string" format:"date-time" example:"2023-01-01T00:00:00Z"`
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
func (mlcm *MultiLevelCacheMonitor) RegisterRoutes(router *gin.RouterGroup) {
	// DEPRECATED: This functionality has been merged into CacheMonitoringHandler
	// These routes are no longer registered to avoid conflicts
	// The unified CacheMonitoringHandler now provides all cache monitoring endpoints
}

// GetHealth returns comprehensive health status (DEPRECATED - use CacheMonitoringHandler.GetMultiLevelHealth)
func (mlcm *MultiLevelCacheMonitor) GetHealth(c *gin.Context) {
	health := mlcm.checkHealth(c.Request.Context())
	response.Success(c, health)
}

// GetMetrics returns detailed cache metrics (DEPRECATED - use CacheMonitoringHandler.GetMultiLevelMetrics)
func (mlcm *MultiLevelCacheMonitor) GetMetrics(c *gin.Context) {
	metrics := mlcm.multilevelCache.GetMetrics()
	response.Success(c, metrics)
}

// GetPerformanceMetrics returns performance analysis (DEPRECATED - use CacheMonitoringHandler.GetPerformanceMetrics)
func (mlcm *MultiLevelCacheMonitor) GetPerformanceMetrics(c *gin.Context) {
	mlcm.updatePerformanceMetrics()

	mlcm.mu.RLock()
	performance := *mlcm.performanceMetrics
	mlcm.mu.RUnlock()

	response.Success(c, performance)
}

// GetDashboard returns dashboard data (DEPRECATED - use CacheMonitoringHandler.GetDashboard)
func (mlcm *MultiLevelCacheMonitor) GetDashboard(c *gin.Context) {
	dashboard := map[string]interface{}{
		"health":       mlcm.checkHealth(c.Request.Context()),
		"metrics":      mlcm.multilevelCache.GetMetrics(),
		"performance":  mlcm.getPerformanceSnapshot(),
		"warming":      mlcm.getWarmingSnapshot(),
		"invalidation": mlcm.getInvalidationSnapshot(),
		"alerts":       mlcm.checkAlerts(),
	}

	response.Success(c, dashboard)
}

// GetAlerts checks and returns active alerts (DEPRECATED - use CacheMonitoringHandler.GetAlerts)
func (mlcm *MultiLevelCacheMonitor) GetAlerts(c *gin.Context) {
	alerts := mlcm.checkAlerts()
	response.Success(c, alerts)
}

// RunBenchmark runs a performance benchmark (DEPRECATED - use CacheMonitoringHandler.RunBenchmark)
func (mlcm *MultiLevelCacheMonitor) RunBenchmark(c *gin.Context) {
	result := mlcm.runBenchmark(c.Request.Context())
	response.Success(c, result)
}

// GetWarmingStatus returns cache warming status (DEPRECATED - use CacheMonitoringHandler.GetWarmingStatus)
func (mlcm *MultiLevelCacheMonitor) GetWarmingStatus(c *gin.Context) {
	if mlcm.warmer != nil {
		metrics := mlcm.warmer.GetMetrics()
		response.Success(c, metrics)
	} else {
		response.Success(c, map[string]string{"status": "warming not configured"})
	}
}

// TriggerWarming manually triggers cache warming (DEPRECATED - use CacheMonitoringHandler.TriggerWarming)
func (mlcm *MultiLevelCacheMonitor) TriggerWarming(c *gin.Context) {
	var req struct {
		Prefixes []string `json:"prefixes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	if mlcm.warmer == nil {
		response.Error(c, http.StatusServiceUnavailable, 503, "Cache warming not configured")
		return
	}

	results := make(map[string]interface{})
	for _, prefix := range req.Prefixes {
		err := mlcm.warmer.WarmPrefix(c.Request.Context(), prefix)
		if err != nil {
			results[prefix] = map[string]string{"status": "failed", "error": err.Error()}
		} else {
			results[prefix] = map[string]string{"status": "success"}
		}
	}

	response.Success(c, map[string]interface{}{
		"triggered_at": time.Now().Format(time.RFC3339),
		"results":      results,
	})
}

// GetInvalidationMetrics returns cache invalidation metrics (DEPRECATED - use CacheMonitoringHandler.GetInvalidationMetrics)
func (mlcm *MultiLevelCacheMonitor) GetInvalidationMetrics(c *gin.Context) {
	if mlcm.invalidator != nil {
		metrics := mlcm.invalidator.GetMetrics()
		response.Success(c, metrics)
	} else {
		response.Success(c, map[string]string{"status": "invalidation not configured"})
	}
}

// Private methods

func (mlcm *MultiLevelCacheMonitor) checkHealth(ctx context.Context) *DeprecatedCacheHealthStatus {
	health := &DeprecatedCacheHealthStatus{
		Components: make(map[string]string),
		Issues:     make([]string, 0),
		Timestamp:  time.Now(),
	}

	// Check L1 cache health
	if mlcm.multilevelCache.config.EnableL1 && mlcm.multilevelCache.l1Cache != nil {
		l1Metrics := mlcm.multilevelCache.l1Cache.GetMetrics()
		if l1Metrics.entryCount > 0 {
			health.Components["L1_cache"] = "healthy"
		} else {
			health.Components["L1_cache"] = "empty"
		}

		// Check memory usage
		memoryUsage := float64(l1Metrics.currentSize) / float64(l1Metrics.maxSize) * 100
		if memoryUsage > mlcm.alertThresholds.MaxMemoryUsagePercent {
			health.Issues = append(health.Issues, fmt.Sprintf("L1 memory usage high: %.1f%%", memoryUsage))
			health.Components["L1_cache"] = "warning"
		}
	} else {
		health.Components["L1_cache"] = "disabled"
	}

	// Check L2 cache health
	if mlcm.multilevelCache.config.EnableL2 && mlcm.multilevelCache.l2Cache != nil {
		// Try a test operation
		testKey := "health_check_" + fmt.Sprintf("%d", time.Now().Unix())
		err := mlcm.multilevelCache.l2Cache.Set(ctx, testKey, []byte("test"), 1*time.Second)
		if err == nil {
			health.Components["L2_cache"] = "healthy"
			_ = mlcm.multilevelCache.l2Cache.Delete(ctx, testKey)
		} else {
			health.Components["L2_cache"] = "unhealthy"
			health.Issues = append(health.Issues, fmt.Sprintf("L2 cache error: %v", err))
		}
	} else {
		health.Components["L2_cache"] = "disabled"
	}

	// Check overall metrics
	metrics := mlcm.multilevelCache.GetMetrics()
	health.Metrics = metrics

	// Check performance
	mlcm.updatePerformanceMetrics()
	mlcm.mu.RLock()
	health.Performance = mlcm.performanceMetrics
	mlcm.mu.RUnlock()

	// Check alerts
	alerts := mlcm.checkAlerts()
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

func (mlcm *MultiLevelCacheMonitor) updatePerformanceMetrics() {
	mlcm.mu.Lock()
	defer mlcm.mu.Unlock()

	metrics := mlcm.multilevelCache.GetMetrics()

	// Calculate efficiency metrics
	totalOps := metrics.L1Hits + metrics.L2Hits + metrics.TotalMisses
	if totalOps > 0 {
		mlcm.performanceMetrics.MemoryEfficiency = float64(metrics.L1Hits+metrics.L2Hits) / float64(totalOps) * 100
	}

	// Update hourly trends (simplified - should use proper time windows)
	currentHour := time.Now().Hour()
	if totalOps > 0 {
		mlcm.performanceMetrics.HourlyHitRates[currentHour] = metrics.OverallHitRate
		mlcm.performanceMetrics.HourlyMissRates[currentHour] = 100 - metrics.OverallHitRate
	}

	// Update response times (placeholder - would need actual measurement)
	mlcm.performanceMetrics.L1AvgResponseTime = 0.5 // 0.5ms avg for memory cache
	mlcm.performanceMetrics.L2AvgResponseTime = 5.0 // 5ms avg for Redis cache

	// Calculate data freshness (placeholder)
	mlcm.performanceMetrics.DataFreshness = 85.0 // 85% fresh data

	// Calculate cache consistency
	if metrics.L1Metrics != nil && metrics.L2Metrics != nil {
		// Simple consistency score based on similar hit rates
		hitRateDiff := metrics.L1HitRate - metrics.L2HitRate
		if hitRateDiff < 0 {
			hitRateDiff = -hitRateDiff
		}
		mlcm.performanceMetrics.CacheConsistency = 100 - hitRateDiff
	}

	// Update promotion accuracy (placeholder)
	if metrics.Promotions > 0 {
		mlcm.performanceMetrics.PromotionAccuracy = 88.5 // 88.5% accuracy
	}
}

func (mlcm *MultiLevelCacheMonitor) checkAlerts() []string {
	alerts := make([]string, 0)

	metrics := mlcm.multilevelCache.GetMetrics()

	// Check hit rate drop
	if metrics.OverallHitRate < 50.0 {
		alerts = append(alerts, fmt.Sprintf("Low overall hit rate: %.1f%%", metrics.OverallHitRate))
	}

	// Check error rates
	if metrics.L2Metrics != nil && metrics.L2Metrics.ErrorRate > mlcm.alertThresholds.MaxErrorRatePercent {
		alerts = append(alerts, fmt.Sprintf("High L2 error rate: %.1f%%", metrics.L2Metrics.ErrorRate))
	}

	// Check memory usage
	if metrics.L1Metrics != nil {
		memoryUsage := float64(metrics.L1Metrics.currentSize) / float64(metrics.L1Metrics.maxSize) * 100
		if memoryUsage > mlcm.alertThresholds.MaxMemoryUsagePercent {
			alerts = append(alerts, fmt.Sprintf("High L1 memory usage: %.1f%%", memoryUsage))
		}
	}

	return alerts
}

func (mlcm *MultiLevelCacheMonitor) getPerformanceSnapshot() map[string]interface{} {
	mlcm.updatePerformanceMetrics()

	mlcm.mu.RLock()
	defer mlcm.mu.RUnlock()

	return map[string]interface{}{
		"memory_efficiency":  mlcm.performanceMetrics.MemoryEfficiency,
		"l1_response_time":   mlcm.performanceMetrics.L1AvgResponseTime,
		"l2_response_time":   mlcm.performanceMetrics.L2AvgResponseTime,
		"data_freshness":     mlcm.performanceMetrics.DataFreshness,
		"cache_consistency":  mlcm.performanceMetrics.CacheConsistency,
		"promotion_accuracy": mlcm.performanceMetrics.PromotionAccuracy,
	}
}

func (mlcm *MultiLevelCacheMonitor) getWarmingSnapshot() map[string]interface{} {
	if mlcm.warmer == nil {
		return map[string]interface{}{"status": "disabled"}
	}

	metrics := mlcm.warmer.GetMetrics()
	return map[string]interface{}{
		"total_warmed":     metrics.TotalWarmed,
		"success_count":    metrics.SuccessCount,
		"error_count":      metrics.ErrorCount,
		"last_warm_time":   metrics.LastWarmTime,
		"warming_duration": metrics.WarmingDuration,
		"items_per_second": metrics.ItemsPerSecond,
	}
}

func (mlcm *MultiLevelCacheMonitor) getInvalidationSnapshot() map[string]interface{} {
	if mlcm.invalidator == nil {
		return map[string]interface{}{"status": "disabled"}
	}

	metrics := mlcm.invalidator.GetMetrics()
	result := make(map[string]interface{})
	for k, v := range metrics {
		result[k] = v
	}
	return result
}

func (mlcm *MultiLevelCacheMonitor) runBenchmark(ctx context.Context) map[string]interface{} {
	startTime := time.Now()

	// Run simple benchmark
	numOps := 1000
	var getTime, setTime time.Duration

	// Benchmark sets
	setStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		value := []byte(fmt.Sprintf("benchmark_value_%d", i))
		_ = mlcm.multilevelCache.Set(ctx, key, value, 1*time.Minute)
	}
	setTime = time.Since(setStart)

	// Benchmark gets
	getStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		_, _ = mlcm.multilevelCache.Get(ctx, key)
	}
	getTime = time.Since(getStart)

	// Cleanup
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		_ = mlcm.multilevelCache.Delete(ctx, key)
	}

	totalTime := time.Since(startTime)

	return map[string]interface{}{
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
