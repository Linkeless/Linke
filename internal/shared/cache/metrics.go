package cache

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type MetricType string

const (
	MetricHit       MetricType = "hit"
	MetricMiss      MetricType = "miss"
	MetricSet       MetricType = "set"
	MetricDelete    MetricType = "delete"
	MetricError     MetricType = "error"
	MetricEviction  MetricType = "eviction"
)

type MetricsCollector interface {
	RecordHit(ctx context.Context, key string)
	RecordMiss(ctx context.Context, key string)
	RecordSet(ctx context.Context, key string, ttl time.Duration)
	RecordDelete(ctx context.Context, key string)
	RecordError(ctx context.Context, key string, operation string)
	RecordEviction(ctx context.Context, key string)
	GetMetrics() *Metrics
	GetMetricsByPrefix(prefix string) *Metrics
	Reset()
}

type Metrics struct {
	Hits       int64   `json:"hits"`
	Misses     int64   `json:"misses"`
	Sets       int64   `json:"sets"`
	Deletes    int64   `json:"deletes"`
	Errors     int64   `json:"errors"`
	Evictions  int64   `json:"evictions"`
	HitRate    float64 `json:"hit_rate"`
	MissRate   float64 `json:"miss_rate"`
	ErrorRate  float64 `json:"error_rate"`
	TotalOps   int64   `json:"total_operations"`
}

type DefaultMetricsCollector struct {
	mu sync.RWMutex
	
	hits      atomic.Int64
	misses    atomic.Int64
	sets      atomic.Int64
	deletes   atomic.Int64
	errors    atomic.Int64
	evictions atomic.Int64
	
	prefixMetrics map[string]*prefixMetric
}

type prefixMetric struct {
	hits      atomic.Int64
	misses    atomic.Int64
	sets      atomic.Int64
	deletes   atomic.Int64
	errors    atomic.Int64
	evictions atomic.Int64
}

func NewMetricsCollector() MetricsCollector {
	return &DefaultMetricsCollector{
		prefixMetrics: make(map[string]*prefixMetric),
	}
}

func (mc *DefaultMetricsCollector) RecordHit(ctx context.Context, key string) {
	mc.hits.Add(1)
	mc.recordPrefixMetric(key, MetricHit)
}

func (mc *DefaultMetricsCollector) RecordMiss(ctx context.Context, key string) {
	mc.misses.Add(1)
	mc.recordPrefixMetric(key, MetricMiss)
}

func (mc *DefaultMetricsCollector) RecordSet(ctx context.Context, key string, ttl time.Duration) {
	mc.sets.Add(1)
	mc.recordPrefixMetric(key, MetricSet)
}

func (mc *DefaultMetricsCollector) RecordDelete(ctx context.Context, key string) {
	mc.deletes.Add(1)
	mc.recordPrefixMetric(key, MetricDelete)
}

func (mc *DefaultMetricsCollector) RecordError(ctx context.Context, key string, operation string) {
	mc.errors.Add(1)
	mc.recordPrefixMetric(key, MetricError)
}

func (mc *DefaultMetricsCollector) RecordEviction(ctx context.Context, key string) {
	mc.evictions.Add(1)
	mc.recordPrefixMetric(key, MetricEviction)
}

func (mc *DefaultMetricsCollector) GetMetrics() *Metrics {
	hits := mc.hits.Load()
	misses := mc.misses.Load()
	sets := mc.sets.Load()
	deletes := mc.deletes.Load()
	errors := mc.errors.Load()
	evictions := mc.evictions.Load()
	
	totalReads := hits + misses
	totalOps := totalReads + sets + deletes
	
	var hitRate, missRate, errorRate float64
	if totalReads > 0 {
		hitRate = float64(hits) / float64(totalReads) * 100
		missRate = float64(misses) / float64(totalReads) * 100
	}
	if totalOps > 0 {
		errorRate = float64(errors) / float64(totalOps) * 100
	}
	
	return &Metrics{
		Hits:      hits,
		Misses:    misses,
		Sets:      sets,
		Deletes:   deletes,
		Errors:    errors,
		Evictions: evictions,
		HitRate:   hitRate,
		MissRate:  missRate,
		ErrorRate: errorRate,
		TotalOps:  totalOps,
	}
}

func (mc *DefaultMetricsCollector) GetMetricsByPrefix(prefix string) *Metrics {
	mc.mu.RLock()
	pm, exists := mc.prefixMetrics[prefix]
	mc.mu.RUnlock()
	
	if !exists {
		return &Metrics{}
	}
	
	hits := pm.hits.Load()
	misses := pm.misses.Load()
	sets := pm.sets.Load()
	deletes := pm.deletes.Load()
	errors := pm.errors.Load()
	evictions := pm.evictions.Load()
	
	totalReads := hits + misses
	totalOps := totalReads + sets + deletes
	
	var hitRate, missRate, errorRate float64
	if totalReads > 0 {
		hitRate = float64(hits) / float64(totalReads) * 100
		missRate = float64(misses) / float64(totalReads) * 100
	}
	if totalOps > 0 {
		errorRate = float64(errors) / float64(totalOps) * 100
	}
	
	return &Metrics{
		Hits:      hits,
		Misses:    misses,
		Sets:      sets,
		Deletes:   deletes,
		Errors:    errors,
		Evictions: evictions,
		HitRate:   hitRate,
		MissRate:  missRate,
		ErrorRate: errorRate,
		TotalOps:  totalOps,
	}
}

func (mc *DefaultMetricsCollector) Reset() {
	mc.hits.Store(0)
	mc.misses.Store(0)
	mc.sets.Store(0)
	mc.deletes.Store(0)
	mc.errors.Store(0)
	mc.evictions.Store(0)
	
	mc.mu.Lock()
	mc.prefixMetrics = make(map[string]*prefixMetric)
	mc.mu.Unlock()
}

func (mc *DefaultMetricsCollector) recordPrefixMetric(key string, metricType MetricType) {
	prefix := extractPrefix(key)
	
	mc.mu.RLock()
	pm, exists := mc.prefixMetrics[prefix]
	mc.mu.RUnlock()
	
	if !exists {
		mc.mu.Lock()
		pm, exists = mc.prefixMetrics[prefix]
		if !exists {
			pm = &prefixMetric{}
			mc.prefixMetrics[prefix] = pm
		}
		mc.mu.Unlock()
	}
	
	switch metricType {
	case MetricHit:
		pm.hits.Add(1)
	case MetricMiss:
		pm.misses.Add(1)
	case MetricSet:
		pm.sets.Add(1)
	case MetricDelete:
		pm.deletes.Add(1)
	case MetricError:
		pm.errors.Add(1)
	case MetricEviction:
		pm.evictions.Add(1)
	}
}

func extractPrefix(key string) string {
	for i, ch := range key {
		if ch == ':' {
			return key[:i]
		}
	}
	return "unknown"
}

type MetricsCacheWrapper struct {
	cache     Cache
	collector MetricsCollector
}

func NewMetricsCacheWrapper(cache Cache, collector MetricsCollector) Cache {
	return &MetricsCacheWrapper{
		cache:     cache,
		collector: collector,
	}
}

func (mw *MetricsCacheWrapper) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := mw.cache.Get(ctx, key)
	
	if err != nil {
		mw.collector.RecordError(ctx, key, "get")
		return nil, err
	}
	
	if result == nil {
		mw.collector.RecordMiss(ctx, key)
	} else {
		mw.collector.RecordHit(ctx, key)
	}
	
	return result, nil
}

func (mw *MetricsCacheWrapper) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := mw.cache.Set(ctx, key, value, ttl)
	
	if err != nil {
		mw.collector.RecordError(ctx, key, "set")
		return err
	}
	
	mw.collector.RecordSet(ctx, key, ttl)
	return nil
}

func (mw *MetricsCacheWrapper) Delete(ctx context.Context, key string) error {
	err := mw.cache.Delete(ctx, key)
	
	if err != nil {
		mw.collector.RecordError(ctx, key, "delete")
		return err
	}
	
	mw.collector.RecordDelete(ctx, key)
	return nil
}

func (mw *MetricsCacheWrapper) DeleteByPattern(ctx context.Context, pattern string) error {
	err := mw.cache.DeleteByPattern(ctx, pattern)
	
	if err != nil {
		mw.collector.RecordError(ctx, pattern, "deleteByPattern")
		return err
	}
	
	return nil
}

func (mw *MetricsCacheWrapper) Exists(ctx context.Context, key string) (bool, error) {
	return mw.cache.Exists(ctx, key)
}

func (mw *MetricsCacheWrapper) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return mw.cache.Expire(ctx, key, ttl)
}

func (mw *MetricsCacheWrapper) Flush(ctx context.Context) error {
	return mw.cache.Flush(ctx)
}

type MetricsReport struct {
	Timestamp     time.Time              `json:"timestamp"`
	GlobalMetrics *Metrics               `json:"global_metrics"`
	PrefixMetrics map[string]*Metrics    `json:"prefix_metrics"`
	TopPrefixes   []PrefixMetricSummary  `json:"top_prefixes"`
}

type PrefixMetricSummary struct {
	Prefix    string  `json:"prefix"`
	HitRate   float64 `json:"hit_rate"`
	TotalOps  int64   `json:"total_operations"`
	ErrorRate float64 `json:"error_rate"`
}

func GenerateMetricsReport(collector MetricsCollector) *MetricsReport {
	globalMetrics := collector.GetMetrics()
	
	prefixMap := make(map[string]*Metrics)
	prefixes := []string{
		CachePrefixUser,
		CachePrefixSubscription,
		CachePrefixPayment,
		CachePrefixAuth,
		CachePrefixPlan,
		CachePrefixInvoice,
		CachePrefixServer,
		CachePrefixCoupon,
		CachePrefixSession,
		CachePrefixConfig,
	}
	
	var topPrefixes []PrefixMetricSummary
	for _, prefix := range prefixes {
		metrics := collector.GetMetricsByPrefix(prefix)
		if metrics.TotalOps > 0 {
			prefixMap[prefix] = metrics
			topPrefixes = append(topPrefixes, PrefixMetricSummary{
				Prefix:    prefix,
				HitRate:   metrics.HitRate,
				TotalOps:  metrics.TotalOps,
				ErrorRate: metrics.ErrorRate,
			})
		}
	}
	
	return &MetricsReport{
		Timestamp:     time.Now(),
		GlobalMetrics: globalMetrics,
		PrefixMetrics: prefixMap,
		TopPrefixes:   topPrefixes,
	}
}

func (mr *MetricsReport) String() string {
	return fmt.Sprintf(
		"Cache Metrics Report (Generated: %s)\n"+
			"============================================\n"+
			"Global Statistics:\n"+
			"  Total Operations: %d\n"+
			"  Hit Rate: %.2f%% (%d hits / %d misses)\n"+
			"  Error Rate: %.2f%% (%d errors)\n"+
			"  Sets: %d, Deletes: %d, Evictions: %d\n"+
			"============================================\n",
		mr.Timestamp.Format("2006-01-02 15:04:05"),
		mr.GlobalMetrics.TotalOps,
		mr.GlobalMetrics.HitRate, mr.GlobalMetrics.Hits, mr.GlobalMetrics.Misses,
		mr.GlobalMetrics.ErrorRate, mr.GlobalMetrics.Errors,
		mr.GlobalMetrics.Sets, mr.GlobalMetrics.Deletes, mr.GlobalMetrics.Evictions,
	)
}