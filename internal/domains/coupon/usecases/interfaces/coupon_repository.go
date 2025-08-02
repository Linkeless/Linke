package interfaces

import (
	"context"
	"linke/internal/domains/coupon/entities"
	"time"
)

// CouponRepository defines the interface for coupon data access operations
type CouponRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, coupon *entities.Coupon) error
	GetByID(ctx context.Context, id uint64) (*entities.Coupon, error)
	GetByCode(ctx context.Context, code string) (*entities.Coupon, error)
	Update(ctx context.Context, coupon *entities.Coupon) error
	Delete(ctx context.Context, id uint64) error

	// Soft delete operations
	SoftDelete(ctx context.Context, id uint64) error
	Restore(ctx context.Context, id uint64) error
	HardDelete(ctx context.Context, id uint64) error

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)
	ListActive(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)
	ListDeleted(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)

	// Status filtering
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Coupon, int64, error)
	ListPublic(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)
	ListPrivate(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)

	// Validity filtering
	ListValid(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)
	ListExpired(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)
	ListExpiringBefore(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.Coupon, int64, error)

	// Type and value filtering
	ListByType(ctx context.Context, couponType string, limit, offset int) ([]*entities.Coupon, int64, error)
	ListByValueRange(ctx context.Context, minValue, maxValue float64, limit, offset int) ([]*entities.Coupon, int64, error)

	// Usage filtering
	ListByUsageCount(ctx context.Context, minUsed, maxUsed int, limit, offset int) ([]*entities.Coupon, int64, error)
	ListAvailable(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)
	ListExhausted(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)

	// Creator filtering
	ListByCreator(ctx context.Context, creatorID uint64, limit, offset int) ([]*entities.Coupon, int64, error)

	// Currency filtering
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.Coupon, int64, error)

	// Plan filtering
	ListByPlan(ctx context.Context, planID uint64, limit, offset int) ([]*entities.Coupon, int64, error)
	ListForAnyPlan(ctx context.Context, limit, offset int) ([]*entities.Coupon, int64, error)

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Coupon, int64, error)

	// Status management
	UpdateStatus(ctx context.Context, id uint64, status string) error
	UpdateUsageCount(ctx context.Context, id uint64, usedCount int) error
	IncrementUsageCount(ctx context.Context, id uint64) error

	// Batch operations
	BatchUpdateStatus(ctx context.Context, ids []uint64, status string) (int, []uint64, error)
	BatchDelete(ctx context.Context, ids []uint64) (int, []uint64, error)

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountValid(ctx context.Context) (int64, error)
	CountExpired(ctx context.Context) (int64, error)
	CountByCreator(ctx context.Context, creatorID uint64) (int64, error)

	// Existence checks
	ExistsByCode(ctx context.Context, code string) (bool, error)
	ExistsByID(ctx context.Context, id uint64) (bool, error)

	// Time-based operations
	MarkExpiredCoupons(ctx context.Context) (int64, error)
	ListRecentlyCreated(ctx context.Context, since time.Time, limit, offset int) ([]*entities.Coupon, int64, error)

	// Advanced filtering
	ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.Coupon, int64, error)
}

// CouponUsageRepository defines the interface for coupon usage data access operations
type CouponUsageRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, usage *entities.CouponUsage) error
	GetByID(ctx context.Context, id uint64) (*entities.CouponUsage, error)
	Update(ctx context.Context, usage *entities.CouponUsage) error
	Delete(ctx context.Context, id uint64) error

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.CouponUsage, int64, error)
	ListByCoupon(ctx context.Context, couponID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error)
	ListByUser(ctx context.Context, userID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error)
	ListByOrder(ctx context.Context, orderID uint64, limit, offset int) ([]*entities.CouponUsage, int64, error)

	// User-specific operations
	GetUserUsageForCoupon(ctx context.Context, userID, couponID uint64) ([]*entities.CouponUsage, error)
	CountUserUsageForCoupon(ctx context.Context, userID, couponID uint64) (int64, error)
	GetUserTotalSavings(ctx context.Context, userID uint64, currency string) (float64, error)

	// Coupon-specific operations
	GetCouponUsageStats(ctx context.Context, couponID uint64) (map[string]interface{}, error)
	GetTopUsersByCoupon(ctx context.Context, couponID uint64, limit int) ([]*entities.CouponUsage, error)

	// Time-based queries
	ListByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.CouponUsage, int64, error)
	ListRecentUsage(ctx context.Context, since time.Time, limit, offset int) ([]*entities.CouponUsage, int64, error)

	// Currency operations
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.CouponUsage, int64, error)
	GetTotalSavingsByCurrency(ctx context.Context, currency string, since time.Time) (float64, error)

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountByCoupon(ctx context.Context, couponID uint64) (int64, error)
	CountByUser(ctx context.Context, userID uint64) (int64, error)
	GetUsageStats(ctx context.Context, since time.Time) (map[string]interface{}, error)

	// Advanced analytics
	GetTopCoupons(ctx context.Context, limit int, since time.Time) ([]*entities.CouponUsage, error)
	GetUsageTrends(ctx context.Context, days int) (map[string]int64, error)

	// Existence checks
	UserHasUsedCoupon(ctx context.Context, userID, couponID uint64) (bool, error)
	OrderHasUsedCoupon(ctx context.Context, orderID uint64) (bool, error)
}
