package interfaces

import (
	"context"
	"time"

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
	GetRetryStatsByGateway(ctx context.Context, gateway string, fromDate, toDate time.Time) (*RetryStatistics, error)
	GetRetryStatsByDate(ctx context.Context, fromDate, toDate time.Time) ([]*DailyRetryStats, error)
	GetRetrySuccessRate(ctx context.Context, gateway string, days int) (float64, error)

	// Admin operations
	GetAllRetries(ctx context.Context, filters *RetryFilters, limit, offset int) ([]*entities.PaymentRetry, int64, error)
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
	GetAttemptStatistics(ctx context.Context, retryID uint) (*AttemptStatistics, error)
	GetFailurePatterns(ctx context.Context, gateway string, days int) ([]*FailurePattern, error)
}

// Data structures for filters and statistics

// RetryFilters represents filters for querying payment retries
type RetryFilters struct {
	UserID        *uint      `json:"user_id,omitempty"`
	Gateway       *string    `json:"gateway,omitempty"`
	PaymentMethod *string    `json:"payment_method,omitempty"`
	Status        *string    `json:"status,omitempty"`
	FailureType   *string    `json:"failure_type,omitempty"`
	FromDate      *time.Time `json:"from_date,omitempty"`
	ToDate        *time.Time `json:"to_date,omitempty"`
	MinAttempts   *int       `json:"min_attempts,omitempty"`
	MaxAttempts   *int       `json:"max_attempts,omitempty"`
}

// RetryStatistics represents aggregate retry statistics
type RetryStatistics struct {
	Gateway           string    `json:"gateway"`
	PaymentMethod     string    `json:"payment_method,omitempty"`
	TotalRetries      int64     `json:"total_retries"`
	SuccessfulRetries int64     `json:"successful_retries"`
	FailedRetries     int64     `json:"failed_retries"`
	CancelledRetries  int64     `json:"cancelled_retries"`
	SuccessRate       float64   `json:"success_rate"`
	AverageAttempts   float64   `json:"average_attempts"`
	AverageDelayTime  float64   `json:"average_delay_time"`
	FromDate          time.Time `json:"from_date"`
	ToDate            time.Time `json:"to_date"`
}

// DailyRetryStats represents daily retry statistics
type DailyRetryStats struct {
	Date              time.Time `json:"date"`
	TotalRetries      int64     `json:"total_retries"`
	SuccessfulRetries int64     `json:"successful_retries"`
	FailedRetries     int64     `json:"failed_retries"`
	SuccessRate       float64   `json:"success_rate"`
}

// AttemptStatistics represents statistics for retry attempts
type AttemptStatistics struct {
	RetryID         uint    `json:"retry_id"`
	TotalAttempts   int     `json:"total_attempts"`
	SuccessfulCount int     `json:"successful_count"`
	FailedCount     int     `json:"failed_count"`
	TimeoutCount    int     `json:"timeout_count"`
	ErrorCount      int     `json:"error_count"`
	AverageDuration float64 `json:"average_duration"`
	TotalDuration   int     `json:"total_duration"`
}

// FailurePattern represents common failure patterns
type FailurePattern struct {
	ErrorType       string  `json:"error_type"`
	FailureReason   string  `json:"failure_reason"`
	Count           int64   `json:"count"`
	SuccessRate     float64 `json:"success_rate"`
	AverageAttempts float64 `json:"average_attempts"`
}
