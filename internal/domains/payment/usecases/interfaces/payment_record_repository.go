package interfaces

import (
	"context"
	"linke/internal/domains/payment/entities"
	"linke/internal/shared/framework"
	"time"
)

// PaymentRecordRepository defines the interface for payment record data access operations
// It extends UserScopedRepository and TimeBasedRepository with PaymentRecord-specific methods
type PaymentRecordRepository interface {
	framework.UserScopedRepository[entities.PaymentRecord, uint]
	framework.TimeBasedRepository[entities.PaymentRecord, uint]

	// Payment-specific query methods
	GetByPaymentNo(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error)
	GetByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*entities.PaymentRecord, error)

	// User-specific operations (extending UserScopedRepository)
	GetUserPaymentHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetUserCompletedPayments(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetUserTotalPaid(ctx context.Context, userID uint, currency string) (float64, error)

	// Payment-specific status filtering
	ListPendingPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListCompletedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListFailedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListExpiredPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Gateway and method filtering
	ListByGateway(ctx context.Context, gateway string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListByPaymentMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListByGatewayAndMethod(ctx context.Context, gateway, paymentMethod string, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Time-based queries (extending TimeBasedRepository)
	ListRecentPayments(ctx context.Context, since time.Time, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListPaymentsByPeriod(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Order-related operations
	ListBySubscriptionOrder(ctx context.Context, orderID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetOrderPayments(ctx context.Context, orderID uint) ([]*entities.PaymentRecord, error)
	GetOrderCompletedPayments(ctx context.Context, orderID uint) ([]*entities.PaymentRecord, error)

	// Refund operations
	ListRefundedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListByRefundStatus(ctx context.Context, refundStatus string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListRefundablePayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Currency operations
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetSupportedCurrencies(ctx context.Context) ([]string, error)

	// Payment-specific status management
	UpdatePaymentStatus(ctx context.Context, id uint, paymentStatus string) error
	MarkAsCompleted(ctx context.Context, id uint, transactionID string, paidAt time.Time) error
	MarkAsFailed(ctx context.Context, id uint, reason string) error
	MarkAsRefunded(ctx context.Context, id uint, refundAmount float64, refundReason string, refundedAt time.Time) error

	// Notification management
	UpdateNotification(ctx context.Context, id uint, notifiedAt time.Time, notifyHash string) error
	IncrementNotifyCount(ctx context.Context, id uint) error
	GetByNotifyHash(ctx context.Context, notifyHash string) (*entities.PaymentRecord, error)

	// Expiry management
	ListExpiringPayments(ctx context.Context, beforeTime time.Time, limit int) ([]*entities.PaymentRecord, error)
	MarkExpiredPayments(ctx context.Context, beforeTime time.Time) (int64, error)

	// Payment-specific search operations
	SearchByUserEmail(ctx context.Context, email string, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Payment-specific statistics
	CountByGateway(ctx context.Context, gateway string) (int64, error)
	CountCompletedPayments(ctx context.Context, since time.Time) (int64, error)
	CountFailedPayments(ctx context.Context, since time.Time) (int64, error)

	// Revenue statistics
	GetTotalRevenue(ctx context.Context, currency string, since time.Time) (float64, error)
	GetRevenueByGateway(ctx context.Context, gateway, currency string, since time.Time) (float64, error)
	GetRevenueByMethod(ctx context.Context, paymentMethod, currency string, since time.Time) (float64, error)
	GetDailyRevenue(ctx context.Context, currency string, days int) (map[string]float64, error)
	GetMonthlyRevenue(ctx context.Context, currency string, months int) (map[string]float64, error)

	// Amount-based queries
	ListByAmountRange(ctx context.Context, minAmount, maxAmount float64, currency string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetAveragePaymentAmount(ctx context.Context, currency string, since time.Time) (float64, error)
	GetLargestPayment(ctx context.Context, currency string, since time.Time) (*entities.PaymentRecord, error)

	// Payment-specific existence checks
	ExistsByPaymentNo(ctx context.Context, paymentNo string) (bool, error)
	ExistsByOutTradeNo(ctx context.Context, outTradeNo string) (bool, error)
	ExistsByTransactionID(ctx context.Context, transactionID string) (bool, error)

	// Payment number generation support
	GetLastPaymentNumber(ctx context.Context, prefix string) (string, error)

	// Duplicate prevention
	GetPendingPaymentByUserAndAmount(ctx context.Context, userID uint, amount float64, currency string) (*entities.PaymentRecord, error)
	HasRecentPayment(ctx context.Context, userID uint, amount float64, currency string, within time.Duration) (bool, error)
}
