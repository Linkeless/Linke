package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/logger"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// paymentRecordRepository implements the PaymentRecordRepository interface
type paymentRecordRepository struct {
	*repository.UserScopedTimeBasedRepositoryImpl[entities.PaymentRecord, uint]
}

// paymentConfigRepository implements the simplified PaymentConfigRepository interface
type paymentConfigRepository struct {
	*repository.BaseRepositoryImpl[entities.PaymentConfig, uint]
}

// NewPaymentRecordRepository creates a new PaymentRecordRepository implementation
func NewPaymentRecordRepository(db *gorm.DB, logger framework.Logger) interfaces.PaymentRecordRepository {
	return &paymentRecordRepository{
		UserScopedTimeBasedRepositoryImpl: repository.NewUserScopedTimeBasedRepository[entities.PaymentRecord, uint](db, logger),
	}
}

// NewPaymentConfigRepository creates a new PaymentConfigRepository implementation
func NewPaymentConfigRepository(db *gorm.DB, logger framework.Logger) interfaces.PaymentConfigRepository {
	return &paymentConfigRepository{
		BaseRepositoryImpl: repository.NewBaseRepository[entities.PaymentConfig, uint](db, logger),
	}
}

// === PaymentRecordRepository Implementation ===

// GetRevenueStats calculates revenue statistics for the given parameters
func (r *paymentRecordRepository) GetRevenueStats(ctx context.Context, currency string, startDate, endDate *time.Time) (*interfaces.RevenueStats, error) {
	stats := &interfaces.RevenueStats{}

	// Build where condition
	conditions := []string{"status = ?"}
	args := []interface{}{constants.PaymentRecordStatusCompleted}

	if currency != "" {
		conditions = append(conditions, "currency = ?")
		args = append(args, currency)
	}

	if startDate != nil {
		conditions = append(conditions, "paid_at >= ?")
		args = append(args, *startDate)
	}

	if endDate != nil {
		conditions = append(conditions, "paid_at <= ?")
		args = append(args, *endDate)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Get basic stats
	query := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(amount), 0) as total_revenue,
			COUNT(*) as payment_count,
			COALESCE(AVG(amount), 0) as average_amount
		FROM payment_records 
		WHERE %s
	`, whereClause)

	row := r.GetDB().WithContext(ctx).Raw(query, args...).Row()
	if err := row.Scan(&stats.TotalRevenue, &stats.PaymentCount, &stats.AverageAmount); err != nil {
		return nil, fmt.Errorf("failed to get basic revenue stats: %w", err)
	}

	// Get revenue by gateway
	stats.RevenueByGateway = make(map[string]float64)
	gatewayQuery := fmt.Sprintf(`
		SELECT gateway, COALESCE(SUM(amount), 0) as revenue
		FROM payment_records 
		WHERE %s
		GROUP BY gateway
	`, whereClause)

	gatewayRows, err := r.GetDB().WithContext(ctx).Raw(gatewayQuery, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue by gateway: %w", err)
	}
	defer gatewayRows.Close()

	for gatewayRows.Next() {
		var gateway string
		var revenue float64
		if err := gatewayRows.Scan(&gateway, &revenue); err != nil {
			return nil, fmt.Errorf("failed to scan gateway revenue: %w", err)
		}
		stats.RevenueByGateway[gateway] = revenue
	}

	// Get revenue by method
	stats.RevenueByMethod = make(map[string]float64)
	methodQuery := fmt.Sprintf(`
		SELECT method, COALESCE(SUM(amount), 0) as revenue
		FROM payment_records 
		WHERE %s
		GROUP BY method
	`, whereClause)

	methodRows, err := r.GetDB().WithContext(ctx).Raw(methodQuery, args...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue by method: %w", err)
	}
	defer methodRows.Close()

	for methodRows.Next() {
		var method string
		var revenue float64
		if err := methodRows.Scan(&method, &revenue); err != nil {
			return nil, fmt.Errorf("failed to scan method revenue: %w", err)
		}
		stats.RevenueByMethod[method] = revenue
	}

	return stats, nil
}

// ListWithFilter provides flexible filtering for payment records
func (r *paymentRecordRepository) ListWithFilter(ctx context.Context, filter interfaces.PaymentRecordFilter, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Build dynamic query
	conditions := []string{}
	args := []interface{}{}

	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}

	if filter.Gateway != "" {
		conditions = append(conditions, "gateway = ?")
		args = append(args, filter.Gateway)
	}

	if filter.Method != "" {
		conditions = append(conditions, "method = ?")
		args = append(args, filter.Method)
	}

	if filter.Currency != "" {
		conditions = append(conditions, "currency = ?")
		args = append(args, filter.Currency)
	}

	if filter.MinAmount != nil {
		conditions = append(conditions, "amount >= ?")
		args = append(args, *filter.MinAmount)
	}

	if filter.MaxAmount != nil {
		conditions = append(conditions, "amount <= ?")
		args = append(args, *filter.MaxAmount)
	}

	if filter.StartDate != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *filter.EndDate)
	}

	if filter.OrderID != nil {
		conditions = append(conditions, "order_id = ?")
		args = append(args, *filter.OrderID)
	}

	// Build query
	query := r.GetDB().WithContext(ctx)
	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " AND "), args...)
	}

	// Count total
	if err := query.Model(&entities.PaymentRecord{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records with filter", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to count payment records: %w", err)
	}

	// Get records with pagination
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list payment records with filter", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to list payment records: %w", err)
	}

	return payments, total, nil
}

// UpdatePaymentStatus updates the payment status and related fields
func (r *paymentRecordRepository) UpdatePaymentStatus(ctx context.Context, id uint, status string, transactionID string, paidAt *time.Time) error {
	updateData := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if transactionID != "" {
		updateData["transaction_id"] = transactionID
	}

	if paidAt != nil {
		updateData["paid_at"] = paidAt
	}

	result := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		logger.Error("Failed to update payment status",
			logger.Uint("payment_id", id),
			logger.String("status", status),
			logger.ErrorField(result.Error),
		)
		return fmt.Errorf("failed to update payment status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}

	return nil
}

// GetByPaymentNo retrieves a payment record by payment number
func (r *paymentRecordRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		logger.Error("Failed to get payment record by payment number",
			logger.String("payment_no", paymentNo),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get payment record: %w", err)
	}
	return &payment, nil
}

// GetByOutTradeNo retrieves a payment record by merchant trade number
func (r *paymentRecordRepository) GetByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("out_trade_no = ?", outTradeNo).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		logger.Error("Failed to get payment record by out trade number",
			logger.String("out_trade_no", outTradeNo),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get payment record: %w", err)
	}
	return &payment, nil
}

// GetByTransactionID retrieves a payment record by transaction ID
func (r *paymentRecordRepository) GetByTransactionID(ctx context.Context, transactionID string) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("transaction_id = ?", transactionID).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		logger.Error("Failed to get payment record by transaction ID",
			logger.String("transaction_id", transactionID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get payment record: %w", err)
	}
	return &payment, nil
}

// ExistsByPaymentNo checks if a payment record exists by payment number
func (r *paymentRecordRepository) ExistsByPaymentNo(ctx context.Context, paymentNo string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("payment_no = ?", paymentNo).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment existence by payment_no: %w", err)
	}
	return count > 0, nil
}

// ExistsByOutTradeNo checks if a payment record exists by out trade number
func (r *paymentRecordRepository) ExistsByOutTradeNo(ctx context.Context, outTradeNo string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("out_trade_no = ?", outTradeNo).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment existence by out_trade_no: %w", err)
	}
	return count > 0, nil
}

// HasRecentPayment checks if user has made a similar payment recently
func (r *paymentRecordRepository) HasRecentPayment(ctx context.Context, userID uint, amount float64, currency string, within time.Duration) (bool, error) {
	since := time.Now().Add(-within)
	var count int64

	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("user_id = ? AND amount = ? AND currency = ? AND created_at >= ?", userID, amount, currency, since).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check recent payment: %w", err)
	}

	return count > 0, nil
}

// === PaymentConfigRepository Implementation ===

// GetByMethod retrieves a payment config by method
func (r *paymentConfigRepository) GetByMethod(ctx context.Context, method string) (*entities.PaymentConfig, error) {
	var config entities.PaymentConfig
	if err := r.GetDB().WithContext(ctx).Where("method = ?", method).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment config not found")
		}
		logger.Error("Failed to get payment config by method",
			logger.String("method", method),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get payment config: %w", err)
	}
	return &config, nil
}

// ListWithFilter provides flexible filtering for payment configs
func (r *paymentConfigRepository) ListWithFilter(ctx context.Context, filter interfaces.PaymentConfigFilter, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64

	// Build dynamic query
	conditions := []string{}
	args := []interface{}{}

	if filter.Enabled != nil {
		conditions = append(conditions, "is_enabled = ?")
		args = append(args, *filter.Enabled)
	}

	if filter.Currency != "" {
		conditions = append(conditions, "(supported_currencies = ? OR supported_currencies = 'ALL' OR supported_currencies = '*')")
		args = append(args, filter.Currency)
	}

	if filter.Method != "" {
		conditions = append(conditions, "method = ?")
		args = append(args, filter.Method)
	}

	if filter.Gateway != "" {
		conditions = append(conditions, "gateway = ?")
		args = append(args, filter.Gateway)
	}

	if filter.MinAmount != nil {
		conditions = append(conditions, "min_amount >= ?")
		args = append(args, *filter.MinAmount)
	}

	if filter.MaxAmount != nil {
		conditions = append(conditions, "max_amount <= ?")
		args = append(args, *filter.MaxAmount)
	}

	if filter.SearchQuery != "" {
		conditions = append(conditions, "(method LIKE ? OR display_name LIKE ? OR description LIKE ?)")
		likeQuery := "%" + filter.SearchQuery + "%"
		args = append(args, likeQuery, likeQuery, likeQuery)
	}

	// Build query
	query := r.GetDB().WithContext(ctx)
	if len(conditions) > 0 {
		query = query.Where(strings.Join(conditions, " AND "), args...)
	}

	// Count total
	if err := query.Model(&entities.PaymentConfig{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment configs with filter", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to count payment configs: %w", err)
	}

	// Get configs with pagination
	if err := query.Order("sort_order ASC, created_at ASC").Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		logger.Error("Failed to list payment configs with filter", logger.ErrorField(err))
		return nil, 0, fmt.Errorf("failed to list payment configs: %w", err)
	}

	return configs, total, nil
}

// GetEnabledConfigs gets all enabled payment configs for a specific currency
func (r *paymentConfigRepository) GetEnabledConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error) {
	var configs []*entities.PaymentConfig
	query := r.GetDB().WithContext(ctx).Where("is_enabled = ?", true)

	if currency != "" {
		query = query.Where("(supported_currencies = ? OR supported_currencies = 'ALL' OR supported_currencies = '*')", currency)
	}

	if err := query.Order("sort_order ASC").Find(&configs).Error; err != nil {
		logger.Error("Failed to get enabled payment configs",
			logger.String("currency", currency),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get enabled payment configs: %w", err)
	}

	return configs, nil
}

// ExistsByMethod checks if a payment config with the given method exists
func (r *paymentConfigRepository) ExistsByMethod(ctx context.Context, method string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentConfig{}).Where("method = ?", method).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment config existence by method: %w", err)
	}
	return count > 0, nil
}
