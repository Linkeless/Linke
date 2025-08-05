package interfaces

import (
	"context"
	"linke/internal/domains/referral/entities"
	"time"
)

// ReferralCampaignService defines the interface for referral campaign operations
type ReferralCampaignService interface {
	// Campaign CRUD operations
	CreateReferralCampaign(ctx context.Context, req *CreateReferralCampaignRequest) (*entities.ReferralCampaign, error)
	GetReferralCampaign(ctx context.Context, campaignID uint) (*entities.ReferralCampaign, error)
	UpdateReferralCampaign(ctx context.Context, campaignID uint, req *UpdateReferralCampaignRequest) (*entities.ReferralCampaign, error)
	DeleteReferralCampaign(ctx context.Context, campaignID uint) error

	// Campaign listing and filtering
	GetReferralCampaigns(ctx context.Context, req *GetReferralCampaignsRequest) ([]*entities.ReferralCampaign, int64, error)
	GetActiveCampaigns(ctx context.Context) ([]*entities.ReferralCampaign, error)

	// Campaign management
	ActivateCampaign(ctx context.Context, campaignID uint) error
	DeactivateCampaign(ctx context.Context, campaignID uint) error
	ExpireCampaign(ctx context.Context, campaignID uint) error

	// Campaign statistics
	GetCampaignStatistics(ctx context.Context, campaignID uint) (map[string]any, error)
	GetCampaignPerformance(ctx context.Context, campaignID uint, fromDate, toDate time.Time) (map[string]any, error)
}

// CreateReferralCampaignRequest represents the request to create a referral campaign
type CreateReferralCampaignRequest struct {
	Name                 string     `json:"name" binding:"required,max=255" example:"Summer Referral Campaign"`
	Description          string     `json:"description,omitempty" binding:"max=1000" example:"Refer friends and earn rewards"`
	Type                 string     `json:"type" binding:"required,oneof=standard bonus limited" example:"standard"`
	Status               string     `json:"status,omitempty" example:"active"`
	StartDate            *time.Time `json:"start_date,omitempty" example:"2024-06-01T00:00:00Z"`
	EndDate              *time.Time `json:"end_date,omitempty" example:"2024-08-31T23:59:59Z"`
	ReferrerRewardAmount float64    `json:"referrer_reward_amount" binding:"required,min=0" example:"10.00"`
	RefereeRewardAmount  float64    `json:"referee_reward_amount,omitempty" example:"5.00"`
	RewardCurrency       string     `json:"reward_currency,omitempty" example:"USD"`
	MaxRewards           int        `json:"max_rewards,omitempty" example:"1000"`
	MinReferrals         int        `json:"min_referrals,omitempty" example:"1"`
	MaxReferrals         int        `json:"max_referrals,omitempty" example:"10"`
	Terms                string     `json:"terms,omitempty" binding:"max=5000" example:"Terms and conditions"`
	IsPublic             bool       `json:"is_public,omitempty" example:"true"`
}

// UpdateReferralCampaignRequest represents the request to update a referral campaign
type UpdateReferralCampaignRequest struct {
	Name                 *string    `json:"name,omitempty" binding:"omitempty,max=255"`
	Description          *string    `json:"description,omitempty" binding:"omitempty,max=1000"`
	Status               *string    `json:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	ReferrerRewardAmount *float64   `json:"referrer_reward_amount,omitempty" binding:"omitempty,min=0"`
	RefereeRewardAmount  *float64   `json:"referee_reward_amount,omitempty" binding:"omitempty,min=0"`
	MaxRewards           *int       `json:"max_rewards,omitempty"`
	MaxReferrals         *int       `json:"max_referrals,omitempty"`
	Terms                *string    `json:"terms,omitempty" binding:"omitempty,max=5000"`
	IsPublic             *bool      `json:"is_public,omitempty"`
}

// GetReferralCampaignsRequest represents the request to get referral campaigns with filters
type GetReferralCampaignsRequest struct {
	Status   string `form:"status,omitempty" binding:"omitempty,oneof=active inactive expired"`
	Type     string `form:"type,omitempty" binding:"omitempty,oneof=standard bonus limited"`
	IsPublic *bool  `form:"is_public,omitempty"`
	Limit    int    `form:"limit,omitempty" binding:"omitempty,min=1,max=100"`
	Offset   int    `form:"offset,omitempty" binding:"omitempty,min=0"`
}
