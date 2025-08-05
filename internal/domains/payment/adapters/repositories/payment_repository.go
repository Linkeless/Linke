package repositories

import (
	"context"
	"fmt"
	"time"

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

// paymentConfigRepository implements the PaymentConfigRepository interface
type paymentConfigRepository struct {
	db *gorm.DB
}

// NewPaymentRecordRepository creates a new PaymentRecordRepository implementation
func NewPaymentRecordRepository(db *gorm.DB, logger framework.Logger) interfaces.PaymentRecordRepository {
	return &paymentRecordRepository{
		UserScopedTimeBasedRepositoryImpl: repository.NewUserScopedTimeBasedRepository[entities.PaymentRecord, uint](db, logger),
	}
}

// NewPaymentConfigRepository creates a new PaymentConfigRepository implementation
func NewPaymentConfigRepository(db *gorm.DB) interfaces.PaymentConfigRepository {
	return &paymentConfigRepository{
		db: db,
	}
}

// === PaymentRecordRepository Implementation ===

// Payment-specific methods that extend the base repository functionality

// GetByPaymentNo retrieves a payment record by payment number
func (r *paymentRecordRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		logger.Error("Failed to get payment record by payment number",
			logger.String("payment_no", paymentNo),
			logger.Error2("error", err),
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
			logger.Error2("error", err),
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
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get payment record: %w", err)
	}
	return &payment, nil
}

// GetUserPaymentHistory retrieves user's payment history
func (r *paymentRecordRepository) GetUserPaymentHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	return r.ListByUser(ctx, userID, limit, offset)
}

// GetUserCompletedPayments retrieves user's completed payments
func (r *paymentRecordRepository) GetUserCompletedPayments(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	condition := "user_id = ? AND status = ?"
	args := []interface{}{userID, entities.PaymentRecordStatusCompleted}

	// Count total completed payments for user
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count completed payment records by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count completed payment records by user: %w", err)
	}

	// Get completed payments with pagination
	if err := r.GetDB().WithContext(ctx).Where(condition, args...).
		Order("paid_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list completed payment records by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list completed payment records by user: %w", err)
	}

	return payments, total, nil
}

// GetUserTotalPaid calculates total amount paid by user in a specific currency
func (r *paymentRecordRepository) GetUserTotalPaid(ctx context.Context, userID uint, currency string) (float64, error) {
	var total float64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("user_id = ? AND currency = ? AND status = ?", userID, currency, entities.PaymentRecordStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		logger.Error("Failed to get user total paid",
			logger.Uint("user_id", userID),
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to get user total paid: %w", err)
	}
	return total, nil
}

// ListPendingPayments lists pending payment records
func (r *paymentRecordRepository) ListPendingPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	return r.ListByStatus(ctx, entities.PaymentRecordStatusPending, limit, offset)
}

// ListCompletedPayments lists completed payment records
func (r *paymentRecordRepository) ListCompletedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	return r.ListByStatus(ctx, entities.PaymentRecordStatusCompleted, limit, offset)
}

// ListFailedPayments lists failed payment records
func (r *paymentRecordRepository) ListFailedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	return r.ListByStatus(ctx, entities.PaymentRecordStatusFailed, limit, offset)
}

// ListExpiredPayments lists expired payment records
func (r *paymentRecordRepository) ListExpiredPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	condition := "expired_at IS NOT NULL AND expired_at < ?"
	args := []interface{}{time.Now()}

	// Count total expired payments
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count expired payment records", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count expired payment records: %w", err)
	}

	// Get expired payments with pagination
	if err := r.GetDB().WithContext(ctx).Where(condition, args...).
		Order("expired_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list expired payment records", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list expired payment records: %w", err)
	}

	return payments, total, nil
}

// ListByGateway lists payment records by gateway
func (r *paymentRecordRepository) ListByGateway(ctx context.Context, gateway string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Count total payments by gateway
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("gateway = ?", gateway).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by gateway",
			logger.String("gateway", gateway),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by gateway: %w", err)
	}

	// Get payments with pagination
	if err := r.GetDB().WithContext(ctx).Where("gateway = ?", gateway).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list payment records by gateway",
			logger.String("gateway", gateway),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list payment records by gateway: %w", err)
	}

	return payments, total, nil
}

// ListByPaymentMethod lists payment records by payment method
func (r *paymentRecordRepository) ListByPaymentMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Count total payments by method
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("payment_method = ?", paymentMethod).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by payment method",
			logger.String("payment_method", paymentMethod),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by payment method: %w", err)
	}

	// Get payments with pagination
	if err := r.GetDB().WithContext(ctx).Where("payment_method = ?", paymentMethod).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list payment records by payment method",
			logger.String("payment_method", paymentMethod),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list payment records by payment method: %w", err)
	}

	return payments, total, nil
}

// ListByGatewayAndMethod lists payment records by gateway and payment method
func (r *paymentRecordRepository) ListByGatewayAndMethod(ctx context.Context, gateway, paymentMethod string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	condition := "gateway = ? AND payment_method = ?"
	args := []interface{}{gateway, paymentMethod}

	// Count total payments by gateway and method
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by gateway and method",
			logger.String("gateway", gateway),
			logger.String("payment_method", paymentMethod),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by gateway and method: %w", err)
	}

	// Get payments with pagination
	if err := r.GetDB().WithContext(ctx).Where(condition, args...).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list payment records by gateway and method",
			logger.String("gateway", gateway),
			logger.String("payment_method", paymentMethod),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list payment records by gateway and method: %w", err)
	}

	return payments, total, nil
}

// ListRecentPayments lists recent payment records since a specific time
func (r *paymentRecordRepository) ListRecentPayments(ctx context.Context, since time.Time, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Count total recent payments
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("created_at >= ?", since).Count(&total).Error; err != nil {
		logger.Error("Failed to count recent payment records",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count recent payment records: %w", err)
	}

	// Get recent payments with pagination
	if err := r.GetDB().WithContext(ctx).Where("created_at >= ?", since).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list recent payment records",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list recent payment records: %w", err)
	}

	return payments, total, nil
}

// MarkAsCompleted marks a payment as completed
func (r *paymentRecordRepository) MarkAsCompleted(ctx context.Context, id uint, transactionID string, paidAt time.Time) error {
	updates := map[string]interface{}{
		"status":         entities.PaymentRecordStatusCompleted,
		"transaction_id": transactionID,
		"paid_at":        paidAt,
		"updated_at":     time.Now(),
	}

	result := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		logger.Error("Failed to mark payment as completed",
			logger.Uint("payment_id", id),
			logger.String("transaction_id", transactionID),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to mark payment as completed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}

	logger.Info("Payment marked as completed",
		logger.Uint("payment_id", id),
		logger.String("transaction_id", transactionID),
	)
	return nil
}

// MarkAsFailed marks a payment as failed
func (r *paymentRecordRepository) MarkAsFailed(ctx context.Context, id uint, reason string) error {
	updates := map[string]interface{}{
		"status":         entities.PaymentRecordStatusFailed,
		"payment_status": reason,
		"updated_at":     time.Now(),
	}

	result := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		logger.Error("Failed to mark payment as failed",
			logger.Uint("payment_id", id),
			logger.String("reason", reason),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to mark payment as failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}

	logger.Info("Payment marked as failed",
		logger.Uint("payment_id", id),
		logger.String("reason", reason),
	)
	return nil
}

// GetTotalRevenue calculates total revenue for a currency since a specific time
func (r *paymentRecordRepository) GetTotalRevenue(ctx context.Context, currency string, since time.Time) (float64, error) {
	var total float64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("currency = ? AND status = ? AND paid_at >= ?", currency, entities.PaymentRecordStatusCompleted, since).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		logger.Error("Failed to get total revenue",
			logger.String("currency", currency),
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to get total revenue: %w", err)
	}
	return total, nil
}

// ExistsByPaymentNo checks if a payment record with the given payment number exists
func (r *paymentRecordRepository) ExistsByPaymentNo(ctx context.Context, paymentNo string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("payment_no = ?", paymentNo).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment record existence by payment number: %w", err)
	}
	return count > 0, nil
}

// ExistsByOutTradeNo checks if a payment record with the given out trade number exists
func (r *paymentRecordRepository) ExistsByOutTradeNo(ctx context.Context, outTradeNo string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("out_trade_no = ?", outTradeNo).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment record existence by out trade number: %w", err)
	}
	return count > 0, nil
}

// ExistsByTransactionID checks if a payment record with the given transaction ID exists
func (r *paymentRecordRepository) ExistsByTransactionID(ctx context.Context, transactionID string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("transaction_id = ?", transactionID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment record existence by transaction ID: %w", err)
	}
	return count > 0, nil
}

// ListBySubscriptionOrder lists payment records by subscription order
func (r *paymentRecordRepository) ListBySubscriptionOrder(ctx context.Context, orderID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("subscription_order_id = ?", orderID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by subscription order: %w", err)
	}
	if err := r.GetDB().WithContext(ctx).Where("subscription_order_id = ?", orderID).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments by subscription order: %w", err)
	}
	return payments, total, nil
}

// GetOrderPayments gets all payment records for an order
func (r *paymentRecordRepository) GetOrderPayments(ctx context.Context, orderID uint) ([]*entities.PaymentRecord, error) {
	var payments []*entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("subscription_order_id = ?", orderID).Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get order payments: %w", err)
	}
	return payments, nil
}

// GetOrderCompletedPayments gets completed payment records for an order
func (r *paymentRecordRepository) GetOrderCompletedPayments(ctx context.Context, orderID uint) ([]*entities.PaymentRecord, error) {
	var payments []*entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("subscription_order_id = ? AND status = ?", orderID, entities.PaymentRecordStatusCompleted).Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get order completed payments: %w", err)
	}
	return payments, nil
}

// ListRefundedPayments lists refunded payment records
func (r *paymentRecordRepository) ListRefundedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	condition := "refund_amount > 0"
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where(condition).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count refunded payments: %w", err)
	}
	if err := r.GetDB().WithContext(ctx).Where(condition).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list refunded payments: %w", err)
	}
	return payments, total, nil
}

// ListByRefundStatus lists payment records by refund status
func (r *paymentRecordRepository) ListByRefundStatus(ctx context.Context, refundStatus string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("refund_status = ?", refundStatus).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by refund status: %w", err)
	}
	if err := r.GetDB().WithContext(ctx).Where("refund_status = ?", refundStatus).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments by refund status: %w", err)
	}
	return payments, total, nil
}

// ListRefundablePayments lists refundable payment records
func (r *paymentRecordRepository) ListRefundablePayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	condition := "status = ? AND refund_amount = 0"
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where(condition, entities.PaymentRecordStatusCompleted).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count refundable payments: %w", err)
	}
	if err := r.GetDB().WithContext(ctx).Where(condition, entities.PaymentRecordStatusCompleted).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list refundable payments: %w", err)
	}
	return payments, total, nil
}

// ListByCurrency lists payment records by currency
func (r *paymentRecordRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("currency = ?", currency).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by currency: %w", err)
	}
	if err := r.GetDB().WithContext(ctx).Where("currency = ?", currency).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments by currency: %w", err)
	}
	return payments, total, nil
}

// GetSupportedCurrencies gets list of supported currencies
func (r *paymentRecordRepository) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	var currencies []string
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Distinct("currency").Pluck("currency", &currencies).Error; err != nil {
		return nil, fmt.Errorf("failed to get supported currencies: %w", err)
	}
	return currencies, nil
}

// UpdatePaymentStatus updates payment status
func (r *paymentRecordRepository) UpdatePaymentStatus(ctx context.Context, id uint, paymentStatus string) error {
	result := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Update("payment_status", paymentStatus)
	if result.Error != nil {
		return fmt.Errorf("failed to update payment status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}
	return nil
}

// MarkAsRefunded marks a payment as refunded
func (r *paymentRecordRepository) MarkAsRefunded(ctx context.Context, id uint, refundAmount float64, refundReason string, refundedAt time.Time) error {
	updates := map[string]interface{}{
		"refund_amount": refundAmount,
		"refund_reason": refundReason,
		"refunded_at":   refundedAt,
		"updated_at":    time.Now(),
	}
	result := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to mark payment as refunded: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}
	return nil
}

// UpdateNotification updates notification info
func (r *paymentRecordRepository) UpdateNotification(ctx context.Context, id uint, notifiedAt time.Time, notifyHash string) error {
	updates := map[string]interface{}{
		"notified_at": notifiedAt,
		"notify_hash": notifyHash,
		"updated_at":  time.Now(),
	}
	result := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification: %w", result.Error)
	}
	return nil
}

// IncrementNotifyCount increments notification count
func (r *paymentRecordRepository) IncrementNotifyCount(ctx context.Context, id uint) error {
	result := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).UpdateColumn("notify_count", "notify_count + 1")
	if result.Error != nil {
		return fmt.Errorf("failed to increment notify count: %w", result.Error)
	}
	return nil
}

// ListExpiringPayments lists payments expiring before specified time
func (r *paymentRecordRepository) ListExpiringPayments(ctx context.Context, beforeTime time.Time, limit int) ([]*entities.PaymentRecord, error) {
	var payments []*entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("expired_at IS NOT NULL AND expired_at < ? AND status = ?", beforeTime, entities.PaymentRecordStatusPending).Limit(limit).Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to list expiring payments: %w", err)
	}
	return payments, nil
}

// MarkExpiredPayments marks payments as expired
func (r *paymentRecordRepository) MarkExpiredPayments(ctx context.Context, beforeTime time.Time) (int64, error) {
	result := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("expired_at IS NOT NULL AND expired_at < ? AND status = ?", beforeTime, entities.PaymentRecordStatusPending).Update("status", entities.PaymentRecordStatusFailed)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to mark expired payments: %w", result.Error)
	}
	return result.RowsAffected, nil
}

// SearchByUserEmail searches payment records by user email
func (r *paymentRecordRepository) SearchByUserEmail(ctx context.Context, email string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	// TODO: This requires joining with user table - implement based on actual user table structure
	return nil, 0, fmt.Errorf("SearchByUserEmail not implemented - requires user table join")
}

// GetRevenueByGateway gets revenue by gateway
func (r *paymentRecordRepository) GetRevenueByGateway(ctx context.Context, gateway, currency string, since time.Time) (float64, error) {
	var total float64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("gateway = ? AND currency = ? AND status = ? AND paid_at >= ?", gateway, currency, entities.PaymentRecordStatusCompleted, since).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to get revenue by gateway: %w", err)
	}
	return total, nil
}

// GetRevenueByMethod gets revenue by payment method
func (r *paymentRecordRepository) GetRevenueByMethod(ctx context.Context, paymentMethod, currency string, since time.Time) (float64, error) {
	var total float64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("payment_method = ? AND currency = ? AND status = ? AND paid_at >= ?", paymentMethod, currency, entities.PaymentRecordStatusCompleted, since).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to get revenue by method: %w", err)
	}
	return total, nil
}

// GetDailyRevenue gets daily revenue data
func (r *paymentRecordRepository) GetDailyRevenue(ctx context.Context, currency string, days int) (map[string]float64, error) {
	// TODO: Implement daily revenue aggregation
	return make(map[string]float64), fmt.Errorf("GetDailyRevenue not implemented")
}

// GetMonthlyRevenue gets monthly revenue data
func (r *paymentRecordRepository) GetMonthlyRevenue(ctx context.Context, currency string, months int) (map[string]float64, error) {
	// TODO: Implement monthly revenue aggregation
	return make(map[string]float64), fmt.Errorf("GetMonthlyRevenue not implemented")
}

// ListByAmountRange lists payments by amount range
func (r *paymentRecordRepository) ListByAmountRange(ctx context.Context, minAmount, maxAmount float64, currency string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	condition := "currency = ? AND amount BETWEEN ? AND ?"
	args := []interface{}{currency, minAmount, maxAmount}
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where(condition, args...).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by amount range: %w", err)
	}
	if err := r.GetDB().WithContext(ctx).Where(condition, args...).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments by amount range: %w", err)
	}
	return payments, total, nil
}

// GetLargestPayment gets the largest payment record
func (r *paymentRecordRepository) GetLargestPayment(ctx context.Context, currency string, since time.Time) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("currency = ? AND status = ? AND paid_at >= ?", currency, entities.PaymentRecordStatusCompleted, since).Order("amount DESC").First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no payment found")
		}
		return nil, fmt.Errorf("failed to get largest payment: %w", err)
	}
	return &payment, nil
}

// GetLastPaymentNumber gets the last payment number with prefix
func (r *paymentRecordRepository) GetLastPaymentNumber(ctx context.Context, prefix string) (string, error) {
	var paymentNo string
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("payment_no LIKE ?", prefix+"%").Order("payment_no DESC").Limit(1).Pluck("payment_no", &paymentNo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", fmt.Errorf("failed to get last payment number: %w", err)
	}
	return paymentNo, nil
}

// GetPendingPaymentByUserAndAmount gets pending payment by user and amount
func (r *paymentRecordRepository) GetPendingPaymentByUserAndAmount(ctx context.Context, userID uint, amount float64, currency string) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("user_id = ? AND amount = ? AND currency = ? AND status = ?", userID, amount, currency, entities.PaymentRecordStatusPending).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get pending payment: %w", err)
	}
	return &payment, nil
}

// HasRecentPayment checks if user has recent payment
func (r *paymentRecordRepository) HasRecentPayment(ctx context.Context, userID uint, amount float64, currency string, within time.Duration) (bool, error) {
	since := time.Now().Add(-within)
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("user_id = ? AND amount = ? AND currency = ? AND created_at >= ?", userID, amount, currency, since).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check recent payment: %w", err)
	}
	return count > 0, nil
}

// CountByGateway counts payment records by gateway
func (r *paymentRecordRepository) CountByGateway(ctx context.Context, gateway string) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("gateway = ?", gateway).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count payment records by gateway: %w", err)
	}
	return count, nil
}

// CountCompletedPayments counts completed payment records since a specific time
func (r *paymentRecordRepository) CountCompletedPayments(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("status = ? AND paid_at >= ?", entities.PaymentRecordStatusCompleted, since).Count(&count).Error; err != nil {
		logger.Error("Failed to count completed payment records",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count completed payment records: %w", err)
	}
	return count, nil
}

// CountFailedPayments counts failed payment records since a specific time
func (r *paymentRecordRepository) CountFailedPayments(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).Where("status = ? AND created_at >= ?", entities.PaymentRecordStatusFailed, since).Count(&count).Error; err != nil {
		logger.Error("Failed to count failed payment records",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count failed payment records: %w", err)
	}
	return count, nil
}

// GetAveragePaymentAmount calculates average payment amount for a currency since a specific time
func (r *paymentRecordRepository) GetAveragePaymentAmount(ctx context.Context, currency string, since time.Time) (float64, error) {
	var avg float64
	if err := r.GetDB().WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("currency = ? AND status = ? AND paid_at >= ?", currency, entities.PaymentRecordStatusCompleted, since).
		Select("COALESCE(AVG(amount), 0)").Scan(&avg).Error; err != nil {
		logger.Error("Failed to get average payment amount",
			logger.String("currency", currency),
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to get average payment amount: %w", err)
	}
	return avg, nil
}

// GetByNotifyHash retrieves payment record by notify hash
func (r *paymentRecordRepository) GetByNotifyHash(ctx context.Context, notifyHash string) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.GetDB().WithContext(ctx).Where("notify_hash = ?", notifyHash).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		return nil, fmt.Errorf("failed to get payment record by notify hash: %w", err)
	}
	return &payment, nil
}

// ListPaymentsByPeriod lists payment records by period
func (r *paymentRecordRepository) ListPaymentsByPeriod(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	// Use the inherited method from base class
	return r.ListByDateRange(ctx, field, start, end, limit, offset)
}

// === PaymentConfigRepository Implementation ===

// Create creates a new payment config
func (r *paymentConfigRepository) Create(ctx context.Context, config *entities.PaymentConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		logger.Error("Failed to create payment config",
			logger.String("gateway", config.Gateway),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to create payment config: %w", err)
	}

	logger.Debug("Payment config created successfully",
		logger.Uint("config_id", config.ID),
		logger.String("gateway", config.Gateway),
	)
	return nil
}

// GetByID retrieves a payment config by ID
func (r *paymentConfigRepository) GetByID(ctx context.Context, id uint) (*entities.PaymentConfig, error) {
	var config entities.PaymentConfig
	if err := r.db.WithContext(ctx).First(&config, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment config not found")
		}
		logger.Error("Failed to get payment config by ID",
			logger.Uint("config_id", id),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get payment config: %w", err)
	}
	return &config, nil
}

// GetByGateway retrieves a payment config by gateway
func (r *paymentConfigRepository) GetByGateway(ctx context.Context, gateway string) (*entities.PaymentConfig, error) {
	var config entities.PaymentConfig
	if err := r.db.WithContext(ctx).Where("gateway = ?", gateway).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment config not found")
		}
		logger.Error("Failed to get payment config by gateway",
			logger.String("gateway", gateway),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get payment config: %w", err)
	}
	return &config, nil
}

// ListActive lists active payment configs
func (r *paymentConfigRepository) ListActive(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64

	// Count total active configs
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).
		Where("is_enabled = ?", true).Count(&total).Error; err != nil {
		logger.Error("Failed to count active payment configs", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count active payment configs: %w", err)
	}

	// Get active configs with pagination
	if err := r.db.WithContext(ctx).Where("is_enabled = ?", true).
		Order("sort_order ASC, created_at ASC").
		Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		logger.Error("Failed to list active payment configs", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list active payment configs: %w", err)
	}

	return configs, total, nil
}

// GetEnabledByCurrency gets enabled payment configs supporting a specific currency
func (r *paymentConfigRepository) GetEnabledByCurrency(ctx context.Context, currency string) ([]*entities.PaymentConfig, error) {
	var configs []*entities.PaymentConfig
	if err := r.db.WithContext(ctx).
		Where("is_enabled = ? AND (supported_currencies = ? OR supported_currencies = 'ALL' OR supported_currencies = '*')",
			true, currency).
		Order("sort_order ASC").Find(&configs).Error; err != nil {
		logger.Error("Failed to get enabled payment configs by currency",
			logger.String("currency", currency),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get enabled payment configs by currency: %w", err)
	}
	return configs, nil
}

// GetAvailableForPayment gets payment configs available for a specific currency and amount
func (r *paymentConfigRepository) GetAvailableForPayment(ctx context.Context, currency string, amount float64) ([]*entities.PaymentConfig, error) {
	var configs []*entities.PaymentConfig
	if err := r.db.WithContext(ctx).
		Where("is_enabled = ? AND (supported_currencies = ? OR supported_currencies = 'ALL' OR supported_currencies = '*') AND min_amount <= ? AND max_amount >= ?",
			true, currency, amount, amount).
		Order("sort_order ASC").Find(&configs).Error; err != nil {
		logger.Error("Failed to get available payment configs",
			logger.String("currency", currency),
			logger.String("amount", fmt.Sprintf("%.2f", amount)),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get available payment configs: %w", err)
	}
	return configs, nil
}

// UpdateStatus updates a payment config's enabled status
func (r *paymentConfigRepository) UpdateStatus(ctx context.Context, id uint, isEnabled bool) error {
	result := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("id = ?", id).Update("is_enabled", isEnabled)
	if result.Error != nil {
		logger.Error("Failed to update payment config status",
			logger.Uint("config_id", id),
			logger.String("is_enabled", fmt.Sprintf("%t", isEnabled)),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update payment config status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment config not found")
	}

	logger.Debug("Payment config status updated successfully",
		logger.Uint("config_id", id),
		logger.String("is_enabled", fmt.Sprintf("%t", isEnabled)),
	)
	return nil
}

// Delete performs soft delete on a payment config
func (r *paymentConfigRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.PaymentConfig{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete payment config",
			logger.Uint("config_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete payment config: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment config not found")
	}

	logger.Debug("Payment config deleted successfully",
		logger.Uint("config_id", id),
	)
	return nil
}

// ExistsByGateway checks if a payment config with the given gateway exists
func (r *paymentConfigRepository) ExistsByGateway(ctx context.Context, gateway string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("gateway = ?", gateway).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment config existence by gateway: %w", err)
	}
	return count > 0, nil
}

// ExistsByID checks if a payment config with the given ID exists
func (r *paymentConfigRepository) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment config existence by ID: %w", err)
	}
	return count > 0, nil
}

// BatchDelete deletes multiple payment configs by IDs
func (r *paymentConfigRepository) BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	var successCount int
	var failedIDs []uint

	for _, id := range ids {
		if err := r.db.WithContext(ctx).Delete(&entities.PaymentConfig{}, id).Error; err != nil {
			logger.Error("Failed to delete payment config in batch",
				logger.Uint("id", id),
				logger.String("error", err.Error()))
			failedIDs = append(failedIDs, id)
		} else {
			successCount++
		}
	}

	return successCount, failedIDs, nil
}

// BatchUpdateSortOrder updates the sort order of multiple payment configs
func (r *paymentConfigRepository) BatchUpdateSortOrder(ctx context.Context, updates map[uint]int) error {
	if len(updates) == 0 {
		return nil
	}

	tx := r.db.WithContext(ctx).Begin()
	for id, sortOrder := range updates {
		if err := tx.Model(&entities.PaymentConfig{}).
			Where("id = ?", id).
			Update("sort_order", sortOrder).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update sort order for config %d: %w", id, err)
		}
	}

	return tx.Commit().Error
}

// Stub implementations for methods not yet implemented
// TODO: Implement these methods or refactor PaymentConfigRepository to use base repository

func (r *paymentConfigRepository) Update(ctx context.Context, config *entities.PaymentConfig) error {
	return fmt.Errorf("Update method not implemented")
}

func (r *paymentConfigRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.Delete(ctx, id)
}

func (r *paymentConfigRepository) Restore(ctx context.Context, id uint) error {
	return fmt.Errorf("Restore method not implemented")
}

func (r *paymentConfigRepository) HardDelete(ctx context.Context, id uint) error {
	return fmt.Errorf("HardDelete method not implemented")
}

func (r *paymentConfigRepository) List(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("List method not implemented")
}

func (r *paymentConfigRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListDeleted method not implemented")
}

func (r *paymentConfigRepository) ListByStatus(ctx context.Context, isEnabled bool, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListByStatus method not implemented")
}

func (r *paymentConfigRepository) ListEnabled(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListEnabled method not implemented")
}

func (r *paymentConfigRepository) ListDisabled(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListDisabled method not implemented")
}

func (r *paymentConfigRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListByCurrency method not implemented")
}

func (r *paymentConfigRepository) ListSupportingCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListSupportingCurrency method not implemented")
}

func (r *paymentConfigRepository) ListByMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListByMethod method not implemented")
}

func (r *paymentConfigRepository) ListSupportingMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListSupportingMethod method not implemented")
}

func (r *paymentConfigRepository) GetEnabledByMethod(ctx context.Context, paymentMethod string) ([]*entities.PaymentConfig, error) {
	return nil, fmt.Errorf("GetEnabledByMethod method not implemented")
}

func (r *paymentConfigRepository) GetEnabledByCurrencyAndMethod(ctx context.Context, currency, paymentMethod string) ([]*entities.PaymentConfig, error) {
	return nil, fmt.Errorf("GetEnabledByCurrencyAndMethod method not implemented")
}

func (r *paymentConfigRepository) ListByAmountRange(ctx context.Context, minAmount, maxAmount float64, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListByAmountRange method not implemented")
}

func (r *paymentConfigRepository) GetConfigsForAmount(ctx context.Context, amount float64, currency string) ([]*entities.PaymentConfig, error) {
	return r.GetAvailableForPayment(ctx, currency, amount)
}

func (r *paymentConfigRepository) ListByFeeRange(ctx context.Context, minFee, maxFee float64, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListByFeeRange method not implemented")
}

func (r *paymentConfigRepository) GetLowestFeeConfig(ctx context.Context, currency string, amount float64) (*entities.PaymentConfig, error) {
	return nil, fmt.Errorf("GetLowestFeeConfig method not implemented")
}

func (r *paymentConfigRepository) UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error {
	return fmt.Errorf("UpdateSortOrder method not implemented")
}

func (r *paymentConfigRepository) UpdateConfig(ctx context.Context, id uint, config string) error {
	return fmt.Errorf("UpdateConfig method not implemented")
}

func (r *paymentConfigRepository) UpdateMethods(ctx context.Context, id uint, methods []entities.Method) error {
	return fmt.Errorf("UpdateMethods method not implemented")
}

func (r *paymentConfigRepository) BatchUpdateStatus(ctx context.Context, ids []uint, isEnabled bool) (int, []uint, error) {
	return 0, nil, fmt.Errorf("BatchUpdateStatus method not implemented")
}

func (r *paymentConfigRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("Search method not implemented")
}

func (r *paymentConfigRepository) CountTotal(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("CountTotal method not implemented")
}

func (r *paymentConfigRepository) CountEnabled(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("CountEnabled method not implemented")
}

func (r *paymentConfigRepository) CountDisabled(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("CountDisabled method not implemented")
}

func (r *paymentConfigRepository) CountByGateway(ctx context.Context, gateway string) (int64, error) {
	return 0, fmt.Errorf("CountByGateway method not implemented")
}

func (r *paymentConfigRepository) ListPublic(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListPublic method not implemented")
}

func (r *paymentConfigRepository) ListPublicByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListPublicByCurrency method not implemented")
}

func (r *paymentConfigRepository) GetPublicEnabledConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error) {
	return r.GetEnabledByCurrency(ctx, currency)
}

func (r *paymentConfigRepository) GetOrderedConfigs(ctx context.Context, orderBy string, ascending bool, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("GetOrderedConfigs method not implemented")
}

func (r *paymentConfigRepository) ListBySortOrder(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListBySortOrder method not implemented")
}

func (r *paymentConfigRepository) ListByGatewayType(ctx context.Context, gatewayType string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListByGatewayType method not implemented")
}

func (r *paymentConfigRepository) GetSupportedGateways(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("GetSupportedGateways method not implemented")
}

func (r *paymentConfigRepository) GetSupportedMethods(ctx context.Context) ([]string, error) {
	return nil, fmt.Errorf("GetSupportedMethods method not implemented")
}

func (r *paymentConfigRepository) ValidateGatewayConfig(ctx context.Context, gateway string, config string) (bool, error) {
	return true, nil
}

func (r *paymentConfigRepository) TestGatewayConnection(ctx context.Context, id uint) (bool, error) {
	return true, nil
}

func (r *paymentConfigRepository) GetMethodByCode(ctx context.Context, configID uint, methodCode string) (*entities.Method, error) {
	return nil, fmt.Errorf("GetMethodByCode method not implemented")
}

func (r *paymentConfigRepository) UpdateMethodStatus(ctx context.Context, configID uint, methodCode string, isEnabled bool) error {
	return fmt.Errorf("UpdateMethodStatus method not implemented")
}

func (r *paymentConfigRepository) GetEnabledMethods(ctx context.Context, configID uint) ([]entities.Method, error) {
	return nil, fmt.Errorf("GetEnabledMethods method not implemented")
}

func (r *paymentConfigRepository) GetCurrencyStats(ctx context.Context) (map[string]int64, error) {
	return make(map[string]int64), fmt.Errorf("GetCurrencyStats method not implemented")
}

func (r *paymentConfigRepository) GetMethodStats(ctx context.Context) (map[string]int64, error) {
	return make(map[string]int64), fmt.Errorf("GetMethodStats method not implemented")
}

func (r *paymentConfigRepository) GetGatewayStats(ctx context.Context) (map[string]int64, error) {
	return make(map[string]int64), fmt.Errorf("GetGatewayStats method not implemented")
}

func (r *paymentConfigRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListWithFilters method not implemented")
}

func (r *paymentConfigRepository) ExportConfigs(ctx context.Context) ([]*entities.PaymentConfig, error) {
	return nil, fmt.Errorf("ExportConfigs method not implemented")
}

func (r *paymentConfigRepository) ImportConfigs(ctx context.Context, configs []*entities.PaymentConfig) error {
	return fmt.Errorf("ImportConfigs method not implemented")
}

func (r *paymentConfigRepository) ListByEnvironment(ctx context.Context, environment string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return nil, 0, fmt.Errorf("ListByEnvironment method not implemented")
}

func (r *paymentConfigRepository) GetProductionConfigs(ctx context.Context) ([]*entities.PaymentConfig, error) {
	return nil, fmt.Errorf("GetProductionConfigs method not implemented")
}

func (r *paymentConfigRepository) GetTestConfigs(ctx context.Context) ([]*entities.PaymentConfig, error) {
	return nil, fmt.Errorf("GetTestConfigs method not implemented")
}

// Note: PaymentConfigRepository has extensive interface requirements.
// This repository should be refactored to inherit from a base repository class in the future.
// The above are stub implementations to satisfy compilation requirements.