package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"
)

// paymentRetryRepository implements PaymentRetryRepository interface
type paymentRetryRepository struct {
	*repository.TimeBasedRepositoryImpl[entities.PaymentRetry, uint]
}

// NewPaymentRetryRepository creates a new payment retry repository
func NewPaymentRetryRepository(db *gorm.DB, logger framework.Logger) interfaces.PaymentRetryRepository {
	return &paymentRetryRepository{
		TimeBasedRepositoryImpl: repository.NewTimeBasedRepository[entities.PaymentRetry, uint](db, logger),
	}
}

func (r *paymentRetryRepository) GetByPaymentRecordID(ctx context.Context, paymentRecordID uint) (*entities.PaymentRetry, error) {
	var retry entities.PaymentRetry
	if err := r.GetDB().WithContext(ctx).Where("payment_record_id = ?", paymentRecordID).First(&retry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment retry not found for payment record: %d", paymentRecordID)
		}
		return nil, fmt.Errorf("failed to get payment retry: %w", err)
	}

	return &retry, nil
}

// Query operations

func (r *paymentRetryRepository) GetPendingRetries(ctx context.Context, limit int) ([]*entities.PaymentRetry, error) {
	var retries []*entities.PaymentRetry
	query := r.GetDB().WithContext(ctx).
		Where("status = ?", constants.PaymentRetryStatusPending).
		Order("next_retry_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&retries).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending retries: %w", err)
	}

	return retries, nil
}

func (r *paymentRetryRepository) GetRetriesDueForProcessing(ctx context.Context, beforeTime time.Time, limit int) ([]*entities.PaymentRetry, error) {
	var retries []*entities.PaymentRetry
	query := r.GetDB().WithContext(ctx).
		Where("status = ? AND next_retry_at <= ?", constants.PaymentRetryStatusPending, beforeTime).
		Order("next_retry_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&retries).Error; err != nil {
		return nil, fmt.Errorf("failed to get retries due for processing: %w", err)
	}

	return retries, nil
}

func (r *paymentRetryRepository) GetActiveRetriesForGateway(ctx context.Context, gateway string) ([]*entities.PaymentRetry, error) {
	var retries []*entities.PaymentRetry
	if err := r.GetDB().WithContext(ctx).
		Joins("JOIN payment_records ON payment_retries.payment_record_id = payment_records.id").
		Where("payment_records.gateway = ? AND payment_retries.status IN (?)",
			gateway, []string{constants.PaymentRetryStatusPending, constants.PaymentRetryStatusInProgress}).
		Find(&retries).Error; err != nil {
		return nil, fmt.Errorf("failed to get active retries for gateway: %w", err)
	}

	return retries, nil
}

func (r *paymentRetryRepository) GetRetriesForUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	var retries []*entities.PaymentRetry
	var total int64

	// Count total
	countQuery := r.GetDB().WithContext(ctx).Model(&entities.PaymentRetry{}).
		Joins("JOIN payment_records ON payment_retries.payment_record_id = payment_records.id").
		Where("payment_records.user_id = ?", userID)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count retries for user: %w", err)
	}

	// Get records
	query := r.GetDB().WithContext(ctx).
		Joins("JOIN payment_records ON payment_retries.payment_record_id = payment_records.id").
		Where("payment_records.user_id = ?", userID).
		Order("payment_retries.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&retries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get retries for user: %w", err)
	}

	return retries, total, nil
}

// Statistics and reporting

func (r *paymentRetryRepository) GetRetryStatsByGateway(ctx context.Context, gateway string, fromDate, toDate time.Time) (*dto.RetryStatistics, error) {
	var stats struct {
		TotalRetries      int64   `gorm:"column:total_retries"`
		SuccessfulRetries int64   `gorm:"column:successful_retries"`
		FailedRetries     int64   `gorm:"column:failed_retries"`
		CancelledRetries  int64   `gorm:"column:cancelled_retries"`
		AvgAttempts       float64 `gorm:"column:avg_attempts"`
		AvgDelayTime      float64 `gorm:"column:avg_delay_time"`
	}

	query := `
		SELECT 
			COUNT(*) as total_retries,
			COUNT(CASE WHEN pr.status = 'completed' THEN 1 END) as successful_retries,
			COUNT(CASE WHEN pr.status = 'failed' THEN 1 END) as failed_retries,
			COUNT(CASE WHEN pr.status = 'cancelled' THEN 1 END) as cancelled_retries,
			AVG(pr.attempt_number) as avg_attempts,
			AVG(pr.total_delay_time) as avg_delay_time
		FROM payment_retries pr
		JOIN payment_records p ON pr.payment_record_id = p.id
		WHERE p.gateway = ? AND pr.created_at BETWEEN ? AND ?
	`

	if err := r.GetDB().WithContext(ctx).Raw(query, gateway, fromDate, toDate).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get retry stats: %w", err)
	}

	successRate := float64(0)
	if stats.TotalRetries > 0 {
		successRate = float64(stats.SuccessfulRetries) / float64(stats.TotalRetries) * 100
	}

	return &dto.RetryStatistics{
		Gateway:           gateway,
		TotalRetries:      stats.TotalRetries,
		SuccessfulRetries: stats.SuccessfulRetries,
		FailedRetries:     stats.FailedRetries,
		CancelledRetries:  stats.CancelledRetries,
		SuccessRate:       successRate,
		AverageAttempts:   stats.AvgAttempts,
		AverageDelayTime:  stats.AvgDelayTime,
		FromDate:          fromDate,
		ToDate:            toDate,
	}, nil
}

func (r *paymentRetryRepository) GetRetryStatsByDate(ctx context.Context, fromDate, toDate time.Time) ([]*dto.DailyRetryStats, error) {
	var stats []*dto.DailyRetryStats

	query := `
		SELECT 
			DATE(pr.created_at) as date,
			COUNT(*) as total_retries,
			COUNT(CASE WHEN pr.status = 'completed' THEN 1 END) as successful_retries,
			COUNT(CASE WHEN pr.status = 'failed' THEN 1 END) as failed_retries
		FROM payment_retries pr
		WHERE pr.created_at BETWEEN ? AND ?
		GROUP BY DATE(pr.created_at)
		ORDER BY date DESC
	`

	rows, err := r.GetDB().WithContext(ctx).Raw(query, fromDate, toDate).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get retry stats by date: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat dto.DailyRetryStats
		if err := rows.Scan(&stat.Date, &stat.TotalRetries, &stat.SuccessfulRetries, &stat.FailedRetries); err != nil {
			continue
		}

		if stat.TotalRetries > 0 {
			stat.SuccessRate = float64(stat.SuccessfulRetries) / float64(stat.TotalRetries) * 100
		}

		stats = append(stats, &stat)
	}

	return stats, nil
}

func (r *paymentRetryRepository) GetRetrySuccessRate(ctx context.Context, gateway string, days int) (float64, error) {
	var result struct {
		TotalRetries      int64 `gorm:"column:total_retries"`
		SuccessfulRetries int64 `gorm:"column:successful_retries"`
	}

	fromDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT 
			COUNT(*) as total_retries,
			COUNT(CASE WHEN pr.status = 'completed' THEN 1 END) as successful_retries
		FROM payment_retries pr
		JOIN payment_records p ON pr.payment_record_id = p.id
		WHERE p.gateway = ? AND pr.created_at >= ?
	`

	if err := r.GetDB().WithContext(ctx).Raw(query, gateway, fromDate).Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to get retry success rate: %w", err)
	}

	if result.TotalRetries == 0 {
		return 0, nil
	}

	return float64(result.SuccessfulRetries) / float64(result.TotalRetries) * 100, nil
}

// Admin operations

func (r *paymentRetryRepository) GetAllRetries(ctx context.Context, filters *dto.RetryFilters, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	var retries []*entities.PaymentRetry
	var total int64

	query := r.GetDB().WithContext(ctx).Model(&entities.PaymentRetry{})

	// Apply filters
	if filters != nil {
		if filters.UserID != nil {
			query = query.Joins("JOIN payment_records ON payment_retries.payment_record_id = payment_records.id").
				Where("payment_records.user_id = ?", *filters.UserID)
		}
		if filters.Gateway != nil {
			if filters.UserID == nil {
				query = query.Joins("JOIN payment_records ON payment_retries.payment_record_id = payment_records.id")
			}
			query = query.Where("payment_records.gateway = ?", *filters.Gateway)
		}
		if filters.PaymentMethod != nil {
			if filters.UserID == nil && filters.Gateway == nil {
				query = query.Joins("JOIN payment_records ON payment_retries.payment_record_id = payment_records.id")
			}
			query = query.Where("payment_records.payment_method = ?", *filters.PaymentMethod)
		}
		if filters.Status != nil {
			query = query.Where("payment_retries.status = ?", *filters.Status)
		}
		if filters.FailureType != nil {
			query = query.Where("payment_retries.failure_type = ?", *filters.FailureType)
		}
		if filters.FromDate != nil {
			query = query.Where("payment_retries.created_at >= ?", *filters.FromDate)
		}
		if filters.ToDate != nil {
			query = query.Where("payment_retries.created_at <= ?", *filters.ToDate)
		}
		if filters.MinAttempts != nil {
			query = query.Where("payment_retries.attempt_number >= ?", *filters.MinAttempts)
		}
		if filters.MaxAttempts != nil {
			query = query.Where("payment_retries.attempt_number <= ?", *filters.MaxAttempts)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count retries: %w", err)
	}

	// Get records
	query = query.Order("payment_retries.created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&retries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get retries: %w", err)
	}

	return retries, total, nil
}

func (r *paymentRetryRepository) CancelRetry(ctx context.Context, id uint, reason string) error {
	now := time.Now()
	updates := map[string]any{
		"status":       constants.PaymentRetryStatusCancelled,
		"cancelled_at": &now,
		"completed_at": &now,
		"notes":        reason,
	}

	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRetry{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to cancel retry: %w", err)
	}

	return nil
}

func (r *paymentRetryRepository) ResetRetry(ctx context.Context, id uint) error {
	updates := map[string]any{
		"attempt_number":   0,
		"status":           constants.PaymentRetryStatusPending,
		"next_retry_at":    time.Now().Add(time.Hour), // Reset to 1 hour from now
		"last_attempt_at":  time.Now(),
		"completed_at":     nil,
		"cancelled_at":     nil,
		"successful_at":    nil,
		"total_delay_time": 0,
	}

	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRetry{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to reset retry: %w", err)
	}

	return nil
}

// Bulk operations

func (r *paymentRetryRepository) MarkRetriesAsInProgress(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRetry{}).
		Where("id IN (?)", ids).
		Update("status", constants.PaymentRetryStatusInProgress).Error; err != nil {
		return fmt.Errorf("failed to mark retries as in progress: %w", err)
	}

	return nil
}
