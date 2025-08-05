package interfaces

import (
	"context"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/framework"
)

// SubscriptionPlanRepository defines the interface for subscription plan data access operations
// It extends GenericRepository with SubscriptionPlan-specific methods
type SubscriptionPlanRepository interface {
	framework.GenericRepository[entities.SubscriptionPlan, uint]

	// Subscription plan specific query methods
	GetByCode(ctx context.Context, code string) (*entities.SubscriptionPlan, error)

	// Filter operations specific to subscription plans
	ListActive(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListVisible(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListByBillingCycle(ctx context.Context, billingCycle string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListPopular(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListRecommended(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)

	// Public list operations (for non-authenticated users)
	ListPublic(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListPublicByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)

	// Price range operations
	ListByPriceRange(ctx context.Context, minPrice, maxPrice float64, currency string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	GetCheapest(ctx context.Context, currency string) (*entities.SubscriptionPlan, error)
	GetMostExpensive(ctx context.Context, currency string) (*entities.SubscriptionPlan, error)

	// Subscription plan specific status management
	UpdateVisibility(ctx context.Context, id uint, isVisible bool) error
	UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error
	UpdatePopularFlag(ctx context.Context, id uint, isPopular bool) error
	UpdateRecommendedFlag(ctx context.Context, id uint, isRecommended bool) error

	// Subscription plan specific batch operations
	BatchUpdateVisibility(ctx context.Context, ids []uint, isVisible bool) (int, []uint, error)

	// Subscription plan specific statistics
	CountVisible(ctx context.Context) (int64, error)
	CountByCurrency(ctx context.Context, currency string) (int64, error)
	CountByBillingCycle(ctx context.Context, billingCycle string) (int64, error)

	// Subscription plan specific existence checks
	ExistsByCode(ctx context.Context, code string) (bool, error)

	// Ordering operations
	GetOrderedPlans(ctx context.Context, orderBy string, ascending bool, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	GetPlansSortedByPrice(ctx context.Context, currency string, ascending bool, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
}
