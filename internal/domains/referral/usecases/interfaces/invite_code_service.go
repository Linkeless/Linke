package interfaces

import (
	"context"
	"linke/internal/domains/referral/entities"
)

// InviteCodeService defines the interface for invite code operations
type InviteCodeService interface {
	// Invite code CRUD operations
	CreateInviteCode(ctx context.Context, createdByID uint, req *CreateInviteCodeRequest) (*entities.InviteCode, error)
	GetInviteCode(ctx context.Context, inviteCodeID uint) (*entities.InviteCode, error)
	GetInviteCodeByCode(ctx context.Context, code string) (*entities.InviteCode, error)
	UpdateInviteCode(ctx context.Context, inviteCodeID uint, req *UpdateInviteCodeRequest) (*entities.InviteCode, error)
	DeleteInviteCode(ctx context.Context, inviteCodeID uint) error

	// Invite code listing and filtering
	GetInviteCodes(ctx context.Context, req *GetInviteCodesRequest) ([]*entities.InviteCode, int64, error)
	GetUserInviteCodes(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCode, int64, error)

	// Invite code validation and usage
	ValidateInviteCode(ctx context.Context, code string) (*entities.InviteCode, error)
	UseInviteCode(ctx context.Context, code string, userID uint, ipAddress, userAgent string) (*entities.InviteCodeUsage, error)

	// Invite code management
	ActivateInviteCode(ctx context.Context, inviteCodeID uint) error
	DeactivateInviteCode(ctx context.Context, inviteCodeID uint) error
	ExpireInviteCode(ctx context.Context, inviteCodeID uint) error

	// Usage tracking
	GetInviteCodeUsage(ctx context.Context, inviteCodeID uint, limit, offset int) ([]*entities.InviteCodeUsage, int64, error)
	GetUserInviteCodeUsage(ctx context.Context, userID uint, limit, offset int) ([]*entities.InviteCodeUsage, int64, error)

	// Invite code statistics
	GetInviteCodeStatistics(ctx context.Context, inviteCodeID uint) (map[string]interface{}, error)
	GetUserInviteCodeStatistics(ctx context.Context, userID uint) (map[string]interface{}, error)

	// Utility
	GenerateInviteCode() (string, error)
}

// CreateInviteCodeRequest represents the request to create an invite code
type CreateInviteCodeRequest struct {
	MaxUses              int     `json:"max_uses" binding:"min=1,max=100" example:"10"`                  // Maximum number of times the code can be used
	Description          string  `json:"description" binding:"max=255" example:"Friend invitation code"` // Description of the invite code
	ReferralCampaignID   *uint   `json:"referral_campaign_id,omitempty" example:"1"`                     // Associated referral campaign ID
	ReferralRewardAmount float64 `json:"referral_reward_amount,omitempty" example:"5.00"`                // Referral reward amount
}

// UpdateInviteCodeRequest represents the request to update an invite code
type UpdateInviteCodeRequest struct {
	MaxUses              *int     `json:"max_uses,omitempty" binding:"omitempty,min=1,max=100"`
	Description          *string  `json:"description,omitempty" binding:"omitempty,max=255"`
	Status               *string  `json:"status,omitempty"`
	ReferralRewardAmount *float64 `json:"referral_reward_amount,omitempty"`
}

// GetInviteCodesRequest represents the request to get invite codes with filters
type GetInviteCodesRequest struct {
	CreatedByID        uint   `form:"created_by_id,omitempty"`
	Status             string `form:"status,omitempty"`
	ReferralCampaignID *uint  `form:"referral_campaign_id,omitempty"`
	Limit              int    `form:"limit,omitempty"`
	Offset             int    `form:"offset,omitempty"`
}
