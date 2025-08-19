package repositories

import (
	"context"
	"time"

	"linke/internal/domains/referral/entities"
	"linke/internal/domains/referral/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// ReferralCampaignRepository implements the ReferralCampaignRepository interface
type ReferralCampaignRepository struct {
	*repository.TimeBasedRepositoryImpl[entities.ReferralCampaign, uint]
}

// NewReferralCampaignRepository creates a new referral campaign repository
func NewReferralCampaignRepository(db *gorm.DB, frameworkLogger framework.Logger) interfaces.ReferralCampaignRepository {
	return &ReferralCampaignRepository{
		TimeBasedRepositoryImpl: repository.NewTimeBasedRepository[entities.ReferralCampaign, uint](db, frameworkLogger),
	}
}

// List lists referral campaigns
func (r *ReferralCampaignRepository) List(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error) {
	return r.TimeBasedRepositoryImpl.List(ctx, limit, offset)
}

// ListWithFilters lists referral campaigns with filters
func (r *ReferralCampaignRepository) ListWithFilters(ctx context.Context, filters map[string]any, limit, offset int) ([]*entities.ReferralCampaign, int64, error) {
	var campaigns []*entities.ReferralCampaign
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.ReferralCampaign{})

	// Apply filters
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if name, ok := filters["name"]; ok && name != "" {
		query = query.Where("name LIKE ?", "%"+name.(string)+"%")
	}
	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query = query.Where("start_date >= ?", startDate)
	}
	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query = query.Where("end_date <= ?", endDate)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get results with pagination
	if err := query.Limit(limit).Offset(offset).Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// ListActive gets all active referral campaigns
func (r *ReferralCampaignRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error) {
	var campaigns []*entities.ReferralCampaign
	var total int64

	query := r.GetDB().WithContext(ctx).
		Where("status = ? AND start_date <= NOW() AND (end_date IS NULL OR end_date >= NOW())", "active")

	if err := query.Model(&entities.ReferralCampaign{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// ListCurrent gets current referral campaigns
func (r *ReferralCampaignRepository) ListCurrent(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error) {
	var campaigns []*entities.ReferralCampaign
	var total int64

	query := r.GetDB().WithContext(ctx).
		Where("start_date <= NOW() AND (end_date IS NULL OR end_date >= NOW())")

	if err := query.Model(&entities.ReferralCampaign{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// ListExpired gets expired referral campaigns
func (r *ReferralCampaignRepository) ListExpired(ctx context.Context, limit, offset int) ([]*entities.ReferralCampaign, int64, error) {
	var campaigns []*entities.ReferralCampaign
	var total int64

	query := r.GetDB().WithContext(ctx).
		Where("end_date IS NOT NULL AND end_date < NOW()")

	if err := query.Model(&entities.ReferralCampaign{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&campaigns).Error; err != nil {
		return nil, 0, err
	}

	return campaigns, total, nil
}

// ListByDateRange lists campaigns by date range for specified field (implements TimeBasedRepository interface)
func (r *ReferralCampaignRepository) ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*entities.ReferralCampaign, int64, error) {
	return r.TimeBasedRepositoryImpl.ListByDateRange(ctx, field, start, end, limit, offset)
}

// ListCreatedAfter lists campaigns created after specified time (implements TimeBasedRepository interface)
func (r *ReferralCampaignRepository) ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.ReferralCampaign, int64, error) {
	return r.TimeBasedRepositoryImpl.ListCreatedAfter(ctx, after, limit, offset)
}

// ListUpdatedAfter lists campaigns updated after specified time (implements TimeBasedRepository interface)
func (r *ReferralCampaignRepository) ListUpdatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.ReferralCampaign, int64, error) {
	return r.TimeBasedRepositoryImpl.ListUpdatedAfter(ctx, after, limit, offset)
}
