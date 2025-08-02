package interfaces

import (
	"context"
	"linke/internal/domains/payment/entities"
	"time"
)

// PaymentRecordRepository defines the interface for payment record data access operations
type PaymentRecordRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, payment *entities.PaymentRecord) error
	GetByID(ctx context.Context, id uint) (*entities.PaymentRecord, error)
	GetByPaymentNo(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error)
	GetByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*entities.PaymentRecord, error)
	Update(ctx context.Context, payment *entities.PaymentRecord) error
	Delete(ctx context.Context, id uint) error

	// Soft delete operations
	SoftDelete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error
	HardDelete(ctx context.Context, id uint) error

	// User-specific operations
	ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetUserPaymentHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetUserCompletedPayments(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetUserTotalPaid(ctx context.Context, userID uint, currency string) (float64, error)

	// Status filtering
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListPendingPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListCompletedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListFailedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListExpiredPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Gateway and method filtering
	ListByGateway(ctx context.Context, gateway string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListByPaymentMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListByGatewayAndMethod(ctx context.Context, gateway, paymentMethod string, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Time-based queries
	ListByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.PaymentRecord, int64, error)
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

	// Status management
	UpdateStatus(ctx context.Context, id uint, status string) error
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

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	SearchByUserEmail(ctx context.Context, email string, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Statistics and reporting
	CountTotal(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	CountByUser(ctx context.Context, userID uint) (int64, error)
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

	// Batch operations
	BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error)
	BatchDelete(ctx context.Context, ids []uint) (int, []uint, error)

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	ListDeleted(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Existence checks
	ExistsByPaymentNo(ctx context.Context, paymentNo string) (bool, error)
	ExistsByOutTradeNo(ctx context.Context, outTradeNo string) (bool, error)
	ExistsByTransactionID(ctx context.Context, transactionID string) (bool, error)

	// Advanced filtering
	ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Payment number generation support
	GetLastPaymentNumber(ctx context.Context, prefix string) (string, error)

	// Duplicate prevention
	GetPendingPaymentByUserAndAmount(ctx context.Context, userID uint, amount float64, currency string) (*entities.PaymentRecord, error)
	HasRecentPayment(ctx context.Context, userID uint, amount float64, currency string, within time.Duration) (bool, error)
}
