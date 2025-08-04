package cache

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"linke/internal/shared/response"
)

type CacheMonitoringHandler struct {
	manager   CacheManager
	collector MetricsCollector
}

func NewCacheMonitoringHandler(manager CacheManager, collector MetricsCollector) *CacheMonitoringHandler {
	return &CacheMonitoringHandler{
		manager:   manager,
		collector: collector,
	}
}

func (h *CacheMonitoringHandler) RegisterRoutes(router *gin.RouterGroup) {
	cacheGroup := router.Group("/cache")
	{
		cacheGroup.GET("/metrics", h.GetMetrics)
		cacheGroup.GET("/metrics/:prefix", h.GetMetricsByPrefix)
		cacheGroup.GET("/stats", h.GetCacheStats)
		cacheGroup.POST("/reset-metrics", h.ResetMetrics)
		cacheGroup.DELETE("/flush", h.FlushCache)
		cacheGroup.DELETE("/pattern/:pattern", h.DeleteByPattern)
	}
}

// GetMetrics returns overall cache metrics
// @Summary Get cache metrics
// @Description Get comprehensive cache performance metrics
// @Tags Cache
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=MetricsReport}
// @Router /api/v1/admin/cache/metrics [get]
func (h *CacheMonitoringHandler) GetMetrics(c *gin.Context) {
	report := GenerateMetricsReport(h.collector)
	response.Success(c, report)
}

// GetMetricsByPrefix returns metrics for a specific cache prefix
// @Summary Get cache metrics by prefix
// @Description Get cache performance metrics for a specific prefix
// @Tags Cache
// @Accept json
// @Produce json
// @Param prefix path string true "Cache prefix (user, subscription, payment, etc.)"
// @Success 200 {object} response.Response{data=Metrics}
// @Router /api/v1/admin/cache/metrics/{prefix} [get]
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
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=CacheStats}
// @Router /api/v1/admin/cache/stats [get]
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
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/admin/cache/reset-metrics [post]
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
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/admin/cache/flush [delete]
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
// @Accept json
// @Produce json
// @Param pattern path string true "Cache key pattern (e.g., user:*, subscription:123:*)"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/cache/pattern/{pattern} [delete]
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
