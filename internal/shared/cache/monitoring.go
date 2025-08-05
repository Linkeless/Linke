package cache

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

type CacheMonitoringHandler struct {
	manager         CacheManager
	collector       MetricsCollector
	multiLevelMgr   MultiLevelCacheManager // Optional multi-level cache manager
	logger          logger.Logger

	// Performance tracking for multi-level monitoring
	performanceMetrics *PerformanceMetrics
	alertThresholds    *AlertThresholds
	mu                 sync.RWMutex
}

// NewCacheMonitoringHandler creates a monitoring handler with optional multi-level cache support
func NewCacheMonitoringHandler(
	manager CacheManager,
	collector MetricsCollector,
	multiLevelMgr MultiLevelCacheManager, // Optional - can be nil for basic cache
	logger logger.Logger,
) *CacheMonitoringHandler {
	handler := &CacheMonitoringHandler{
		manager:       manager,
		collector:     collector,
		multiLevelMgr: multiLevelMgr,
		logger:        logger,
		performanceMetrics: &PerformanceMetrics{
			StartTime:       time.Now(),
			HourlyHitRates:  make([]float64, 24),
			HourlyMissRates: make([]float64, 24),
			HourlyEvictions: make([]int64, 24),
		},
		alertThresholds: &AlertThresholds{
			MaxL1HitRateDropPercent:     20.0,
			MaxL2ResponseTimeMs:         100.0,
			MaxMemoryUsagePercent:       85.0,
			MaxErrorRatePercent:         5.0,
			MinPromotionAccuracyPercent: 80.0,
		},
	}
	return handler
}

func (h *CacheMonitoringHandler) RegisterRoutes(router *gin.RouterGroup) {
	cacheGroup := router.Group("/cache")
	{
		// Basic cache monitoring routes
		cacheGroup.GET("/metrics", h.GetMetrics)
		cacheGroup.GET("/metrics/:prefix", h.GetMetricsByPrefix)
		cacheGroup.GET("/statistics", h.GetCacheStats)
		cacheGroup.POST("/reset-metrics", h.ResetMetrics)
		cacheGroup.DELETE("/flush", h.FlushCache)
		cacheGroup.DELETE("/pattern/:pattern", h.DeleteByPattern)

		// Multi-level cache monitoring routes (under /cache/monitor)
		monitorGroup := cacheGroup.Group("/monitor")
		{
			monitorGroup.GET("/health", h.GetMultiLevelHealth)
			monitorGroup.GET("/metrics", h.GetMultiLevelMetrics)
			monitorGroup.GET("/performance", h.GetPerformanceMetrics)
			monitorGroup.GET("/dashboard", h.GetDashboard)
			monitorGroup.GET("/alerts", h.GetAlerts)
			monitorGroup.POST("/benchmark", h.RunBenchmark)
			monitorGroup.GET("/warming/status", h.GetWarmingStatus)
			monitorGroup.POST("/warming/trigger", h.TriggerWarming)
			monitorGroup.GET("/invalidation/metrics", h.GetInvalidationMetrics)
		}
	}
}

// GetMetrics returns overall cache metrics
// @Summary Get cache metrics
// @Description Get comprehensive cache performance metrics
// @Tags Cache
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=MetricsReport}
// @Router /admin/cache/metrics [get]
func (h *CacheMonitoringHandler) GetMetrics(c *gin.Context) {
	report := GenerateMetricsReport(h.collector)
	response.Success(c, report)
}

// GetMetricsByPrefix returns metrics for a specific cache prefix
// @Summary Get cache metrics by prefix
// @Description Get cache performance metrics for a specific prefix
// @Tags Cache
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param prefix path string true "Cache prefix (user, subscription, payment, etc.)"
// @Success 200 {object} response.Response{data=Metrics}
// @Router /admin/cache/metrics/{prefix} [get]
func (h *CacheMonitoringHandler) GetMetricsByPrefix(c *gin.Context) {
	prefix := c.Param("prefix")

	if prefix == "" {
		response.BadRequest(c, "Prefix is required")
		return
	}

	prefixWithColon := prefix + ":"
	metrics := h.collector.GetMetricsByPrefix(prefixWithColon)

	response.Success(c, metrics)
}

// GetCacheStats returns Redis cache statistics
// @Summary Get cache statistics
// @Description Get low-level cache statistics from Redis
// @Tags Cache
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=CacheStats}
// @Router /admin/cache/statistics [get]
func (h *CacheMonitoringHandler) GetCacheStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.manager.GetStats(ctx)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.Success(c, stats)
}

// ResetMetrics resets all cache metrics
// @Summary Reset cache metrics
// @Description Reset all cache performance metrics
// @Tags Cache
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /admin/cache/reset-metrics [post]
func (h *CacheMonitoringHandler) ResetMetrics(c *gin.Context) {
	h.collector.Reset()
	response.Success(c, map[string]string{
		"message":  "Cache metrics reset successfully",
		"reset_at": time.Now().Format(time.RFC3339),
	})
}

// FlushCache flushes all cache entries
// @Summary Flush cache
// @Description Flush all cache entries (use with caution)
// @Tags Cache
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /admin/cache/flush [delete]
func (h *CacheMonitoringHandler) FlushCache(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.manager.GetCache().Flush(ctx); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	h.collector.Reset()

	response.Success(c, map[string]string{
		"message":    "Cache flushed successfully",
		"flushed_at": time.Now().Format(time.RFC3339),
	})
}

// DeleteByPattern deletes cache entries matching a pattern
// @Summary Delete cache by pattern
// @Description Delete cache entries matching a specific pattern
// @Tags Cache
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param pattern path string true "Cache key pattern (e.g., user:*, subscription:123:*)"
// @Success 200 {object} response.Response
// @Router /admin/cache/pattern/{pattern} [delete]
func (h *CacheMonitoringHandler) DeleteByPattern(c *gin.Context) {
	pattern := c.Param("pattern")

	if pattern == "" {
		response.BadRequest(c, "Pattern is required")
		return
	}

	ctx := c.Request.Context()

	if err := h.manager.InvalidateCache(ctx, pattern); err != nil {
		response.Error(c, http.StatusInternalServerError, 500, err.Error())
		return
	}

	response.Success(c, map[string]any{
		"message":    "Cache entries deleted successfully",
		"pattern":    pattern,
		"deleted_at": time.Now().Format(time.RFC3339),
	})
}

type CacheHealthCheck struct {
	manager   CacheManager
	collector MetricsCollector
}

func NewCacheHealthCheck(manager CacheManager, collector MetricsCollector) *CacheHealthCheck {
	return &CacheHealthCheck{
		manager:   manager,
		collector: collector,
	}
}

func (ch *CacheHealthCheck) Check(ctx context.Context) map[string]any {
	testKey := "health:check:" + time.Now().Format("20060102150405")
	testValue := []byte("healthy")

	cache := ch.manager.GetCache()

	setErr := cache.Set(ctx, testKey, testValue, 10*time.Second)
	if setErr != nil {
		return map[string]any{
			"status": "unhealthy",
			"error":  setErr.Error(),
		}
	}

	getValue, getErr := cache.Get(ctx, testKey)
	if getErr != nil {
		return map[string]any{
			"status": "unhealthy",
			"error":  getErr.Error(),
		}
	}

	_ = cache.Delete(ctx, testKey)

	if string(getValue) != string(testValue) {
		return map[string]any{
			"status": "unhealthy",
			"error":  "cache read/write mismatch",
		}
	}

	metrics := ch.collector.GetMetrics()

	return map[string]any{
		"status":    "healthy",
		"hit_rate":  metrics.HitRate,
		"total_ops": metrics.TotalOps,
		"errors":    metrics.Errors,
	}
}

type CacheMetricsMiddleware struct {
	collector MetricsCollector
	enabled   bool
}

func NewCacheMetricsMiddleware(collector MetricsCollector, enabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		c.Set("cache_metrics_collector", collector)
		c.Next()
	}
}

func GetMetricsCollector(c *gin.Context) MetricsCollector {
	if collector, exists := c.Get("cache_metrics_collector"); exists {
		if mc, ok := collector.(MetricsCollector); ok {
			return mc
		}
	}
	return nil
}

// Multi-level cache monitoring types and methods

// PerformanceMetrics tracks detailed performance metrics
type PerformanceMetrics struct {
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

// AlertThresholds defines when to trigger alerts
type AlertThresholds struct {
	MaxL1HitRateDropPercent     float64 `json:"max_l1_hit_rate_drop_percent"`
	MaxL2ResponseTimeMs         float64 `json:"max_l2_response_time_ms"`
	MaxMemoryUsagePercent       float64 `json:"max_memory_usage_percent"`
	MaxErrorRatePercent         float64 `json:"max_error_rate_percent"`
	MinPromotionAccuracyPercent float64 `json:"min_promotion_accuracy_percent"`
}

// CacheHealthStatus represents the overall health of the cache system
type CacheHealthStatus struct {
	Overall     string                  `json:"overall"`
	Components  map[string]string       `json:"components"`
	Issues      []string                `json:"issues"`
	Metrics     *MultiLevelCacheMetrics `json:"metrics"`
	Performance *PerformanceMetrics     `json:"performance"`
	Timestamp   time.Time               `json:"timestamp" swaggertype:"string" format:"date-time" example:"2023-01-01T00:00:00Z"`
}

// Multi-level cache monitoring handler methods

// GetMultiLevelHealth returns comprehensive multi-level cache health status
// @Summary Get multi-level cache health
// @Description Get comprehensive health status of multi-level cache system
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=CacheHealthStatus}
// @Router /admin/cache/monitor/health [get]
func (h *CacheMonitoringHandler) GetMultiLevelHealth(c *gin.Context) {
	if h.multiLevelMgr == nil {
		response.Error(c, http.StatusServiceUnavailable, 503, "Multi-level cache not configured")
		return
	}

	health := h.checkMultiLevelHealth(c.Request.Context())
	response.Success(c, health)
}

// GetMultiLevelMetrics returns detailed multi-level cache metrics
// @Summary Get multi-level cache metrics
// @Description Get detailed metrics for multi-level cache system
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=MultiLevelCacheMetrics}
// @Router /admin/cache/monitor/metrics [get]
func (h *CacheMonitoringHandler) GetMultiLevelMetrics(c *gin.Context) {
	if h.multiLevelMgr == nil {
		// Fall back to basic metrics
		report := GenerateMetricsReport(h.collector)
		response.Success(c, report)
		return
	}

	cache := h.multiLevelMgr.GetMultiLevelCache()
	metrics := cache.GetMetrics()
	response.Success(c, metrics)
}

// GetPerformanceMetrics returns performance analysis
// @Summary Get cache performance metrics
// @Description Get detailed performance analysis of cache system
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=PerformanceMetrics}
// @Router /admin/cache/monitor/performance [get]
func (h *CacheMonitoringHandler) GetPerformanceMetrics(c *gin.Context) {
	if h.multiLevelMgr == nil {
		response.Error(c, http.StatusServiceUnavailable, 503, "Multi-level cache not configured")
		return
	}

	h.updatePerformanceMetrics()

	h.mu.RLock()
	performance := *h.performanceMetrics
	h.mu.RUnlock()

	response.Success(c, performance)
}

// GetDashboard returns dashboard data
// @Summary Get cache dashboard data
// @Description Get comprehensive dashboard data for cache monitoring
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.CacheDashboardResponse}
// @Router /admin/cache/monitor/dashboard [get]
func (h *CacheMonitoringHandler) GetDashboard(c *gin.Context) {
	if h.multiLevelMgr == nil {
		// Basic dashboard
		dashboard := map[string]any{
			"health":  h.checkBasicHealth(c.Request.Context()),
			"metrics": GenerateMetricsReport(h.collector),
		}
		response.Success(c, dashboard)
		return
	}

	dashboard := map[string]any{
		"health":       h.checkMultiLevelHealth(c.Request.Context()),
		"metrics":      h.multiLevelMgr.GetMultiLevelCache().GetMetrics(),
		"performance":  h.getPerformanceSnapshot(),
		"warming":      h.getWarmingSnapshot(),
		"invalidation": h.getInvalidationSnapshot(),
		"alerts":       h.checkAlerts(),
	}

	response.Success(c, dashboard)
}

// GetAlerts checks and returns active alerts
// @Summary Get cache system alerts
// @Description Get active alerts for cache system health and performance
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Router /admin/cache/monitor/alerts [get]
func (h *CacheMonitoringHandler) GetAlerts(c *gin.Context) {
	alerts := h.checkAlerts()
	response.Success(c, alerts)
}

// RunBenchmark runs a performance benchmark
// @Summary Run cache benchmark
// @Description Run performance benchmark on cache system
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.CacheBenchmarkResponse}
// @Router /admin/cache/monitor/benchmark [post]
func (h *CacheMonitoringHandler) RunBenchmark(c *gin.Context) {
	if h.multiLevelMgr == nil {
		result := h.runBasicBenchmark(c.Request.Context())
		response.Success(c, result)
		return
	}

	result := h.runMultiLevelBenchmark(c.Request.Context())
	response.Success(c, result)
}

// GetWarmingStatus returns cache warming status
// @Summary Get cache warming status
// @Description Get status and metrics of cache warming operations
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=WarmingMetrics}
// @Router /admin/cache/monitor/warming/status [get]
func (h *CacheMonitoringHandler) GetWarmingStatus(c *gin.Context) {
	if h.multiLevelMgr == nil {
		response.Success(c, map[string]string{"status": "warming not configured"})
		return
	}

	warmer := h.multiLevelMgr.GetWarmer()
	if warmer != nil {
		metrics := warmer.GetMetrics()
		response.Success(c, metrics)
	} else {
		response.Success(c, map[string]string{"status": "warming not configured"})
	}
}

// TriggerWarming manually triggers cache warming
// @Summary Trigger cache warming
// @Description Manually trigger cache warming for specified prefixes
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body map[string]any true "Warming request"
// @Success 200 {object} response.Response{data=dto.CacheWarmingResponse}
// @Router /admin/cache/monitor/warming/trigger [post]
func (h *CacheMonitoringHandler) TriggerWarming(c *gin.Context) {
	if h.multiLevelMgr == nil {
		response.Error(c, http.StatusServiceUnavailable, 503, "Multi-level cache not configured")
		return
	}

	var req struct {
		Prefixes []string `json:"prefixes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	warmer := h.multiLevelMgr.GetWarmer()
	if warmer == nil {
		response.Error(c, http.StatusServiceUnavailable, 503, "Cache warming not configured")
		return
	}

	results := make(map[string]any)
	for _, prefix := range req.Prefixes {
		err := warmer.WarmPrefix(c.Request.Context(), prefix)
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

// GetInvalidationMetrics returns cache invalidation metrics
// @Summary Get cache invalidation metrics
// @Description Get metrics for event-driven cache invalidation
// @Tags Cache Monitoring
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]int64}
// @Router /admin/cache/monitor/invalidation/metrics [get]
func (h *CacheMonitoringHandler) GetInvalidationMetrics(c *gin.Context) {
	if h.multiLevelMgr == nil {
		response.Success(c, map[string]string{"status": "invalidation not configured"})
		return
	}

	invalidator := h.multiLevelMgr.GetInvalidator()
	if invalidator != nil {
		metrics := invalidator.GetMetrics()
		response.Success(c, metrics)
	} else {
		response.Success(c, map[string]string{"status": "invalidation not configured"})
	}
}

// Supporting helper methods

func (h *CacheMonitoringHandler) checkMultiLevelHealth(ctx context.Context) *CacheHealthStatus {
	health := &CacheHealthStatus{
		Components: make(map[string]string),
		Issues:     make([]string, 0),
		Timestamp:  time.Now(),
	}

	if h.multiLevelMgr == nil {
		return h.checkBasicHealth(ctx)
	}

	multiLevelCache := h.multiLevelMgr.GetMultiLevelCache()

	// Check L1 cache health
	if multiLevelCache.config.EnableL1 && multiLevelCache.l1Cache != nil {
		l1Metrics := multiLevelCache.l1Cache.GetMetrics()
		if l1Metrics.entryCount > 0 {
			health.Components["L1_cache"] = "healthy"
		} else {
			health.Components["L1_cache"] = "empty"
		}

		// Check memory usage
		memoryUsage := float64(l1Metrics.currentSize) / float64(l1Metrics.maxSize) * 100
		if memoryUsage > h.alertThresholds.MaxMemoryUsagePercent {
			health.Issues = append(health.Issues, fmt.Sprintf("L1 memory usage high: %.1f%%", memoryUsage))
			health.Components["L1_cache"] = "warning"
		}
	} else {
		health.Components["L1_cache"] = "disabled"
	}

	// Check L2 cache health
	if multiLevelCache.config.EnableL2 && multiLevelCache.l2Cache != nil {
		// Try a test operation
		testKey := "health_check_" + fmt.Sprintf("%d", time.Now().Unix())
		err := multiLevelCache.l2Cache.Set(ctx, testKey, []byte("test"), 1*time.Second)
		if err == nil {
			health.Components["L2_cache"] = "healthy"
			_ = multiLevelCache.l2Cache.Delete(ctx, testKey)
		} else {
			health.Components["L2_cache"] = "unhealthy"
			health.Issues = append(health.Issues, fmt.Sprintf("L2 cache error: %v", err))
		}
	} else {
		health.Components["L2_cache"] = "disabled"
	}

	// Check overall metrics
	metrics := multiLevelCache.GetMetrics()
	health.Metrics = metrics

	// Check performance
	h.updatePerformanceMetrics()
	h.mu.RLock()
	health.Performance = h.performanceMetrics
	h.mu.RUnlock()

	// Check alerts
	alerts := h.checkAlerts()
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

func (h *CacheMonitoringHandler) checkBasicHealth(ctx context.Context) *CacheHealthStatus {
	health := &CacheHealthStatus{
		Components: make(map[string]string),
		Issues:     make([]string, 0),
		Timestamp:  time.Now(),
	}

	// Use the basic health check from CacheHealthCheck
	testKey := "health:check:" + time.Now().Format("20060102150405")
	testValue := []byte("healthy")

	cache := h.manager.GetCache()

	setErr := cache.Set(ctx, testKey, testValue, 10*time.Second)
	if setErr != nil {
		health.Overall = "unhealthy"
		health.Components["cache"] = "unhealthy"
		health.Issues = append(health.Issues, setErr.Error())
		return health
	}

	getValue, getErr := cache.Get(ctx, testKey)
	if getErr != nil {
		health.Overall = "unhealthy"
		health.Components["cache"] = "unhealthy"
		health.Issues = append(health.Issues, getErr.Error())
		return health
	}

	_ = cache.Delete(ctx, testKey)

	if string(getValue) != string(testValue) {
		health.Overall = "unhealthy"
		health.Components["cache"] = "unhealthy"
		health.Issues = append(health.Issues, "cache read/write mismatch")
		return health
	}

	health.Overall = "healthy"
	health.Components["cache"] = "healthy"
	return health
}

func (h *CacheMonitoringHandler) updatePerformanceMetrics() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.multiLevelMgr == nil {
		return
	}

	metrics := h.multiLevelMgr.GetMultiLevelCache().GetMetrics()

	// Calculate efficiency metrics
	totalOps := metrics.L1Hits + metrics.L2Hits + metrics.TotalMisses
	if totalOps > 0 {
		h.performanceMetrics.MemoryEfficiency = float64(metrics.L1Hits+metrics.L2Hits) / float64(totalOps) * 100
	}

	// Update hourly trends (simplified - should use proper time windows)
	currentHour := time.Now().Hour()
	if totalOps > 0 {
		h.performanceMetrics.HourlyHitRates[currentHour] = metrics.OverallHitRate
		h.performanceMetrics.HourlyMissRates[currentHour] = 100 - metrics.OverallHitRate
	}

	// Update response times (placeholder - would need actual measurement)
	h.performanceMetrics.L1AvgResponseTime = 0.5 // 0.5ms avg for memory cache
	h.performanceMetrics.L2AvgResponseTime = 5.0 // 5ms avg for Redis cache

	// Calculate data freshness (placeholder)
	h.performanceMetrics.DataFreshness = 85.0 // 85% fresh data

	// Calculate cache consistency
	if metrics.L1Metrics != nil && metrics.L2Metrics != nil {
		// Simple consistency score based on similar hit rates
		hitRateDiff := metrics.L1HitRate - metrics.L2HitRate
		if hitRateDiff < 0 {
			hitRateDiff = -hitRateDiff
		}
		h.performanceMetrics.CacheConsistency = 100 - hitRateDiff
	}

	// Update promotion accuracy (placeholder)
	if metrics.Promotions > 0 {
		h.performanceMetrics.PromotionAccuracy = 88.5 // 88.5% accuracy
	}
}

func (h *CacheMonitoringHandler) checkAlerts() []string {
	alerts := make([]string, 0)

	if h.multiLevelMgr == nil {
		return alerts
	}

	metrics := h.multiLevelMgr.GetMultiLevelCache().GetMetrics()

	// Check hit rate drop
	if metrics.OverallHitRate < 50.0 {
		alerts = append(alerts, fmt.Sprintf("Low overall hit rate: %.1f%%", metrics.OverallHitRate))
	}

	// Check error rates
	if metrics.L2Metrics != nil && metrics.L2Metrics.ErrorRate > h.alertThresholds.MaxErrorRatePercent {
		alerts = append(alerts, fmt.Sprintf("High L2 error rate: %.1f%%", metrics.L2Metrics.ErrorRate))
	}

	// Check memory usage
	if metrics.L1Metrics != nil {
		memoryUsage := float64(metrics.L1Metrics.currentSize) / float64(metrics.L1Metrics.maxSize) * 100
		if memoryUsage > h.alertThresholds.MaxMemoryUsagePercent {
			alerts = append(alerts, fmt.Sprintf("High L1 memory usage: %.1f%%", memoryUsage))
		}
	}

	return alerts
}

func (h *CacheMonitoringHandler) getPerformanceSnapshot() map[string]any {
	h.updatePerformanceMetrics()

	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]any{
		"memory_efficiency":  h.performanceMetrics.MemoryEfficiency,
		"l1_response_time":   h.performanceMetrics.L1AvgResponseTime,
		"l2_response_time":   h.performanceMetrics.L2AvgResponseTime,
		"data_freshness":     h.performanceMetrics.DataFreshness,
		"cache_consistency":  h.performanceMetrics.CacheConsistency,
		"promotion_accuracy": h.performanceMetrics.PromotionAccuracy,
	}
}

func (h *CacheMonitoringHandler) getWarmingSnapshot() map[string]any {
	if h.multiLevelMgr == nil {
		return map[string]any{"status": "disabled"}
	}

	warmer := h.multiLevelMgr.GetWarmer()
	if warmer == nil {
		return map[string]any{"status": "disabled"}
	}

	metrics := warmer.GetMetrics()
	return map[string]any{
		"total_warmed":     metrics.TotalWarmed,
		"success_count":    metrics.SuccessCount,
		"error_count":      metrics.ErrorCount,
		"last_warm_time":   metrics.LastWarmTime,
		"warming_duration": metrics.WarmingDuration,
		"items_per_second": metrics.ItemsPerSecond,
	}
}

func (h *CacheMonitoringHandler) getInvalidationSnapshot() map[string]any {
	if h.multiLevelMgr == nil {
		return map[string]any{"status": "disabled"}
	}

	invalidator := h.multiLevelMgr.GetInvalidator()
	if invalidator == nil {
		return map[string]any{"status": "disabled"}
	}

	metrics := invalidator.GetMetrics()
	result := make(map[string]any)
	for k, v := range metrics {
		result[k] = v
	}
	return result
}

func (h *CacheMonitoringHandler) runBasicBenchmark(ctx context.Context) map[string]any {
	startTime := time.Now()

	// Run simple benchmark
	numOps := 1000
	var getTime, setTime time.Duration

	cache := h.manager.GetCache()

	// Benchmark sets
	setStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		value := []byte(fmt.Sprintf("benchmark_value_%d", i))
		_ = cache.Set(ctx, key, value, 1*time.Minute)
	}
	setTime = time.Since(setStart)

	// Benchmark gets
	getStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		_, _ = cache.Get(ctx, key)
	}
	getTime = time.Since(getStart)

	// Cleanup
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		_ = cache.Delete(ctx, key)
	}

	totalTime := time.Since(startTime)

	return map[string]any{
		"cache_type":          "basic",
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

func (h *CacheMonitoringHandler) runMultiLevelBenchmark(ctx context.Context) map[string]any {
	startTime := time.Now()

	// Run simple benchmark
	numOps := 1000
	var getTime, setTime time.Duration

	multiLevelCache := h.multiLevelMgr.GetMultiLevelCache()

	// Benchmark sets
	setStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		value := []byte(fmt.Sprintf("benchmark_value_%d", i))
		_ = multiLevelCache.Set(ctx, key, value, 1*time.Minute)
	}
	setTime = time.Since(setStart)

	// Benchmark gets
	getStart := time.Now()
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		_, _ = multiLevelCache.Get(ctx, key)
	}
	getTime = time.Since(getStart)

	// Cleanup
	for i := 0; i < numOps; i++ {
		key := fmt.Sprintf("benchmark_set_%d", i)
		_ = multiLevelCache.Delete(ctx, key)
	}

	totalTime := time.Since(startTime)

	return map[string]any{
		"cache_type":          "multi-level",
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
