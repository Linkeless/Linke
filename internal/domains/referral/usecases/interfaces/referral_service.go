package interfaces

import (
	"context"
	"linke/internal/domains/referral/dto"
	"linke/internal/domains/referral/entities"
)

// ReferralService defines the interface for referral operations
type ReferralService interface {
	// Referral CRUD operations
	CreateReferral(ctx context.Context, req *dto.CreateReferralRequest) (*entities.Referral, error)
	GetReferral(ctx context.Context, referralID uint) (*entities.Referral, error)
	GetReferralByCode(ctx context.Context, code string) (*entities.Referral, error)
	UpdateReferral(ctx context.Context, referralID uint, req *dto.UpdateReferralRequest) (*entities.Referral, error)

	// Referral listing and filtering
	GetReferrals(ctx context.Context, req *dto.GetReferralsRequest) ([]*entities.Referral, int64, error)
	GetUserReferrals(ctx context.Context, userID uint, limit, offset int) ([]*entities.Referral, int64, error)
	GetRefereeReferrals(ctx context.Context, userID uint, limit, offset int) ([]*entities.Referral, int64, error)

	// Referral status management
	ConfirmReferral(ctx context.Context, referralID uint) error
	ProcessReferralReward(ctx context.Context, referralID uint, rewardAmount float64) error
	MarkReferralAsPaid(ctx context.Context, referralID uint) error

	// Referral statistics
	GetReferralStatistics(ctx context.Context, userID uint) (map[string]any, error)
	GetSystemReferralStatistics(ctx context.Context) (map[string]any, error)
}

