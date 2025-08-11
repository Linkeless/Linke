package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
	"linke/internal/shared/framework"
)

// PaymentRetryRepository defines the interface for payment retry data access
type PaymentRetryRepository interface {
	framework.TimeBasedRepository[entities.PaymentRetry, uint]

	// Specific queries
	GetByPaymentRecordID(ctx context.Context, paymentRecordID uint) (*entities.PaymentRetry, error)
	GetPendingRetries(ctx context.Context, limit int) ([]*entities.PaymentRetry, error)
	GetRetriesDueForProcessing(ctx context.Context, beforeTime time.Time, limit int) ([]*entities.PaymentRetry, error)
	GetActiveRetriesForGateway(ctx context.Context, gateway string) ([]*entities.PaymentRetry, error)
	GetRetriesForUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRetry, int64, error)

	// Statistics and reporting
	GetRetryStatsByGateway(ctx context.Context, gateway string, fromDate, toDate time.Time) (*dto.RetryStatistics, error)
	GetRetryStatsByDate(ctx context.Context, fromDate, toDate time.Time) ([]*dto.DailyRetryStats, error)
	GetRetrySuccessRate(ctx context.Context, gateway string, days int) (float64, error)

	// Admin operations
	GetAllRetries(ctx context.Context, filters *dto.RetryFilters, limit, offset int) ([]*entities.PaymentRetry, int64, error)
	CancelRetry(ctx context.Context, id uint, reason string) error
	ResetRetry(ctx context.Context, id uint) error

	// Bulk operations
	MarkRetriesAsInProgress(ctx context.Context, ids []uint) error
}

// PaymentRetryHistoryRepository defines the interface for payment retry history data access
type PaymentRetryHistoryRepository interface {
	framework.TimeBasedRepository[entities.PaymentRetryHistory, uint]

	// Specific queries
	GetByRetryID(ctx context.Context, retryID uint) ([]*entities.PaymentRetryHistory, error)
	GetByPaymentRecordID(ctx context.Context, paymentRecordID uint) ([]*entities.PaymentRetryHistory, error)
	GetRecentAttempts(ctx context.Context, retryID uint, limit int) ([]*entities.PaymentRetryHistory, error)
	GetAttemptsForPayment(ctx context.Context, paymentRecordID uint) ([]*entities.PaymentRetryHistory, error)

	// Statistics
	GetAttemptStatistics(ctx context.Context, retryID uint) (*dto.AttemptStatistics, error)
	GetFailurePatterns(ctx context.Context, gateway string, days int) ([]*dto.FailurePattern, error)
}

