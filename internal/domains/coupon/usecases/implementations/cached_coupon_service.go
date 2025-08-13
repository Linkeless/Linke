package implementations

import (
	"context"
	"encoding/json"
	"time"

	"linke/internal/domains/coupon/dto"
	"linke/internal/domains/coupon/entities"
	"linke/internal/domains/coupon/usecases/interfaces"
	"linke/internal/shared/cache"
	"linke/internal/shared/logger"
)

// CachedCouponService wraps CouponService with intelligent caching for performance optimization
// Caching Strategy:
// ✅ Cache: Static coupon data, public coupons, usage statistics (non-transactional data)
// ❌ No Cache: Real-time operations like validation, usage count updates (transactional data)
type CachedCouponService struct {
	base         interfaces.CouponService
	cacheManager cache.CacheManager
	cacheKeys    *cache.CouponCacheKeys
}

// NewCachedCouponService creates a new cached coupon service with optimal caching strategy
func NewCachedCouponService(
	base interfaces.CouponService,
	cacheManager cache.CacheManager,
	allKeys *cache.AllCacheKeys,
) interfaces.CouponService {
	return &CachedCouponService{
		base:         base,
		cacheManager: cacheManager,
		cacheKeys:    allKeys.Coupon,
	}
}

// CreateCoupon creates a new coupon (no caching - always transactional)
func (cs *CachedCouponService) CreateCoupon(ctx context.Context, creatorID uint64, req *dto.CreateCouponRequest) (*entities.Coupon, error) {
	coupon, err := cs.base.CreateCoupon(ctx, creatorID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate related caches
	cs.invalidateCouponCaches(coupon.ID, coupon.Code)

	return coupon, nil
}

// GetCoupon gets a coupon by ID with caching for static data
func (cs *CachedCouponService) GetCoupon(ctx context.Context, couponID uint64) (*entities.Coupon, error) {
	cacheKey := cs.cacheKeys.CouponByID(uint(couponID))

	// Try cache first
	if cached, err := cs.cacheManager.GetCache().Get(ctx, cacheKey); err == nil && cached != nil {
		var coupon entities.Coupon
		if err := json.Unmarshal(cached, &coupon); err == nil {
			return &coupon, nil
		}
		// Log but don't fail on cache deserialization error
		logger.Warn("Failed to deserialize cached coupon", logger.Uint64("couponID", couponID))
	}

	// Fetch from base service
	coupon, err := cs.base.GetCoupon(ctx, couponID)
	if err != nil {
		return nil, err
	}

	// Cache the result for 15 minutes (warm data strategy)
	if data, err := json.Marshal(coupon); err == nil {
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return coupon, nil
}

// GetCouponByCode gets a coupon by code with caching
func (cs *CachedCouponService) GetCouponByCode(ctx context.Context, code string) (*entities.Coupon, error) {
	cacheKey := cs.cacheKeys.CouponByCode(code)

	// Try cache first
	if cached, err := cs.cacheManager.GetCache().Get(ctx, cacheKey); err == nil && cached != nil {
		var coupon entities.Coupon
		if err := json.Unmarshal(cached, &coupon); err == nil {
			return &coupon, nil
		}
		logger.Warn("Failed to deserialize cached coupon by code", logger.String("code", code))
	}

	// Fetch from base service
	coupon, err := cs.base.GetCouponByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// Cache the result for 15 minutes
	if data, err := json.Marshal(coupon); err == nil {
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return coupon, nil
}

// GetCoupons gets coupons with filtering and pagination (no caching for search queries)
func (cs *CachedCouponService) GetCoupons(ctx context.Context, req *dto.GetCouponsRequest) ([]*entities.Coupon, int64, error) {
	// Search queries are too variable to cache effectively
	return cs.base.GetCoupons(ctx, req)
}

// GetPublicCoupons gets active public coupons with caching
func (cs *CachedCouponService) GetPublicCoupons(ctx context.Context, limit int) ([]*entities.Coupon, error) {
	cacheKey := cs.cacheKeys.PublicCoupons()

	// Try cache first
	if cached, err := cs.cacheManager.GetCache().Get(ctx, cacheKey); err == nil && cached != nil {
		var coupons []*entities.Coupon
		if err := json.Unmarshal(cached, &coupons); err == nil {
			// Apply limit if specified
			if limit > 0 && len(coupons) > limit {
				return coupons[:limit], nil
			}
			return coupons, nil
		}
		logger.Warn("Failed to deserialize cached public coupons")
	}

	// Fetch from base service
	coupons, err := cs.base.GetPublicCoupons(ctx, limit)
	if err != nil {
		return nil, err
	}

	// Cache for 1 hour (configuration data strategy) - public coupons change infrequently
	if data, err := json.Marshal(coupons); err == nil {
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.LongCacheTTL)
	}

	return coupons, nil
}

// GetActiveCoupons gets all active coupons with caching
func (cs *CachedCouponService) GetActiveCoupons(ctx context.Context) ([]*entities.Coupon, error) {
	cacheKey := cs.cacheKeys.ActiveCoupons()

	// Try cache first
	if cached, err := cs.cacheManager.GetCache().Get(ctx, cacheKey); err == nil && cached != nil {
		var coupons []*entities.Coupon
		if err := json.Unmarshal(cached, &coupons); err == nil {
			return coupons, nil
		}
		logger.Warn("Failed to deserialize cached active coupons")
	}

	// Fetch from base service
	coupons, err := cs.base.GetActiveCoupons(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for 15 minutes (warm data strategy)
	if data, err := json.Marshal(coupons); err == nil {
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.MediumCacheTTL)
	}

	return coupons, nil
}

// UpdateCoupon updates a coupon (invalidate caches)
func (cs *CachedCouponService) UpdateCoupon(ctx context.Context, couponID uint64, req *dto.UpdateCouponRequest) (*entities.Coupon, error) {
	coupon, err := cs.base.UpdateCoupon(ctx, couponID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate all related caches
	cs.invalidateCouponCaches(coupon.ID, coupon.Code)

	return coupon, nil
}

// DeleteCoupon soft deletes a coupon (invalidate caches)
func (cs *CachedCouponService) DeleteCoupon(ctx context.Context, couponID uint64) error {
	// Get coupon info for cache invalidation
	coupon, _ := cs.base.GetCoupon(ctx, couponID)

	err := cs.base.DeleteCoupon(ctx, couponID)
	if err != nil {
		return err
	}

	// Invalidate caches
	if coupon != nil {
		cs.invalidateCouponCaches(coupon.ID, coupon.Code)
	} else {
		// Fallback: invalidate by ID only
		cs.invalidateCouponCaches(couponID, "")
	}

	return nil
}

// ValidateCoupon validates a coupon (real-time validation - no caching)
func (cs *CachedCouponService) ValidateCoupon(ctx context.Context, req *dto.ValidateCouponRequest) (*dto.ValidateCouponResponse, error) {
	// Coupon validation must always be real-time for security and accuracy
	// No caching to ensure fresh usage counts and status checks
	return cs.base.ValidateCoupon(ctx, req)
}

// UseCoupon records coupon usage (transactional - no caching, with atomic operations)
func (cs *CachedCouponService) UseCoupon(ctx context.Context, couponID, userID uint64, orderAmount float64, orderID *uint64) (*entities.CouponUsage, error) {
	usage, err := cs.base.UseCoupon(ctx, couponID, userID, orderAmount, orderID)
	if err != nil {
		return nil, err
	}

	// Invalidate usage-related caches
	cs.invalidateUsageCaches(couponID, userID)

	return usage, nil
}

// GetCouponUsages gets coupon usage records (no caching - usage data changes frequently)
// This is a helper method that delegates to the appropriate base service method
func (cs *CachedCouponService) GetCouponUsages(ctx context.Context, couponID *uint64, userID *uint64, limit, offset int) ([]*entities.CouponUsage, int64, error) {
	if couponID != nil {
		return cs.base.GetCouponUsage(ctx, *couponID, limit, offset)
	} else if userID != nil {
		return cs.base.GetUserCouponUsage(ctx, *userID, limit, offset)
	}
	// If neither ID is provided, this could be extended to support general queries
	// For now, return empty results
	return []*entities.CouponUsage{}, 0, nil
}

// ActivateCoupon activates a coupon (invalidate caches)
func (cs *CachedCouponService) ActivateCoupon(ctx context.Context, couponID uint64) error {
	err := cs.base.ActivateCoupon(ctx, couponID)
	if err != nil {
		return err
	}

	// Invalidate caches since status changed
	cs.invalidateCouponCaches(couponID, "")
	return nil
}

// DeactivateCoupon deactivates a coupon (invalidate caches)
func (cs *CachedCouponService) DeactivateCoupon(ctx context.Context, couponID uint64) error {
	err := cs.base.DeactivateCoupon(ctx, couponID)
	if err != nil {
		return err
	}

	// Invalidate caches since status changed
	cs.invalidateCouponCaches(couponID, "")
	return nil
}

// ExpireCoupon expires a coupon (invalidate caches)
func (cs *CachedCouponService) ExpireCoupon(ctx context.Context, couponID uint64) error {
	err := cs.base.ExpireCoupon(ctx, couponID)
	if err != nil {
		return err
	}

	// Invalidate caches since status changed
	cs.invalidateCouponCaches(couponID, "")
	return nil
}

// GetCouponUsage gets usage records for a specific coupon (no caching)
func (cs *CachedCouponService) GetCouponUsage(ctx context.Context, couponID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error) {
	return cs.base.GetCouponUsage(ctx, couponID, limit, offset)
}

// GetUserCouponUsage gets usage records for a specific user (no caching)
func (cs *CachedCouponService) GetUserCouponUsage(ctx context.Context, userID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error) {
	return cs.base.GetUserCouponUsage(ctx, userID, limit, offset)
}

// GetCouponStatistics gets statistics for a specific coupon with caching
func (cs *CachedCouponService) GetCouponStatistics(ctx context.Context, couponID uint64) (map[string]any, error) {
	cacheKey := cs.cacheKeys.CouponStatistics(uint(couponID))

	// Try cache first (5 minute TTL for statistics)
	if cached, err := cs.cacheManager.GetCache().Get(ctx, cacheKey); err == nil && cached != nil {
		var stats map[string]any
		if err := json.Unmarshal(cached, &stats); err == nil {
			return stats, nil
		}
		logger.Warn("Failed to deserialize cached coupon statistics", logger.Uint64("couponID", couponID))
	}

	// Fetch from base service
	stats, err := cs.base.GetCouponStatistics(ctx, couponID)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes (statistics change but not too frequently)
	if data, err := json.Marshal(stats); err == nil {
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.ShortCacheTTL)
	}

	return stats, nil
}

// GetCouponSystemStatistics gets system-wide coupon statistics with caching
func (cs *CachedCouponService) GetCouponSystemStatistics(ctx context.Context) (map[string]any, error) {
	cacheKey := cs.cacheKeys.SystemStatistics()

	// Try cache first (10 minute TTL for system statistics)
	if cached, err := cs.cacheManager.GetCache().Get(ctx, cacheKey); err == nil && cached != nil {
		var stats map[string]any
		if err := json.Unmarshal(cached, &stats); err == nil {
			return stats, nil
		}
		logger.Warn("Failed to deserialize cached system statistics")
	}

	// Fetch from base service
	stats, err := cs.base.GetCouponSystemStatistics(ctx)
	if err != nil {
		return nil, err
	}

	// Cache for 10 minutes
	if data, err := json.Marshal(stats); err == nil {
		_ = cs.cacheManager.GetCache().Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return stats, nil
}

// Cache invalidation helpers

// invalidateCouponCaches invalidates all caches related to a specific coupon
func (cs *CachedCouponService) invalidateCouponCaches(couponID uint64, code string) {
	ctx := context.Background()

	// Individual coupon caches
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.CouponByID(uint(couponID)))
	if code != "" {
		_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.CouponByCode(code))
	}

	// List caches (these need to be refreshed when any coupon changes)
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.ActiveCoupons())
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.PublicCoupons())

	// Statistics caches
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.CouponStatistics(uint(couponID)))
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.SystemStatistics())

	logger.Debug("Invalidated coupon caches",
		logger.Uint64("couponID", couponID),
		logger.String("code", code))
}

// invalidateUsageCaches invalidates caches related to coupon usage
func (cs *CachedCouponService) invalidateUsageCaches(couponID, userID uint64) {
	ctx := context.Background()

	// Usage-related caches
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.CouponUsage(uint(couponID)))
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.UserCouponUsage(uint(userID)))

	// Statistics caches (usage affects statistics)
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.CouponStatistics(uint(couponID)))
	_ = cs.cacheManager.GetCache().Delete(ctx, cs.cacheKeys.SystemStatistics())

	logger.Debug("Invalidated usage caches",
		logger.Uint64("couponID", couponID),
		logger.Uint64("userID", userID))
}
