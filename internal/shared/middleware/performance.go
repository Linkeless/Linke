package middleware

import (
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	atomicCounters "linke/internal/shared/atomic"

	"github.com/gin-gonic/gin"
)

// PerformanceConfig configures the performance monitoring middleware
type PerformanceConfig struct {
	EnableMetrics       bool          `json:"enable_metrics"`
	EnableMemoryMetrics bool          `json:"enable_memory_metrics"`
	EnableGCMetrics     bool          `json:"enable_gc_metrics"`
	SlowRequestThreshold time.Duration `json:"slow_request_threshold"`
	MemorySampleInterval time.Duration `json:"memory_sample_interval"`
	MetricsHeader       string        `json:"metrics_header"`
	DetailedMetrics     bool          `json:"detailed_metrics"`
}

// DefaultPerformanceConfig returns default performance monitoring configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		EnableMetrics:        true,
		EnableMemoryMetrics:  true,
		EnableGCMetrics:      true,
		SlowRequestThreshold: 1 * time.Second,
		MemorySampleInterval: 10 * time.Second,
		MetricsHeader:        "X-Performance-Metrics",
		DetailedMetrics:      false,
	}
}

// PerformanceMetrics holds performance monitoring data
type PerformanceMetrics struct {
	// Request metrics
	requestCount     *atomicCounters.Counter
	requestDuration  *atomicCounters.Counter // in nanoseconds
	slowRequestCount *atomicCounters.Counter
	errorCount       *atomicCounters.Counter

	// Memory metrics
	allocBytes     *atomicCounters.Counter
	totalAllocated *atomicCounters.Counter
	heapObjects    *atomicCounters.Counter
	heapBytes      *atomicCounters.Counter

	// GC metrics
	gcRuns       *atomicCounters.Counter
	gcPauseTotal *atomicCounters.Counter // in nanoseconds
	nextGC       *atomicCounters.Counter

	// System metrics
	goroutines *atomicCounters.Counter
	startTime  time.Time

	config *PerformanceConfig
}

// NewPerformanceMetrics creates a new performance metrics instance
func NewPerformanceMetrics(config *PerformanceConfig) *PerformanceMetrics {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	group := atomicCounters.NewCounterGroup("performance")

	pm := &PerformanceMetrics{
		requestCount:     group.AddCounter("request_count"),
		requestDuration:  group.AddCounter("request_duration_ns"),
		slowRequestCount: group.AddCounter("slow_request_count"),
		errorCount:       group.AddCounter("error_count"),
		allocBytes:       group.AddCounter("alloc_bytes"),
		totalAllocated:   group.AddCounter("total_allocated_bytes"),
		heapObjects:      group.AddCounter("heap_objects"),
		heapBytes:        group.AddCounter("heap_bytes"),
		gcRuns:           group.AddCounter("gc_runs"),
		gcPauseTotal:     group.AddCounter("gc_pause_total_ns"),
		nextGC:           group.AddCounter("next_gc_bytes"),
		goroutines:       group.AddCounter("goroutines"),
		startTime:        time.Now(),
		config:           config,
	}

	// Start background metrics collection if enabled
	if config.EnableMemoryMetrics || config.EnableGCMetrics {
		go pm.collectSystemMetrics()
	}

	return pm
}

// PerformanceMiddleware returns a Gin middleware for performance monitoring
func PerformanceMiddleware(config *PerformanceConfig) gin.HandlerFunc {
	metrics := NewPerformanceMetrics(config)

	return func(c *gin.Context) {
		if !config.EnableMetrics {
			c.Next()
			return
		}

		start := time.Now()
		
		// Capture memory state before request if enabled
		var memBefore runtime.MemStats
		if config.EnableMemoryMetrics {
			runtime.ReadMemStats(&memBefore)
		}

		// Process request
		c.Next()

		// Calculate metrics after request
		duration := time.Since(start)
		metrics.requestCount.Increment()
		metrics.requestDuration.IncrementBy(duration.Nanoseconds())

		// Check for slow requests
		if duration > config.SlowRequestThreshold {
			metrics.slowRequestCount.Increment()
		}

		// Count errors
		if c.Writer.Status() >= 400 {
			metrics.errorCount.Increment()
		}

		// Capture memory metrics if enabled
		if config.EnableMemoryMetrics {
			var memAfter runtime.MemStats
			runtime.ReadMemStats(&memAfter)
			
			// Track memory allocation during request
			allocDiff := int64(memAfter.Alloc - memBefore.Alloc)
			if allocDiff > 0 {
				metrics.allocBytes.IncrementBy(allocDiff)
			}
		}

		// Add performance headers if enabled
		if config.MetricsHeader != "" {
			if config.DetailedMetrics {
				c.Header(config.MetricsHeader, metrics.getDetailedMetricsHeader())
			} else {
				c.Header(config.MetricsHeader, fmt.Sprintf("duration=%dms", duration.Milliseconds()))
			}
		}
	}
}

// collectSystemMetrics runs in background to collect system-level metrics
func (pm *PerformanceMetrics) collectSystemMetrics() {
	ticker := time.NewTicker(pm.config.MemorySampleInterval)
	defer ticker.Stop()

	for range ticker.C {
		if pm.config.EnableMemoryMetrics || pm.config.EnableGCMetrics {
			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)

			if pm.config.EnableMemoryMetrics {
				pm.totalAllocated.Store(int64(mem.TotalAlloc))
				pm.heapObjects.Store(int64(mem.HeapObjects))
				pm.heapBytes.Store(int64(mem.HeapInuse))
				pm.allocBytes.Store(int64(mem.Alloc))
			}

			if pm.config.EnableGCMetrics {
				pm.gcRuns.Store(int64(mem.NumGC))
				pm.gcPauseTotal.Store(int64(mem.PauseTotalNs))
				pm.nextGC.Store(int64(mem.NextGC))
			}
		}

		// Update goroutine count
		pm.goroutines.Store(int64(runtime.NumGoroutine()))
	}
}

// getDetailedMetricsHeader returns detailed metrics as a header string
func (pm *PerformanceMetrics) getDetailedMetricsHeader() string {
	return fmt.Sprintf("requests=%d,errors=%d,slow=%d,goroutines=%d,alloc=%dKB",
		pm.requestCount.Load(),
		pm.errorCount.Load(),
		pm.slowRequestCount.Load(),
		pm.goroutines.Load(),
		pm.allocBytes.Load()/1024,
	)
}

// GetMetrics returns current performance metrics
func (pm *PerformanceMetrics) GetMetrics() PerformanceSnapshot {
	return PerformanceSnapshot{
		Timestamp: time.Now(),
		Uptime:    time.Since(pm.startTime),
		Requests: RequestMetrics{
			Total:         pm.requestCount.Load(),
			TotalDuration: time.Duration(pm.requestDuration.Load()),
			AverageDuration: func() time.Duration {
				total := pm.requestCount.Load()
				if total == 0 {
					return 0
				}
				return time.Duration(pm.requestDuration.Load() / total)
			}(),
			SlowRequests: pm.slowRequestCount.Load(),
			Errors:       pm.errorCount.Load(),
		},
		Memory: MemoryMetrics{
			AllocatedBytes: pm.allocBytes.Load(),
			TotalAllocated: pm.totalAllocated.Load(),
			HeapObjects:    pm.heapObjects.Load(),
			HeapBytes:      pm.heapBytes.Load(),
		},
		GC: GCMetrics{
			Runs:       pm.gcRuns.Load(),
			PauseTotal: time.Duration(pm.gcPauseTotal.Load()),
			NextGC:     pm.nextGC.Load(),
		},
		System: SystemMetrics{
			Goroutines: pm.goroutines.Load(),
			NumCPU:     int64(runtime.NumCPU()),
		},
	}
}

// PerformanceSnapshot represents a snapshot of performance metrics
type PerformanceSnapshot struct {
	Timestamp time.Time      `json:"timestamp"`
	Uptime    time.Duration  `json:"uptime"`
	Requests  RequestMetrics `json:"requests"`
	Memory    MemoryMetrics  `json:"memory"`
	GC        GCMetrics      `json:"gc"`
	System    SystemMetrics  `json:"system"`
}

// RequestMetrics holds request-related performance metrics
type RequestMetrics struct {
	Total           int64         `json:"total"`
	TotalDuration   time.Duration `json:"total_duration"`
	AverageDuration time.Duration `json:"average_duration"`
	SlowRequests    int64         `json:"slow_requests"`
	Errors          int64         `json:"errors"`
}

// MemoryMetrics holds memory-related performance metrics
type MemoryMetrics struct {
	AllocatedBytes int64 `json:"allocated_bytes"`
	TotalAllocated int64 `json:"total_allocated"`
	HeapObjects    int64 `json:"heap_objects"`
	HeapBytes      int64 `json:"heap_bytes"`
}

// GCMetrics holds garbage collection performance metrics
type GCMetrics struct {
	Runs       int64         `json:"runs"`
	PauseTotal time.Duration `json:"pause_total"`
	NextGC     int64         `json:"next_gc"`
}

// SystemMetrics holds system-level performance metrics
type SystemMetrics struct {
	Goroutines int64 `json:"goroutines"`
	NumCPU     int64 `json:"num_cpu"`
}

// Reset resets all performance metrics
func (pm *PerformanceMetrics) Reset() {
	pm.requestCount.Reset()
	pm.requestDuration.Reset()
	pm.slowRequestCount.Reset()
	pm.errorCount.Reset()
	pm.allocBytes.Reset()
	pm.totalAllocated.Reset()
	pm.heapObjects.Reset()
	pm.heapBytes.Reset()
	pm.gcRuns.Reset()
	pm.gcPauseTotal.Reset()
	pm.nextGC.Reset()
	pm.goroutines.Reset()
	pm.startTime = time.Now()
}

// PerformanceHandler returns a Gin handler for exposing performance metrics
func PerformanceHandler(metrics *PerformanceMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		snapshot := metrics.GetMetrics()
		c.JSON(200, snapshot)
	}
}

// MemoryStatsHandler returns a Gin handler for detailed memory statistics
func MemoryStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		stats := map[string]interface{}{
			"alloc_bytes":        mem.Alloc,
			"total_alloc_bytes":  mem.TotalAlloc,
			"sys_bytes":          mem.Sys,
			"heap_alloc_bytes":   mem.HeapAlloc,
			"heap_sys_bytes":     mem.HeapSys,
			"heap_objects":       mem.HeapObjects,
			"stack_inuse_bytes":  mem.StackInuse,
			"stack_sys_bytes":    mem.StackSys,
			"gc_runs":            mem.NumGC,
			"gc_pause_total_ns":  mem.PauseTotalNs,
			"gc_pause_recent_ns": func() uint64 {
				if mem.NumGC > 0 {
					return mem.PauseNs[(mem.NumGC+255)%256]
				}
				return 0
			}(),
			"next_gc_bytes":   mem.NextGC,
			"last_gc_time":    time.Unix(0, int64(mem.LastGC)),
			"forced_gc_runs":  mem.NumForcedGC,
			"gc_cpu_fraction": mem.GCCPUFraction,
		}

		c.JSON(200, stats)
	}
}

// GCStatsHandler returns a Gin handler for garbage collection statistics
func GCStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		// Calculate GC statistics
		var gcStats struct {
			NumGC         uint32        `json:"num_gc"`
			NumForcedGC   uint32        `json:"num_forced_gc"`
			TotalPauseNs  uint64        `json:"total_pause_ns"`
			LastGC        time.Time     `json:"last_gc"`
			NextGC        uint64        `json:"next_gc"`
			GCCPUFraction float64       `json:"gc_cpu_fraction"`
			RecentPauses  []uint64      `json:"recent_pauses"`
			AveragePause  time.Duration `json:"average_pause"`
		}

		gcStats.NumGC = mem.NumGC
		gcStats.NumForcedGC = mem.NumForcedGC
		gcStats.TotalPauseNs = mem.PauseTotalNs
		gcStats.LastGC = time.Unix(0, int64(mem.LastGC))
		gcStats.NextGC = mem.NextGC
		gcStats.GCCPUFraction = mem.GCCPUFraction

		// Get recent pause times (last 10)
		recentCount := 10
		if int(mem.NumGC) < recentCount {
			recentCount = int(mem.NumGC)
		}

		gcStats.RecentPauses = make([]uint64, recentCount)
		for i := 0; i < recentCount; i++ {
			idx := (mem.NumGC - uint32(i) + 255) % 256
			gcStats.RecentPauses[i] = mem.PauseNs[idx]
		}

		// Calculate average pause time
		if mem.NumGC > 0 {
			gcStats.AveragePause = time.Duration(mem.PauseTotalNs / uint64(mem.NumGC))
		}

		c.JSON(200, gcStats)
	}
}

// SystemStatsHandler returns a Gin handler for system-level statistics
func SystemStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := map[string]interface{}{
			"goroutines":     runtime.NumGoroutine(),
			"num_cpu":        runtime.NumCPU(),
			"go_version":     runtime.Version(),
			"go_os":          runtime.GOOS,
			"go_arch":        runtime.GOARCH,
			"max_procs":      runtime.GOMAXPROCS(0),
			"cgo_calls":      runtime.NumCgoCall(),
		}

		c.JSON(200, stats)
	}
}

// RequestLatencyMiddleware tracks request latency with percentiles
func RequestLatencyMiddleware() gin.HandlerFunc {
	var (
		requestCount   int64
		totalDuration  int64
		recentLatencies = make([]int64, 0, 1000) // Keep last 1000 requests
		latencyMutex   *atomicCounters.AtomicBool
	)
	
	latencyMutex = atomicCounters.NewAtomicBool("latency_mutex")

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Update metrics atomically
		atomic.AddInt64(&requestCount, 1)
		atomic.AddInt64(&totalDuration, duration.Nanoseconds())

		// Update recent latencies (with simple lock)
		if latencyMutex.CompareAndSwap(false, true) {
			if len(recentLatencies) >= 1000 {
				// Remove oldest entry
				copy(recentLatencies, recentLatencies[1:])
				recentLatencies = recentLatencies[:999]
			}
			recentLatencies = append(recentLatencies, duration.Nanoseconds())
			latencyMutex.Store(false)
		}

		// Add latency header
		c.Header("X-Response-Time", duration.String())
	}
}

// LatencyStatsHandler returns detailed latency statistics
func LatencyStatsHandler() gin.HandlerFunc {
	// This would need a more sophisticated implementation for real percentiles
	// This is a simplified version
	return func(c *gin.Context) {
		stats := map[string]interface{}{
			"message": "Latency statistics endpoint - implement with proper percentile calculation",
			"note":    "This is a placeholder for detailed latency metrics",
		}
		c.JSON(200, stats)
	}
}

// PerformanceConfig methods for runtime configuration

// SetEnableMetrics enables or disables metrics collection
func (pc *PerformanceConfig) SetEnableMetrics(enable bool) {
	pc.EnableMetrics = enable
}

// SetSlowRequestThreshold sets the threshold for slow request detection
func (pc *PerformanceConfig) SetSlowRequestThreshold(threshold time.Duration) {
	pc.SlowRequestThreshold = threshold
}

// SetMetricsHeader sets the header name for performance metrics
func (pc *PerformanceConfig) SetMetricsHeader(header string) {
	pc.MetricsHeader = header
}

// SetDetailedMetrics enables or disables detailed metrics in headers
func (pc *PerformanceConfig) SetDetailedMetrics(enable bool) {
	pc.DetailedMetrics = enable
}