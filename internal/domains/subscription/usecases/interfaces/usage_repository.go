package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/subscription/entities"
)

// UsageRepository defines the contract for usage data storage operations
type UsageRepository interface {
	// Usage Records Management
	CreateUsageRecord(ctx context.Context, record *entities.UsageRecord) error
	GetUsageRecord(ctx context.Context, id uint) (*entities.UsageRecord, error)
	GetUsageRecords(ctx context.Context, filter UsageRecordFilter) ([]*entities.UsageRecord, error)
	UpdateUsageRecord(ctx context.Context, record *entities.UsageRecord) error
	DeleteUsageRecord(ctx context.Context, id uint) error

	// Usage Aggregations
	GetUsageAggregation(ctx context.Context, filter UsageAggregationFilter) (*entities.UsageAggregation, error)
	GetUsageAggregations(ctx context.Context, filter UsageAggregationFilter) ([]*entities.UsageAggregation, error)
	GetUsageSummary(ctx context.Context, filter UsageSummaryFilter) (*entities.UsageSummary, error)

	// Bulk Operations
	CreateUsageRecordsBatch(ctx context.Context, records []*entities.UsageRecord) error
	DeleteOldUsageRecords(ctx context.Context, olderThan time.Time) (int64, error)

	// Current Usage Queries
	GetCurrentUsage(ctx context.Context, subscriptionID uint, usageType string) (int64, error)
	GetCurrentUsageByPeriod(ctx context.Context, subscriptionID uint, usageType string, periodStart, periodEnd time.Time) (int64, error)
	GetUsageByTimeRange(ctx context.Context, subscriptionID uint, usageType string, startTime, endTime time.Time) ([]*entities.UsageRecord, error)

	// Statistics
	GetUsageStatistics(ctx context.Context, filter UsageStatsFilter) (*UsageStatistics, error)
	GetTopUsageSubscriptions(ctx context.Context, usageType string, limit int, period time.Duration) ([]*TopUsageSubscription, error)
}

// AlertRepository defines the contract for alert configuration and alert data storage
type AlertRepository interface {
	// Alert Configurations
	CreateAlertConfiguration(ctx context.Context, config *entities.AlertConfiguration) error
	GetAlertConfiguration(ctx context.Context, id uint) (*entities.AlertConfiguration, error)
	GetAlertConfigurations(ctx context.Context, filter AlertConfigurationFilter) ([]*entities.AlertConfiguration, error)
	UpdateAlertConfiguration(ctx context.Context, config *entities.AlertConfiguration) error
	DeleteAlertConfiguration(ctx context.Context, id uint) error

	// Usage Alerts
	CreateUsageAlert(ctx context.Context, alert *entities.UsageAlert) error
	GetUsageAlert(ctx context.Context, id uint) (*entities.UsageAlert, error)
	GetUsageAlerts(ctx context.Context, filter UsageAlertFilter) ([]*entities.UsageAlert, error)
	UpdateUsageAlert(ctx context.Context, alert *entities.UsageAlert) error
	DeleteUsageAlert(ctx context.Context, id uint) error

	// Alert Status Operations
	ResolveAlert(ctx context.Context, alertID uint) error
	AcknowledgeAlert(ctx context.Context, alertID uint) error
	SuppressAlert(ctx context.Context, alertID uint, duration time.Duration) error

	// Alert Queries
	GetActiveAlerts(ctx context.Context, subscriptionID uint) ([]*entities.UsageAlert, error)
	GetAlertsForConfiguration(ctx context.Context, configID uint, limit int) ([]*entities.UsageAlert, error)
	GetActiveAlertConfigurations(ctx context.Context, subscriptionID uint, usageType string) ([]*entities.AlertConfiguration, error)

	// Bulk Operations
	ResolveAlertsForSubscription(ctx context.Context, subscriptionID uint, usageType string) error
	CleanupOldAlerts(ctx context.Context, olderThan time.Time) (int64, error)
}

// Filter structures for repository queries

// UsageRecordFilter defines filtering options for usage record queries
type UsageRecordFilter struct {
	UserSubscriptionID uint
	UsageType          string
	SourceType         string
	SourceID           string
	IPAddress          string
	StartTime          *time.Time
	EndTime            *time.Time
	Limit              int
	Offset             int
	OrderBy            string // timestamp, amount, created_at
	OrderDirection     string // asc, desc
	IncludeDeleted     bool
}

// UsageAggregationFilter defines filtering options for usage aggregation queries
type UsageAggregationFilter struct {
	UserSubscriptionID uint
	UsageType          string
	Period             string // hourly, daily, weekly, monthly
	StartTime          *time.Time
	EndTime            *time.Time
	GroupBy            []string // usage_type, source_type, hour, day, week, month
	OrderBy            string
	OrderDirection     string
	Limit              int
	Offset             int
}

// UsageSummaryFilter defines filtering options for usage summary queries
type UsageSummaryFilter struct {
	UserSubscriptionID uint
	UsageTypes         []string
	Period             string // daily, weekly, monthly, custom
	PeriodStart        *time.Time
	PeriodEnd          *time.Time
	IncludeBreakdown   bool // Include breakdown by source, time, etc.
	IncludePredictions bool // Include usage predictions
}

// AlertConfigurationFilter defines filtering options for alert configuration queries
type AlertConfigurationFilter struct {
	UserSubscriptionID uint
	UsageType          string
	ThresholdType      string
	IsEnabled          *bool
	Priority           string
	Limit              int
	Offset             int
	OrderBy            string
	OrderDirection     string
	IncludeDeleted     bool
}

// UsageAlertFilter defines filtering options for usage alert queries
type UsageAlertFilter struct {
	UserSubscriptionID   uint
	AlertConfigurationID uint
	UsageType            string
	Status               string
	Severity             string
	StartTime            *time.Time
	EndTime              *time.Time
	IsResolved           *bool
	IsActive             *bool
	Limit                int
	Offset               int
	OrderBy              string // fired_at, resolved_at, severity, usage_percent
	OrderDirection       string
	IncludeDeleted       bool
}

// UsageStatsFilter defines filtering options for usage statistics queries
type UsageStatsFilter struct {
	UserSubscriptionID uint
	UsageType          string
	Period             string
	StartTime          *time.Time
	EndTime            *time.Time
	GroupBy            []string
}

// Response structures for repository queries

// UsageStatistics represents aggregated usage statistics
type UsageStatistics struct {
	UserSubscriptionID uint                                `json:"user_subscription_id"`
	Period             string                              `json:"period"`
	TotalUsage         int64                               `json:"total_usage"`
	TotalRecords       int64                               `json:"total_records"`
	AverageUsage       float64                             `json:"average_usage"`
	PeakUsage          int64                               `json:"peak_usage"`
	PeakUsageTime      time.Time                           `json:"peak_usage_time"`
	UsageByType        map[string]*entities.UsageTypeStats `json:"usage_by_type"`
	UsageByHour        map[int]int64                       `json:"usage_by_hour"`        // Hour of day (0-23) -> usage
	UsageByDayOfWeek   map[int]int64                       `json:"usage_by_day_of_week"` // Day of week (0-6) -> usage
	UsageBySource      map[string]int64                    `json:"usage_by_source"`      // Source type -> usage
	StartTime          time.Time                           `json:"start_time"`
	EndTime            time.Time                           `json:"end_time"`
	GeneratedAt        time.Time                           `json:"generated_at"`
}

// TopUsageSubscription represents a subscription with high usage
type TopUsageSubscription struct {
	UserSubscriptionID uint      `json:"user_subscription_id"`
	UserID             uint      `json:"user_id"`
	TotalUsage         int64     `json:"total_usage"`
	UsageType          string    `json:"usage_type"`
	Period             string    `json:"period"`
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	RecordCount        int64     `json:"record_count"`
	AverageUsage       float64   `json:"average_usage"`
	PeakUsage          int64     `json:"peak_usage"`
}

// Repository method constants

// Order by options for usage records
const (
	OrderByTimestamp = "timestamp"
	OrderByAmount    = "amount"
	OrderByCreatedAt = "created_at"
)

// Order by options for alerts
const (
	OrderByFiredAt      = "fired_at"
	OrderByResolvedAt   = "resolved_at"
	OrderBySeverity     = "severity"
	OrderByUsagePercent = "usage_percent"
)

// Order directions
const (
	OrderDirectionAsc  = "asc"
	OrderDirectionDesc = "desc"
)

// Period options
const (
	PeriodHourly  = "hourly"
	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
	PeriodCustom  = "custom"
)

// Group by options
const (
	GroupByUsageType  = "usage_type"
	GroupBySourceType = "source_type"
	GroupByHour       = "hour"
	GroupByDay        = "day"
	GroupByWeek       = "week"
	GroupByMonth      = "month"
)
