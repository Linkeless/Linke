package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/subscription/entities"
)

// UsageTrackingService defines the contract for usage tracking operations
type UsageTrackingService interface {
	// Real-time Usage Tracking
	RecordUsage(ctx context.Context, req *RecordUsageRequest) error
	RecordUsageBatch(ctx context.Context, requests []*RecordUsageRequest) error
	GetCurrentUsage(ctx context.Context, subscriptionID uint, usageType string) (*CurrentUsageResponse, error)
	GetUsageHistory(ctx context.Context, req *UsageHistoryRequest) (*UsageHistoryResponse, error)

	// Usage Aggregation and Reporting
	GetUsageSummary(ctx context.Context, req *UsageSummaryRequest) (*entities.UsageSummary, error)
	GetUsageStatistics(ctx context.Context, req *UsageStatsRequest) (*UsageStatistics, error)
	GetUsageTrends(ctx context.Context, req *UsageTrendsRequest) (*UsageTrendsResponse, error)

	// Usage Predictions
	PredictUsage(ctx context.Context, req *UsagePredictionRequest) (*entities.UsagePrediction, error)
	GetUsagePredictions(ctx context.Context, subscriptionID uint, usageType string) ([]*entities.UsagePrediction, error)

	// Subscription Integration
	UpdateSubscriptionUsage(ctx context.Context, subscriptionID uint) error
	ResetUsageForSubscription(ctx context.Context, subscriptionID uint, usageType string) error
	SyncSubscriptionLimits(ctx context.Context, subscriptionID uint) error

	// Admin Operations
	GetTopUsageSubscriptions(ctx context.Context, req *TopUsageRequest) (*TopUsageResponse, error)
	ExportUsageData(ctx context.Context, req *ExportUsageRequest) (*ExportUsageResponse, error)
	CleanupOldUsageData(ctx context.Context, olderThan time.Time) (*CleanupResult, error)

	// Real-time Monitoring
	GetRealTimeUsage(ctx context.Context, subscriptionID uint) (*RealTimeUsageResponse, error)
	GetUsageAlerts(ctx context.Context, subscriptionID uint) ([]*entities.UsageAlert, error)
}

// UsageAlertService defines the contract for usage alert operations
type UsageAlertService interface {
	// Alert Configuration Management
	CreateAlertConfiguration(ctx context.Context, req *CreateAlertConfigRequest) (*entities.AlertConfiguration, error)
	UpdateAlertConfiguration(ctx context.Context, req *UpdateAlertConfigRequest) (*entities.AlertConfiguration, error)
	DeleteAlertConfiguration(ctx context.Context, configID uint) error
	GetAlertConfiguration(ctx context.Context, configID uint) (*entities.AlertConfiguration, error)
	GetAlertConfigurations(ctx context.Context, req *GetAlertConfigsRequest) (*GetAlertConfigsResponse, error)

	// Alert Processing
	CheckUsageThresholds(ctx context.Context, subscriptionID uint, usageType string) ([]*entities.UsageAlert, error)
	ProcessUsageUpdate(ctx context.Context, subscriptionID uint, usageType string, currentUsage int64) error
	FireAlert(ctx context.Context, req *FireAlertRequest) (*entities.UsageAlert, error)
	ResolveAlert(ctx context.Context, alertID uint, reason string) error

	// Alert Management
	GetUsageAlerts(ctx context.Context, req *GetUsageAlertsRequest) (*GetUsageAlertsResponse, error)
	AcknowledgeAlert(ctx context.Context, alertID uint, acknowledgedBy uint) error
	SuppressAlert(ctx context.Context, alertID uint, duration time.Duration, reason string) error
	BulkResolveAlerts(ctx context.Context, req *BulkResolveAlertsRequest) (*BulkResolveAlertsResponse, error)

	// Notification Management
	SendNotification(ctx context.Context, alert *entities.UsageAlert, channels []entities.NotificationChannel) error
	TestNotificationChannel(ctx context.Context, req *TestNotificationRequest) (*TestNotificationResponse, error)
	UpdateNotificationPreferences(ctx context.Context, req *UpdateNotificationPrefsRequest) error

	// Alert Analytics
	GetAlertStatistics(ctx context.Context, req *AlertStatsRequest) (*AlertStatisticsResponse, error)
	GetAlertHistory(ctx context.Context, req *AlertHistoryRequest) (*AlertHistoryResponse, error)

	// Default Configurations
	CreateDefaultAlertConfigurations(ctx context.Context, subscriptionID uint) error
	CopyAlertConfigurationsFromPlan(ctx context.Context, subscriptionID uint, planID uint) error
}

// Request/Response structures for usage tracking service

// RecordUsageRequest represents a request to record usage
type RecordUsageRequest struct {
	UserSubscriptionID uint      `json:"user_subscription_id" validate:"required"`
	UsageType          string    `json:"usage_type" validate:"required"`
	Amount             int64     `json:"amount" validate:"min=0"`
	Unit               string    `json:"unit"`
	Timestamp          time.Time `json:"timestamp"`
	SourceType         string    `json:"source_type" validate:"required"`
	SourceID           string    `json:"source_id"`
	Metadata           string    `json:"metadata,omitempty"`
	UserAgent          string    `json:"user_agent,omitempty"`
	IPAddress          string    `json:"ip_address,omitempty"`
}

// CurrentUsageResponse represents current usage information
type CurrentUsageResponse struct {
	UserSubscriptionID uint                            `json:"user_subscription_id"`
	UsageByType        map[string]*CurrentUsageDetails `json:"usage_by_type"`
	TotalUsage         int64                           `json:"total_usage"`
	LastUpdated        time.Time                       `json:"last_updated"`
	PeriodStart        time.Time                       `json:"period_start"`
	PeriodEnd          time.Time                       `json:"period_end"`
}

// CurrentUsageDetails represents current usage details for a specific type
type CurrentUsageDetails struct {
	UsageType      string    `json:"usage_type"`
	CurrentUsage   int64     `json:"current_usage"`
	UsageLimit     int64     `json:"usage_limit"`
	UsagePercent   float64   `json:"usage_percent"`
	RemainingUsage int64     `json:"remaining_usage"`
	Unit           string    `json:"unit"`
	LastRecorded   time.Time `json:"last_recorded"`
	RecordCount    int64     `json:"record_count"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	IsUnlimited    bool      `json:"is_unlimited"`
	IsExceeded     bool      `json:"is_exceeded"`
}

// UsageHistoryRequest represents a request for usage history
type UsageHistoryRequest struct {
	UserSubscriptionID uint       `json:"user_subscription_id" validate:"required"`
	UsageType          string     `json:"usage_type,omitempty"`
	StartTime          *time.Time `json:"start_time,omitempty"`
	EndTime            *time.Time `json:"end_time,omitempty"`
	Granularity        string     `json:"granularity"` // hourly, daily, weekly, monthly
	Limit              int        `json:"limit"`
	Offset             int        `json:"offset"`
	IncludeDetails     bool       `json:"include_details"`
	SourceType         string     `json:"source_type,omitempty"`
}

// UsageHistoryResponse represents usage history data
type UsageHistoryResponse struct {
	UserSubscriptionID uint                   `json:"user_subscription_id"`
	UsageHistory       []*UsageHistoryEntry   `json:"usage_history"`
	Summary            *entities.UsageSummary `json:"summary"`
	Pagination         *PaginationInfo        `json:"pagination"`
	Granularity        string                 `json:"granularity"`
	TotalRecords       int64                  `json:"total_records"`
}

// UsageHistoryEntry represents a single entry in usage history
type UsageHistoryEntry struct {
	Period      string                              `json:"period"`
	PeriodStart time.Time                           `json:"period_start"`
	PeriodEnd   time.Time                           `json:"period_end"`
	UsageByType map[string]*entities.UsageTypeStats `json:"usage_by_type"`
	TotalUsage  int64                               `json:"total_usage"`
	RecordCount int64                               `json:"record_count"`
}

// UsageSummaryRequest represents a request for usage summary
type UsageSummaryRequest struct {
	UserSubscriptionID  uint       `json:"user_subscription_id" validate:"required"`
	UsageTypes          []string   `json:"usage_types,omitempty"`
	Period              string     `json:"period"` // daily, weekly, monthly, custom
	PeriodStart         *time.Time `json:"period_start,omitempty"`
	PeriodEnd           *time.Time `json:"period_end,omitempty"`
	IncludeBreakdown    bool       `json:"include_breakdown"`
	IncludePredictions  bool       `json:"include_predictions"`
	CompareWithPrevious bool       `json:"compare_with_previous"`
}

// UsageStatsRequest represents a request for usage statistics
type UsageStatsRequest struct {
	UserSubscriptionID uint       `json:"user_subscription_id" validate:"required"`
	UsageType          string     `json:"usage_type,omitempty"`
	Period             string     `json:"period"`
	StartTime          *time.Time `json:"start_time,omitempty"`
	EndTime            *time.Time `json:"end_time,omitempty"`
	GroupBy            []string   `json:"group_by,omitempty"`
	IncludeComparison  bool       `json:"include_comparison"`
}

// UsageTrendsRequest represents a request for usage trends
type UsageTrendsRequest struct {
	UserSubscriptionID uint   `json:"user_subscription_id" validate:"required"`
	UsageType          string `json:"usage_type,omitempty"`
	Period             string `json:"period"`      // 7d, 30d, 90d, 365d
	Granularity        string `json:"granularity"` // hourly, daily, weekly
	IncludePredictions bool   `json:"include_predictions"`
	IncludeAnomalies   bool   `json:"include_anomalies"`
}

// UsageTrendsResponse represents usage trends data
type UsageTrendsResponse struct {
	UserSubscriptionID uint                        `json:"user_subscription_id"`
	Trends             []*UsageTrendEntry          `json:"trends"`
	Predictions        []*entities.UsagePrediction `json:"predictions,omitempty"`
	Anomalies          []*UsageAnomaly             `json:"anomalies,omitempty"`
	Summary            *TrendSummary               `json:"summary"`
	Period             string                      `json:"period"`
	Granularity        string                      `json:"granularity"`
}

// UsageTrendEntry represents a single trend data point
type UsageTrendEntry struct {
	Timestamp      time.Time `json:"timestamp"`
	UsageAmount    int64     `json:"usage_amount"`
	UsageType      string    `json:"usage_type"`
	RecordCount    int64     `json:"record_count"`
	MovingAvg      float64   `json:"moving_avg"`
	TrendDirection string    `json:"trend_direction"` // up, down, stable
}

// UsageAnomaly represents detected usage anomalies
type UsageAnomaly struct {
	Timestamp      time.Time `json:"timestamp"`
	UsageType      string    `json:"usage_type"`
	ActualUsage    int64     `json:"actual_usage"`
	ExpectedUsage  int64     `json:"expected_usage"`
	DeviationScore float64   `json:"deviation_score"`
	AnomalyType    string    `json:"anomaly_type"` // spike, drop, pattern_change
	Confidence     float64   `json:"confidence"`
}

// TrendSummary represents summary of trend analysis
type TrendSummary struct {
	OverallTrend        string    `json:"overall_trend"` // increasing, decreasing, stable
	AverageUsage        float64   `json:"average_usage"`
	PeakUsage           int64     `json:"peak_usage"`
	PeakUsageTime       time.Time `json:"peak_usage_time"`
	LowestUsage         int64     `json:"lowest_usage"`
	LowestUsageTime     time.Time `json:"lowest_usage_time"`
	TrendStrength       float64   `json:"trend_strength"`    // 0-1 scale
	VariabilityScore    float64   `json:"variability_score"` // 0-1 scale
	AnomalyCount        int       `json:"anomaly_count"`
	PredictedGrowthRate float64   `json:"predicted_growth_rate"` // % per period
}

// UsagePredictionRequest represents a request for usage prediction
type UsagePredictionRequest struct {
	UserSubscriptionID uint   `json:"user_subscription_id" validate:"required"`
	UsageType          string `json:"usage_type" validate:"required"`
	PredictionType     string `json:"prediction_type"` // daily, weekly, monthly, period_end
	BasedOnDays        int    `json:"based_on_days"`   // Number of days of historical data to use
	IncludeConfidence  bool   `json:"include_confidence"`
	IncludeTrends      bool   `json:"include_trends"`
}

// TopUsageRequest represents a request for top usage subscriptions
type TopUsageRequest struct {
	UsageType   string        `json:"usage_type,omitempty"`
	Period      time.Duration `json:"period"` // Period to look back (e.g., 24h, 7d, 30d)
	Limit       int           `json:"limit"`
	OrderBy     string        `json:"order_by"`     // total_usage, average_usage, peak_usage
	IncludeZero bool          `json:"include_zero"` // Include subscriptions with zero usage
}

// TopUsageResponse represents response for top usage query
type TopUsageResponse struct {
	TopSubscriptions []*TopUsageSubscription `json:"top_subscriptions"`
	Period           string                  `json:"period"`
	GeneratedAt      time.Time               `json:"generated_at"`
	TotalCount       int64                   `json:"total_count"`
}

// ExportUsageRequest represents a request to export usage data
type ExportUsageRequest struct {
	UserSubscriptionID uint       `json:"user_subscription_id"`
	UsageType          string     `json:"usage_type,omitempty"`
	StartTime          *time.Time `json:"start_time,omitempty"`
	EndTime            *time.Time `json:"end_time,omitempty"`
	Format             string     `json:"format"` // csv, json, xlsx
	IncludeMetadata    bool       `json:"include_metadata"`
	IncludeRawData     bool       `json:"include_raw_data"`
	Granularity        string     `json:"granularity"`
}

// ExportUsageResponse represents response for usage data export
type ExportUsageResponse struct {
	DownloadURL string    `json:"download_url"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	Format      string    `json:"format"`
	RecordCount int64     `json:"record_count"`
	GeneratedAt time.Time `json:"generated_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// RealTimeUsageResponse represents real-time usage information
type RealTimeUsageResponse struct {
	UserSubscriptionID uint                          `json:"user_subscription_id"`
	UsageByType        map[string]*RealTimeUsageData `json:"usage_by_type"`
	LastUpdated        time.Time                     `json:"last_updated"`
	UpdateFrequency    string                        `json:"update_frequency" swaggertype:"string"`
	AlertCount         int                           `json:"alert_count"`
	Predictions        []*entities.UsagePrediction   `json:"predictions,omitempty"`
}

// RealTimeUsageData represents real-time usage data for a specific type
type RealTimeUsageData struct {
	UsageType         string    `json:"usage_type"`
	CurrentUsage      int64     `json:"current_usage"`
	UsageLimit        int64     `json:"usage_limit"`
	UsagePercent      float64   `json:"usage_percent"`
	RemainingUsage    int64     `json:"remaining_usage"`
	RecentUsage       int64     `json:"recent_usage"`                             // Usage in last hour
	UsageRate         float64   `json:"usage_rate"`                               // Usage per hour
	EstimatedTimeLeft string    `json:"estimated_time_left" swaggertype:"string"` // Time until limit reached (as string)
	LastUpdated       time.Time `json:"last_updated"`
	TrendDirection    string    `json:"trend_direction"`
	IsNearLimit       bool      `json:"is_near_limit"`
	IsExceeded        bool      `json:"is_exceeded"`
}

// CleanupResult represents the result of cleanup operations
type CleanupResult struct {
	RecordsDeleted int64     `json:"records_deleted"`
	AlertsDeleted  int64     `json:"alerts_deleted"`
	SpaceFreed     int64     `json:"space_freed"` // in bytes
	OperationTime  string    `json:"operation_time" swaggertype:"string"`
	CompletedAt    time.Time `json:"completed_at"`
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalPages  int   `json:"total_pages"`
	TotalCount  int64 `json:"total_count"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// Common validation and constraints
const (
	// Granularity options
	GranularityHourly  = "hourly"
	GranularityDaily   = "daily"
	GranularityWeekly  = "weekly"
	GranularityMonthly = "monthly"

	// Prediction types
	PredictionTypeDaily     = "daily"
	PredictionTypeWeekly    = "weekly"
	PredictionTypeMonthly   = "monthly"
	PredictionTypePeriodEnd = "period_end"

	// Export formats
	FormatCSV  = "csv"
	FormatJSON = "json"
	FormatXLSX = "xlsx"

	// Trend directions
	TrendDirectionUp     = "up"
	TrendDirectionDown   = "down"
	TrendDirectionStable = "stable"

	// Anomaly types
	AnomalyTypeSpike         = "spike"
	AnomalyTypeDrop          = "drop"
	AnomalyTypePatternChange = "pattern_change"

	// Default limits
	DefaultPageSize       = 50
	MaxPageSize           = 1000
	DefaultPredictionDays = 30
	MaxPredictionDays     = 90
)
