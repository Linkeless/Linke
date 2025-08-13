package entities

import (
	"slices"
	"time"

	"gorm.io/gorm"
)

// UserAccountBinding represents a user's third-party account binding
type UserAccountBinding struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Key (no constraint, managed by application)
	UserID uint `json:"user_id" gorm:"not null;index"`

	// Provider Information
	Provider       string `json:"provider" gorm:"size:20;not null;index"` // google, github, telegram
	ProviderUserID string `json:"provider_user_id" gorm:"size:100;not null;index"`

	// Provider User Data
	ProviderEmail    *string `json:"provider_email,omitempty" gorm:"size:255;index"`
	ProviderUsername *string `json:"provider_username,omitempty" gorm:"size:100"`
	ProviderName     *string `json:"provider_name,omitempty" gorm:"size:255"`
	ProviderAvatar   *string `json:"provider_avatar,omitempty" gorm:"size:500"`
	ProviderData     *string `json:"provider_data,omitempty" gorm:"type:json"`

	// Binding Status
	IsPrimary  bool       `json:"is_primary" gorm:"default:false"`
	BoundAt    time.Time  `json:"bound_at" gorm:"not null;default:CURRENT_TIMESTAMP"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" gorm:"index"`

	// Timestamp Fields (GORM convention order)
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for UserAccountBinding model
func (UserAccountBinding) TableName() string {
	return "user_account_bindings"
}

// Provider constants
const (
	BindingProviderGoogle   = "google"
	BindingProviderGitHub   = "github"
	BindingProviderTelegram = "telegram"
)

// ValidProviders returns a list of valid provider names
func ValidProviders() []string {
	return []string{
		BindingProviderGoogle,
		BindingProviderGitHub,
		BindingProviderTelegram,
	}
}

// IsValidProvider checks if a provider is valid
func IsValidProvider(provider string) bool {
	return slices.Contains(ValidProviders(), provider)
}

// IsDeleted checks if the binding is soft deleted
func (uab *UserAccountBinding) IsDeleted() bool {
	return uab.DeletedAt.Valid
}

// UpdateLastUsed updates the last used timestamp
func (uab *UserAccountBinding) UpdateLastUsed(db *gorm.DB) error {
	now := time.Now()
	uab.LastUsedAt = &now
	return db.Model(uab).Update("last_used_at", now).Error
}

// SetAsPrimary sets this binding as the primary one for the user
// Note: This should be used within a transaction to ensure consistency
func (uab *UserAccountBinding) SetAsPrimary(db *gorm.DB) error {
	// First, unset all other bindings for this user
	if err := db.Model(&UserAccountBinding{}).
		Where("user_id = ? AND id != ?", uab.UserID, uab.ID).
		Update("is_primary", false).Error; err != nil {
		return err
	}

	// Then set this one as primary
	uab.IsPrimary = true
	return db.Model(uab).Update("is_primary", true).Error
}

// SoftDelete performs soft delete on the binding
func (uab *UserAccountBinding) SoftDelete(db *gorm.DB) error {
	return db.Delete(uab).Error
}

// UserAccountBindingResponse represents the binding data structure for API responses
type UserAccountBindingResponse struct {
	// Primary Key
	ID uint `json:"id"`

	// Provider Information
	Provider       string `json:"provider"`
	ProviderUserID string `json:"provider_user_id"`

	// Provider User Data (only show non-sensitive data)
	ProviderEmail    *string `json:"provider_email,omitempty"`
	ProviderUsername *string `json:"provider_username,omitempty"`
	ProviderName     *string `json:"provider_name,omitempty"`
	ProviderAvatar   *string `json:"provider_avatar,omitempty"`

	// Binding Status
	IsPrimary  bool       `json:"is_primary"`
	BoundAt    time.Time  `json:"bound_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`

	// Timestamp Fields
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ToResponse converts UserAccountBinding to UserAccountBindingResponse
func (uab *UserAccountBinding) ToResponse() *UserAccountBindingResponse {
	resp := &UserAccountBindingResponse{
		// Primary Key
		ID: uab.ID,

		// Provider Information
		Provider:       uab.Provider,
		ProviderUserID: uab.ProviderUserID,

		// Provider User Data
		ProviderEmail:    uab.ProviderEmail,
		ProviderUsername: uab.ProviderUsername,
		ProviderName:     uab.ProviderName,
		ProviderAvatar:   uab.ProviderAvatar,

		// Binding Status
		IsPrimary:  uab.IsPrimary,
		BoundAt:    uab.BoundAt,
		LastUsedAt: uab.LastUsedAt,

		// Timestamp Fields
		CreatedAt: uab.CreatedAt,
		UpdatedAt: uab.UpdatedAt,
	}

	// Set DeletedAt only if valid
	if uab.DeletedAt.Valid {
		resp.DeletedAt = &uab.DeletedAt.Time
	}

	return resp
}

// CreateBindingRequest represents the request structure for creating a new binding
type CreateBindingRequest struct {
	Provider         string  `json:"provider" binding:"required,oneof=google github telegram" example:"google"`
	ProviderUserID   string  `json:"provider_user_id" binding:"required,max=100" example:"123456789"`
	ProviderEmail    *string `json:"provider_email,omitempty" binding:"omitempty,email,max=255" example:"user@example.com"`
	ProviderUsername *string `json:"provider_username,omitempty" binding:"omitempty,max=100" example:"username"`
	ProviderName     *string `json:"provider_name,omitempty" binding:"omitempty,max=255" example:"User Name"`
	ProviderAvatar   *string `json:"provider_avatar,omitempty" binding:"omitempty,max=500" example:"https://example.com/avatar.jpg"`
	ProviderData     *string `json:"provider_data,omitempty" example:"{\"extra\": \"data\"}"`
	IsPrimary        *bool   `json:"is_primary,omitempty" example:"false"`
}

// UpdateBindingRequest represents the request structure for updating a binding
type UpdateBindingRequest struct {
	ProviderEmail    *string `json:"provider_email,omitempty" binding:"omitempty,email,max=255" example:"user@example.com"`
	ProviderUsername *string `json:"provider_username,omitempty" binding:"omitempty,max=100" example:"username"`
	ProviderName     *string `json:"provider_name,omitempty" binding:"omitempty,max=255" example:"User Name"`
	ProviderAvatar   *string `json:"provider_avatar,omitempty" binding:"omitempty,max=500" example:"https://example.com/avatar.jpg"`
	ProviderData     *string `json:"provider_data,omitempty" example:"{\"extra\": \"data\"}"`
	IsPrimary        *bool   `json:"is_primary,omitempty" example:"false"`
}

// BindingListResponse represents the response structure for listing bindings
type BindingListResponse struct {
	Bindings []*UserAccountBindingResponse `json:"bindings"`
	Total    int64                         `json:"total"`
}

// Enhanced structures for new features

// BatchBindingRequest represents a batch operation request
type BatchBindingRequest struct {
	Operation string             `json:"operation" binding:"required,oneof=create update delete activate deactivate" example:"create"`
	Bindings  []BatchBindingItem `json:"bindings" binding:"required,min=1,max=100"`
}

// BatchBindingItem represents an individual item in a batch operation
type BatchBindingItem struct {
	UserID           *uint   `json:"user_id,omitempty" binding:"omitempty,min=1"`
	Provider         string  `json:"provider" binding:"required,oneof=google github telegram"`
	ProviderUserID   string  `json:"provider_user_id" binding:"required,max=100"`
	ProviderEmail    *string `json:"provider_email,omitempty" binding:"omitempty,email,max=255"`
	ProviderUsername *string `json:"provider_username,omitempty" binding:"omitempty,max=100"`
	ProviderName     *string `json:"provider_name,omitempty" binding:"omitempty,max=255"`
	ProviderAvatar   *string `json:"provider_avatar,omitempty" binding:"omitempty,max=500"`
	ProviderData     *string `json:"provider_data,omitempty"`
	IsPrimary        *bool   `json:"is_primary,omitempty"`
	BindingID        *uint   `json:"binding_id,omitempty" binding:"omitempty,min=1"`
}

// BatchBindingResponse represents the response for batch operations
type BatchBindingResponse struct {
	SuccessCount int64                         `json:"success_count"`
	FailureCount int64                         `json:"failure_count"`
	Results      []BatchBindingOperationResult `json:"results"`
	Errors       []BatchBindingOperationError  `json:"errors,omitempty"`
}

// BatchBindingOperationResult represents a successful batch operation result
type BatchBindingOperationResult struct {
	Index   int                         `json:"index"`
	Binding *UserAccountBindingResponse `json:"binding,omitempty"`
	Message string                      `json:"message"`
}

// BatchBindingOperationError represents a failed batch operation
type BatchBindingOperationError struct {
	Index int              `json:"index"`
	Error string           `json:"error"`
	Item  BatchBindingItem `json:"item"`
}

// BindingAnalytics represents comprehensive binding analytics
type BindingAnalytics struct {
	Overview          BindingOverview          `json:"overview"`
	ProviderAnalytics map[string]ProviderStats `json:"provider_analytics"`
	UserAnalytics     UserBindingAnalytics     `json:"user_analytics"`
	SecurityMetrics   SecurityMetrics          `json:"security_metrics"`
	TrendAnalysis     TrendAnalysis            `json:"trend_analysis"`
}

// BindingOverview represents overall binding statistics
type BindingOverview struct {
	TotalBindings          int64   `json:"total_bindings"`
	ActiveBindings         int64   `json:"active_bindings"`
	InactiveBindings       int64   `json:"inactive_bindings"`
	UsersWithBindings      int64   `json:"users_with_bindings"`
	AverageBindingsPerUser float64 `json:"average_bindings_per_user"`
}

// ProviderStats represents statistics for a specific provider
type ProviderStats struct {
	Provider       string  `json:"provider"`
	BindingCount   int64   `json:"binding_count"`
	ActiveCount    int64   `json:"active_count"`
	PrimaryCount   int64   `json:"primary_count"`
	RecentActivity int64   `json:"recent_activity"` // last 30 days
	GrowthRate     float64 `json:"growth_rate"`     // month-over-month
}

// UserBindingAnalytics represents user-centric binding analytics
type UserBindingAnalytics struct {
	SingleBindingUsers   int64   `json:"single_binding_users"`
	MultipleBindingUsers int64   `json:"multiple_binding_users"`
	MaxBindingsPerUser   int64   `json:"max_bindings_per_user"`
	AverageBindingAge    float64 `json:"average_binding_age_days"`
}

// SecurityMetrics represents security-related metrics
type SecurityMetrics struct {
	SuspiciousBindings   int64   `json:"suspicious_bindings"`
	RecentFailedAttempts int64   `json:"recent_failed_attempts"`
	ConflictAttempts     int64   `json:"conflict_attempts"`
	SecurityScore        float64 `json:"security_score"` // 0-100
}

// TrendAnalysis represents binding trend data
type TrendAnalysis struct {
	DailyBindings  []DailyBindingTrend  `json:"daily_bindings"`
	MonthlyGrowth  []MonthlyGrowthTrend `json:"monthly_growth"`
	ProviderTrends []ProviderTrendData  `json:"provider_trends"`
}

// DailyBindingTrend represents daily binding trend data
type DailyBindingTrend struct {
	Date     string `json:"date"`
	Count    int64  `json:"count"`
	Provider string `json:"provider"`
}

// MonthlyGrowthTrend represents monthly growth data
type MonthlyGrowthTrend struct {
	Month       string  `json:"month"`
	NewBindings int64   `json:"new_bindings"`
	GrowthRate  float64 `json:"growth_rate"`
}

// ProviderTrendData represents provider-specific trend data
type ProviderTrendData struct {
	Provider   string  `json:"provider"`
	Trend      string  `json:"trend"` // "increasing", "decreasing", "stable"
	GrowthRate float64 `json:"growth_rate"`
	Popularity float64 `json:"popularity"` // percentage of total bindings
}

// BindingAuditLog represents audit log entry for binding operations
type BindingAuditLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	AdminID      *uint     `json:"admin_id,omitempty" gorm:"index"`
	BindingID    *uint     `json:"binding_id,omitempty" gorm:"index"`
	Operation    string    `json:"operation" gorm:"size:50;not null"`
	Provider     string    `json:"provider" gorm:"size:20;not null;index"`
	Details      string    `json:"details" gorm:"type:text"`
	IPAddress    string    `json:"ip_address" gorm:"size:45"`
	UserAgent    string    `json:"user_agent" gorm:"size:500"`
	Status       string    `json:"status" gorm:"size:20;not null;index"` // success, failure, warning
	ErrorMessage *string   `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null;index"`
}

// TableName returns the table name for BindingAuditLog model
func (BindingAuditLog) TableName() string {
	return "user_binding_audit_logs"
}

// BindingSecurityEvent represents security-related binding events
type BindingSecurityEvent struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	EventType      string     `json:"event_type" gorm:"size:50;not null;index"`
	Severity       string     `json:"severity" gorm:"size:20;not null;index"`
	UserID         *uint      `json:"user_id,omitempty" gorm:"index"`
	Provider       string     `json:"provider" gorm:"size:20;not null;index"`
	ProviderUserID string     `json:"provider_user_id" gorm:"size:100;index"`
	IPAddress      string     `json:"ip_address" gorm:"size:45;index"`
	UserAgent      string     `json:"user_agent" gorm:"size:500"`
	Description    string     `json:"description" gorm:"type:text;not null"`
	Metadata       *string    `json:"metadata,omitempty" gorm:"type:json"`
	Resolved       bool       `json:"resolved" gorm:"default:false;index"`
	ResolvedBy     *uint      `json:"resolved_by,omitempty" gorm:"index"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" gorm:"index"`
	CreatedAt      time.Time  `json:"created_at" gorm:"not null;index"`
}

// TableName returns the table name for BindingSecurityEvent model
func (BindingSecurityEvent) TableName() string {
	return "user_binding_security_events"
}

// Security event types and severities
const (
	// Event Types
	SecurityEventSuspiciousBinding  = "suspicious_binding"
	SecurityEventDuplicateAttempt   = "duplicate_attempt"
	SecurityEventRapidBinding       = "rapid_binding"
	SecurityEventUnusualProvider    = "unusual_provider"
	SecurityEventFailedValidation   = "failed_validation"
	SecurityEventMassBatchOperation = "mass_batch_operation"

	// Severity Levels
	SecuritySeverityLow      = "low"
	SecuritySeverityMedium   = "medium"
	SecuritySeverityHigh     = "high"
	SecuritySeverityCritical = "critical"
)

// BindingValidationConfig represents validation configuration
type BindingValidationConfig struct {
	MaxBindingsPerUser       int     `json:"max_bindings_per_user"`
	MaxBindingsPerProvider   int     `json:"max_bindings_per_provider"`
	RequireEmailVerification bool    `json:"require_email_verification"`
	BlockSuspiciousProviders bool    `json:"block_suspicious_providers"`
	EnableRateLimit          bool    `json:"enable_rate_limit"`
	RateLimitWindow          int     `json:"rate_limit_window_minutes"`
	RateLimitMaxAttempts     int     `json:"rate_limit_max_attempts"`
	SecurityScoreThreshold   float64 `json:"security_score_threshold"`
}

// DefaultBindingValidationConfig returns default validation configuration
func DefaultBindingValidationConfig() BindingValidationConfig {
	return BindingValidationConfig{
		MaxBindingsPerUser:       10,
		MaxBindingsPerProvider:   1,
		RequireEmailVerification: true,
		BlockSuspiciousProviders: true,
		EnableRateLimit:          true,
		RateLimitWindow:          60,
		RateLimitMaxAttempts:     5,
		SecurityScoreThreshold:   70.0,
	}
}
