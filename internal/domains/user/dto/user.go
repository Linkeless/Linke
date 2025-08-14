package dto

import (
	"time"
)

// Core user request structures

// CreateUserRequest represents the request structure for creating a new user
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email,max=255" example:"user@example.com"`
	Username string `json:"username" binding:"omitempty,max=100" example:"johndoe"`
	Name     string `json:"name" binding:"omitempty,max=255" example:"John Doe"`
	Password string `json:"password" binding:"omitempty,min=6,max=255" example:"password123"`
	Role     string `json:"role" binding:"omitempty,oneof=user admin" example:"user"`
	Status   string `json:"status" binding:"omitempty,oneof=active inactive banned" example:"active"`
}

// UpdateUserRequest represents the request structure for updating a user
type UpdateUserRequest struct {
	Email    *string `json:"email,omitempty" binding:"omitempty,email,max=255" example:"user@example.com"`
	Username *string `json:"username,omitempty" binding:"omitempty,max=100" example:"johndoe"`
	Name     *string `json:"name,omitempty" binding:"omitempty,max=255" example:"John Doe"`
	Role     *string `json:"role,omitempty" binding:"omitempty,oneof=user admin" example:"user"`
	Status   *string `json:"status,omitempty" binding:"omitempty,oneof=active inactive banned" example:"active"`
}

// PatchUserRequest represents the request body for patching user fields
type PatchUserRequest struct {
	Name     *string `json:"name,omitempty" example:"John Doe"`
	Email    *string `json:"email,omitempty" example:"user@example.com"`
	Username *string `json:"username,omitempty" example:"johndoe"`
	Role     *string `json:"role,omitempty" example:"user" enums:"user,admin"`
	Status   *string `json:"status,omitempty" example:"active" enums:"active,inactive,banned"`
}

// UpdateUserRoleRequest represents the request body for updating user role
type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user admin" example:"user"`
}

// UpdateUserStatusRequest represents the request body for updating user status
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive banned" example:"active"`
}

// BatchUserIDsRequest represents the request body for batch operations on users
type BatchUserIDsRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=100"`
}

// ResetPasswordRequest represents the request structure for admin password reset
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6,max=255" example:"newSecurePassword123"`
}

// UserProfileUpdateRequest represents the structure for profile updates
type UserProfileUpdateRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

// AdvancedUserSearchRequest represents advanced search parameters for users
type AdvancedUserSearchRequest struct {
	Query         string `form:"query" json:"query"`
	Status        string `form:"status" json:"status"`
	Provider      string `form:"provider" json:"provider"`
	Role          string `form:"role" json:"role"`
	EmailVerified *bool  `form:"email_verified" json:"email_verified"`
	Limit         int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=100"`
	Offset        int    `form:"offset" json:"offset" binding:"omitempty,min=0"`
}

// User response structures

// UserResponse represents the user data structure for API responses
// Fields are ordered to match the User model for consistency
type UserResponse struct {
	// Primary Key
	ID uint `json:"id"`

	// Core Identity Fields
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`

	// Authentication Fields (excluding password)
	Provider string `json:"provider"`
	Status   string `json:"status"`
	Role     string `json:"role"`

	// OAuth Provider IDs (only show if not empty)
	GoogleID   *string `json:"google_id,omitempty"`
	GitHubID   *string `json:"github_id,omitempty"`
	TelegramID *string `json:"telegram_id,omitempty"`

	// Provider Metadata (only show if not empty)
	ProviderData *string `json:"provider_data,omitempty"`

	// Invite Code Fields
	InviteCodeID   *uint   `json:"invite_code_id,omitempty"`
	InviteCodeUsed *string `json:"invite_code_used,omitempty"`

	// Timestamp Fields
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserProfile represents a user profile with computed fields
type UserProfile struct {
	User               UserResponse `json:"user"`
	IsEmailVerified    bool         `json:"is_email_verified"`
	AccountAge         int          `json:"account_age_days"`
	LastLoginDaysAgo   int          `json:"last_login_days_ago"`
	HasProfilePicture  bool         `json:"has_profile_picture"`
	SubscriptionStatus string       `json:"subscription_status,omitempty"`
	TotalOrders        int64        `json:"total_orders,omitempty"`
	TotalSpent         float64      `json:"total_spent,omitempty"`
}

// Statistics and analysis structures

// UserStats represents user statistics
type UserStats struct {
	TotalUsers    int64            `json:"total_users"`
	ActiveUsers   int64            `json:"active_users"`
	InactiveUsers int64            `json:"inactive_users"`
	BannedUsers   int64            `json:"banned_users"`
	DeletedUsers  int64            `json:"deleted_users"`
	ByProvider    map[string]int64 `json:"by_provider"`
	RecentSignups int64            `json:"recent_signups"`
}

// BatchOperationResult represents the result of batch operations
type BatchOperationResult struct {
	DeletedCount  int    `json:"deleted_count,omitempty"`
	RestoredCount int    `json:"restored_count,omitempty"`
	FailedIDs     []uint `json:"failed_ids,omitempty"`
}

// Admin-specific structures

// AdminUserListRequest represents admin request for listing users
type AdminUserListRequest struct {
	Page     int    `form:"page" json:"page" binding:"omitempty,min=1"`
	Limit    int    `form:"limit" json:"limit" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" json:"search"`
	Status   string `form:"status" json:"status"`
	Role     string `form:"role" json:"role"`
	Provider string `form:"provider" json:"provider"`
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Users []*UserResponse `json:"users"`
	Total int64           `json:"total"`
}

// Enhanced user analytics structures

// UserAnalytics represents comprehensive user analytics
type UserAnalytics struct {
	Overview      UserOverview      `json:"overview"`
	Demographics  UserDemographics  `json:"demographics"`
	ActivityStats UserActivityStats `json:"activity_stats"`
	TrendAnalysis UserTrendAnalysis `json:"trend_analysis"`
}

// UserOverview represents overall user statistics
type UserOverview struct {
	TotalUsers        int64   `json:"total_users"`
	ActiveUsers       int64   `json:"active_users"`
	NewUsersToday     int64   `json:"new_users_today"`
	NewUsersThisWeek  int64   `json:"new_users_this_week"`
	NewUsersThisMonth int64   `json:"new_users_this_month"`
	GrowthRate        float64 `json:"growth_rate"`
}

// UserDemographics represents user demographic data
type UserDemographics struct {
	ByProvider      map[string]int64 `json:"by_provider"`
	ByRole          map[string]int64 `json:"by_role"`
	ByStatus        map[string]int64 `json:"by_status"`
	EmailVerified   int64            `json:"email_verified"`
	EmailUnverified int64            `json:"email_unverified"`
}

// UserActivityStats represents user activity statistics
type UserActivityStats struct {
	ActiveToday            int64   `json:"active_today"`
	ActiveThisWeek         int64   `json:"active_this_week"`
	ActiveThisMonth        int64   `json:"active_this_month"`
	AverageSessionDuration float64 `json:"average_session_duration_minutes"`
	RetentionRate          float64 `json:"retention_rate"`
}

// UserTrendAnalysis represents user trend data
type UserTrendAnalysis struct {
	DailyRegistrations []DailyRegistrationTrend `json:"daily_registrations"`
	MonthlyGrowth      []MonthlyGrowthTrend     `json:"monthly_growth"`
	ProviderTrends     []ProviderTrendData      `json:"provider_trends"`
	ActivityTrends     []ActivityTrendData      `json:"activity_trends"`
}

// DailyRegistrationTrend represents daily registration trend data
type DailyRegistrationTrend struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// MonthlyGrowthTrend represents monthly growth data
type MonthlyGrowthTrend struct {
	Month         string  `json:"month"`
	NewUsers      int64   `json:"new_users"`
	GrowthRate    float64 `json:"growth_rate"`
	RetentionRate float64 `json:"retention_rate"`
}

// ProviderTrendData represents provider-specific trend data
type ProviderTrendData struct {
	Provider   string  `json:"provider"`
	Trend      string  `json:"trend"` // "increasing", "decreasing", "stable"
	GrowthRate float64 `json:"growth_rate"`
	Popularity float64 `json:"popularity"` // percentage of total users
}

// ActivityTrendData represents user activity trend data
type ActivityTrendData struct {
	Date           string  `json:"date"`
	ActiveUsers    int64   `json:"active_users"`
	SessionCount   int64   `json:"session_count"`
	AverageSession float64 `json:"average_session_duration"`
}
