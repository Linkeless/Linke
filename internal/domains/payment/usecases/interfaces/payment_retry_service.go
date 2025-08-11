package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
)

// PaymentRetryService defines the interface for payment retry operations
type PaymentRetryService interface {
	// Retry management
	InitiateRetry(ctx context.Context, paymentRecord *entities.PaymentRecord, failureType, failureCode, errorMessage string) (*entities.PaymentRetry, error)
	ProcessPendingRetries(ctx context.Context, batchSize int) (int, error)
	ProcessRetry(ctx context.Context, retryID uint) (*dto.RetryResult, error)
	CancelRetry(ctx context.Context, retryID uint, reason string) error
	ResetRetry(ctx context.Context, retryID uint) error

	// Configuration and strategy
	GetRetryStrategy(ctx context.Context, gateway, paymentMethod string) (*entities.RetryStrategyConfig, error)
	UpdateRetryStrategy(ctx context.Context, gateway, paymentMethod string, config *entities.RetryStrategyConfig) error
	CalculateNextRetryTime(ctx context.Context, retry *entities.PaymentRetry) time.Time

	// Query operations
	GetRetryByPaymentID(ctx context.Context, paymentRecordID uint) (*entities.PaymentRetry, error)
	GetRetryWithHistory(ctx context.Context, retryID uint) (*dto.RetryWithHistory, error)
	GetActiveRetries(ctx context.Context, filters *dto.RetryFilters, limit, offset int) ([]*entities.PaymentRetry, int64, error)
	GetRetryHistory(ctx context.Context, retryID uint) ([]*entities.PaymentRetryHistory, error)

	// Statistics and monitoring
	GetRetryStatistics(ctx context.Context, gateway string, days int) (*dto.RetryStatistics, error)
	GetFailureAnalysis(ctx context.Context, gateway string, days int) (*dto.FailureAnalysis, error)
	GetRetryHealthMetrics(ctx context.Context) (*dto.RetryHealthMetrics, error)

	// Admin operations
	GetRetriesForAdmin(ctx context.Context, filters *dto.AdminRetryFilters) (*dto.AdminRetryResponse, error)
	BulkCancelRetries(ctx context.Context, retryIDs []uint, reason string) error
	BulkResetRetries(ctx context.Context, retryIDs []uint) error

	// Integration points
	ClassifyFailure(ctx context.Context, gateway, paymentMethod, errorCode, errorMessage string) string
	ShouldRetryPayment(ctx context.Context, paymentRecord *entities.PaymentRecord, errorCode string) bool
	NotifyRetryAttempt(ctx context.Context, retry *entities.PaymentRetry, attempt *entities.PaymentRetryHistory) error
}

