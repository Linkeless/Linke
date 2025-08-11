package interfaces

import (
	"context"
	"linke/internal/domains/referral/dto"
	"linke/internal/domains/referral/entities"
	"time"
)

// ReferralCampaignService defines the interface for referral campaign operations
type ReferralCampaignService interface {
	// Campaign CRUD operations
	CreateReferralCampaign(ctx context.Context, req *dto.CreateReferralCampaignRequest) (*entities.ReferralCampaign, error)
	GetReferralCampaign(ctx context.Context, campaignID uint) (*entities.ReferralCampaign, error)
	UpdateReferralCampaign(ctx context.Context, campaignID uint, req *dto.UpdateReferralCampaignRequest) (*entities.ReferralCampaign, error)
	DeleteReferralCampaign(ctx context.Context, campaignID uint) error

	// Campaign listing and filtering
	GetReferralCampaigns(ctx context.Context, req *dto.GetReferralCampaignsRequest) ([]*entities.ReferralCampaign, int64, error)
	GetActiveCampaigns(ctx context.Context) ([]*entities.ReferralCampaign, error)

	// Campaign management
	ActivateCampaign(ctx context.Context, campaignID uint) error
	DeactivateCampaign(ctx context.Context, campaignID uint) error
	ExpireCampaign(ctx context.Context, campaignID uint) error

	// Campaign statistics
	GetCampaignStatistics(ctx context.Context, campaignID uint) (map[string]any, error)
	GetCampaignPerformance(ctx context.Context, campaignID uint, fromDate, toDate time.Time) (map[string]any, error)
}

