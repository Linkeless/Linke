package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"
)

// paymentRetryRepository implements PaymentRetryRepository interface
type paymentRetryRepository struct {
	db *gorm.DB
}

// NewPaymentRetryRepository creates a new payment retry repository
func NewPaymentRetryRepository(db *gorm.DB) interfaces.PaymentRetryRepository {
	return &paymentRetryRepository{
		db: db,
	}
}

// Basic CRUD operations

func (r *paymentRetryRepository) Create(ctx context.Context, retry *entities.PaymentRetry) error {
	if err := r.db.WithContext(ctx).Create(retry).Error; err != nil {
		logger.Error("Failed to create payment retry",
			logger.Error2("error", err),
			logger.Uint("payment_record_id", retry.PaymentRecordID),
		)
		return fmt.Errorf("failed to create payment retry: %w", err)
	}

	logger.Info("Payment retry created successfully",
		logger.Uint("retry_id", retry.ID),
		logger.Uint("payment_record_id", retry.PaymentRecordID),
		logger.String("strategy", retry.RetryStrategy),
	)

	return nil
}

func (r *paymentRetryRepository) GetByID(ctx context.Context, id uint) (*entities.PaymentRetry, error) {
	var retry entities.PaymentRetry
	if err := r.db.WithContext(ctx).First(&retry, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment retry not found: %d", id)
		}
		logger.Error("Failed to get payment retry by ID",
			logger.Error2("error", err),
			logger.Uint("retry_id", id),
		)
		return nil, fmt.Errorf("failed to get payment retry: %w", err)
	}

	return &retry, nil
}

func (r *paymentRetryRepository) GetByPaymentRecordID(ctx context.Context, paymentRecordID uint) (*entities.PaymentRetry, error) {
	var retry entities.PaymentRetry
	if err := r.db.WithContext(ctx).Where("payment_record_id = ?", paymentRecordID).First(&retry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment retry not found for payment record: %d", paymentRecordID)
		}
		logger.Error("Failed to get payment retry by payment record ID",
			logger.Error2("error", err),
			logger.Uint("payment_record_id", paymentRecordID),
		)
		return nil, fmt.Errorf("failed to get payment retry: %w", err)
	}

	return &retry, nil
}

func (r *paymentRetryRepository) Update(ctx context.Context, retry *entities.PaymentRetry) error {
	if err := r.db.WithContext(ctx).Save(retry).Error; err != nil {
		logger.Error("Failed to update payment retry",
			logger.Error2("error", err),
			logger.Uint("retry_id", retry.ID),
		)
		return fmt.Errorf("failed to update payment retry: %w", err)
	}

	return nil
}

func (r *paymentRetryRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entities.PaymentRetry{}, id).Error; err != nil {
		logger.Error("Failed to delete payment retry",
			logger.Error2("error", err),
			logger.Uint("retry_id", id),
		)
		return fmt.Errorf("failed to delete payment retry: %w", err)
	}

	return nil
}

// Query operations

func (r *paymentRetryRepository) GetPendingRetries(ctx context.Context, limit int) ([]*entities.PaymentRetry, error) {
	var retries []*entities.PaymentRetry
	query := r.db.WithContext(ctx).
		Where("status = ?", entities.PaymentRetryStatusPending).
		Order("next_retry_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&retries).Error; err != nil {
		logger.Error("Failed to get pending retries",
			logger.Error2("error", err),
			logger.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to get pending retries: %w", err)
	}

	return retries, nil
}

func (r *paymentRetryRepository) GetRetriesDueForProcessing(ctx context.Context, beforeTime time.Time, limit int) ([]*entities.PaymentRetry, error) {
	var retries []*entities.PaymentRetry
	query := r.db.WithContext(ctx).
		Where("status = ? AND next_retry_at <= ?", entities.PaymentRetryStatusPending, beforeTime).
		Order("next_retry_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&retries).Error; err != nil {
		logger.Error("Failed to get retries due for processing",
			logger.Error2("error", err),
			logger.String("before_time", beforeTime.String()),
			logger.Int("limit", limit),
		)
		return nil, fmt.Errorf("failed to get retries due for processing: %w", err)
	}

	return retries, nil
}

func (r *paymentRetryRepository) GetActiveRetriesForGateway(ctx context.Context, gateway string) ([]*entities.PaymentRetry, error) {
	var retries []*entities.PaymentRetry
	if err := r.db.WithContext(ctx).
		Joins("JOIN payment_records ON payment_retries.payment_record_id = payment_records.id").
		Where("payment_records.gateway = ? AND payment_retries.status IN (?)",
			gateway, []string{entities.PaymentRetryStatusPending, entities.PaymentRetryStatusInProgress}).
		Find(&retries).Error; err != nil {
		logger.Error("Failed to get active retries for gateway",
			logger.Error2("error", err),
			logger.String("gateway", gateway),
		)
		return nil, fmt.Errorf("failed to get active retries for gateway: %w", err)
	}

	return retries, nil
}

func (r *paymentRetryRepository) GetRetriesForUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	var retries []*entities.PaymentRetry
	var total int64

	// Count total
	countQuery := r.db.WithContext(ctx).Model(&entities.PaymentRetry{}).
		Joins("JOIN payment_records ON payment_retries.payment_record_id = payment_records.id").
		Where("payment_records.user_id = ?", userID)

	if err := countQuery.Count(&total).Error; err != nil {
		logger.Error("Failed to count retries for user",
			logger.Error2("error", err),
			logger.Uint("user_id", userID),
		)
		return nil, 0, fmt.Errorf("failed to count retries for user: %w", err)
	}

	// Get records
	query := r.db.WithContext(ctx).
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
		logger.Error("Failed to get retries for user",
			logger.Error2("error", err),
			logger.Uint("user_id", userID),
		)
		return nil, 0, fmt.Errorf("failed to get retries for user: %w", err)
	}

	return retries, total, nil
}

// Statistics and reporting

func (r *paymentRetryRepository) GetRetryStatsByGateway(ctx context.Context, gateway string, fromDate, toDate time.Time) (*interfaces.RetryStatistics, error) {
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

	if err := r.db.WithContext(ctx).Raw(query, gateway, fromDate, toDate).Scan(&stats).Error; err != nil {
		logger.Error("Failed to get retry stats by gateway",
			logger.Error2("error", err),
			logger.String("gateway", gateway),
		)
		return nil, fmt.Errorf("failed to get retry stats: %w", err)
	}

	successRate := float64(0)
	if stats.TotalRetries > 0 {
		successRate = float64(stats.SuccessfulRetries) / float64(stats.TotalRetries) * 100
	}

	return &interfaces.RetryStatistics{
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

func (r *paymentRetryRepository) GetRetryStatsByDate(ctx context.Context, fromDate, toDate time.Time) ([]*interfaces.DailyRetryStats, error) {
	var stats []*interfaces.DailyRetryStats

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

	rows, err := r.db.WithContext(ctx).Raw(query, fromDate, toDate).Rows()
	if err != nil {
		logger.Error("Failed to get retry stats by date",
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get retry stats by date: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stat interfaces.DailyRetryStats
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

	if err := r.db.WithContext(ctx).Raw(query, gateway, fromDate).Scan(&result).Error; err != nil {
		logger.Error("Failed to get retry success rate",
			logger.Error2("error", err),
			logger.String("gateway", gateway),
		)
		return 0, fmt.Errorf("failed to get retry success rate: %w", err)
	}

	if result.TotalRetries == 0 {
		return 0, nil
	}

	return float64(result.SuccessfulRetries) / float64(result.TotalRetries) * 100, nil
}

// Admin operations

func (r *paymentRetryRepository) GetAllRetries(ctx context.Context, filters *interfaces.RetryFilters, limit, offset int) ([]*entities.PaymentRetry, int64, error) {
	var retries []*entities.PaymentRetry
	var total int64

	query := r.db.WithContext(ctx).Model(&entities.PaymentRetry{})

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
		logger.Error("Failed to count all retries",
			logger.Error2("error", err),
		)
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
		logger.Error("Failed to get all retries",
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to get retries: %w", err)
	}

	return retries, total, nil
}

func (r *paymentRetryRepository) CancelRetry(ctx context.Context, id uint, reason string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       entities.PaymentRetryStatusCancelled,
		"cancelled_at": &now,
		"completed_at": &now,
		"notes":        reason,
	}

	if err := r.db.WithContext(ctx).Model(&entities.PaymentRetry{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Error("Failed to cancel retry",
			logger.Error2("error", err),
			logger.Uint("retry_id", id),
		)
		return fmt.Errorf("failed to cancel retry: %w", err)
	}

	logger.Info("Payment retry cancelled",
		logger.Uint("retry_id", id),
		logger.String("reason", reason),
	)

	return nil
}

func (r *paymentRetryRepository) ResetRetry(ctx context.Context, id uint) error {
	updates := map[string]interface{}{
		"attempt_number":   0,
		"status":           entities.PaymentRetryStatusPending,
		"next_retry_at":    time.Now().Add(time.Hour), // Reset to 1 hour from now
		"last_attempt_at":  time.Now(),
		"completed_at":     nil,
		"cancelled_at":     nil,
		"successful_at":    nil,
		"total_delay_time": 0,
	}

	if err := r.db.WithContext(ctx).Model(&entities.PaymentRetry{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logger.Error("Failed to reset retry",
			logger.Error2("error", err),
			logger.Uint("retry_id", id),
		)
		return fmt.Errorf("failed to reset retry: %w", err)
	}

	logger.Info("Payment retry reset",
		logger.Uint("retry_id", id),
	)

	return nil
}

// Bulk operations

func (r *paymentRetryRepository) MarkRetriesAsInProgress(ctx context.Context, ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).Model(&entities.PaymentRetry{}).
		Where("id IN (?)", ids).
		Update("status", entities.PaymentRetryStatusInProgress).Error; err != nil {
		logger.Error("Failed to mark retries as in progress",
			logger.Error2("error", err),
			logger.Int("count", len(ids)),
		)
		return fmt.Errorf("failed to mark retries as in progress: %w", err)
	}

	return nil
}

func (r *paymentRetryRepository) UpdateRetryStatus(ctx context.Context, id uint, status string) error {
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRetry{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		logger.Error("Failed to update retry status",
			logger.Error2("error", err),
			logger.Uint("retry_id", id),
			logger.String("status", status),
		)
		return fmt.Errorf("failed to update retry status: %w", err)
	}

	return nil
}
