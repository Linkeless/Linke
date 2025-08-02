package interfaces

import (
	"context"
	"linke/internal/domains/referral/entities"
)

// InviteCodeRepository defines the interface for invite code data access operations
type InviteCodeRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, inviteCode *entities.InviteCode) error
	GetByID(ctx context.Context, id uint) (*entities.InviteCode, error)
	GetByCode(ctx context.Context, code string) (*entities.InviteCode, error)
	Update(ctx context.Context, inviteCode *entities.InviteCode) error
	Delete(ctx context.Context, id uint) error

	// Status management
	UpdateStatus(ctx context.Context, id uint, status string) error
	UpdateUsageCount(ctx context.Context, id uint, usedCount int) error
	IncrementUsageCount(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error)
	ListByCreator(ctx context.Context, creatorID uint, limit, offset int) ([]*entities.InviteCode, int64, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.InviteCode, int64, error)
	ListActive(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error)

	// Usage operations
	ListAvailable(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error)
	ListExhausted(ctx context.Context, limit, offset int) ([]*entities.InviteCode, int64, error)

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountByCreator(ctx context.Context, creatorID uint) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)

	// Existence checks
	ExistsByCode(ctx context.Context, code string) (bool, error)
}

// ReferralRepository defines the interface for referral data access operations
type ReferralRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, referral *entities.Referral) error
	GetByID(ctx context.Context, id uint) (*entities.Referral, error)
	Update(ctx context.Context, referral *entities.Referral) error
	Delete(ctx context.Context, id uint) error

	// Referral relationship operations
	GetByReferrer(ctx context.Context, referrerID uint, limit, offset int) ([]*entities.Referral, int64, error)
	GetByReferee(ctx context.Context, refereeID uint) (*entities.Referral, error)
	GetReferralChain(ctx context.Context, userID uint, depth int) ([]*entities.Referral, error)

	// Status operations
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Referral, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error

	// Statistics
	CountByReferrer(ctx context.Context, referrerID uint) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	GetReferralStats(ctx context.Context, referrerID uint) (map[string]interface{}, error)

	// List operations
	List(ctx context.Context, limit, offset int) ([]*entities.Referral, int64, error)
}

// ReferralCampaignRepository defines the interface for referral campaign data access operations
type ReferralCampaignRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, campaign *entities.ReferralCampaign) error
	GetByID(ctx context.Context, id uint) (*entities.ReferralCampaign, error)
	Update(ctx context.Context, campaign *entities.ReferralCampaign) error
	Delete(ctx context.Context, id uint) error

	// Status operations
	ListActive(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.ReferralCampaign, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error

	// Time-based operations
	ListCurrent(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error)
	ListExpired(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error)

	// List operations
	List(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error)
	CountTotal(ctx context.Context) (int64, error)
}