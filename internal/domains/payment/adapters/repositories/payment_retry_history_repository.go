package repositories

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"
)

// paymentRetryHistoryRepository implements PaymentRetryHistoryRepository interface
type paymentRetryHistoryRepository struct {
	db *gorm.DB
}

// NewPaymentRetryHistoryRepository creates a new payment retry history repository
func NewPaymentRetryHistoryRepository(db *gorm.DB) interfaces.PaymentRetryHistoryRepository {
	return &paymentRetryHistoryRepository{
		db: db,
	}
}

// Basic CRUD operations

func (r *paymentRetryHistoryRepository) Create(ctx context.Context, history *entities.PaymentRetryHistory) error {
	if err := r.db.WithContext(ctx).Create(history).Error; err != nil {
		logger.Error("Failed to create payment retry history",
			logger.Error2("error", err),
			logger.Uint("payment_retry_id", history.PaymentRetryID),
			logger.Int("attempt_number", history.AttemptNumber),
		)
		return fmt.Errorf("failed to create payment retry history: %w", err)
	}

	logger.Debug("Payment retry history created",
		logger.Uint("history_id", history.ID),
		logger.Uint("payment_retry_id", history.PaymentRetryID),
		logger.Int("attempt_number", history.AttemptNumber),
		logger.String("status", history.Status),
	)

	return nil
}

func (r *paymentRetryHistoryRepository) GetByID(ctx context.Context, id uint) (*entities.PaymentRetryHistory, error) {
	var history entities.PaymentRetryHistory
	if err := r.db.WithContext(ctx).First(&history, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment retry history not found: %d", id)
		}
		logger.Error("Failed to get payment retry history by ID",
			logger.Error2("error", err),
			logger.Uint("history_id", id),
		)
		return nil, fmt.Errorf("failed to get payment retry history: %w", err)
	}

	return &history, nil
}

func (r *paymentRetryHistoryRepository) GetByRetryID(ctx context.Context, retryID uint) ([]*entities.PaymentRetryHistory, error) {
	var histories []*entities.PaymentRetryHistory
	if err := r.db.WithContext(ctx).
		Where("payment_retry_id = ?", retryID).
		Order("attempt_number ASC").
		Find(&histories).Error; err != nil {
		logger.Error("Failed to get payment retry histories by retry ID",
			logger.Error2("error", err),
			logger.Uint("payment_retry_id", retryID),
		)
		return nil, fmt.Errorf("failed to get payment retry histories: %w", err)
	}

	return histories, nil
}

func (r *paymentRetryHistoryRepository) GetByPaymentRecordID(ctx context.Context, paymentRecordID uint) ([]*entities.PaymentRetryHistory, error) {
	var histories []*entities.PaymentRetryHistory
	if err := r.db.WithContext(ctx).
		Where("payment_record_id = ?", paymentRecordID).
		Order("attempted_at ASC").
		Find(&histories).Error; err != nil {
		logger.Error("Failed to get payment retry histories by payment record ID",
			logger.Error2("error", err),
			logger.Uint("payment_record_id", paymentRecordID),
		)
		return nil, fmt.Errorf("failed to get payment retry histories: %w", err)
	}

	return histories, nil
}

func (r *paymentRetryHistoryRepository) Update(ctx context.Context, history *entities.PaymentRetryHistory) error {
	if err := r.db.WithContext(ctx).Save(history).Error; err != nil {
		logger.Error("Failed to update payment retry history",
			logger.Error2("error", err),
			logger.Uint("history_id", history.ID),
		)
		return fmt.Errorf("failed to update payment retry history: %w", err)
	}

	return nil
}

func (r *paymentRetryHistoryRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entities.PaymentRetryHistory{}, id).Error; err != nil {
		logger.Error("Failed to delete payment retry history",
			logger.Error2("error", err),
			logger.Uint("history_id", id),
		)
		return fmt.Errorf("failed to delete payment retry history: %w", err)
	}

	return nil
}

// Query operations

func (r *paymentRetryHistoryRepository) GetRecentAttempts(ctx context.Context, retryID uint, limit int) ([]*entities.PaymentRetryHistory, error) {
	var histories []*entities.PaymentRetryHistory
	query := r.db.WithContext(ctx).
		Where("payment_retry_id = ?", retryID).
		Order("attempted_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&histories).Error; err != nil {
		logger.Error("Failed to get recent retry attempts",
			logger.Error2("error", err),
			logger.Uint("payment_retry_id", retryID),
			logger.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to get recent retry attempts: %w", err)
	}

	return histories, nil
}

func (r *paymentRetryHistoryRepository) GetAttemptsForPayment(ctx context.Context, paymentRecordID uint) ([]*entities.PaymentRetryHistory, error) {
	var histories []*entities.PaymentRetryHistory
	if err := r.db.WithContext(ctx).
		Where("payment_record_id = ?", paymentRecordID).
		Order("attempted_at ASC").
		Find(&histories).Error; err != nil {
		logger.Error("Failed to get attempts for payment",
			logger.Error2("error", err),
			logger.Uint("payment_record_id", paymentRecordID),
		)
		return nil, fmt.Errorf("failed to get attempts for payment: %w", err)
	}

	return histories, nil
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
			COUNT(CASE WHEN status = 'success' THEN 1 END) as successful_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
			COUNT(CASE WHEN status = 'timeout' THEN 1 END) as timeout_count,
			COUNT(CASE WHEN status = 'error' THEN 1 END) as error_count,
			AVG(duration) as average_duration,
			SUM(duration) as total_duration
		FROM payment_retry_histories
		WHERE payment_retry_id = ?
	`

	if err := r.db.WithContext(ctx).Raw(query, retryID).Scan(&stats).Error; err != nil {
		logger.Error("Failed to get attempt statistics",
			logger.Error2("error", err),
			logger.Uint("payment_retry_id", retryID),
		)
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
			prh.error_type,
			prh.failure_reason,
			COUNT(*) as count,
			COUNT(CASE WHEN prh.status = 'success' THEN 1 END) * 100.0 / COUNT(*) as success_rate,
			AVG(pr.attempt_number) as average_attempts
		FROM payment_retry_histories prh
		JOIN payment_retries pr ON prh.payment_retry_id = pr.id
		JOIN payment_records p ON pr.payment_record_id = p.id
		WHERE p.gateway = ? 
		  AND prh.attempted_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
		  AND prh.error_type IS NOT NULL
		  AND prh.failure_reason IS NOT NULL
		GROUP BY prh.error_type, prh.failure_reason
		HAVING count >= 2
		ORDER BY count DESC, success_rate ASC
		LIMIT 50
	`

	rows, err := r.db.WithContext(ctx).Raw(query, gateway, days).Rows()
	if err != nil {
		logger.Error("Failed to get failure patterns",
			logger.Error2("error", err),
			logger.String("gateway", gateway),
			logger.Int("days", days),
		)
		return nil, fmt.Errorf("failed to get failure patterns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pattern interfaces.FailurePattern
		if err := rows.Scan(
			&pattern.ErrorType,
			&pattern.FailureReason,
			&pattern.Count,
			&pattern.SuccessRate,
			&pattern.AverageAttempts,
		); err != nil {
			logger.Warn("Failed to scan failure pattern row",
				logger.Error2("error", err),
			)
			continue
		}

		patterns = append(patterns, &pattern)
	}

	return patterns, nil
}
