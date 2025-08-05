package interfaces

import (
	"context"
	"linke/internal/domains/subscription/entities"
	"time"
)

// UserSubscriptionService defines the interface for user subscription operations
type UserSubscriptionService interface {
	// Subscription CRUD operations
	CreateUserSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*entities.UserSubscription, error)
	GetUserSubscription(ctx context.Context, subscriptionID uint) (*entities.UserSubscription, error)
	GetActiveUserSubscription(ctx context.Context, userID, planID uint) (*entities.UserSubscription, error)
	UpdateUserSubscription(ctx context.Context, subscriptionID uint, req *UpdateSubscriptionRequest) (*entities.UserSubscription, error)

	// Subscription listing
	GetUserSubscriptions(ctx context.Context, req *GetUserSubscriptionsRequest) ([]*entities.UserSubscription, int64, error)
	GetUserActiveSubscriptions(ctx context.Context, userID uint) ([]*entities.UserSubscription, error)

	// Subscription management
	CancelUserSubscription(ctx context.Context, subscriptionID uint, reason string, cancelAtPeriodEnd bool) error
	RenewUserSubscription(ctx context.Context, subscriptionID uint) error
	PauseUserSubscription(ctx context.Context, subscriptionID uint, req *PauseSubscriptionRequest, adminUserID uint) (*entities.UserSubscription, error)
	ResumeUserSubscription(ctx context.Context, subscriptionID uint, req *ResumeSubscriptionRequest, adminUserID uint) (*entities.UserSubscription, error)

	// Traffic and usage management
	UpdateTrafficUsage(ctx context.Context, subscriptionID uint, usedBytes int64) error
	ResetTrafficUsage(ctx context.Context, subscriptionID uint, adminUserID uint) (*entities.UserSubscription, error)
	GetSubscriptionTrafficStats(ctx context.Context, subscriptionID uint) (map[string]any, error)

	// Subscription expiry management
	CheckAndProcessExpiredSubscriptions(ctx context.Context) error
	ExtendSubscription(ctx context.Context, subscriptionID uint, extendByDays int, reason string) error

	// Statistics
	GetSubscriptionStatistics(ctx context.Context) (map[string]any, error)
}

// CreateSubscriptionRequest represents the request to create a user subscription
type CreateSubscriptionRequest struct {
	UserID             uint   `json:"user_id" binding:"required" example:"1"`
	SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required" example:"1"`
	StartDate          string `json:"start_date,omitempty" binding:"omitempty" example:"2024-01-01T00:00:00Z"`
	UseTrial           bool   `json:"use_trial,omitempty" example:"true"`
	ServerGroupIDs     []uint `json:"server_group_ids,omitempty"`

	// Custom Traffic Configuration (optional, overrides plan defaults)
	CustomTrafficLimit      *int64  `json:"custom_traffic_limit,omitempty" example:"107374182400"`  // Custom traffic limit in bytes
	CustomTrafficResetCycle *string `json:"custom_traffic_reset_cycle,omitempty" example:"monthly"` // Custom reset cycle
	DisableTrafficLimit     *bool   `json:"disable_traffic_limit,omitempty" example:"false"`        // Disable traffic limit for this subscription
}

// UpdateSubscriptionRequest represents the request to update a user subscription
type UpdateSubscriptionRequest struct {
	Status             *string    `json:"status,omitempty" binding:"omitempty,oneof=active paused cancelled expired trial" example:"active"`
	EndDate            *time.Time `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
	CancellationReason *string    `json:"cancellation_reason,omitempty" binding:"omitempty,max=255" example:"User request"`
	CancelAtPeriodEnd  *bool      `json:"cancel_at_period_end,omitempty" example:"true"`
	AutoRenew          *bool      `json:"auto_renew,omitempty" example:"true"`
	Notes              *string    `json:"notes,omitempty" binding:"omitempty,max=1000" example:"Customer feedback notes"`
	ServerGroupIDs     *[]uint    `json:"server_group_ids,omitempty"`
}

// GetUserSubscriptionsRequest represents the request to get user subscriptions
type GetUserSubscriptionsRequest struct {
	UserID uint   `form:"user_id,omitempty" example:"1"`
	Status string `form:"status,omitempty" binding:"omitempty,oneof=active paused cancelled expired trial" example:"active"`
	Limit  int    `form:"limit,omitempty" binding:"omitempty,min=1,max=100" example:"10"`
	Offset int    `form:"offset,omitempty" binding:"omitempty,min=0" example:"0"`
}

// PauseSubscriptionRequest represents the request to pause a user subscription
type PauseSubscriptionRequest struct {
	Reason           string `json:"reason" binding:"required,max=255" example:"User requested pause"`
	MaxPauseDuration *int   `json:"max_pause_duration,omitempty" binding:"omitempty,min=1,max=365" example:"90"` // Override default max pause duration
}

// ResumeSubscriptionRequest represents the request to resume a user subscription
type ResumeSubscriptionRequest struct {
	AdjustBillingDate bool `json:"adjust_billing_date,omitempty" example:"true"` // Whether to adjust billing dates based on pause duration
}
