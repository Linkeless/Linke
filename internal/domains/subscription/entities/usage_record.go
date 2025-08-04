package entities

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// UsageRecord represents detailed usage tracking data
type UsageRecord struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Key
	UserSubscriptionID uint `json:"user_subscription_id" gorm:"not null;index"`

	// Usage Details
	UsageType string    `json:"usage_type" gorm:"size:50;not null;index;comment:Type of usage (traffic, api_calls, storage, etc.)"`
	Amount    int64     `json:"amount" gorm:"not null;comment:Usage amount in bytes or count"`
	Unit      string    `json:"unit" gorm:"size:20;not null;default:'bytes';comment:Unit of measurement (bytes, count, etc.)"`
	Timestamp time.Time `json:"timestamp" gorm:"not null;index;comment:When the usage occurred"`

	// Source Information
	SourceType string `json:"source_type" gorm:"size:50;not null;comment:Source of usage (server, api, admin, etc.)"`
	SourceID   string `json:"source_id" gorm:"size:100;index;comment:ID of the source (server_id, api_endpoint, etc.)"`

	// Additional Context
	Metadata  string `json:"metadata,omitempty" gorm:"type:text;comment:Additional metadata as JSON"`
	UserAgent string `json:"user_agent,omitempty" gorm:"type:text;comment:User agent if applicable"`
	IPAddress string `json:"ip_address,omitempty" gorm:"size:45;index;comment:IP address if applicable"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for UsageRecord model
func (UsageRecord) TableName() string {
	return "usage_records"
}

// Usage type constants
const (
	UsageTypeTraffic     = "traffic"
	UsageTypeAPICall     = "api_call"
	UsageTypeStorage     = "storage"
	UsageTypeBandwidth   = "bandwidth"
	UsageTypeConnections = "connections"
)

// Unit constants
const (
	UnitBytes = "bytes"
	UnitCount = "count"
	UnitMB    = "mb"
	UnitGB    = "gb"
	UnitTB    = "tb"
)

// Source type constants
const (
	SourceTypeServer = "server"
	SourceTypeAPI    = "api"
	SourceTypeAdmin  = "admin"
	SourceTypeSystem = "system"
	SourceTypeUser   = "user"
)

// UsageAggregation represents aggregated usage data for reporting
type UsageAggregation struct {
	UserSubscriptionID uint      `json:"user_subscription_id"`
	UsageType          string    `json:"usage_type"`
	Period             string    `json:"period"` // hourly, daily, weekly, monthly
	PeriodStart        time.Time `json:"period_start"`
	PeriodEnd          time.Time `json:"period_end"`
	TotalAmount        int64     `json:"total_amount"`
	AverageAmount      float64   `json:"average_amount"`
	MinAmount          int64     `json:"min_amount"`
	MaxAmount          int64     `json:"max_amount"`
	RecordCount        int64     `json:"record_count"`
	FirstUsage         time.Time `json:"first_usage"`
	LastUsage          time.Time `json:"last_usage"`
}

// UsageSummary represents a usage summary for API responses
type UsageSummary struct {
	UserSubscriptionID uint                       `json:"user_subscription_id"`
	Period             string                     `json:"period"`
	PeriodStart        time.Time                  `json:"period_start"`
	PeriodEnd          time.Time                  `json:"period_end"`
	UsageByType        map[string]*UsageTypeStats `json:"usage_by_type"`
	TotalUsage         int64                      `json:"total_usage"`
	TotalRecords       int64                      `json:"total_records"`
	GeneratedAt        time.Time                  `json:"generated_at"`
}

// UsageTypeStats represents statistics for a specific usage type
type UsageTypeStats struct {
	UsageType     string    `json:"usage_type"`
	TotalAmount   int64     `json:"total_amount"`
	AverageAmount float64   `json:"average_amount"`
	MinAmount     int64     `json:"min_amount"`
	MaxAmount     int64     `json:"max_amount"`
	RecordCount   int64     `json:"record_count"`
	FirstUsage    time.Time `json:"first_usage"`
	LastUsage     time.Time `json:"last_usage"`
	Unit          string    `json:"unit"`
}

// UsagePrediction represents usage prediction data
type UsagePrediction struct {
	UserSubscriptionID uint      `json:"user_subscription_id"`
	UsageType          string    `json:"usage_type"`
	PredictionType     string    `json:"prediction_type"` // daily, weekly, monthly, period_end
	CurrentUsage       int64     `json:"current_usage"`
	PredictedUsage     int64     `json:"predicted_usage"`
	Confidence         float64   `json:"confidence"`
	PredictionDate     time.Time `json:"prediction_date"`
	BasedOnDays        int       `json:"based_on_days"`
	Trend              string    `json:"trend"`               // increasing, decreasing, stable
	EstimatedDaysLeft  int       `json:"estimated_days_left"` // Days until limit reached
}

// Prediction type constants
const (
	PredictionTypeDaily     = "daily"
	PredictionTypeWeekly    = "weekly"
	PredictionTypeMonthly   = "monthly"
	PredictionTypePeriodEnd = "period_end"
)

// Trend constants
const (
	TrendIncreasing = "increasing"
	TrendDecreasing = "decreasing"
	TrendStable     = "stable"
)

// IsDeleted checks if the usage record is soft deleted
func (ur *UsageRecord) IsDeleted() bool {
	return ur.DeletedAt.Valid
}

// GetFormattedAmount returns the usage amount in a human-readable format
func (ur *UsageRecord) GetFormattedAmount() string {
	switch ur.Unit {
	case UnitBytes:
		return formatBytes(ur.Amount)
	case UnitMB:
		return formatMB(ur.Amount)
	case UnitGB:
		return formatGB(ur.Amount)
	case UnitTB:
		return formatTB(ur.Amount)
	default:
		return formatCount(ur.Amount)
	}
}

// Helper functions for formatting
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatMB(mb int64) string {
	return fmt.Sprintf("%.2f MB", float64(mb))
}

func formatGB(gb int64) string {
	return fmt.Sprintf("%.2f GB", float64(gb))
}

func formatTB(tb int64) string {
	return fmt.Sprintf("%.2f TB", float64(tb))
}

func formatCount(count int64) string {
	return fmt.Sprintf("%d", count)
}

// ToResponse converts UsageRecord to a response structure
func (ur *UsageRecord) ToResponse() *UsageRecordResponse {
	return &UsageRecordResponse{
		ID:                 ur.ID,
		UserSubscriptionID: ur.UserSubscriptionID,
		UsageType:          ur.UsageType,
		Amount:             ur.Amount,
		Unit:               ur.Unit,
		FormattedAmount:    ur.GetFormattedAmount(),
		Timestamp:          ur.Timestamp,
		SourceType:         ur.SourceType,
		SourceID:           ur.SourceID,
		Metadata:           ur.Metadata,
		UserAgent:          ur.UserAgent,
		IPAddress:          ur.IPAddress,
		CreatedAt:          ur.CreatedAt,
		UpdatedAt:          ur.UpdatedAt,
	}
}

// UsageRecordResponse represents the usage record data structure for API responses
type UsageRecordResponse struct {
	ID                 uint      `json:"id" example:"1"`                                          // Record ID
	UserSubscriptionID uint      `json:"user_subscription_id" example:"1"`                        // Subscription ID
	UsageType          string    `json:"usage_type" example:"traffic"`                            // Usage type
	Amount             int64     `json:"amount" example:"1048576"`                                // Usage amount
	Unit               string    `json:"unit" example:"bytes"`                                    // Unit
	FormattedAmount    string    `json:"formatted_amount" example:"1.0 MB"`                       // Human readable amount
	Timestamp          time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`                // Usage timestamp
	SourceType         string    `json:"source_type" example:"server"`                            // Source type
	SourceID           string    `json:"source_id" example:"server-123"`                          // Source ID
	Metadata           string    `json:"metadata,omitempty" example:"{\"location\":\"us-west\"}"` // Metadata
	UserAgent          string    `json:"user_agent,omitempty" example:"Shadowsocks/1.0"`          // User agent
	IPAddress          string    `json:"ip_address,omitempty" example:"192.168.1.1"`              // IP address
	CreatedAt          time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`               // Creation time
	UpdatedAt          time.Time `json:"updated_at" example:"2024-01-01T12:00:00Z"`               // Update time
}
