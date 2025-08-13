package constants

// Alert Status Constants
const (
	AlertStatusFired        = "fired"
	AlertStatusResolved     = "resolved"
	AlertStatusSuppressed   = "suppressed"
	AlertStatusAcknowledged = "acknowledged"
)

// Alert Severity Constants
const (
	AlertSeverityInfo     = "info"
	AlertSeverityWarning  = "warning"
	AlertSeverityError    = "error"
	AlertSeverityCritical = "critical"
)

// Priority Constants
const (
	PriorityLow      = "low"
	PriorityMedium   = "medium"
	PriorityHigh     = "high"
	PriorityCritical = "critical"
)

// Threshold Type Constants
const (
	ThresholdTypePercentage = "percentage"
	ThresholdTypeAbsolute   = "absolute"
)

// Notification Channel Type Constants
const (
	NotificationChannelEmail    = "email"
	NotificationChannelWebhook  = "webhook"
	NotificationChannelInApp    = "in_app"
	NotificationChannelSMS      = "sms"
	NotificationChannelSlack    = "slack"
	NotificationChannelDiscord  = "discord"
	NotificationChannelTelegram = "telegram"
)

// Default Alert Values
const (
	DefaultCooldownMinutes = 60
	MaxAlertsPerHour       = 10
	MaxAlertsPerDay        = 100
)
