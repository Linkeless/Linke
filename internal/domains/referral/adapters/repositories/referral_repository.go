package repositories

import (
	"context"

	"linke/internal/domains/referral/entities"
	"linke/internal/domains/referral/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// ReferralRepository implements the ReferralRepository interface
type ReferralRepository struct {
	*repository.UserScopedTimeBasedRepositoryImpl[entities.Referral, uint]
}

// NewReferralRepository creates a new referral repository
func NewReferralRepository(db *gorm.DB, frameworkLogger framework.Logger) interfaces.ReferralRepository {
	return &ReferralRepository{
		UserScopedTimeBasedRepositoryImpl: repository.NewUserScopedTimeBasedRepository[entities.Referral, uint](db, frameworkLogger),
	}
}

// GetByCode gets a referral by referral code
func (r *ReferralRepository) GetByCode(ctx context.Context, code string) (*entities.Referral, error) {
	var referral entities.Referral
	if err := r.GetDB().WithContext(ctx).Where("referral_code = ?", code).First(&referral).Error; err != nil {
		return nil, err
	}
	return &referral, nil
}

// List lists referrals (implements GenericRepository interface)
func (r *ReferralRepository) List(ctx context.Context, limit, offset int) ([]*entities.Referral, int64, error) {
	return r.UserScopedTimeBasedRepositoryImpl.List(ctx, limit, offset)
}

// ListWithFilters lists referrals with filters
func (r *ReferralRepository) ListWithFilters(ctx context.Context, filters map[string]any, limit, offset int) ([]*entities.Referral, int64, error) {
	var referrals []*entities.Referral
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.Referral{})

	// Apply filters
	if referrerID, ok := filters["referrer_id"]; ok && referrerID != uint(0) {
		query = query.Where("referrer_id = ?", referrerID)
	}
	if refereeID, ok := filters["referee_id"]; ok && refereeID != uint(0) {
		query = query.Where("referee_id = ?", refereeID)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if rewardStatus, ok := filters["reward_status"]; ok && rewardStatus != "" {
		query = query.Where("reward_status = ?", rewardStatus)
	}
	if campaignID, ok := filters["campaign_id"]; ok && campaignID != nil {
		query = query.Where("campaign_id = ?", campaignID)
	}
	if dateFrom, ok := filters["date_from"]; ok && dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo, ok := filters["date_to"]; ok && dateTo != "" {
		query = query.Where("created_at <= ?", dateTo)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results with pagination
	if err := query.Limit(limit).Offset(offset).Find(&referrals).Error; err != nil {
		return nil, 0, err
	}

	return referrals, total, nil
}

// GetUserReferrals gets referrals where user is the referrer
func (r *ReferralRepository) GetUserReferrals(ctx context.Context, userID uint, limit, offset int) ([]*entities.Referral, int64, error) {
	var referrals []*entities.Referral
	var total int64

	query := r.GetDB().WithContext(ctx).Where("referrer_id = ?", userID)

	if err := query.Model(&entities.Referral{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&referrals).Error; err != nil {
		return nil, 0, err
	}

	return referrals, total, nil
}

// GetRefereeReferrals gets referrals where user is the referee
func (r *ReferralRepository) GetRefereeReferrals(ctx context.Context, userID uint, limit, offset int) ([]*entities.Referral, int64, error) {
	var referrals []*entities.Referral
	var total int64

	query := r.GetDB().WithContext(ctx).Where("referee_id = ?", userID)

	if err := query.Model(&entities.Referral{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&referrals).Error; err != nil {
		return nil, 0, err
	}

	return referrals, total, nil
}

// GetReferralStatistics gets statistics for a user's referrals
func (r *ReferralRepository) GetReferralStatistics(ctx context.Context, userID uint) (map[string]any, error) {
	var stats struct {
		TotalReferrals     int64   `gorm:"column:total_referrals"`
		ConfirmedReferrals int64   `gorm:"column:confirmed_referrals"`
		TotalRewards       float64 `gorm:"column:total_rewards"`
		PaidRewards        float64 `gorm:"column:paid_rewards"`
	}

	err := r.GetDB().WithContext(ctx).
		Model(&entities.Referral{}).
		Select(`
			COUNT(*) as total_referrals,
			COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as confirmed_referrals,
			COALESCE(SUM(reward_amount), 0) as total_rewards,
			COALESCE(SUM(CASE WHEN reward_status = 'paid' THEN reward_amount ELSE 0 END), 0) as paid_rewards
		`).
		Where("referrer_id = ?", userID).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"total_referrals":     stats.TotalReferrals,
		"confirmed_referrals": stats.ConfirmedReferrals,
		"total_rewards":       stats.TotalRewards,
		"paid_rewards":        stats.PaidRewards,
		"pending_rewards":     stats.TotalRewards - stats.PaidRewards,
	}, nil
}

// GetSystemReferralStatistics gets system-wide referral statistics
func (r *ReferralRepository) GetSystemReferralStatistics(ctx context.Context) (map[string]any, error) {
	var stats struct {
		TotalReferrals     int64   `gorm:"column:total_referrals"`
		ConfirmedReferrals int64   `gorm:"column:confirmed_referrals"`
		TotalRewards       float64 `gorm:"column:total_rewards"`
		PaidRewards        float64 `gorm:"column:paid_rewards"`
		UniqueReferrers    int64   `gorm:"column:unique_referrers"`
	}

	err := r.GetDB().WithContext(ctx).
		Model(&entities.Referral{}).
		Select(`
			COUNT(*) as total_referrals,
			COUNT(CASE WHEN status = 'confirmed' THEN 1 END) as confirmed_referrals,
			COALESCE(SUM(reward_amount), 0) as total_rewards,
			COALESCE(SUM(CASE WHEN reward_status = 'paid' THEN reward_amount ELSE 0 END), 0) as paid_rewards,
			COUNT(DISTINCT referrer_id) as unique_referrers
		`).
		Scan(&stats).Error
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"total_referrals":     stats.TotalReferrals,
		"confirmed_referrals": stats.ConfirmedReferrals,
		"total_rewards":       stats.TotalRewards,
		"paid_rewards":        stats.PaidRewards,
		"pending_rewards":     stats.TotalRewards - stats.PaidRewards,
		"unique_referrers":    stats.UniqueReferrers,
	}, nil
}

// GetByReferrer gets referrals where user is the referrer
func (r *ReferralRepository) GetByReferrer(ctx context.Context, referrerID uint, limit, offset int) ([]*entities.Referral, int64, error) {
	return r.GetUserReferrals(ctx, referrerID, limit, offset)
}

// GetByReferee gets referrals where user is the referee
func (r *ReferralRepository) GetByReferee(ctx context.Context, refereeID uint) (*entities.Referral, error) {
	var referral entities.Referral
	if err := r.GetDB().WithContext(ctx).Where("referee_id = ?", refereeID).First(&referral).Error; err != nil {
		return nil, err
	}
	return &referral, nil
}

// GetReferralChain gets referral chain for a user up to specified depth
func (r *ReferralRepository) GetReferralChain(ctx context.Context, userID uint, depth int) ([]*entities.Referral, error) {
	var chain []*entities.Referral
	currentUserID := userID

	for i := 0; i < depth; i++ {
		var referral entities.Referral
		err := r.GetDB().WithContext(ctx).Where("referee_id = ?", currentUserID).First(&referral).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				break // End of chain
			}
			return nil, err
		}

		chain = append(chain, &referral)
		currentUserID = referral.ReferrerID
	}

	return chain, nil
}

// CountByReferrer counts referrals by referrer
func (r *ReferralRepository) CountByReferrer(ctx context.Context, referrerID uint) (int64, error) {
	var count int64
	err := r.GetDB().WithContext(ctx).Model(&entities.Referral{}).Where("referrer_id = ?", referrerID).Count(&count).Error
	return count, err
}

// GetReferralStats gets referral statistics for a user
func (r *ReferralRepository) GetReferralStats(ctx context.Context, referrerID uint) (map[string]any, error) {
	return r.GetReferralStatistics(ctx, referrerID)
}
