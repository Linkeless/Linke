package dto

// =============================================================================
// Health Check Response Structures
// =============================================================================

// HealthCheckResponse represents the system health check response
type HealthCheckResponse struct {
	Application ApplicationHealthResponse `json:"application"`
	Database    DatabaseHealthResponse    `json:"database"`
}

// ApplicationHealthResponse represents application-level health status
type ApplicationHealthResponse struct {
	Status  string            `json:"status" example:"healthy"`
	Modules map[string]string `json:"modules"`
}

// DatabaseHealthResponse represents database health status
type DatabaseHealthResponse struct {
	MySQL bool `json:"mysql" example:"true"`
	Redis bool `json:"redis" example:"true"`
}

// =============================================================================
// Server API Response Structures
// =============================================================================

// ServerAPIHealthResponse represents server API health response
type ServerAPIHealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"server-api"`
}

// UniProxyPushResponse represents UniProxy push operation response
type UniProxyPushResponse struct {
	Success   bool   `json:"success" example:"true"`
	Message   string `json:"message" example:"Traffic data processed successfully"`
	Processed int    `json:"processed" example:"100"`
	Timestamp int64  `json:"timestamp" example:"1704067200"`
}

// =============================================================================
// Cache Monitoring Response Structures
// =============================================================================

// CacheDashboardResponse represents cache dashboard response
type CacheDashboardResponse struct {
	Health       CacheHealthResponse        `json:"health"`
	Metrics      CacheMetricsResponse       `json:"metrics"`
	Performance  *CachePerformanceResponse  `json:"performance,omitempty"`
	Warming      *CacheWarmingResponse      `json:"warming,omitempty"`
	Invalidation *CacheInvalidationResponse `json:"invalidation,omitempty"`
	Alerts       []string                   `json:"alerts"`
}

// CacheHealthResponse represents cache health status
type CacheHealthResponse struct {
	Overall     string `json:"overall" example:"healthy"`
	MemoryCache string `json:"memory_cache" example:"healthy"`
	RedisCache  string `json:"redis_cache" example:"healthy"`
	LastChecked int64  `json:"last_checked" example:"1704067200"`
}

// CacheMetricsResponse represents cache metrics
type CacheMetricsResponse struct {
	HitRate       float64 `json:"hit_rate" example:"0.85"`
	MissRate      float64 `json:"miss_rate" example:"0.15"`
	TotalHits     int64   `json:"total_hits" example:"1000"`
	TotalMisses   int64   `json:"total_misses" example:"150"`
	TotalRequests int64   `json:"total_requests" example:"1150"`
	AvgLatency    float64 `json:"avg_latency_ms" example:"2.5"`
}

// CachePerformanceResponse represents cache performance metrics
type CachePerformanceResponse struct {
	MemoryUsage   int64   `json:"memory_usage_bytes" example:"1048576"`
	KeyCount      int64   `json:"key_count" example:"500"`
	EvictionCount int64   `json:"eviction_count" example:"10"`
	ResponseTime  float64 `json:"avg_response_time_ms" example:"1.2"`
	ThroughputQPS float64 `json:"throughput_qps" example:"100.5"`
}

// CacheWarmingResponse represents cache warming status
type CacheWarmingResponse struct {
	InProgress  bool     `json:"in_progress" example:"false"`
	LastWarmed  int64    `json:"last_warmed" example:"1704067200"`
	WarmingKeys []string `json:"warming_keys"`
	SuccessRate float64  `json:"success_rate" example:"0.95"`
	Duration    int64    `json:"duration_ms" example:"5000"`
}

// CacheInvalidationResponse represents cache invalidation status
type CacheInvalidationResponse struct {
	PendingKeys         []string `json:"pending_keys"`
	RecentlyInvalidated []string `json:"recently_invalidated"`
	InvalidationRate    float64  `json:"invalidation_rate" example:"0.05"`
	LastInvalidation    int64    `json:"last_invalidation" example:"1704067200"`
}

// CacheBenchmarkResponse represents cache benchmark results
type CacheBenchmarkResponse struct {
	Duration         int64                      `json:"duration_ms" example:"10000"`
	TotalOperations  int64                      `json:"total_operations" example:"10000"`
	OperationsPerSec float64                    `json:"operations_per_sec" example:"1000.0"`
	AvgLatency       float64                    `json:"avg_latency_ms" example:"1.0"`
	MaxLatency       float64                    `json:"max_latency_ms" example:"10.0"`
	MinLatency       float64                    `json:"min_latency_ms" example:"0.1"`
	Results          map[string]BenchmarkResult `json:"results"`
}

// BenchmarkResult represents individual benchmark operation result
type BenchmarkResult struct {
	OperationType   string  `json:"operation_type" example:"get"`
	TotalOperations int64   `json:"total_operations" example:"5000"`
	SuccessCount    int64   `json:"success_count" example:"4950"`
	FailureCount    int64   `json:"failure_count" example:"50"`
	AvgLatency      float64 `json:"avg_latency_ms" example:"1.2"`
	Throughput      float64 `json:"throughput_ops" example:"500.0"`
}

// =============================================================================
// Statistics Response Structures
// =============================================================================

// TrafficStatsResponse represents subscription traffic statistics
type TrafficStatsResponse struct {
	SubscriptionID uint    `json:"subscription_id" example:"1"`
	UsedBytes      int64   `json:"used_bytes" example:"1073741824"`
	TotalBytes     int64   `json:"total_bytes" example:"10737418240"`
	UsagePercent   float64 `json:"usage_percent" example:"10.0"`
	RemainingBytes int64   `json:"remaining_bytes" example:"9663676416"`
	ResetDate      string  `json:"reset_date" example:"2024-02-01"`
	Period         string  `json:"period" example:"monthly"`
	Status         string  `json:"status" example:"active"`
	LastUpdated    string  `json:"last_updated" example:"2024-01-15T10:30:00Z"`
}

// InvoiceStatisticsResponse represents invoice statistics
type InvoiceStatisticsResponse struct {
	TotalInvoices    int64                  `json:"total_invoices" example:"150"`
	TotalAmount      float64                `json:"total_amount" example:"15000.00"`
	PaidAmount       float64                `json:"paid_amount" example:"12000.00"`
	UnpaidAmount     float64                `json:"unpaid_amount" example:"3000.00"`
	OverdueAmount    float64                `json:"overdue_amount" example:"500.00"`
	StatusBreakdown  InvoiceStatusBreakdown `json:"status_breakdown"`
	MonthlyBreakdown []MonthlyInvoiceStats  `json:"monthly_breakdown"`
	Period           InvoiceStatsPeriod     `json:"period"`
}

// InvoiceStatusBreakdown represents invoice status breakdown
type InvoiceStatusBreakdown struct {
	Draft     int64 `json:"draft" example:"5"`
	Pending   int64 `json:"pending" example:"20"`
	Paid      int64 `json:"paid" example:"120"`
	Overdue   int64 `json:"overdue" example:"3"`
	Cancelled int64 `json:"cancelled" example:"2"`
}

// MonthlyInvoiceStats represents monthly invoice statistics
type MonthlyInvoiceStats struct {
	Month       string  `json:"month" example:"2024-01"`
	Count       int64   `json:"count" example:"12"`
	TotalAmount float64 `json:"total_amount" example:"1200.00"`
	PaidAmount  float64 `json:"paid_amount" example:"1000.00"`
	PaidCount   int64   `json:"paid_count" example:"10"`
}

// InvoiceStatsPeriod represents the period for invoice statistics
type InvoiceStatsPeriod struct {
	FromDate string `json:"from_date" example:"2024-01-01"`
	ToDate   string `json:"to_date" example:"2024-01-31"`
	Duration int    `json:"duration_days" example:"31"`
}
