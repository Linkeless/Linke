package interfaces

import (
	"context"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/framework"
	"time"
)

// UserSubscriptionRepository defines the interface for user subscription data access operations
// It extends UserScopedRepository and TimeBasedRepository with UserSubscription-specific methods
type UserSubscriptionRepository interface {
	framework.UserScopedRepository[entities.UserSubscription, uint]
	framework.TimeBasedRepository[entities.UserSubscription, uint]

	// Subscription-specific query methods
	GetByUUID(ctx context.Context, uuid string) (*entities.UserSubscription, error)
	GetActiveByUser(ctx context.Context, userID uint) ([]*entities.UserSubscription, error)
	GetUserCurrentSubscription(ctx context.Context, userID uint) (*entities.UserSubscription, error)
	GetUserActiveSubscriptions(ctx context.Context, userID uint) ([]*entities.UserSubscription, error)
	GetUserExpiredSubscriptions(ctx context.Context, userID uint, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Plan-specific operations
	ListByPlan(ctx context.Context, planID uint, limit, offset int) ([]*entities.UserSubscription, int64, error)
	CountByPlan(ctx context.Context, planID uint) (int64, error)
	GetActivePlanSubscriptions(ctx context.Context, planID uint, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Status filtering (extending base status operations)
	ListActive(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListExpired(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListCancelled(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListInTrial(ctx context.Context, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Expiry and renewal operations
	ListExpiringBefore(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.UserSubscription, int64, error)
	ListForRenewal(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.UserSubscription, error)
	ListOverdueRenewals(ctx context.Context, limit int) ([]*entities.UserSubscription, error)
	ListTrialsExpiring(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.UserSubscription, error)

	// Additional time-based queries (extending TimeBasedRepository)
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

	// Subscription-specific batch operations (extending base batch operations)
	BatchCancel(ctx context.Context, ids []uint, reason string) (int, []uint, error)
	BatchResetTraffic(ctx context.Context, ids []uint, resetDate time.Time) (int, []uint, error)

	// Subscription-specific search operations (extending base search)
	SearchByUserEmail(ctx context.Context, email string, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Subscription-specific statistics (extending base statistics)
	CountActiveSubscriptions(ctx context.Context) (int64, error)
	CountExpiredSubscriptions(ctx context.Context) (int64, error)
	CountTrialSubscriptions(ctx context.Context) (int64, error)

	// Revenue statistics
	GetSubscriptionStats(ctx context.Context, since time.Time) (map[string]any, error)
	GetChurnRate(ctx context.Context, period time.Duration) (float64, error)
	GetRetentionRate(ctx context.Context, period time.Duration) (float64, error)

	// Subscription-specific existence checks (extending base existence checks)
	ExistsByUUID(ctx context.Context, uuid string) (bool, error)
	UserHasActiveSubscription(ctx context.Context, userID uint) (bool, error)
	UserHasSubscriptionToPlan(ctx context.Context, userID, planID uint) (bool, error)

	// Currency operations
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.UserSubscription, int64, error)

	// Lifecycle management
	GetSubscriptionsNeedingAttention(ctx context.Context, limit int) ([]*entities.UserSubscription, error)
	GetSubscriptionsForMaintenance(ctx context.Context, maintenanceType string, limit int) ([]*entities.UserSubscription, error)
}
