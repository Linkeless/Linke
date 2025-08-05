package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"
)

// paymentRetryHistoryRepository implements PaymentRetryHistoryRepository interface
type paymentRetryHistoryRepository struct {
	*repository.TimeBasedRepositoryImpl[entities.PaymentRetryHistory, uint]
}

// NewPaymentRetryHistoryRepository creates a new payment retry history repository
func NewPaymentRetryHistoryRepository(db *gorm.DB, logger framework.Logger) interfaces.PaymentRetryHistoryRepository {
	return &paymentRetryHistoryRepository{
		TimeBasedRepositoryImpl: repository.NewTimeBasedRepository[entities.PaymentRetryHistory, uint](db, logger),
	}
}

// Query operations

func (r *paymentRetryHistoryRepository) GetByRetryID(ctx context.Context, retryID uint) ([]*entities.PaymentRetryHistory, error) {
	var histories []*entities.PaymentRetryHistory
	if err := r.GetDB().WithContext(ctx).Where("payment_retry_id = ?", retryID).
		Order("created_at ASC").Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment retry history by retry ID: %w", err)
	}
	return histories, nil
}

func (r *paymentRetryHistoryRepository) GetByPaymentRecordID(ctx context.Context, paymentRecordID uint) ([]*entities.PaymentRetryHistory, error) {
	var histories []*entities.PaymentRetryHistory
	if err := r.GetDB().WithContext(ctx).
		Joins("JOIN payment_retries ON payment_retry_histories.payment_retry_id = payment_retries.id").
		Where("payment_retries.payment_record_id = ?", paymentRecordID).
		Order("payment_retry_histories.created_at ASC").
		Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("failed to get payment retry history by payment record ID: %w", err)
	}
	return histories, nil
}

func (r *paymentRetryHistoryRepository) GetRecentAttempts(ctx context.Context, retryID uint, limit int) ([]*entities.PaymentRetryHistory, error) {
	var histories []*entities.PaymentRetryHistory
	query := r.GetDB().WithContext(ctx).Where("payment_retry_id = ?", retryID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&histories).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent attempts: %w", err)
	}
	return histories, nil
}

func (r *paymentRetryHistoryRepository) GetAttemptsForPayment(ctx context.Context, paymentRecordID uint) ([]*entities.PaymentRetryHistory, error) {
	return r.GetByPaymentRecordID(ctx, paymentRecordID)
}

// Statistics

func (r *paymentRetryHistoryRepository) GetAttemptStatistics(ctx context.Context, retryID uint) (*interfaces.AttemptStatistics, error) {
	var stats struct {
		TotalAttempts   int     `gorm:"column:total_attempts"`
		SuccessfulCount int     `gorm:"column:successful_count"`
		FailedCount     int     `gorm:"column:failed_count"`
		TimeoutCount    int     `gorm:"column:timeout_count"`
		ErrorCount      int     `gorm:"column:error_count"`
		AverageDuration float64 `gorm:"column:average_duration"`
		TotalDuration   int     `gorm:"column:total_duration"`
	}

	query := `
		SELECT 
			COUNT(*) as total_attempts,
			COUNT(CASE WHEN response_status = 'success' THEN 1 END) as successful_count,
			COUNT(CASE WHEN response_status = 'failed' THEN 1 END) as failed_count,
			COUNT(CASE WHEN response_status = 'timeout' THEN 1 END) as timeout_count,
			COUNT(CASE WHEN response_status = 'error' THEN 1 END) as error_count,
			AVG(duration_ms) as average_duration,
			SUM(duration_ms) as total_duration
		FROM payment_retry_histories
		WHERE payment_retry_id = ?
	`

	if err := r.GetDB().WithContext(ctx).Raw(query, retryID).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get attempt statistics: %w", err)
	}

	return &interfaces.AttemptStatistics{
		RetryID:         retryID,
		TotalAttempts:   stats.TotalAttempts,
		SuccessfulCount: stats.SuccessfulCount,
		FailedCount:     stats.FailedCount,
		TimeoutCount:    stats.TimeoutCount,
		ErrorCount:      stats.ErrorCount,
		AverageDuration: stats.AverageDuration,
		TotalDuration:   stats.TotalDuration,
	}, nil
}

func (r *paymentRetryHistoryRepository) GetFailurePatterns(ctx context.Context, gateway string, days int) ([]*interfaces.FailurePattern, error) {
	var patterns []*interfaces.FailurePattern

	query := `
		SELECT 
			response_error_type as error_type,
			response_error_message as failure_reason,
			COUNT(*) as count,
			AVG(CASE WHEN response_status = 'success' THEN 1.0 ELSE 0.0 END) * 100 as success_rate,
			AVG(attempt_number) as average_attempts
		FROM payment_retry_histories prh
		JOIN payment_retries pr ON prh.payment_retry_id = pr.id
		JOIN payment_records p ON pr.payment_record_id = p.id
		WHERE p.gateway = ? 
			AND prh.created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
			AND response_error_type IS NOT NULL
		GROUP BY response_error_type, response_error_message
		ORDER BY count DESC
		LIMIT 50
	`

	rows, err := r.GetDB().WithContext(ctx).Raw(query, gateway, days).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get failure patterns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pattern interfaces.FailurePattern
		if err := rows.Scan(&pattern.ErrorType, &pattern.FailureReason, &pattern.Count, &pattern.SuccessRate, &pattern.AverageAttempts); err != nil {
			continue
		}
		patterns = append(patterns, &pattern)
	}

	return patterns, nil
}
