package interfaces

import (
	"context"

	"linke/internal/domains/referral/entities"
	"linke/internal/shared/framework"
)

// InviteCodeRepository defines the interface for invite code data access operations
type InviteCodeRepository interface {
	framework.UserScopedRepository[entities.InviteCode, uint]

	// Code-specific operations
	GetByCode(ctx context.Context, code string) (*entities.InviteCode, error)
	ExistsByCode(ctx context.Context, code string) (bool, error)

	// Usage management
	UpdateUsageCount(ctx context.Context, id uint, usedCount int) error
	IncrementUsageCount(ctx context.Context, id uint) error

	// Status-specific queries
	ListByCreator(ctx context.Context, creatorID uint, limit, offset int) ([]*entities.InviteCode, int64, error)
	ListActive(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error)
	ListAvailable(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error)
	ListExhausted(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error)

	// Creator-specific statistics
	CountByCreator(ctx context.Context, creatorID uint) (int64, error)
}

// ReferralRepository defines the interface for referral data access operations
type ReferralRepository interface {
	framework.UserScopedTimeBasedRepository[entities.Referral, uint]

	// Referral relationship operations
	GetByReferrer(ctx context.Context, referrerID uint, limit, offset int) ([]*entities.Referral, int64, error)
	GetByReferee(ctx context.Context, refereeID uint) (*entities.Referral, error)
	GetReferralChain(ctx context.Context, userID uint, depth int) ([]*entities.Referral, error)

	// Statistics and analytics
	CountByReferrer(ctx context.Context, referrerID uint) (int64, error)
	GetReferralStats(ctx context.Context, referrerID uint) (map[string]any, error)
}

// ReferralCampaignRepository defines the interface for referral campaign data access operations
type ReferralCampaignRepository interface {
	framework.TimeBasedRepository[entities.ReferralCampaign, uint]

	// Campaign-specific time operations
	ListActive(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error)
	ListCurrent(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error)
	ListExpired(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error)
}
