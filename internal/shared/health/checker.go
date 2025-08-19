package health

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
	StatusUnknown   Status = "unknown"
)

// HealthCheck interface defines the contract for health checks
type HealthCheck interface {
	// Name returns the name of the health check
	Name() string
	
	// Check performs the health check and returns the result
	Check(ctx context.Context) HealthResult
	
	// Timeout returns the maximum duration for this health check
	Timeout() time.Duration
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Name         string                 `json:"name"`
	Status       Status                 `json:"status"`
	Message      string                 `json:"message,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Duration     time.Duration          `json:"duration"`
	ComponentID  string                 `json:"component_id,omitempty"`
	Version      string                 `json:"version,omitempty"`
}

// OverallHealth represents the overall health status of the system
type OverallHealth struct {
	Status      Status                   `json:"status"`
	Timestamp   time.Time                `json:"timestamp"`
	Duration    time.Duration            `json:"duration"`
	Checks      map[string]HealthResult  `json:"checks"`
	Summary     HealthSummary            `json:"summary"`
	SystemInfo  SystemInfo               `json:"system_info,omitempty"`
}

// HealthSummary provides aggregated health information
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Unhealthy int `json:"unhealthy"`
	Degraded  int `json:"degraded"`
	Unknown   int `json:"unknown"`
}

// SystemInfo provides basic system information
type SystemInfo struct {
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	StartTime time.Time `json:"start_time"`
	Uptime    string    `json:"uptime"`
}

// HealthChecker manages and executes health checks
type HealthChecker struct {
	checks     map[string]HealthCheck
	mutex      sync.RWMutex
	systemInfo SystemInfo
	startTime  time.Time
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(service, version string) *HealthChecker {
	startTime := time.Now()
	return &HealthChecker{
		checks:    make(map[string]HealthCheck),
		startTime: startTime,
		systemInfo: SystemInfo{
			Service:   service,
			Version:   version,
			StartTime: startTime,
		},
	}
}

// Register registers a health check
func (hc *HealthChecker) Register(check HealthCheck) error {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	name := check.Name()
	if name == "" {
		return fmt.Errorf("health check name cannot be empty")
	}
	
	if _, exists := hc.checks[name]; exists {
		return fmt.Errorf("health check with name %s already registered", name)
	}
	
	hc.checks[name] = check
	return nil
}

// Unregister removes a health check
func (hc *HealthChecker) Unregister(name string) bool {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	if _, exists := hc.checks[name]; exists {
		delete(hc.checks, name)
		return true
	}
	return false
}

// Check performs all registered health checks
func (hc *HealthChecker) Check(ctx context.Context) OverallHealth {
	startTime := time.Now()
	
	hc.mutex.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mutex.RUnlock()
	
	// Execute health checks concurrently
	results := hc.executeChecks(ctx, checks)
	
	// Calculate overall status and summary
	overallStatus := hc.calculateOverallStatus(results)
	summary := hc.calculateSummary(results)
	
	// Update system info
	systemInfo := hc.systemInfo
	systemInfo.Uptime = time.Since(hc.startTime).String()
	
	return OverallHealth{
		Status:     overallStatus,
		Timestamp:  startTime,
		Duration:   time.Since(startTime),
		Checks:     results,
		Summary:    summary,
		SystemInfo: systemInfo,
	}
}

// CheckSingle performs a specific health check by name
func (hc *HealthChecker) CheckSingle(ctx context.Context, name string) (HealthResult, error) {
	hc.mutex.RLock()
	check, exists := hc.checks[name]
	hc.mutex.RUnlock()
	
	if !exists {
		return HealthResult{}, fmt.Errorf("health check %s not found", name)
	}
	
	return hc.executeCheck(ctx, check), nil
}

// ListChecks returns the names of all registered health checks
func (hc *HealthChecker) ListChecks() []string {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	names := make([]string, 0, len(hc.checks))
	for name := range hc.checks {
		names = append(names, name)
	}
	return names
}

// executeChecks runs all health checks concurrently
func (hc *HealthChecker) executeChecks(ctx context.Context, checks map[string]HealthCheck) map[string]HealthResult {
	results := make(map[string]HealthResult)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	
	for name, check := range checks {
		wg.Add(1)
		go func(name string, check HealthCheck) {
			defer wg.Done()
			
			result := hc.executeCheck(ctx, check)
			
			resultsMutex.Lock()
			results[name] = result
			resultsMutex.Unlock()
		}(name, check)
	}
	
	wg.Wait()
	return results
}

// executeCheck runs a single health check with timeout
func (hc *HealthChecker) executeCheck(ctx context.Context, check HealthCheck) HealthResult {
	startTime := time.Now()
	
	// Create timeout context
	timeout := check.Timeout()
	if timeout <= 0 {
		timeout = 30 * time.Second // Default timeout
	}
	
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Execute the check
	result := check.Check(timeoutCtx)
	
	// Ensure timestamp and duration are set
	result.Timestamp = startTime
	result.Duration = time.Since(startTime)
	
	// Handle timeout
	if timeoutCtx.Err() == context.DeadlineExceeded {
		result.Status = StatusUnhealthy
		result.Error = fmt.Sprintf("health check timed out after %v", timeout)
	}
	
	return result
}

// calculateOverallStatus determines the overall system health
func (hc *HealthChecker) calculateOverallStatus(results map[string]HealthResult) Status {
	if len(results) == 0 {
		return StatusUnknown
	}
	
	hasUnhealthy := false
	hasDegraded := false
	hasUnknown := false
	
	for _, result := range results {
		switch result.Status {
		case StatusUnhealthy:
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		case StatusUnknown:
			hasUnknown = true
		}
	}
	
	// Priority: unhealthy > degraded > unknown > healthy
	if hasUnhealthy {
		return StatusUnhealthy
	}
	if hasDegraded {
		return StatusDegraded
	}
	if hasUnknown {
		return StatusUnknown
	}
	
	return StatusHealthy
}

// calculateSummary calculates health check summary statistics
func (hc *HealthChecker) calculateSummary(results map[string]HealthResult) HealthSummary {
	summary := HealthSummary{Total: len(results)}
	
	for _, result := range results {
		switch result.Status {
		case StatusHealthy:
			summary.Healthy++
		case StatusUnhealthy:
			summary.Unhealthy++
		case StatusDegraded:
			summary.Degraded++
		case StatusUnknown:
			summary.Unknown++
		}
	}
	
	return summary
}

// JSON returns the JSON representation of overall health
func (oh OverallHealth) JSON() ([]byte, error) {
	return json.MarshalIndent(oh, "", "  ")
}

// IsHealthy returns true if the overall status is healthy
func (oh OverallHealth) IsHealthy() bool {
	return oh.Status == StatusHealthy
}

// BaseHealthCheck provides a base implementation for health checks
type BaseHealthCheck struct {
	name      string
	timeout   time.Duration
	checkFunc func(ctx context.Context) HealthResult
}

// NewBaseHealthCheck creates a new base health check
func NewBaseHealthCheck(name string, timeout time.Duration, checkFunc func(ctx context.Context) HealthResult) *BaseHealthCheck {
	return &BaseHealthCheck{
		name:      name,
		timeout:   timeout,
		checkFunc: checkFunc,
	}
}

func (bhc *BaseHealthCheck) Name() string {
	return bhc.name
}

func (bhc *BaseHealthCheck) Timeout() time.Duration {
	return bhc.timeout
}

func (bhc *BaseHealthCheck) Check(ctx context.Context) HealthResult {
	if bhc.checkFunc != nil {
		return bhc.checkFunc(ctx)
	}
	
	return HealthResult{
		Name:    bhc.name,
		Status:  StatusHealthy,
		Message: "Default healthy status",
	}
}

// DatabaseHealthCheck checks database connectivity
type DatabaseHealthCheck struct {
	*BaseHealthCheck
	pingFunc func(ctx context.Context) error
}

// NewDatabaseHealthCheck creates a new database health check
func NewDatabaseHealthCheck(name string, pingFunc func(ctx context.Context) error) *DatabaseHealthCheck {
	base := NewBaseHealthCheck(name, 10*time.Second, nil)
	return &DatabaseHealthCheck{
		BaseHealthCheck: base,
		pingFunc:        pingFunc,
	}
}

func (dhc *DatabaseHealthCheck) Check(ctx context.Context) HealthResult {
	result := HealthResult{
		Name: dhc.name,
	}
	
	if dhc.pingFunc == nil {
		result.Status = StatusUnknown
		result.Error = "ping function not provided"
		return result
	}
	
	err := dhc.pingFunc(ctx)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Database ping failed"
	} else {
		result.Status = StatusHealthy
		result.Message = "Database connection healthy"
	}
	
	return result
}

// RedisHealthCheck checks Redis connectivity
type RedisHealthCheck struct {
	*BaseHealthCheck
	pingFunc func(ctx context.Context) error
}

// NewRedisHealthCheck creates a new Redis health check
func NewRedisHealthCheck(name string, pingFunc func(ctx context.Context) error) *RedisHealthCheck {
	base := NewBaseHealthCheck(name, 5*time.Second, nil)
	return &RedisHealthCheck{
		BaseHealthCheck: base,
		pingFunc:        pingFunc,
	}
}

func (rhc *RedisHealthCheck) Check(ctx context.Context) HealthResult {
	result := HealthResult{
		Name: rhc.name,
	}
	
	if rhc.pingFunc == nil {
		result.Status = StatusUnknown
		result.Error = "ping function not provided"
		return result
	}
	
	err := rhc.pingFunc(ctx)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis ping failed"
	} else {
		result.Status = StatusHealthy
		result.Message = "Redis connection healthy"
	}
	
	return result
}

// URLHealthCheck checks HTTP endpoint availability
type URLHealthCheck struct {
	*BaseHealthCheck
	url        string
	httpClient interface {
		Get(string) (int, error)
	}
}

// NewURLHealthCheck creates a new URL health check
func NewURLHealthCheck(name, url string, httpClient interface{ Get(string) (int, error) }) *URLHealthCheck {
	base := NewBaseHealthCheck(name, 15*time.Second, nil)
	return &URLHealthCheck{
		BaseHealthCheck: base,
		url:             url,
		httpClient:      httpClient,
	}
}

func (uhc *URLHealthCheck) Check(ctx context.Context) HealthResult {
	result := HealthResult{
		Name: uhc.name,
		Details: map[string]interface{}{
			"url": uhc.url,
		},
	}
	
	if uhc.httpClient == nil {
		result.Status = StatusUnknown
		result.Error = "HTTP client not provided"
		return result
	}
	
	statusCode, err := uhc.httpClient.Get(uhc.url)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "HTTP request failed"
	} else if statusCode >= 200 && statusCode < 400 {
		result.Status = StatusHealthy
		result.Message = "URL is accessible"
		result.Details["status_code"] = statusCode
	} else {
		result.Status = StatusUnhealthy
		result.Message = "URL returned unhealthy status code"
		result.Details["status_code"] = statusCode
	}
	
	return result
}

// MemoryHealthCheck checks memory usage
type MemoryHealthCheck struct {
	*BaseHealthCheck
	maxMemoryMB uint64
	getMemUsage func() uint64
}

// NewMemoryHealthCheck creates a new memory health check
func NewMemoryHealthCheck(name string, maxMemoryMB uint64, getMemUsage func() uint64) *MemoryHealthCheck {
	base := NewBaseHealthCheck(name, 5*time.Second, nil)
	return &MemoryHealthCheck{
		BaseHealthCheck: base,
		maxMemoryMB:     maxMemoryMB,
		getMemUsage:     getMemUsage,
	}
}

func (mhc *MemoryHealthCheck) Check(ctx context.Context) HealthResult {
	result := HealthResult{
		Name: mhc.name,
	}
	
	if mhc.getMemUsage == nil {
		result.Status = StatusUnknown
		result.Error = "memory usage function not provided"
		return result
	}
	
	currentMemMB := mhc.getMemUsage()
	usagePercent := float64(currentMemMB) / float64(mhc.maxMemoryMB) * 100
	
	result.Details = map[string]interface{}{
		"current_memory_mb": currentMemMB,
		"max_memory_mb":     mhc.maxMemoryMB,
		"usage_percent":     usagePercent,
	}
	
	if usagePercent >= 90 {
		result.Status = StatusUnhealthy
		result.Message = "Memory usage critically high"
	} else if usagePercent >= 80 {
		result.Status = StatusDegraded
		result.Message = "Memory usage high"
	} else {
		result.Status = StatusHealthy
		result.Message = "Memory usage normal"
	}
	
	return result
}

// DiskHealthCheck checks disk space usage
type DiskHealthCheck struct {
	*BaseHealthCheck
	path            string
	maxUsagePercent float64
	getDiskUsage    func(string) (used, total uint64, err error)
}

// NewDiskHealthCheck creates a new disk health check
func NewDiskHealthCheck(name, path string, maxUsagePercent float64, getDiskUsage func(string) (used, total uint64, err error)) *DiskHealthCheck {
	base := NewBaseHealthCheck(name, 5*time.Second, nil)
	return &DiskHealthCheck{
		BaseHealthCheck: base,
		path:            path,
		maxUsagePercent: maxUsagePercent,
		getDiskUsage:    getDiskUsage,
	}
}

func (dhc *DiskHealthCheck) Check(ctx context.Context) HealthResult {
	result := HealthResult{
		Name: dhc.name,
		Details: map[string]interface{}{
			"path": dhc.path,
		},
	}
	
	if dhc.getDiskUsage == nil {
		result.Status = StatusUnknown
		result.Error = "disk usage function not provided"
		return result
	}
	
	used, total, err := dhc.getDiskUsage(dhc.path)
	if err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Message = "Failed to get disk usage"
		return result
	}
	
	usagePercent := float64(used) / float64(total) * 100
	
	result.Details["used_bytes"] = used
	result.Details["total_bytes"] = total
	result.Details["usage_percent"] = usagePercent
	
	if usagePercent >= dhc.maxUsagePercent {
		result.Status = StatusUnhealthy
		result.Message = "Disk usage critically high"
	} else if usagePercent >= dhc.maxUsagePercent*0.8 {
		result.Status = StatusDegraded
		result.Message = "Disk usage high"
	} else {
		result.Status = StatusHealthy
		result.Message = "Disk usage normal"
	}
	
	return result
}