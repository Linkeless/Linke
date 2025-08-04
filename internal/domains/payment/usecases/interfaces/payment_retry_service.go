package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/payment/entities"
)

// PaymentRetryService defines the interface for payment retry operations
type PaymentRetryService interface {
	// Retry management
	InitiateRetry(ctx context.Context, paymentRecord *entities.PaymentRecord, failureType, failureCode, errorMessage string) (*entities.PaymentRetry, error)
	ProcessPendingRetries(ctx context.Context, batchSize int) (int, error)
	ProcessRetry(ctx context.Context, retryID uint) (*RetryResult, error)
	CancelRetry(ctx context.Context, retryID uint, reason string) error
	ResetRetry(ctx context.Context, retryID uint) error

	// Configuration and strategy
	GetRetryStrategy(ctx context.Context, gateway, paymentMethod string) (*entities.RetryStrategyConfig, error)
	UpdateRetryStrategy(ctx context.Context, gateway, paymentMethod string, config *entities.RetryStrategyConfig) error
	CalculateNextRetryTime(ctx context.Context, retry *entities.PaymentRetry) time.Time

	// Query operations
	GetRetryByPaymentID(ctx context.Context, paymentRecordID uint) (*entities.PaymentRetry, error)
	GetRetryWithHistory(ctx context.Context, retryID uint) (*RetryWithHistory, error)
	GetActiveRetries(ctx context.Context, filters *RetryFilters, limit, offset int) ([]*entities.PaymentRetry, int64, error)
	GetRetryHistory(ctx context.Context, retryID uint) ([]*entities.PaymentRetryHistory, error)

	// Statistics and monitoring
	GetRetryStatistics(ctx context.Context, gateway string, days int) (*RetryStatistics, error)
	GetFailureAnalysis(ctx context.Context, gateway string, days int) (*FailureAnalysis, error)
	GetRetryHealthMetrics(ctx context.Context) (*RetryHealthMetrics, error)

	// Admin operations
	GetRetriesForAdmin(ctx context.Context, filters *AdminRetryFilters) (*AdminRetryResponse, error)
	BulkCancelRetries(ctx context.Context, retryIDs []uint, reason string) error
	BulkResetRetries(ctx context.Context, retryIDs []uint) error

	// Integration points
	ClassifyFailure(ctx context.Context, gateway, paymentMethod, errorCode, errorMessage string) string
	ShouldRetryPayment(ctx context.Context, paymentRecord *entities.PaymentRecord, errorCode string) bool
	NotifyRetryAttempt(ctx context.Context, retry *entities.PaymentRetry, attempt *entities.PaymentRetryHistory) error
}

// Data structures for service operations

// RetryResult represents the result of a retry attempt
type RetryResult struct {
	Success       bool                          `json:"success"`
	RetryID       uint                          `json:"retry_id"`
	AttemptNumber int                           `json:"attempt_number"`
	Status        string                        `json:"status"`
	NextRetryAt   *time.Time                    `json:"next_retry_at,omitempty"`
	CompletedAt   *time.Time                    `json:"completed_at,omitempty"`
	ErrorMessage  string                        `json:"error_message,omitempty"`
	PaymentResult *PaymentProcessResult         `json:"payment_result,omitempty"`
	History       *entities.PaymentRetryHistory `json:"history,omitempty"`
}

// PaymentProcessResult represents the result of payment processing
type PaymentProcessResult struct {
	PaymentRecordID uint       `json:"payment_record_id"`
	Status          string     `json:"status"`
	TransactionID   string     `json:"transaction_id,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	ErrorCode       string     `json:"error_code,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
}

// RetryWithHistory represents a retry with its complete history
type RetryWithHistory struct {
	Retry   *entities.PaymentRetry          `json:"retry"`
	History []*entities.PaymentRetryHistory `json:"history"`
}

// FailureAnalysis represents analysis of payment failures
type FailureAnalysis struct {
	Gateway             string            `json:"gateway"`
	TotalFailures       int64             `json:"total_failures"`
	TemporaryFailures   int64             `json:"temporary_failures"`
	PermanentFailures   int64             `json:"permanent_failures"`
	NetworkFailures     int64             `json:"network_failures"`
	GatewayFailures     int64             `json:"gateway_failures"`
	BusinessFailures    int64             `json:"business_failures"`
	TopFailureReasons   []*FailureReason  `json:"top_failure_reasons"`
	FailurePatterns     []*FailurePattern `json:"failure_patterns"`
	RecoveryRate        float64           `json:"recovery_rate"`
	AverageRecoveryTime float64           `json:"average_recovery_time"`
	RecommendedActions  []string          `json:"recommended_actions"`
}

// FailureReason represents a common failure reason
type FailureReason struct {
	Reason           string  `json:"reason"`
	Count            int64   `json:"count"`
	Percentage       float64 `json:"percentage"`
	RetrySuccessRate float64 `json:"retry_success_rate"`
}

// RetryHealthMetrics represents overall retry system health
type RetryHealthMetrics struct {
	TotalActiveRetries    int64                  `json:"total_active_retries"`
	OverdueRetries        int64                  `json:"overdue_retries"`
	SuccessRate24h        float64                `json:"success_rate_24h"`
	SuccessRate7d         float64                `json:"success_rate_7d"`
	AverageRetryDelay     float64                `json:"average_retry_delay"`
	GatewayHealth         []*GatewayHealthMetric `json:"gateway_health"`
	AlertsTriggered       []string               `json:"alerts_triggered"`
	SystemRecommendations []string               `json:"system_recommendations"`
}

// GatewayHealthMetric represents health metrics for a specific gateway
type GatewayHealthMetric struct {
	Gateway         string  `json:"gateway"`
	ActiveRetries   int64   `json:"active_retries"`
	SuccessRate     float64 `json:"success_rate"`
	AverageAttempts float64 `json:"average_attempts"`
	QueueDepth      int64   `json:"queue_depth"`
	ProcessingRate  float64 `json:"processing_rate"`
	HealthStatus    string  `json:"health_status"` // healthy, degraded, critical
}

// AdminRetryFilters represents filters for admin retry queries
type AdminRetryFilters struct {
	*RetryFilters
	IncludeHistory   bool     `json:"include_history,omitempty"`
	SortBy           string   `json:"sort_by,omitempty"`    // created_at, next_retry_at, attempt_number
	SortOrder        string   `json:"sort_order,omitempty"` // asc, desc
	PaymentRecordIDs []uint   `json:"payment_record_ids,omitempty"`
	UserIDs          []uint   `json:"user_ids,omitempty"`
	RetryStatuses    []string `json:"retry_statuses,omitempty"`
	FailureTypes     []string `json:"failure_types,omitempty"`
	Gateways         []string `json:"gateways,omitempty"`
	PaymentMethods   []string `json:"payment_methods,omitempty"`
	Limit            int      `json:"limit,omitempty"`
	Offset           int      `json:"offset,omitempty"`
}

// AdminRetryResponse represents response for admin retry queries
type AdminRetryResponse struct {
	Retries    []*entities.PaymentRetry `json:"retries"`
	TotalCount int64                    `json:"total_count"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
	Statistics *AdminRetryStatistics    `json:"statistics,omitempty"`
}

// AdminRetryStatistics represents statistics for admin interface
type AdminRetryStatistics struct {
	TotalRetries       int64   `json:"total_retries"`
	PendingRetries     int64   `json:"pending_retries"`
	InProgressRetries  int64   `json:"in_progress_retries"`
	CompletedRetries   int64   `json:"completed_retries"`
	FailedRetries      int64   `json:"failed_retries"`
	CancelledRetries   int64   `json:"cancelled_retries"`
	OverallSuccessRate float64 `json:"overall_success_rate"`
	AverageAttempts    float64 `json:"average_attempts"`
	AverageDelayTime   float64 `json:"average_delay_time"`
}

// RetryConfiguration represents system-wide retry configuration
type RetryConfiguration struct {
	Enabled                    bool                                     `json:"enabled"`
	MaxConcurrentRetries       int                                      `json:"max_concurrent_retries"`
	ProcessingInterval         time.Duration                            `json:"processing_interval"`
	HealthCheckInterval        time.Duration                            `json:"health_check_interval"`
	DefaultStrategy            *entities.RetryStrategyConfig            `json:"default_strategy"`
	GatewayStrategies          map[string]*entities.RetryStrategyConfig `json:"gateway_strategies"`
	FailureClassificationRules map[string]*FailureClassificationRule    `json:"failure_classification_rules"`
	NotificationSettings       *RetryNotificationSettings               `json:"notification_settings"`
	MonitoringSettings         *RetryMonitoringSettings                 `json:"monitoring_settings"`
}

// FailureClassificationRule represents rules for classifying payment failures
type FailureClassificationRule struct {
	Gateway       string   `json:"gateway"`
	PaymentMethod string   `json:"payment_method,omitempty"`
	ErrorCodes    []string `json:"error_codes"`
	ErrorPatterns []string `json:"error_patterns"`
	FailureType   string   `json:"failure_type"`
	ShouldRetry   bool     `json:"should_retry"`
	Priority      int      `json:"priority"`
}

// RetryNotificationSettings represents notification settings for retries
type RetryNotificationSettings struct {
	Enabled             bool     `json:"enabled"`
	NotifyOnFailure     bool     `json:"notify_on_failure"`
	NotifyOnSuccess     bool     `json:"notify_on_success"`
	NotifyOnMaxAttempts bool     `json:"notify_on_max_attempts"`
	EmailTemplates      []string `json:"email_templates"`
	SMSTemplates        []string `json:"sms_templates"`
	WebhookURLs         []string `json:"webhook_urls"`
}

// RetryMonitoringSettings represents monitoring settings for retry system
type RetryMonitoringSettings struct {
	MetricsEnabled     bool             `json:"metrics_enabled"`
	AlertsEnabled      bool             `json:"alerts_enabled"`
	HealthCheckEnabled bool             `json:"health_check_enabled"`
	LogLevel           string           `json:"log_level"`
	RetentionPeriod    time.Duration    `json:"retention_period"`
	AlertThresholds    *AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds represents alert thresholds for retry monitoring
type AlertThresholds struct {
	MaxPendingRetries  int     `json:"max_pending_retries"`
	MaxOverdueRetries  int     `json:"max_overdue_retries"`
	MinSuccessRate     float64 `json:"min_success_rate"`
	MaxAverageAttempts float64 `json:"max_average_attempts"`
	MaxProcessingDelay int     `json:"max_processing_delay"` // seconds
}
