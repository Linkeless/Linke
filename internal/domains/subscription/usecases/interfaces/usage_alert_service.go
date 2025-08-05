package interfaces

import (
	"time"

	"linke/internal/domains/subscription/entities"
)

// Request/Response structures for usage alert service

// CreateAlertConfigRequest represents a request to create an alert configuration
type CreateAlertConfigRequest struct {
	UserSubscriptionID   uint                           `json:"user_subscription_id" validate:"required"`
	UsageType            string                         `json:"usage_type" validate:"required"`
	ThresholdType        string                         `json:"threshold_type" validate:"required,oneof=percentage absolute"`
	Threshold            float64                        `json:"threshold" validate:"required,min=0"`
	Name                 string                         `json:"name" validate:"required,max=100"`
	Description          string                         `json:"description,omitempty"`
	Priority             string                         `json:"priority" validate:"oneof=low medium high critical"`
	IsEnabled            bool                           `json:"is_enabled"`
	NotificationChannels []entities.NotificationChannel `json:"notification_channels"`
	CooldownMinutes      int                            `json:"cooldown_minutes" validate:"min=1,max=10080"` // 1 minute to 1 week
}

// UpdateAlertConfigRequest represents a request to update an alert configuration
type UpdateAlertConfigRequest struct {
	ID                   uint                           `json:"id" validate:"required"`
	ThresholdType        *string                        `json:"threshold_type,omitempty" validate:"omitempty,oneof=percentage absolute"`
	Threshold            *float64                       `json:"threshold,omitempty" validate:"omitempty,min=0"`
	Name                 *string                        `json:"name,omitempty" validate:"omitempty,max=100"`
	Description          *string                        `json:"description,omitempty"`
	Priority             *string                        `json:"priority,omitempty" validate:"omitempty,oneof=low medium high critical"`
	IsEnabled            *bool                          `json:"is_enabled,omitempty"`
	NotificationChannels []entities.NotificationChannel `json:"notification_channels,omitempty"`
	CooldownMinutes      *int                           `json:"cooldown_minutes,omitempty" validate:"omitempty,min=1,max=10080"`
}

// GetAlertConfigsRequest represents a request to get alert configurations
type GetAlertConfigsRequest struct {
	UserSubscriptionID uint   `json:"user_subscription_id" validate:"required"`
	UsageType          string `json:"usage_type,omitempty"`
	IsEnabled          *bool  `json:"is_enabled,omitempty"`
	Priority           string `json:"priority,omitempty"`
	Limit              int    `json:"limit" validate:"min=1,max=1000"`
	Offset             int    `json:"offset" validate:"min=0"`
	OrderBy            string `json:"order_by,omitempty" validate:"omitempty,oneof=name threshold priority created_at"`
	OrderDirection     string `json:"order_direction,omitempty" validate:"omitempty,oneof=asc desc"`
}

// GetAlertConfigsResponse represents response for getting alert configurations
type GetAlertConfigsResponse struct {
	AlertConfigurations []*entities.AlertConfigurationResponse `json:"alert_configurations"`
	Pagination          *PaginationInfo                        `json:"pagination"`
	TotalCount          int64                                  `json:"total_count"`
}

// FireAlertRequest represents a request to fire an alert manually
type FireAlertRequest struct {
	UserSubscriptionID   uint   `json:"user_subscription_id" validate:"required"`
	AlertConfigurationID uint   `json:"alert_configuration_id" validate:"required"`
	CurrentUsage         int64  `json:"current_usage" validate:"min=0"`
	UsageLimit           int64  `json:"usage_limit" validate:"min=0"`
	Message              string `json:"message,omitempty"`
	Severity             string `json:"severity,omitempty" validate:"omitempty,oneof=info warning error critical"`
	ForceNotification    bool   `json:"force_notification"` // Send notification even if in cooldown
}

// GetUsageAlertsRequest represents a request to get usage alerts
type GetUsageAlertsRequest struct {
	UserSubscriptionID   uint       `json:"user_subscription_id,omitempty"`
	AlertConfigurationID uint       `json:"alert_configuration_id,omitempty"`
	UsageType            string     `json:"usage_type,omitempty"`
	Status               string     `json:"status,omitempty" validate:"omitempty,oneof=fired resolved suppressed acknowledged"`
	Severity             string     `json:"severity,omitempty" validate:"omitempty,oneof=info warning error critical"`
	IsActive             *bool      `json:"is_active,omitempty"`
	IsResolved           *bool      `json:"is_resolved,omitempty"`
	StartTime            *time.Time `json:"start_time,omitempty"`
	EndTime              *time.Time `json:"end_time,omitempty"`
	Limit                int        `json:"limit" validate:"min=1,max=1000"`
	Offset               int        `json:"offset" validate:"min=0"`
	OrderBy              string     `json:"order_by,omitempty" validate:"omitempty,oneof=fired_at resolved_at severity usage_percent"`
	OrderDirection       string     `json:"order_direction,omitempty" validate:"omitempty,oneof=asc desc"`
}

// GetUsageAlertsResponse represents response for getting usage alerts
type GetUsageAlertsResponse struct {
	UsageAlerts []*entities.UsageAlertResponse `json:"usage_alerts"`
	Pagination  *PaginationInfo                `json:"pagination"`
	Summary     *AlertsSummary                 `json:"summary"`
	TotalCount  int64                          `json:"total_count"`
}

// AlertsSummary represents summary of alerts
type AlertsSummary struct {
	TotalAlerts     int64 `json:"total_alerts"`
	ActiveAlerts    int64 `json:"active_alerts"`
	ResolvedAlerts  int64 `json:"resolved_alerts"`
	CriticalAlerts  int64 `json:"critical_alerts"`
	HighAlerts      int64 `json:"high_alerts"`
	MediumAlerts    int64 `json:"medium_alerts"`
	LowAlerts       int64 `json:"low_alerts"`
	AlertsToday     int64 `json:"alerts_today"`
	AlertsThisWeek  int64 `json:"alerts_this_week"`
	AlertsThisMonth int64 `json:"alerts_this_month"`
}

// BulkResolveAlertsRequest represents a request to resolve multiple alerts
type BulkResolveAlertsRequest struct {
	AlertIDs           []uint `json:"alert_ids" validate:"required,min=1"`
	UserSubscriptionID uint   `json:"user_subscription_id,omitempty"`
	UsageType          string `json:"usage_type,omitempty"`
	Reason             string `json:"reason,omitempty"`
	ResolvedBy         uint   `json:"resolved_by,omitempty"`
}

// BulkResolveAlertsResponse represents response for bulk resolve operation
type BulkResolveAlertsResponse struct {
	ResolvedCount int64    `json:"resolved_count"`
	FailedIDs     []uint   `json:"failed_ids,omitempty"`
	Errors        []string `json:"errors,omitempty"`
}

// TestNotificationRequest represents a request to test notification channels
type TestNotificationRequest struct {
	UserSubscriptionID uint                         `json:"user_subscription_id" validate:"required"`
	Channel            entities.NotificationChannel `json:"channel" validate:"required"`
	TestMessage        string                       `json:"test_message,omitempty"`
}

// TestNotificationResponse represents response for notification test
type TestNotificationResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	ResponseTime string `json:"response_time" swaggertype:"string"`
	Details      string `json:"details,omitempty"`
	Error        string `json:"error,omitempty"`
}

// UpdateNotificationPrefsRequest represents a request to update notification preferences
type UpdateNotificationPrefsRequest struct {
	UserSubscriptionID uint                          `json:"user_subscription_id" validate:"required"`
	GlobalSettings     *NotificationGlobalSettings   `json:"global_settings,omitempty"`
	ChannelSettings    []NotificationChannelSettings `json:"channel_settings,omitempty"`
	QuietHours         *QuietHoursSettings           `json:"quiet_hours,omitempty"`
}

// NotificationGlobalSettings represents global notification settings
type NotificationGlobalSettings struct {
	Enabled          bool     `json:"enabled"`
	MaxAlertsPerHour int      `json:"max_alerts_per_hour" validate:"min=1,max=100"`
	MaxAlertsPerDay  int      `json:"max_alerts_per_day" validate:"min=1,max=1000"`
	SeverityFilter   []string `json:"severity_filter"`   // Only send alerts of these severities
	UsageTypeFilter  []string `json:"usage_type_filter"` // Only send alerts for these usage types
}

// NotificationChannelSettings represents settings for a specific notification channel
type NotificationChannelSettings struct {
	Type            string            `json:"type" validate:"required"`
	Enabled         bool              `json:"enabled"`
	SeverityFilter  []string          `json:"severity_filter,omitempty"`
	CooldownMinutes int               `json:"cooldown_minutes" validate:"min=1,max=1440"`
	Settings        map[string]string `json:"settings,omitempty"`
}

// QuietHoursSettings represents quiet hours configuration
type QuietHoursSettings struct {
	Enabled   bool   `json:"enabled"`
	StartHour int    `json:"start_hour" validate:"min=0,max=23"`
	EndHour   int    `json:"end_hour" validate:"min=0,max=23"`
	TimeZone  string `json:"time_zone"`
	Days      []int  `json:"days"` // 0=Sunday, 1=Monday, etc.
}

// AlertStatsRequest represents a request for alert statistics
type AlertStatsRequest struct {
	UserSubscriptionID uint       `json:"user_subscription_id,omitempty"`
	UsageType          string     `json:"usage_type,omitempty"`
	Severity           string     `json:"severity,omitempty"`
	Period             string     `json:"period" validate:"required,oneof=24h 7d 30d 90d 365d"`
	StartTime          *time.Time `json:"start_time,omitempty"`
	EndTime            *time.Time `json:"end_time,omitempty"`
	GroupBy            string     `json:"group_by,omitempty" validate:"omitempty,oneof=hour day week month severity usage_type"`
}

// AlertStatisticsResponse represents alert statistics
type AlertStatisticsResponse struct {
	Summary           *AlertsSummary           `json:"summary"`
	Trends            []*AlertTrendEntry       `json:"trends"`
	Distribution      *AlertDistribution       `json:"distribution"`
	TopConfigurations []*TopAlertConfiguration `json:"top_configurations"`
	Period            string                   `json:"period"`
	GeneratedAt       time.Time                `json:"generated_at"`
}

// AlertTrendEntry represents a single point in alert trends
type AlertTrendEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	AlertCount int64     `json:"alert_count"`
	Period     string    `json:"period"`
	Severity   string    `json:"severity,omitempty"`
	UsageType  string    `json:"usage_type,omitempty"`
}

// AlertDistribution represents distribution of alerts
type AlertDistribution struct {
	BySeverity  map[string]int64 `json:"by_severity"`
	ByUsageType map[string]int64 `json:"by_usage_type"`
	ByHour      map[int]int64    `json:"by_hour"`        // Hour of day (0-23)
	ByDayOfWeek map[int]int64    `json:"by_day_of_week"` // Day of week (0-6)
	ByStatus    map[string]int64 `json:"by_status"`
}

// TopAlertConfiguration represents a configuration that generates many alerts
type TopAlertConfiguration struct {
	ConfigurationID       uint      `json:"configuration_id"`
	Name                  string    `json:"name"`
	UsageType             string    `json:"usage_type"`
	Threshold             float64   `json:"threshold"`
	AlertCount            int64     `json:"alert_count"`
	LastAlertFired        time.Time `json:"last_alert_fired"`
	AverageResolutionTime string    `json:"average_resolution_time" swaggertype:"string"`
}

// AlertHistoryRequest represents a request for alert history
type AlertHistoryRequest struct {
	UserSubscriptionID   uint       `json:"user_subscription_id,omitempty"`
	AlertConfigurationID uint       `json:"alert_configuration_id,omitempty"`
	UsageType            string     `json:"usage_type,omitempty"`
	StartTime            *time.Time `json:"start_time,omitempty"`
	EndTime              *time.Time `json:"end_time,omitempty"`
	IncludeResolved      bool       `json:"include_resolved"`
	IncludeNotifications bool       `json:"include_notifications"`
	Limit                int        `json:"limit" validate:"min=1,max=1000"`
	Offset               int        `json:"offset" validate:"min=0"`
}

// AlertHistoryResponse represents alert history data
type AlertHistoryResponse struct {
	AlertHistory []*AlertHistoryEntry `json:"alert_history"`
	Summary      *AlertHistorySummary `json:"summary"`
	Pagination   *PaginationInfo      `json:"pagination"`
	TotalCount   int64                `json:"total_count"`
}

// AlertHistoryEntry represents a single entry in alert history
type AlertHistoryEntry struct {
	Alert               *entities.UsageAlertResponse         `json:"alert"`
	Configuration       *entities.AlertConfigurationResponse `json:"configuration"`
	NotificationHistory []*entities.NotificationResult       `json:"notification_history,omitempty"`
	ResolutionTime      *string                              `json:"resolution_time,omitempty" swaggertype:"string"`
	AcknowledgedBy      *uint                                `json:"acknowledged_by,omitempty"`
	AcknowledgedAt      *time.Time                           `json:"acknowledged_at,omitempty"`
}

// AlertHistorySummary represents summary of alert history
type AlertHistorySummary struct {
	TotalAlerts           int64  `json:"total_alerts"`
	AverageResolutionTime string `json:"average_resolution_time" swaggertype:"string"`
	FastestResolution     string `json:"fastest_resolution" swaggertype:"string"`
	SlowestResolution     string `json:"slowest_resolution" swaggertype:"string"`
	UnresolvedCount       int64  `json:"unresolved_count"`
	NotificationsSent     int64  `json:"notifications_sent"`
	NotificationFailures  int64  `json:"notification_failures"`
}

// Constants for alert service

// Alert status values
const (
	AlertStatusFired        = "fired"
	AlertStatusResolved     = "resolved"
	AlertStatusSuppressed   = "suppressed"
	AlertStatusAcknowledged = "acknowledged"
)

// Alert severity values
const (
	AlertSeverityInfo     = "info"
	AlertSeverityWarning  = "warning"
	AlertSeverityError    = "error"
	AlertSeverityCritical = "critical"
)

// Priority values
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Threshold types
const (
	ThresholdTypePercentage = "percentage"
	ThresholdTypeAbsolute   = "absolute"
)

// Notification channel types
const (
	NotificationChannelEmail   = "email"
	NotificationChannelWebhook = "webhook"
	NotificationChannelInApp   = "in_app"
	NotificationChannelSMS     = "sms"
	NotificationChannelSlack   = "slack"
	NotificationChannelDiscord = "discord"
)

// Default values
const (
	DefaultCooldownMinutes = 60
	MaxAlertsPerHour       = 10
	MaxAlertsPerDay        = 100
)
