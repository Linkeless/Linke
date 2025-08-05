package repositories

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/coupon/entities"
	"linke/internal/domains/coupon/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// couponRepository implements the CouponRepository interface
type couponRepository struct {
	*repository.TimeBasedRepositoryImpl[entities.Coupon, uint64]
}

// NewCouponRepository creates a new CouponRepository implementation
func NewCouponRepository(db *gorm.DB, logger framework.Logger) interfaces.CouponRepository {
	return &couponRepository{
		TimeBasedRepositoryImpl: repository.NewTimeBasedRepository[entities.Coupon, uint64](db, logger),
	}
}

// GetByCode retrieves a coupon by its code
func (r *couponRepository) GetByCode(ctx context.Context, code string) (*entities.Coupon, error) {
	var coupon entities.Coupon
	if err := r.GetDB().WithContext(ctx).Where("code = ?", code).First(&coupon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("coupon with code %s not found", code)
		}
		return nil, fmt.Errorf("failed to get coupon by code: %w", err)
	}
	return &coupon, nil
}

// ListActive retrieves active coupons with pagination
func (r *couponRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error) {
	return r.ListByStatus(ctx, entities.CouponStatusActive, limit, offset)
}

// ListPublic retrieves public coupons with pagination
func (r *couponRepository) ListPublic(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	// Count total public coupons
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("is_public = ?", true).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count public coupons: %w", err)
	}

	// Get paginated public coupons
	query := r.GetDB().WithContext(ctx).Where("is_public = ?", true).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list public coupons: %w", err)
	}

	return coupons, total, nil
}

// ListPrivate retrieves private coupons with pagination
func (r *couponRepository) ListPrivate(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	// Count total private coupons
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("is_public = ?", false).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count private coupons: %w", err)
	}

	// Get paginated private coupons
	query := r.GetDB().WithContext(ctx).Where("is_public = ?", false).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list private coupons: %w", err)
	}

	return coupons, total, nil
}

// ListValid retrieves valid coupons with pagination
func (r *couponRepository) ListValid(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	now := time.Now()
	query := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("status = ?", entities.CouponStatusActive).
		Where("(valid_from IS NULL OR valid_from <= ?)", now).
		Where("(valid_until IS NULL OR valid_until >= ?)", now).
		Where("(max_uses = 0 OR used_count < max_uses)")

	// Count total valid coupons
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count valid coupons: %w", err)
	}

	// Get paginated valid coupons
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list valid coupons: %w", err)
	}

	return coupons, total, nil
}

// ListExpired retrieves expired coupons with pagination
func (r *couponRepository) ListExpired(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	now := time.Now()
	query := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("status = ? OR valid_until < ? OR (max_uses > 0 AND used_count >= max_uses)", 
			entities.CouponStatusExpired, now)

	// Count total expired coupons
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count expired coupons: %w", err)
	}

	// Get paginated expired coupons
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list expired coupons: %w", err)
	}

	return coupons, total, nil
}

// ListExpiringBefore retrieves coupons expiring before a given date
func (r *couponRepository) ListExpiringBefore(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("valid_until IS NOT NULL AND valid_until < ?", beforeDate).
		Where("status = ?", entities.CouponStatusActive)

	// Count total expiring coupons
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count expiring coupons: %w", err)
	}

	// Get paginated expiring coupons
	if err := query.Order("valid_until ASC").Limit(limit).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list expiring coupons: %w", err)
	}

	return coupons, total, nil
}

// ListByType retrieves coupons by type with pagination
func (r *couponRepository) ListByType(ctx context.Context, couponType string, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	// Count total coupons by type
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("type = ?", couponType).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons by type: %w", err)
	}

	// Get paginated coupons by type
	query := r.GetDB().WithContext(ctx).Where("type = ?", couponType).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list coupons by type: %w", err)
	}

	return coupons, total, nil
}

// ListByValueRange retrieves coupons within a value range
func (r *couponRepository) ListByValueRange(ctx context.Context, minValue, maxValue float64, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	// Count total coupons in value range
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("value >= ? AND value <= ?", minValue, maxValue).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons by value range: %w", err)
	}

	// Get paginated coupons in value range
	query := r.GetDB().WithContext(ctx).Where("value >= ? AND value <= ?", minValue, maxValue).Order("value ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list coupons by value range: %w", err)
	}

	return coupons, total, nil
}

// ListByUsageCount retrieves coupons by usage count range
func (r *couponRepository) ListByUsageCount(ctx context.Context, minUsed, maxUsed int, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	// Count total coupons in usage range
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("used_count >= ? AND used_count <= ?", minUsed, maxUsed).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons by usage count: %w", err)
	}

	// Get paginated coupons in usage range
	query := r.GetDB().WithContext(ctx).Where("used_count >= ? AND used_count <= ?", minUsed, maxUsed).Order("used_count DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list coupons by usage count: %w", err)
	}

	return coupons, total, nil
}

// ListAvailable retrieves available coupons (not exhausted)
func (r *couponRepository) ListAvailable(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("status = ?", entities.CouponStatusActive).
		Where("max_uses = 0 OR used_count < max_uses")

	// Count total available coupons
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count available coupons: %w", err)
	}

	// Get paginated available coupons
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list available coupons: %w", err)
	}

	return coupons, total, nil
}

// ListExhausted retrieves exhausted coupons (used up)
func (r *couponRepository) ListExhausted(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("max_uses > 0 AND used_count >= max_uses")

	// Count total exhausted coupons
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count exhausted coupons: %w", err)
	}

	// Get paginated exhausted coupons
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list exhausted coupons: %w", err)
	}

	return coupons, total, nil
}

// ListByCreator retrieves coupons by creator ID
func (r *couponRepository) ListByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	// Count total coupons by creator
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("created_by = ?", creatorID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons by creator: %w", err)
	}

	// Get paginated coupons by creator
	query := r.GetDB().WithContext(ctx).Where("created_by = ?", creatorID).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list coupons by creator: %w", err)
	}

	return coupons, total, nil
}

// ListByCurrency retrieves coupons by currency
func (r *couponRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	// Count total coupons by currency
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("currency = ?", currency).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons by currency: %w", err)
	}

	// Get paginated coupons by currency
	query := r.GetDB().WithContext(ctx).Where("currency = ?", currency).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list coupons by currency: %w", err)
	}

	return coupons, total, nil
}

// ListByPlan retrieves coupons applicable to a specific plan
func (r *couponRepository) ListByPlan(ctx context.Context, planID uint64, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	// For now, this is a placeholder implementation
	// The actual implementation would need to parse the ApplicablePlans JSON field
	planIDStr := fmt.Sprintf("%d", planID)
	query := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("applicable_plans = '' OR applicable_plans IS NULL OR applicable_plans LIKE ?", "%"+planIDStr+"%")

	// Count total coupons for plan
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons by plan: %w", err)
	}

	// Get paginated coupons for plan
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list coupons by plan: %w", err)
	}

	return coupons, total, nil
}

// ListForAnyPlan retrieves coupons applicable to any plan
func (r *couponRepository) ListForAnyPlan(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error) {
	var coupons []*entities.Coupon
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("applicable_plans = '' OR applicable_plans IS NULL")

	// Count total coupons for any plan
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count coupons for any plan: %w", err)
	}

	// Get paginated coupons for any plan
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&coupons).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list coupons for any plan: %w", err)
	}

	return coupons, total, nil
}

// UpdateUsageCount updates the usage count of a coupon
func (r *couponRepository) UpdateUsageCount(ctx context.Context, id uint64, usedCount int) error {
	result := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("id = ?", id).Update("used_count", usedCount)
	if result.Error != nil {
		return fmt.Errorf("failed to update coupon usage count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("coupon with id %d not found", id)
	}
	return nil
}

// IncrementUsageCount increments the usage count of a coupon
func (r *couponRepository) IncrementUsageCount(ctx context.Context, id uint64) error {
	result := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("id = ?", id).
		Update("used_count", gorm.Expr("used_count + 1"))
	if result.Error != nil {
		return fmt.Errorf("failed to increment coupon usage count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("coupon with id %d not found", id)
	}
	return nil
}

// CountValid returns the count of valid coupons
func (r *couponRepository) CountValid(ctx context.Context) (int64, error) {
	var count int64
	now := time.Now()
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("status = ?", entities.CouponStatusActive).
		Where("(valid_from IS NULL OR valid_from <= ?)", now).
		Where("(valid_until IS NULL OR valid_until >= ?)", now).
		Where("(max_uses = 0 OR used_count < max_uses)").
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count valid coupons: %w", err)
	}
	return count, nil
}

// CountExpired returns the count of expired coupons
func (r *couponRepository) CountExpired(ctx context.Context) (int64, error) {
	var count int64
	now := time.Now()
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("status = ? OR valid_until < ? OR (max_uses > 0 AND used_count >= max_uses)", 
			entities.CouponStatusExpired, now).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count expired coupons: %w", err)
	}
	return count, nil
}

// CountByCreator returns the count of coupons created by a specific user
func (r *couponRepository) CountByCreator(ctx context.Context, creatorID uint64) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("created_by = ?", creatorID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count coupons by creator: %w", err)
	}
	return count, nil
}

// ExistsByCode checks if a coupon with the given code exists
func (r *couponRepository) ExistsByCode(ctx context.Context, code string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).Where("code = ?", code).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check coupon exists by code: %w", err)
	}
	return count > 0, nil
}

// MarkExpiredCoupons marks coupons as expired based on validity period
func (r *couponRepository) MarkExpiredCoupons(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.GetDB().WithContext(ctx).Model(&entities.Coupon{}).
		Where("status = ?", entities.CouponStatusActive).
		Where("valid_until IS NOT NULL AND valid_until < ?", now).
		Update("status", entities.CouponStatusExpired)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to mark expired coupons: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// ListRecentlyCreated retrieves coupons created after a specific time
func (r *couponRepository) ListRecentlyCreated(ctx context.Context, since time.Time, limit, offset int) ([]*entities.Coupon, int64, error) {
	return r.ListCreatedAfter(ctx, since, limit, offset)
}