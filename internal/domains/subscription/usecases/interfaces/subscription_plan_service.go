package interfaces

import (
	"context"
	"linke/internal/domains/subscription/entities"
)

// SubscriptionPlanService defines the interface for subscription plan operations
type SubscriptionPlanService interface {
	// Plan CRUD operations
	CreateSubscriptionPlan(ctx context.Context, creatorID uint, req *CreateSubscriptionPlanRequest) (*entities.SubscriptionPlan, error)
	GetSubscriptionPlan(ctx context.Context, planID uint) (*entities.SubscriptionPlan, error)
	GetSubscriptionPlanByCode(ctx context.Context, code string) (*entities.SubscriptionPlan, error)
	UpdateSubscriptionPlan(ctx context.Context, planID uint, req *UpdateSubscriptionPlanRequest) (*entities.SubscriptionPlan, error)
	DeleteSubscriptionPlan(ctx context.Context, planID uint) error

	// Plan listing and filtering
	GetSubscriptionPlans(ctx context.Context, req *GetSubscriptionPlansRequest) ([]*entities.SubscriptionPlan, int64, error)
	GetVisibleSubscriptionPlans(ctx context.Context, currency string) ([]*entities.SubscriptionPlan, error)
	GetPopularSubscriptionPlans(ctx context.Context, limit int) ([]*entities.SubscriptionPlan, error)

	// Plan management
	ToggleSubscriptionPlanStatus(ctx context.Context, planID uint) (*entities.SubscriptionPlan, error)
	ArchiveSubscriptionPlan(ctx context.Context, planID uint) error
}

// CreateSubscriptionPlanRequest represents the request to create a subscription plan
type CreateSubscriptionPlanRequest struct {
	Name            string  `json:"name" binding:"required,min=1,max=100" example:"Premium Plan"`
	Code            string  `json:"code" binding:"required,min=1,max=50" example:"premium-monthly"`
	Description     string  `json:"description" binding:"max=1000" example:"Premium features with monthly billing"`
	Price           float64 `json:"price" binding:"required,min=0" example:"29.99"`
	Currency        string  `json:"currency" binding:"required,len=3" example:"USD"`
	BillingCycle    string  `json:"billing_cycle" binding:"required,oneof=monthly yearly lifetime" example:"monthly"`
	BillingInterval int     `json:"billing_interval" binding:"min=1,max=12" example:"1"`
	TrialPeriodDays int     `json:"trial_period_days" binding:"min=0,max=365" example:"7"`
	Features        string  `json:"features,omitempty" example:"{\"max_projects\": 10, \"storage_gb\": 100}"`
	Limits          string  `json:"limits,omitempty" example:"{\"api_calls_per_month\": 10000}"`
	IsVisible       *bool   `json:"is_visible,omitempty" example:"true"`
	SortOrder       int     `json:"sort_order,omitempty" example:"1"`
	IsPopular       *bool   `json:"is_popular,omitempty" example:"false"`
	IsRecommended   *bool   `json:"is_recommended,omitempty" example:"true"`
	SetupFee        float64 `json:"setup_fee,omitempty" example:"0"`
	CancellationFee float64 `json:"cancellation_fee,omitempty" example:"0"`

	// Traffic Configuration (Required)
	TrafficLimit      int64  `json:"traffic_limit" binding:"required,min=0" example:"107374182400"`                // Traffic limit in bytes (0 = unlimited)
	TrafficResetCycle string `json:"traffic_reset_cycle" binding:"required,oneof=monthly never" example:"monthly"` // Traffic reset cycle

	// Server Group Configuration (Required)
	DefaultServerGroupIDs []uint `json:"default_server_group_ids" binding:"required,min=1" example:"[1]"` // Default server groups for subscriptions
}

// UpdateSubscriptionPlanRequest represents the request to update a subscription plan
type UpdateSubscriptionPlanRequest struct {
	Name            *string  `json:"name,omitempty" binding:"omitempty,min=1,max=100" example:"Premium Plan Updated"`
	Description     *string  `json:"description,omitempty" binding:"omitempty,max=1000" example:"Updated description"`
	Price           *float64 `json:"price,omitempty" binding:"omitempty,min=0" example:"39.99"`
	TrialPeriodDays *int     `json:"trial_period_days,omitempty" binding:"omitempty,min=0,max=365" example:"14"`
	Features        *string  `json:"features,omitempty" example:"{\"max_projects\": 20}"`
	Limits          *string  `json:"limits,omitempty" example:"{\"api_calls_per_month\": 20000}"`
	Status          *string  `json:"status,omitempty" binding:"omitempty,oneof=active inactive archived" example:"active"`
	IsVisible       *bool    `json:"is_visible,omitempty" example:"true"`
	SortOrder       *int     `json:"sort_order,omitempty" example:"2"`
	IsPopular       *bool    `json:"is_popular,omitempty" example:"true"`
	IsRecommended   *bool    `json:"is_recommended,omitempty" example:"false"`
	SetupFee        *float64 `json:"setup_fee,omitempty" example:"10"`
	CancellationFee *float64 `json:"cancellation_fee,omitempty" example:"25"`

	// Traffic Configuration
	TrafficLimit      *int64  `json:"traffic_limit,omitempty" binding:"omitempty,min=0" example:"107374182400"`                // Traffic limit in bytes
	TrafficResetCycle *string `json:"traffic_reset_cycle,omitempty" binding:"omitempty,oneof=monthly never" example:"monthly"` // Traffic reset cycle

	// Server Group Configuration
	DefaultServerGroupIDs *[]uint `json:"default_server_group_ids,omitempty" example:"[1]"` // Default server groups for subscriptions
}

// GetSubscriptionPlansRequest represents the request to get subscription plans
type GetSubscriptionPlansRequest struct {
	Status      string `form:"status" binding:"omitempty,oneof=active inactive archived" example:"active"`
	Currency    string `form:"currency" binding:"omitempty,len=3" example:"USD"`
	Visible     *bool  `form:"visible" example:"true"`
	Popular     *bool  `form:"popular" example:"false"`
	Recommended *bool  `form:"recommended" example:"true"`
	Limit       int    `form:"limit" binding:"omitempty,min=1,max=100" example:"10"`
	Offset      int    `form:"offset" binding:"omitempty,min=0" example:"0"`
}
