package interfaces

import (
	"context"
	"linke/internal/domains/referral/entities"
)

// ReferralService defines the interface for referral operations
type ReferralService interface {
	// Referral CRUD operations
	CreateReferral(ctx context.Context, req *CreateReferralRequest) (*entities.Referral, error)
	GetReferral(ctx context.Context, referralID uint) (*entities.Referral, error)
	GetReferralByCode(ctx context.Context, code string) (*entities.Referral, error)
	UpdateReferral(ctx context.Context, referralID uint, req *UpdateReferralRequest) (*entities.Referral, error)

	// Referral listing and filtering
	GetReferrals(ctx context.Context, req *GetReferralsRequest) ([]*entities.Referral, int64, error)
	GetUserReferrals(ctx context.Context, userID uint, limit, offset int) ([]*entities.Referral, int64, error)
	GetRefereeReferrals(ctx context.Context, userID uint, limit, offset int) ([]*entities.Referral, int64, error)

	// Referral status management
	ConfirmReferral(ctx context.Context, referralID uint) error
	ProcessReferralReward(ctx context.Context, referralID uint, rewardAmount float64) error
	MarkReferralAsPaid(ctx context.Context, referralID uint) error

	// Referral statistics
	GetReferralStatistics(ctx context.Context, userID uint) (map[string]interface{}, error)
	GetSystemReferralStatistics(ctx context.Context) (map[string]interface{}, error)
}

// CreateReferralRequest represents the request to create a referral
type CreateReferralRequest struct {
	ReferrerID       uint                   `json:"referrer_id" binding:"required"`
	RefereeID        uint                   `json:"referee_id" binding:"required"`
	InviteCodeID     *uint                  `json:"invite_code_id,omitempty"`
	ReferralSource   string                 `json:"referral_source" binding:"required"`
	ReferralChannel  string                 `json:"referral_channel,omitempty"`
	ReferralCode     string                 `json:"referral_code,omitempty"`
	CampaignID       *uint                  `json:"campaign_id,omitempty"`
	AttributionData  map[string]interface{} `json:"attribution_data,omitempty"`
	ConversionValue  float64                `json:"conversion_value,omitempty"`
	ConversionType   string                 `json:"conversion_type,omitempty"`
	ExpirationDays   int                    `json:"expiration_days,omitempty"`
}

// UpdateReferralRequest represents the request to update a referral
type UpdateReferralRequest struct {
	Status          *string  `json:"status,omitempty"`
	RefereeStatus   *string  `json:"referee_status,omitempty"`
	RewardStatus    *string  `json:"reward_status,omitempty"`
	RewardAmount    *float64 `json:"reward_amount,omitempty"`
	ConversionValue *float64 `json:"conversion_value,omitempty"`
	ConversionType  *string  `json:"conversion_type,omitempty"`
}

// GetReferralsRequest represents the request to get referrals with filters
type GetReferralsRequest struct {
	ReferrerID    uint   `form:"referrer_id,omitempty"`
	RefereeID     uint   `form:"referee_id,omitempty"`
	Status        string `form:"status,omitempty"`
	RewardStatus  string `form:"reward_status,omitempty"`
	CampaignID    *uint  `form:"campaign_id,omitempty"`
	DateFrom      string `form:"date_from,omitempty"`
	DateTo        string `form:"date_to,omitempty"`
	Limit         int    `form:"limit,omitempty"`
	Offset        int    `form:"offset,omitempty"`
}