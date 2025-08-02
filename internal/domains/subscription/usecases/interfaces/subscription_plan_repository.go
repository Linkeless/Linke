package interfaces

import (
	"context"
	"linke/internal/domains/subscription/entities"
)

// SubscriptionPlanRepository defines the interface for subscription plan data access operations
type SubscriptionPlanRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, plan *entities.SubscriptionPlan) error
	GetByID(ctx context.Context, id uint) (*entities.SubscriptionPlan, error)
	GetByCode(ctx context.Context, code string) (*entities.SubscriptionPlan, error)
	Update(ctx context.Context, plan *entities.SubscriptionPlan) error
	Delete(ctx context.Context, id uint) error

	// Soft delete operations
	SoftDelete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error
	HardDelete(ctx context.Context, id uint) error

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListActive(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListVisible(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	ListDeleted(ctx context.Context, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)

	// Filter operations
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
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

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)

	// Status management
	UpdateStatus(ctx context.Context, id uint, status string) error
	UpdateVisibility(ctx context.Context, id uint, isVisible bool) error
	UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error
	UpdatePopularFlag(ctx context.Context, id uint, isPopular bool) error
	UpdateRecommendedFlag(ctx context.Context, id uint, isRecommended bool) error

	// Batch operations
	BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error)
	BatchUpdateVisibility(ctx context.Context, ids []uint, isVisible bool) (int, []uint, error)
	BatchDelete(ctx context.Context, ids []uint) (int, []uint, error)

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountVisible(ctx context.Context) (int64, error)
	CountByCurrency(ctx context.Context, currency string) (int64, error)
	CountByBillingCycle(ctx context.Context, billingCycle string) (int64, error)

	// Existence checks
	ExistsByCode(ctx context.Context, code string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)

	// Ordering operations
	GetOrderedPlans(ctx context.Context, orderBy string, ascending bool, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
	GetPlansSortedByPrice(ctx context.Context, currency string, ascending bool, limit, offset int) ([]*entities.SubscriptionPlan, int64, error)
}