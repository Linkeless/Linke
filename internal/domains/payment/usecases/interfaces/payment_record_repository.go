package interfaces

import (
	"context"
	"linke/internal/domains/payment/entities"
	"linke/internal/shared/framework"
	"time"
)

// PaymentRecordFilter provides flexible filtering options for payment records
type PaymentRecordFilter struct {
	UserID        *uint
	Status        string
	Gateway       string
	Method        string
	Currency      string
	MinAmount     *float64
	MaxAmount     *float64
	StartDate     *time.Time
	EndDate       *time.Time
	OrderID       *uint
}

// PaymentRecordRepository defines a simplified interface for payment record data access
// Extends framework repositories with payment-specific operations
type PaymentRecordRepository interface {
	framework.UserScopedTimeBasedRepository[entities.PaymentRecord, uint]

	// Core payment queries (most commonly used)
	GetByPaymentNo(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error)
	GetByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*entities.PaymentRecord, error)
	
	// Flexible filtering - replaces many specific list methods
	ListWithFilter(ctx context.Context, filter PaymentRecordFilter, limit, offset int) ([]*entities.PaymentRecord, int64, error)

	// Essential status operations
	UpdatePaymentStatus(ctx context.Context, id uint, status string, transactionID string, paidAt *time.Time) error

	// Statistics (consolidated)
	GetRevenueStats(ctx context.Context, currency string, startDate, endDate *time.Time) (*RevenueStats, error)
	
	// Existence checks for business logic
	ExistsByPaymentNo(ctx context.Context, paymentNo string) (bool, error)
	ExistsByOutTradeNo(ctx context.Context, outTradeNo string) (bool, error)
	
	// Duplicate prevention
	HasRecentPayment(ctx context.Context, userID uint, amount float64, currency string, within time.Duration) (bool, error)
}

// RevenueStats consolidates various revenue statistics
type RevenueStats struct {
	TotalRevenue     float64            `json:"total_revenue"`
	PaymentCount     int64              `json:"payment_count"`
	AverageAmount    float64            `json:"average_amount"`
	RevenueByGateway map[string]float64 `json:"revenue_by_gateway,omitempty"`
	RevenueByMethod  map[string]float64 `json:"revenue_by_method,omitempty"`
	DailyRevenue     map[string]float64 `json:"daily_revenue,omitempty"`
	MonthlyRevenue   map[string]float64 `json:"monthly_revenue,omitempty"`
}
