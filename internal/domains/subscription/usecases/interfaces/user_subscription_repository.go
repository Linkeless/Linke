package interfaces

import (
	"context"
	"time"
	"linke/internal/domains/subscription/entities"
)

// UserSubscriptionRepository defines the interface for user subscription data access operations
type UserSubscriptionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, subscription *entities.UserSubscription) error
	GetByID(ctx context.Context, id uint) (*entities.UserSubscription, error)
	GetByUUID(ctx context.Context, uuid string) (*entities.UserSubscription, error)
	Update(ctx context.Context, subscription *entities.UserSubscription) error
	Delete(ctx context.Context, id uint) error

	// Soft delete operations
	SoftDelete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error
	HardDelete(ctx context.Context, id uint) error

	// User-specific operations
	ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.UserSubscription, int64, error)
	GetActiveByUser(ctx context.Context, userID uint) ([]*entities.UserSubscription, error)
	GetUserCurrentSubscription(ctx context.Context, userID uint) (*entities.UserSubscription, error)
	GetUserActiveSubscriptions(ctx context.Context, userID uint) ([]*entities.UserSubscription, error)
	GetUserExpiredSubscriptions(ctx context.Context, userID uint, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Plan-specific operations
	ListByPlan(ctx context.Context, planID uint, limit, offset int) ([]*entities.UserSubscription, int64, error)
	CountByPlan(ctx context.Context, planID uint) (int64, error)
	GetActivePlanSubscriptions(ctx context.Context, planID uint, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Status filtering
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListActive(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListExpired(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListCancelled(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListInTrial(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Expiry and renewal operations
	ListExpiringBefore(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListForRenewal(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.UserSubscription, error)
	ListOverdueRenewals(ctx context.Context, limit int) ([]*entities.UserSubscription, error)
	ListTrialsExpiring(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.UserSubscription, error)

	// Time-based queries
	ListByDateRange(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListCreatedAfter(ctx context.Context, after time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListLastUsedBefore(ctx context.Context, before time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Traffic management operations
	ListByTrafficUsage(ctx context.Context, minUsage, maxUsage int64, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListTrafficSuspended(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListNearTrafficLimit(ctx context.Context, thresholdPercent float64, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListForTrafficReset(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.UserSubscription, error)

	// Auto-renewal operations
	ListAutoRenewEnabled(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListFailedRenewals(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListPendingCancellations(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Server group access operations
	ListByServerGroupAccess(ctx context.Context, serverGroupID uint, limit, offset int) ([]*entities.UserSubscription, int64, error)
	GetSubscriptionsWithServerAccess(ctx context.Context, serverGroupID uint) ([]*entities.UserSubscription, error)

	// Status management
	UpdateStatus(ctx context.Context, id uint, status string) error
	UpdateLastUsed(ctx context.Context, id uint, lastUsedAt time.Time) error
	UpdateTrafficUsage(ctx context.Context, id uint, trafficUsed int64) error
	AddTrafficUsage(ctx context.Context, id uint, additionalBytes int64) error
	ResetTrafficUsage(ctx context.Context, id uint, resetDate time.Time) error
	SuspendForTrafficLimit(ctx context.Context, id uint) error
	UnsuspendTraffic(ctx context.Context, id uint) error

	// Renewal management
	UpdateNextBillingDate(ctx context.Context, id uint, nextBillingDate time.Time) error
	UpdateRenewalAttempts(ctx context.Context, id uint, attempts int) error
	MarkRenewalFailed(ctx context.Context, id uint, failedAt time.Time, reason string) error
	ResetRenewalAttempts(ctx context.Context, id uint) error

	// Cancellation management
	CancelSubscription(ctx context.Context, id uint, reason string, cancelAtPeriodEnd bool) error
	CancelAtPeriodEnd(ctx context.Context, id uint, reason string) error
	UncancelSubscription(ctx context.Context, id uint) error

	// Server group access management
	UpdateServerGroupAccess(ctx context.Context, id uint, serverGroupIDs []uint) error
	GrantServerGroupAccess(ctx context.Context, id uint, serverGroupID uint) error
	RevokeServerGroupAccess(ctx context.Context, id uint, serverGroupID uint) error

	// Batch operations
	BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error)
	BatchCancel(ctx context.Context, ids []uint, reason string) (int, []uint, error)
	BatchResetTraffic(ctx context.Context, ids []uint, resetDate time.Time) (int, []uint, error)
	BatchDelete(ctx context.Context, ids []uint) (int, []uint, error)

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListDeleted(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.UserSubscription, int64, error)
	SearchByUserEmail(ctx context.Context, email string, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountActiveSubscriptions(ctx context.Context) (int64, error)
	CountExpiredSubscriptions(ctx context.Context) (int64, error)
	CountTrialSubscriptions(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)

	// Revenue statistics
	GetSubscriptionStats(ctx context.Context, since time.Time) (map[string]interface{}, error)
	GetChurnRate(ctx context.Context, period time.Duration) (float64, error)
	GetRetentionRate(ctx context.Context, period time.Duration) (float64, error)

	// Existence checks
	ExistsByUUID(ctx context.Context, uuid string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
	UserHasActiveSubscription(ctx context.Context, userID uint) (bool, error)
	UserHasSubscriptionToPlan(ctx context.Context, userID, planID uint) (bool, error)

	// Currency operations
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Advanced filtering
	ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Lifecycle management
	GetSubscriptionsNeedingAttention(ctx context.Context, limit int) ([]*entities.UserSubscription, error)
	GetSubscriptionsForMaintenance(ctx context.Context, maintenanceType string, limit int) ([]*entities.UserSubscription, error)
}