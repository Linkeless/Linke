package interfaces

import (
	"context"
	"linke/internal/domains/subscription/entities"
	"linke/internal/shared/framework"
	"time"
)

// SubscriptionOrderRepository defines the interface for subscription order data access operations
// It extends UserScopedRepository and TimeBasedRepository with SubscriptionOrder-specific methods
type SubscriptionOrderRepository interface {
	framework.UserScopedRepository[entities.SubscriptionOrder, uint]
	framework.TimeBasedRepository[entities.SubscriptionOrder, uint]
	
	// Subscription order specific query methods
	GetByOrderNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error)

	// User-specific operations (extending UserScopedRepository)
	GetUserOrderHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	GetUserActiveOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)

	// Plan-specific operations
	ListByPlan(ctx context.Context, planID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	GetPlanOrderStats(ctx context.Context, planID uint, since time.Time) (map[string]int64, error)

	// Order type and payment specific filtering
	ListByOrderType(ctx context.Context, orderType string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	ListByPaymentStatus(ctx context.Context, paymentStatus string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)

	// Payment-related operations
	ListPendingPayments(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	ListFailedPayments(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	ListByTransactionID(ctx context.Context, transactionID string) ([]*entities.SubscriptionOrder, error)
	ListByPaymentGateway(ctx context.Context, gateway string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)

	// Time-based queries (extending TimeBasedRepository)
	ListRecentOrders(ctx context.Context, since time.Time, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	ListOrdersForBillingPeriod(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)

	// Coupon and discount operations
	ListByCouponCode(ctx context.Context, couponCode string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	ListWithDiscounts(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)

	// Refund operations
	ListRefundedOrders(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	ListRefundableOrders(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)

	// Invoice operations
	ListByInvoiceStatus(ctx context.Context, invoiceStatus string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	ListUninvoicedOrders(ctx context.Context, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)

	// Subscription order specific status management
	UpdatePaymentStatus(ctx context.Context, id uint, paymentStatus string, transactionID string) error
	UpdateInvoiceStatus(ctx context.Context, id uint, invoiceStatus string, invoiceNumber string) error
	MarkAsPaid(ctx context.Context, id uint, transactionID string, paidAt time.Time) error
	MarkAsRefunded(ctx context.Context, id uint, refundAmount float64, refundReason string, refundedAt time.Time) error

	// Subscription order specific search operations
	SearchByUserEmail(ctx context.Context, email string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)

	// Subscription order specific statistics (extending base statistics)
	CountByPlan(ctx context.Context, planID uint) (int64, error)
	CountPaidOrders(ctx context.Context, since time.Time) (int64, error)
	CountFailedOrders(ctx context.Context, since time.Time) (int64, error)

	// Revenue statistics
	GetTotalRevenue(ctx context.Context, currency string, since time.Time) (float64, error)
	GetRevenueByPlan(ctx context.Context, planID uint, currency string, since time.Time) (float64, error)
	GetRevenueByPeriod(ctx context.Context, currency string, start, end time.Time) (float64, error)
	GetDailyRevenue(ctx context.Context, currency string, days int) (map[string]float64, error)
	GetMonthlyRevenue(ctx context.Context, currency string, months int) (map[string]float64, error)

	// Currency operations
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	GetSupportedCurrencies(ctx context.Context) ([]string, error)

	// Order number generation support
	GetLastOrderNumber(ctx context.Context, prefix string) (string, error)
	ExistsByOrderNumber(ctx context.Context, orderNumber string) (bool, error)

	// Subscription relationship operations
	ListBySubscription(ctx context.Context, subscriptionID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	GetSubscriptionOrders(ctx context.Context, subscriptionID uint) ([]*entities.SubscriptionOrder, error)

	// Subscription order specific filtering and renewal operations
	GetOrdersForRenewal(ctx context.Context, beforeDate time.Time, limit int) ([]*entities.SubscriptionOrder, error)
}
