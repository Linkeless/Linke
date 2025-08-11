package entities

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/subscription/constants"
)

// AlertConfiguration represents alert configuration for usage monitoring
type AlertConfiguration struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Key
	UserSubscriptionID uint `json:"user_subscription_id" gorm:"not null;index"`

	// Alert Settings
	UsageType     string  `json:"usage_type" gorm:"size:50;not null;index;comment:Type of usage to monitor"`
	ThresholdType string  `json:"threshold_type" gorm:"size:20;not null;default:'percentage';comment:Type of threshold (percentage, absolute)"`
	Threshold     float64 `json:"threshold" gorm:"not null;comment:Alert threshold value"`
	IsEnabled     bool    `json:"is_enabled" gorm:"not null;default:true;index;comment:Whether alert is enabled"`

	// Notification Settings
	NotificationChannels string `json:"notification_channels" gorm:"type:text;comment:JSON array of notification channels"`
	CooldownMinutes      int    `json:"cooldown_minutes" gorm:"not null;default:60;comment:Cooldown period between alerts in minutes"`

	// Metadata
	Name        string `json:"name" gorm:"size:100;not null;comment:Human readable name for the alert"`
	Description string `json:"description" gorm:"type:text;comment:Description of what this alert monitors"`
	Priority    string `json:"priority" gorm:"size:20;not null;default:'medium';comment:Alert priority level"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for AlertConfiguration model
func (AlertConfiguration) TableName() string {
	return "alert_configurations"
}

// UsageAlert represents a fired usage alert
type UsageAlert struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserSubscriptionID   uint `json:"user_subscription_id" gorm:"not null;index"`
	AlertConfigurationID uint `json:"alert_configuration_id" gorm:"not null;index"`

	// Alert Details
	UsageType      string  `json:"usage_type" gorm:"size:50;not null;index"`
	CurrentUsage   int64   `json:"current_usage" gorm:"not null;comment:Current usage amount when alert fired"`
	UsageLimit     int64   `json:"usage_limit" gorm:"not null;comment:Usage limit at time of alert"`
	ThresholdValue float64 `json:"threshold_value" gorm:"not null;comment:Threshold that was exceeded"`
	UsagePercent   float64 `json:"usage_percent" gorm:"not null;comment:Percentage of limit used"`

	// Alert State
	Status     string     `json:"status" gorm:"size:20;not null;default:'fired';index;comment:Alert status"`
	Severity   string     `json:"severity" gorm:"size:20;not null;comment:Alert severity level"`
	FiredAt    time.Time  `json:"fired_at" gorm:"not null;index;comment:When the alert was first fired"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty" gorm:"index;comment:When the alert was resolved"`

	// Notification Tracking
	NotificationsSent    int        `json:"notifications_sent" gorm:"not null;default:0;comment:Number of notifications sent"`
	LastNotificationSent *time.Time `json:"last_notification_sent,omitempty" gorm:"index;comment:When last notification was sent"`
	NotificationChannels string     `json:"notification_channels" gorm:"type:text;comment:Channels used for notifications"`
	NotificationResults  string     `json:"notification_results" gorm:"type:text;comment:Results of notification attempts as JSON"`

	// Additional Context
	Message  string `json:"message" gorm:"type:text;comment:Alert message"`
	Metadata string `json:"metadata,omitempty" gorm:"type:text;comment:Additional metadata as JSON"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for UsageAlert model
func (UsageAlert) TableName() string {
	return "usage_alerts"
}

// NotificationChannel represents a notification channel configuration
type NotificationChannel struct {
	Type     string            `json:"type"`     // email, webhook, in_app, sms, telegram
	Target   string            `json:"target"`   // email address, webhook URL, telegram chat ID, etc.
	Settings map[string]string `json:"settings"` // additional settings for the channel
	Enabled  bool              `json:"enabled"`
}

// NotificationResult represents the result of a notification attempt
type NotificationResult struct {
	Channel string    `json:"channel"`
	Target  string    `json:"target"`
	Success bool      `json:"success"`
	Message string    `json:"message"`
	SentAt  time.Time `json:"sent_at" swaggertype:"string" format:"date-time" example:"2024-01-01T12:00:00Z"`
	Error   string    `json:"error,omitempty"`
}

// AlertConfiguration methods

// GetNotificationChannels returns the parsed notification channels
func (ac *AlertConfiguration) GetNotificationChannels() []NotificationChannel {
	if ac.NotificationChannels == "" {
		return nil
	}

	var channels []NotificationChannel
	if err := json.Unmarshal([]byte(ac.NotificationChannels), &channels); err != nil {
		return nil
	}

	return channels
}

// SetNotificationChannels sets the notification channels as JSON
func (ac *AlertConfiguration) SetNotificationChannels(channels []NotificationChannel) error {
	if len(channels) == 0 {
		ac.NotificationChannels = ""
		return nil
	}

	jsonBytes, err := json.Marshal(channels)
	if err != nil {
		return fmt.Errorf("failed to marshal notification channels: %w", err)
	}

	ac.NotificationChannels = string(jsonBytes)
	return nil
}

// IsActive checks if the alert configuration is active
func (ac *AlertConfiguration) IsActive() bool {
	return ac.IsEnabled && !ac.IsDeleted()
}

// IsDeleted checks if the alert configuration is soft deleted
func (ac *AlertConfiguration) IsDeleted() bool {
	return ac.DeletedAt.Valid
}

// ShouldTrigger checks if an alert should trigger based on usage
func (ac *AlertConfiguration) ShouldTrigger(currentUsage, usageLimit int64) bool {
	if !ac.IsActive() {
		return false
	}

	switch ac.ThresholdType {
	case constants.ThresholdTypePercentage:
		if usageLimit == 0 {
			return false // No limit, can't calculate percentage
		}
		percentage := float64(currentUsage) / float64(usageLimit) * 100
		return percentage >= ac.Threshold
	case constants.ThresholdTypeAbsolute:
		return float64(currentUsage) >= ac.Threshold
	default:
		return false
	}
}

// GetSeverityLevel returns the severity level based on usage percentage
func (ac *AlertConfiguration) GetSeverityLevel(currentUsage, usageLimit int64) string {
	if usageLimit == 0 {
		return constants.AlertSeverityInfo
	}

	percentage := float64(currentUsage) / float64(usageLimit) * 100

	switch {
	case percentage >= 100:
		return constants.AlertSeverityCritical
	case percentage >= 90:
		return constants.AlertSeverityError
	case percentage >= 80:
		return constants.AlertSeverityWarning
	default:
		return constants.AlertSeverityInfo
	}
}

// ToResponse converts AlertConfiguration to response structure
func (ac *AlertConfiguration) ToResponse() *AlertConfigurationResponse {
	return &AlertConfigurationResponse{
		ID:                   ac.ID,
		UserSubscriptionID:   ac.UserSubscriptionID,
		UsageType:            ac.UsageType,
		ThresholdType:        ac.ThresholdType,
		Threshold:            ac.Threshold,
		IsEnabled:            ac.IsEnabled,
		NotificationChannels: ac.GetNotificationChannels(),
		CooldownMinutes:      ac.CooldownMinutes,
		Name:                 ac.Name,
		Description:          ac.Description,
		Priority:             ac.Priority,
		CreatedAt:            ac.CreatedAt,
		UpdatedAt:            ac.UpdatedAt,
	}
}

// UsageAlert methods

// GetNotificationResults returns the parsed notification results
func (ua *UsageAlert) GetNotificationResults() []NotificationResult {
	if ua.NotificationResults == "" {
		return nil
	}

	var results []NotificationResult
	if err := json.Unmarshal([]byte(ua.NotificationResults), &results); err != nil {
		return nil
	}

	return results
}

// SetNotificationResults sets the notification results as JSON
func (ua *UsageAlert) SetNotificationResults(results []NotificationResult) error {
	if len(results) == 0 {
		ua.NotificationResults = ""
		return nil
	}

	jsonBytes, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal notification results: %w", err)
	}

	ua.NotificationResults = string(jsonBytes)
	return nil
}

// AddNotificationResult adds a notification result to the alert
func (ua *UsageAlert) AddNotificationResult(result NotificationResult) error {
	results := ua.GetNotificationResults()
	results = append(results, result)
	return ua.SetNotificationResults(results)
}

// IsActive checks if the alert is still active (not resolved)
func (ua *UsageAlert) IsActive() bool {
	return ua.Status == constants.AlertStatusFired && ua.ResolvedAt == nil && !ua.IsDeleted()
}

// IsResolved checks if the alert has been resolved
func (ua *UsageAlert) IsResolved() bool {
	return ua.Status == constants.AlertStatusResolved && ua.ResolvedAt != nil
}

// IsDeleted checks if the alert is soft deleted
func (ua *UsageAlert) IsDeleted() bool {
	return ua.DeletedAt.Valid
}

// Resolve marks the alert as resolved
func (ua *UsageAlert) Resolve() {
	now := time.Now()
	ua.Status = constants.AlertStatusResolved
	ua.ResolvedAt = &now
}

// GetDurationSinceFirstFired returns the duration since the alert was first fired
func (ua *UsageAlert) GetDurationSinceFirstFired() time.Duration {
	return time.Since(ua.FiredAt)
}

// CanSendNotification checks if a notification can be sent based on cooldown
func (ua *UsageAlert) CanSendNotification(cooldownMinutes int) bool {
	if ua.LastNotificationSent == nil {
		return true
	}

	cooldownDuration := time.Duration(cooldownMinutes) * time.Minute
	return time.Since(*ua.LastNotificationSent) >= cooldownDuration
}

// ToResponse converts UsageAlert to response structure
func (ua *UsageAlert) ToResponse() *UsageAlertResponse {
	return &UsageAlertResponse{
		ID:                      ua.ID,
		UserSubscriptionID:      ua.UserSubscriptionID,
		AlertConfigurationID:    ua.AlertConfigurationID,
		UsageType:               ua.UsageType,
		CurrentUsage:            ua.CurrentUsage,
		UsageLimit:              ua.UsageLimit,
		ThresholdValue:          ua.ThresholdValue,
		UsagePercent:            ua.UsagePercent,
		Status:                  ua.Status,
		Severity:                ua.Severity,
		FiredAt:                 ua.FiredAt,
		ResolvedAt:              ua.ResolvedAt,
		NotificationsSent:       ua.NotificationsSent,
		LastNotificationSent:    ua.LastNotificationSent,
		NotificationChannels:    ua.NotificationChannels,
		NotificationResults:     ua.GetNotificationResults(),
		Message:                 ua.Message,
		Metadata:                ua.Metadata,
		IsActive:                ua.IsActive(),
		IsResolved:              ua.IsResolved(),
		DurationSinceFirstFired: ua.GetDurationSinceFirstFired().String(),
		CreatedAt:               ua.CreatedAt,
		UpdatedAt:               ua.UpdatedAt,
	}
}

// Response structures for API

// AlertConfigurationResponse represents the alert configuration data structure for API responses
type AlertConfigurationResponse struct {
	ID                   uint                  `json:"id" example:"1"`                                                                    // Configuration ID
	UserSubscriptionID   uint                  `json:"user_subscription_id" example:"1"`                                                  // Subscription ID
	UsageType            string                `json:"usage_type" example:"traffic"`                                                      // Usage type to monitor
	ThresholdType        string                `json:"threshold_type" example:"percentage"`                                               // Threshold type
	Threshold            float64               `json:"threshold" example:"80.0"`                                                          // Threshold value
	IsEnabled            bool                  `json:"is_enabled" example:"true"`                                                         // Whether enabled
	NotificationChannels []NotificationChannel `json:"notification_channels"`                                                             // Notification channels
	CooldownMinutes      int                   `json:"cooldown_minutes" example:"60"`                                                     // Cooldown in minutes
	Name                 string                `json:"name" example:"Traffic 80% Alert"`                                                  // Alert name
	Description          string                `json:"description" example:"Alert when traffic usage reaches 80%"`                        // Description
	Priority             string                `json:"priority" example:"medium"`                                                         // Priority level
	CreatedAt            time.Time             `json:"created_at" swaggertype:"string" format:"date-time" example:"2024-01-01T00:00:00Z"` // Creation time
	UpdatedAt            time.Time             `json:"updated_at" swaggertype:"string" format:"date-time" example:"2024-01-01T00:00:00Z"` // Update time
}

// UsageAlertResponse represents the usage alert data structure for API responses
type UsageAlertResponse struct {
	ID                      uint                 `json:"id" example:"1"`                                                                                          // Alert ID
	UserSubscriptionID      uint                 `json:"user_subscription_id" example:"1"`                                                                        // Subscription ID
	AlertConfigurationID    uint                 `json:"alert_configuration_id" example:"1"`                                                                      // Configuration ID
	UsageType               string               `json:"usage_type" example:"traffic"`                                                                            // Usage type
	CurrentUsage            int64                `json:"current_usage" example:"8589934592"`                                                                      // Current usage (bytes)
	UsageLimit              int64                `json:"usage_limit" example:"10737418240"`                                                                       // Usage limit (bytes)
	ThresholdValue          float64              `json:"threshold_value" example:"80.0"`                                                                          // Threshold that was exceeded
	UsagePercent            float64              `json:"usage_percent" example:"80.5"`                                                                            // Usage percentage
	Status                  string               `json:"status" example:"fired"`                                                                                  // Alert status
	Severity                string               `json:"severity" example:"warning"`                                                                              // Alert severity
	FiredAt                 time.Time            `json:"fired_at" swaggertype:"string" format:"date-time" example:"2024-01-01T12:00:00Z"`                         // When fired
	ResolvedAt              *time.Time           `json:"resolved_at,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-01T13:00:00Z"`            // When resolved
	NotificationsSent       int                  `json:"notifications_sent" example:"3"`                                                                          // Notifications sent count
	LastNotificationSent    *time.Time           `json:"last_notification_sent,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-01T12:30:00Z"` // Last notification time
	NotificationChannels    string               `json:"notification_channels"`                                                                                   // Notification channels used
	NotificationResults     []NotificationResult `json:"notification_results"`                                                                                    // Notification results
	Message                 string               `json:"message" example:"Traffic usage has reached 80% of your limit"`                                           // Alert message
	Metadata                string               `json:"metadata,omitempty"`                                                                                      // Additional metadata
	IsActive                bool                 `json:"is_active" example:"true"`                                                                                // Whether alert is active
	IsResolved              bool                 `json:"is_resolved" example:"false"`                                                                             // Whether alert is resolved
	DurationSinceFirstFired string               `json:"duration_since_first_fired" example:"30m" swaggertype:"string"`                                           // Duration since first fired (as string)
	CreatedAt               time.Time            `json:"created_at" swaggertype:"string" format:"date-time" example:"2024-01-01T12:00:00Z"`                       // Creation time
	UpdatedAt               time.Time            `json:"updated_at" swaggertype:"string" format:"date-time" example:"2024-01-01T12:00:00Z"`                       // Update time
}
