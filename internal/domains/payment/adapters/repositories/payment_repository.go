package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

// paymentRecordRepository implements the PaymentRecordRepository interface
type paymentRecordRepository struct {
	db *gorm.DB
}

// paymentConfigRepository implements the PaymentConfigRepository interface
type paymentConfigRepository struct {
	db *gorm.DB
}

// NewPaymentRecordRepository creates a new PaymentRecordRepository implementation
func NewPaymentRecordRepository(db *gorm.DB) interfaces.PaymentRecordRepository {
	return &paymentRecordRepository{
		db: db,
	}
}

// NewPaymentConfigRepository creates a new PaymentConfigRepository implementation
func NewPaymentConfigRepository(db *gorm.DB) interfaces.PaymentConfigRepository {
	return &paymentConfigRepository{
		db: db,
	}
}

// === PaymentRecordRepository Implementation ===

// Create creates a new payment record
func (r *paymentRecordRepository) Create(ctx context.Context, payment *entities.PaymentRecord) error {
	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		logger.Error("Failed to create payment record",
			logger.String("payment_no", payment.PaymentNo),
			logger.String("out_trade_no", payment.OutTradeNo),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to create payment record: %w", err)
	}

	logger.Debug("Payment record created successfully",
		logger.Uint("payment_id", payment.ID),
		logger.String("payment_no", payment.PaymentNo),
	)
	return nil
}

// GetByID retrieves a payment record by ID
func (r *paymentRecordRepository) GetByID(ctx context.Context, id uint) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.db.WithContext(ctx).First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		logger.Error("Failed to get payment record by ID",
			logger.Uint("payment_id", id),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get payment record: %w", err)
	}
	return &payment, nil
}

// GetByPaymentNo retrieves a payment record by payment number
func (r *paymentRecordRepository) GetByPaymentNo(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&payment).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("out_trade_no = ?", outTradeNo).First(&payment).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&payment).Error; err != nil {
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

// Update updates a payment record
func (r *paymentRecordRepository) Update(ctx context.Context, payment *entities.PaymentRecord) error {
	if err := r.db.WithContext(ctx).Save(payment).Error; err != nil {
		logger.Error("Failed to update payment record",
			logger.Uint("payment_id", payment.ID),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to update payment record: %w", err)
	}

	logger.Debug("Payment record updated successfully",
		logger.Uint("payment_id", payment.ID),
	)
	return nil
}

// Delete performs soft delete on a payment record
func (r *paymentRecordRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&entities.PaymentRecord{}, id)
	if result.Error != nil {
		logger.Error("Failed to delete payment record",
			logger.Uint("payment_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to delete payment record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}

	logger.Debug("Payment record deleted successfully",
		logger.Uint("payment_id", id),
	)
	return nil
}

// SoftDelete performs soft delete (alias for Delete)
func (r *paymentRecordRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.Delete(ctx, id)
}

// Restore restores a soft deleted payment record
func (r *paymentRecordRepository) Restore(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&entities.PaymentRecord{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		logger.Error("Failed to restore payment record",
			logger.Uint("payment_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to restore payment record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}

	logger.Debug("Payment record restored successfully",
		logger.Uint("payment_id", id),
	)
	return nil
}

// HardDelete permanently deletes a payment record
func (r *paymentRecordRepository) HardDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entities.PaymentRecord{}, id)
	if result.Error != nil {
		logger.Error("Failed to hard delete payment record",
			logger.Uint("payment_id", id),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to permanently delete payment record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}

	logger.Warn("Payment record permanently deleted",
		logger.Uint("payment_id", id),
	)
	return nil
}

// ListByUser lists payment records by user with pagination
func (r *paymentRecordRepository) ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Count total payments for user
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("user_id = ?", userID).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by user: %w", err)
	}

	// Get payments with pagination
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list payment records by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list payment records by user: %w", err)
	}

	return payments, total, nil
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count completed payment records by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count completed payment records by user: %w", err)
	}

	// Get completed payments with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
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

// ListByStatus lists payment records by status with pagination
func (r *paymentRecordRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Count total payments by status
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("status = ?", status).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by status",
			logger.String("status", status),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by status: %w", err)
	}

	// Get payments with pagination
	if err := r.db.WithContext(ctx).Where("status = ?", status).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list payment records by status",
			logger.String("status", status),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list payment records by status: %w", err)
	}

	return payments, total, nil
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count expired payment records", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count expired payment records: %w", err)
	}

	// Get expired payments with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("gateway = ?", gateway).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by gateway",
			logger.String("gateway", gateway),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by gateway: %w", err)
	}

	// Get payments with pagination
	if err := r.db.WithContext(ctx).Where("gateway = ?", gateway).
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("payment_method = ?", paymentMethod).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by payment method",
			logger.String("payment_method", paymentMethod),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by payment method: %w", err)
	}

	// Get payments with pagination
	if err := r.db.WithContext(ctx).Where("payment_method = ?", paymentMethod).
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by gateway and method",
			logger.String("gateway", gateway),
			logger.String("payment_method", paymentMethod),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by gateway and method: %w", err)
	}

	// Get payments with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
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

// ListByDateRange lists payment records within a date range
func (r *paymentRecordRepository) ListByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	condition := "created_at BETWEEN ? AND ?"
	args := []interface{}{start, end}

	// Count total payments in date range
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where(condition, args...).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records by date range",
			logger.String("start", start.Format(time.RFC3339)),
			logger.String("end", end.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count payment records by date range: %w", err)
	}

	// Get payments with pagination
	if err := r.db.WithContext(ctx).Where(condition, args...).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list payment records by date range",
			logger.String("start", start.Format(time.RFC3339)),
			logger.String("end", end.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list payment records by date range: %w", err)
	}

	return payments, total, nil
}

// ListRecentPayments lists recent payment records since a specific time
func (r *paymentRecordRepository) ListRecentPayments(ctx context.Context, since time.Time, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Count total recent payments
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("created_at >= ?", since).Count(&total).Error; err != nil {
		logger.Error("Failed to count recent payment records",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count recent payment records: %w", err)
	}

	// Get recent payments with pagination
	if err := r.db.WithContext(ctx).Where("created_at >= ?", since).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list recent payment records",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to list recent payment records: %w", err)
	}

	return payments, total, nil
}

// UpdateStatus updates a payment record's status
func (r *paymentRecordRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Update("status", status)
	if result.Error != nil {
		logger.Error("Failed to update payment record status",
			logger.Uint("payment_id", id),
			logger.String("status", status),
			logger.Error2("error", result.Error),
		)
		return fmt.Errorf("failed to update payment record status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("payment record not found")
	}

	logger.Debug("Payment record status updated successfully",
		logger.Uint("payment_id", id),
		logger.String("new_status", status),
	)
	return nil
}

// MarkAsCompleted marks a payment as completed
func (r *paymentRecordRepository) MarkAsCompleted(ctx context.Context, id uint, transactionID string, paidAt time.Time) error {
	updates := map[string]interface{}{
		"status":         entities.PaymentRecordStatusCompleted,
		"transaction_id": transactionID,
		"paid_at":        paidAt,
		"updated_at":     time.Now(),
	}

	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updates)
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

	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updates)
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

// List lists all payment records with pagination
func (r *paymentRecordRepository) List(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Count total payments
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count payment records", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count payment records: %w", err)
	}

	// Get payments with pagination
	if err := r.db.WithContext(ctx).Order("created_at DESC").
		Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to list payment records", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to list payment records: %w", err)
	}

	return payments, total, nil
}

// Search searches payment records by payment number, out trade number, or transaction ID
func (r *paymentRecordRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64

	// Prepare search query
	searchQuery := "%" + strings.ToLower(query) + "%"
	whereClause := "LOWER(payment_no) LIKE ? OR LOWER(out_trade_no) LIKE ? OR LOWER(transaction_id) LIKE ?"

	// Count total matching payments
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where(whereClause, searchQuery, searchQuery, searchQuery).Count(&total).Error; err != nil {
		logger.Error("Failed to count search results for payment records",
			logger.String("query", query),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get matching payments with pagination
	if err := r.db.WithContext(ctx).Where(whereClause, searchQuery, searchQuery, searchQuery).
		Order("created_at DESC").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		logger.Error("Failed to search payment records",
			logger.String("query", query),
			logger.Error2("error", err),
		)
		return nil, 0, fmt.Errorf("failed to search payment records: %w", err)
	}

	return payments, total, nil
}

// CountTotal returns the total count of payment records
func (r *paymentRecordRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count total payment records: %w", err)
	}
	return count, nil
}

// GetTotalRevenue calculates total revenue for a currency since a specific time
func (r *paymentRecordRepository) GetTotalRevenue(ctx context.Context, currency string, since time.Time) (float64, error) {
	var total float64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("payment_no = ?", paymentNo).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment record existence by payment number: %w", err)
	}
	return count > 0, nil
}

// ExistsByOutTradeNo checks if a payment record with the given out trade number exists
func (r *paymentRecordRepository) ExistsByOutTradeNo(ctx context.Context, outTradeNo string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("out_trade_no = ?", outTradeNo).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment record existence by out trade number: %w", err)
	}
	return count > 0, nil
}

// ExistsByTransactionID checks if a payment record with the given transaction ID exists
func (r *paymentRecordRepository) ExistsByTransactionID(ctx context.Context, transactionID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("transaction_id = ?", transactionID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment record existence by transaction ID: %w", err)
	}
	return count > 0, nil
}

// Note: This is a comprehensive implementation of the core PaymentRecordRepository methods.
// For brevity, I'm showing the essential methods. The remaining methods (like revenue statistics,
// subscription order operations, refund operations, etc.) would follow similar patterns
// but with their specific business logic and query conditions.

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

// BatchDelete deletes multiple payment records by IDs
func (r *paymentRecordRepository) BatchDelete(ctx context.Context, ids []uint) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	var successCount int
	var failedIDs []uint

	for _, id := range ids {
		if err := r.db.WithContext(ctx).Delete(&entities.PaymentRecord{}, id).Error; err != nil {
			logger.Error("Failed to delete payment record in batch",
				logger.Uint("id", id),
				logger.String("error", err.Error()))
			failedIDs = append(failedIDs, id)
		} else {
			successCount++
		}
	}

	return successCount, failedIDs, nil
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

// BatchUpdateStatus updates the status of multiple payment records
func (r *paymentRecordRepository) BatchUpdateStatus(ctx context.Context, ids []uint, status string) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("id IN ?", ids).
		Update("status", status)

	if result.Error != nil {
		logger.Error("Failed to batch update payment record status",
			logger.String("status", status),
			logger.String("error", result.Error.Error()))
		return 0, ids, result.Error
	}

	return int(result.RowsAffected), nil, nil
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

// CountByGateway counts payment records by gateway
func (r *paymentRecordRepository) CountByGateway(ctx context.Context, gateway string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("gateway = ?", gateway).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count payment records by gateway: %w", err)
	}
	return count, nil
}

// CountByStatus counts payment records by status
func (r *paymentRecordRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("status = ?", status).Count(&count).Error; err != nil {
		logger.Error("Failed to count payment records by status",
			logger.String("status", status),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count payment records by status: %w", err)
	}
	return count, nil
}

// CountByUser counts payment records by user
func (r *paymentRecordRepository) CountByUser(ctx context.Context, userID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		logger.Error("Failed to count payment records by user",
			logger.Uint("user_id", userID),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count payment records by user: %w", err)
	}
	return count, nil
}

// CountCompletedPayments counts completed payment records since a specific time
func (r *paymentRecordRepository) CountCompletedPayments(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("status = ? AND paid_at >= ?", entities.PaymentRecordStatusCompleted, since).Count(&count).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("status = ? AND created_at >= ?", entities.PaymentRecordStatusFailed, since).Count(&count).Error; err != nil {
		logger.Error("Failed to count failed payment records",
			logger.String("since", since.Format(time.RFC3339)),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count failed payment records: %w", err)
	}
	return count, nil
}

// BatchUpdateStatus updates the status of multiple payment configs
func (r *paymentConfigRepository) BatchUpdateStatus(ctx context.Context, ids []uint, isEnabled bool) (int, []uint, error) {
	if len(ids) == 0 {
		return 0, nil, nil
	}

	result := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).
		Where("id IN ?", ids).
		Update("is_enabled", isEnabled)

	if result.Error != nil {
		logger.Error("Failed to batch update payment config status",
			logger.String("is_enabled", fmt.Sprintf("%t", isEnabled)),
			logger.String("error", result.Error.Error()))
		return 0, ids, result.Error
	}

	return int(result.RowsAffected), nil, nil
}

// CountByGateway counts payment configs by gateway
func (r *paymentConfigRepository) CountByGateway(ctx context.Context, gateway string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("gateway = ?", gateway).Count(&count).Error; err != nil {
		logger.Error("Failed to count payment configs by gateway",
			logger.String("gateway", gateway),
			logger.Error2("error", err),
		)
		return 0, fmt.Errorf("failed to count payment configs by gateway: %w", err)
	}
	return count, nil
}

// CountTotal counts all payment configs
func (r *paymentConfigRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Count(&count).Error; err != nil {
		logger.Error("Failed to count total payment configs", logger.Error2("error", err))
		return 0, fmt.Errorf("failed to count total payment configs: %w", err)
	}
	return count, nil
}

// CountEnabled counts enabled payment configs
func (r *paymentConfigRepository) CountEnabled(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("is_enabled = ?", true).Count(&count).Error; err != nil {
		logger.Error("Failed to count enabled payment configs", logger.Error2("error", err))
		return 0, fmt.Errorf("failed to count enabled payment configs: %w", err)
	}
	return count, nil
}

// CountDisabled counts disabled payment configs
func (r *paymentConfigRepository) CountDisabled(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("is_enabled = ?", false).Count(&count).Error; err != nil {
		logger.Error("Failed to count disabled payment configs", logger.Error2("error", err))
		return 0, fmt.Errorf("failed to count disabled payment configs: %w", err)
	}
	return count, nil
}

// GetAveragePaymentAmount calculates average payment amount for a currency since a specific time
func (r *paymentRecordRepository) GetAveragePaymentAmount(ctx context.Context, currency string, since time.Time) (float64, error) {
	var avg float64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
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
	if err := r.db.WithContext(ctx).Where("notify_hash = ?", notifyHash).First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment record not found")
		}
		return nil, fmt.Errorf("failed to get payment record by notify hash: %w", err)
	}
	return &payment, nil
}

// ExportConfigs exports all payment configs
func (r *paymentConfigRepository) ExportConfigs(ctx context.Context) ([]*entities.PaymentConfig, error) {
	var configs []*entities.PaymentConfig
	if err := r.db.WithContext(ctx).Find(&configs).Error; err != nil {
		logger.Error("Failed to export payment configs", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to export payment configs: %w", err)
	}
	return configs, nil
}

// GetConfigsForAmount gets payment configs for a specific amount and currency
func (r *paymentConfigRepository) GetConfigsForAmount(ctx context.Context, amount float64, currency string) ([]*entities.PaymentConfig, error) {
	var configs []*entities.PaymentConfig
	if err := r.db.WithContext(ctx).
		Where("is_enabled = ? AND (supported_currencies = ? OR supported_currencies = 'ALL' OR supported_currencies = '*') AND min_amount <= ? AND max_amount >= ?", 
			true, currency, amount, amount).
		Order("sort_order ASC").Find(&configs).Error; err != nil {
		logger.Error("Failed to get configs for amount",
			logger.String("currency", currency),
			logger.String("amount", fmt.Sprintf("%.2f", amount)),
			logger.Error2("error", err),
		)
		return nil, fmt.Errorf("failed to get configs for amount: %w", err)
	}
	return configs, nil
}

// Note: This is a comprehensive implementation showing the essential methods for both repositories.
// The remaining methods would follow similar patterns with their specific business logic.// Additional missing methods for PaymentRecordRepository

// ListPaymentsByPeriod lists payment records by period
func (r *paymentRecordRepository) ListPaymentsByPeriod(ctx context.Context, field string, start, end time.Time, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	// TODO: Implement based on business requirements
	return nil, 0, fmt.Errorf("ListPaymentsByPeriod not implemented")
}

// ListBySubscriptionOrder lists payment records by subscription order
func (r *paymentRecordRepository) ListBySubscriptionOrder(ctx context.Context, orderID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("subscription_order_id = ?", orderID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by subscription order: %w", err)
	}
	if err := r.db.WithContext(ctx).Where("subscription_order_id = ?", orderID).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments by subscription order: %w", err)
	}
	return payments, total, nil
}

// GetOrderPayments gets all payment records for an order
func (r *paymentRecordRepository) GetOrderPayments(ctx context.Context, orderID uint) ([]*entities.PaymentRecord, error) {
	var payments []*entities.PaymentRecord
	if err := r.db.WithContext(ctx).Where("subscription_order_id = ?", orderID).Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get order payments: %w", err)
	}
	return payments, nil
}

// GetOrderCompletedPayments gets completed payment records for an order
func (r *paymentRecordRepository) GetOrderCompletedPayments(ctx context.Context, orderID uint) ([]*entities.PaymentRecord, error) {
	var payments []*entities.PaymentRecord
	if err := r.db.WithContext(ctx).Where("subscription_order_id = ? AND status = ?", orderID, entities.PaymentRecordStatusCompleted).Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get order completed payments: %w", err)
	}
	return payments, nil
}

// ListRefundedPayments lists refunded payment records
func (r *paymentRecordRepository) ListRefundedPayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	condition := "refund_amount > 0"
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where(condition).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count refunded payments: %w", err)
	}
	if err := r.db.WithContext(ctx).Where(condition).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list refunded payments: %w", err)
	}
	return payments, total, nil
}

// ListByRefundStatus lists payment records by refund status
func (r *paymentRecordRepository) ListByRefundStatus(ctx context.Context, refundStatus string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("refund_status = ?", refundStatus).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by refund status: %w", err)
	}
	if err := r.db.WithContext(ctx).Where("refund_status = ?", refundStatus).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments by refund status: %w", err)
	}
	return payments, total, nil
}

// ListRefundablePayments lists refundable payment records
func (r *paymentRecordRepository) ListRefundablePayments(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	condition := "status = ? AND refund_amount = 0"
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where(condition, entities.PaymentRecordStatusCompleted).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count refundable payments: %w", err)
	}
	if err := r.db.WithContext(ctx).Where(condition, entities.PaymentRecordStatusCompleted).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list refundable payments: %w", err)
	}
	return payments, total, nil
}

// ListByCurrency lists payment records by currency
func (r *paymentRecordRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord  
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("currency = ?", currency).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by currency: %w", err)
	}
	if err := r.db.WithContext(ctx).Where("currency = ?", currency).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments by currency: %w", err)
	}
	return payments, total, nil
}

// GetSupportedCurrencies gets list of supported currencies
func (r *paymentRecordRepository) GetSupportedCurrencies(ctx context.Context) ([]string, error) {
	var currencies []string
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Distinct("currency").Pluck("currency", &currencies).Error; err != nil {
		return nil, fmt.Errorf("failed to get supported currencies: %w", err)
	}
	return currencies, nil
}

// UpdatePaymentStatus updates payment status
func (r *paymentRecordRepository) UpdatePaymentStatus(ctx context.Context, id uint, paymentStatus string) error {
	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Update("payment_status", paymentStatus)
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
	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updates)
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
	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update notification: %w", result.Error)
	}
	return nil
}

// IncrementNotifyCount increments notification count
func (r *paymentRecordRepository) IncrementNotifyCount(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("id = ?", id).UpdateColumn("notify_count", "notify_count + 1")
	if result.Error != nil {
		return fmt.Errorf("failed to increment notify count: %w", result.Error)
	}
	return nil
}

// ListExpiringPayments lists payments expiring before specified time
func (r *paymentRecordRepository) ListExpiringPayments(ctx context.Context, beforeTime time.Time, limit int) ([]*entities.PaymentRecord, error) {
	var payments []*entities.PaymentRecord
	if err := r.db.WithContext(ctx).Where("expired_at IS NOT NULL AND expired_at < ? AND status = ?", beforeTime, entities.PaymentRecordStatusPending).Limit(limit).Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to list expiring payments: %w", err)
	}
	return payments, nil
}

// MarkExpiredPayments marks payments as expired
func (r *paymentRecordRepository) MarkExpiredPayments(ctx context.Context, beforeTime time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("expired_at IS NOT NULL AND expired_at < ? AND status = ?", beforeTime, entities.PaymentRecordStatusPending).Update("status", entities.PaymentRecordStatusFailed)
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
		Where("gateway = ? AND currency = ? AND status = ? AND paid_at >= ?", gateway, currency, entities.PaymentRecordStatusCompleted, since).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("failed to get revenue by gateway: %w", err)
	}
	return total, nil
}

// GetRevenueByMethod gets revenue by payment method
func (r *paymentRecordRepository) GetRevenueByMethod(ctx context.Context, paymentMethod, currency string, since time.Time) (float64, error) {
	var total float64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where(condition, args...).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments by amount range: %w", err)
	}
	if err := r.db.WithContext(ctx).Where(condition, args...).Limit(limit).Offset(offset).Find(&payments).Error; err != nil {  
		return nil, 0, fmt.Errorf("failed to list payments by amount range: %w", err)
	}
	return payments, total, nil
}

// GetLargestPayment gets the largest payment record
func (r *paymentRecordRepository) GetLargestPayment(ctx context.Context, currency string, since time.Time) (*entities.PaymentRecord, error) {
	var payment entities.PaymentRecord
	if err := r.db.WithContext(ctx).Where("currency = ? AND status = ? AND paid_at >= ?", currency, entities.PaymentRecordStatusCompleted, since).Order("amount DESC").First(&payment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no payment found")
		}
		return nil, fmt.Errorf("failed to get largest payment: %w", err)
	}
	return &payment, nil
}

// ListDeleted lists soft deleted payment records
func (r *paymentRecordRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&entities.PaymentRecord{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted payments: %w", err)
	}
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list deleted payments: %w", err)
	}
	return payments, total, nil
}

// ListWithFilters lists payment records with dynamic filters
func (r *paymentRecordRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.PaymentRecord, int64, error) {
	var payments []*entities.PaymentRecord
	var total int64
	db := r.db.WithContext(ctx).Model(&entities.PaymentRecord{})
	for key, value := range filters {
		db = db.Where(key+" = ?", value)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments with filters: %w", err)
	}
	if err := db.Limit(limit).Offset(offset).Find(&payments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payments with filters: %w", err)
	}
	return payments, total, nil
}

// GetLastPaymentNumber gets the last payment number with prefix
func (r *paymentRecordRepository) GetLastPaymentNumber(ctx context.Context, prefix string) (string, error) {
	var paymentNo string
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("payment_no LIKE ?", prefix+"%").Order("payment_no DESC").Limit(1).Pluck("payment_no", &paymentNo).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("user_id = ? AND amount = ? AND currency = ? AND status = ?", userID, amount, currency, entities.PaymentRecordStatusPending).First(&payment).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Model(&entities.PaymentRecord{}).Where("user_id = ? AND amount = ? AND currency = ? AND created_at >= ?", userID, amount, currency, since).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check recent payment: %w", err)
	}
	return count > 0, nil
}// Additional missing methods for PaymentConfigRepository

// Update updates a payment config
func (r *paymentConfigRepository) Update(ctx context.Context, config *entities.PaymentConfig) error {
	if err := r.db.WithContext(ctx).Save(config).Error; err != nil {
		logger.Error("Failed to update payment config",
			logger.Uint("config_id", config.ID),
			logger.Error2("error", err),
		)
		return fmt.Errorf("failed to update payment config: %w", err)
	}
	logger.Debug("Payment config updated successfully", logger.Uint("config_id", config.ID))
	return nil
}

// SoftDelete performs soft delete on a payment config
func (r *paymentConfigRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.Delete(ctx, id)
}

// Restore restores a soft deleted payment config
func (r *paymentConfigRepository) Restore(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Model(&entities.PaymentConfig{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("failed to restore payment config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment config not found")
	}
	return nil
}

// HardDelete permanently deletes a payment config
func (r *paymentConfigRepository) HardDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&entities.PaymentConfig{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to permanently delete payment config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment config not found")
	}
	return nil
}

// List lists all payment configs with pagination
func (r *paymentConfigRepository) List(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payment configs: %w", err)
	}
	if err := r.db.WithContext(ctx).Order("sort_order ASC, created_at ASC").Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payment configs: %w", err)
	}
	return configs, total, nil
}

// ListDeleted lists soft deleted payment configs
func (r *paymentConfigRepository) ListDeleted(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&entities.PaymentConfig{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted payment configs: %w", err)
	}
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list deleted payment configs: %w", err)
	}
	return configs, total, nil
}

// ListByStatus lists payment configs by status
func (r *paymentConfigRepository) ListByStatus(ctx context.Context, isEnabled bool, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("is_enabled = ?", isEnabled).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payment configs by status: %w", err)
	}
	if err := r.db.WithContext(ctx).Where("is_enabled = ?", isEnabled).Order("sort_order ASC").Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payment configs by status: %w", err)
	}
	return configs, total, nil
}

// ListEnabled lists enabled payment configs
func (r *paymentConfigRepository) ListEnabled(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return r.ListByStatus(ctx, true, limit, offset)
}

// ListDisabled lists disabled payment configs
func (r *paymentConfigRepository) ListDisabled(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return r.ListByStatus(ctx, false, limit, offset)
}

// ListByCurrency lists payment configs by currency
func (r *paymentConfigRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	condition := "supported_currencies = ? OR supported_currencies = 'ALL' OR supported_currencies = '*'"
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where(condition, currency).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payment configs by currency: %w", err)
	}
	if err := r.db.WithContext(ctx).Where(condition, currency).Order("sort_order ASC").Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list payment configs by currency: %w", err)
	}
	return configs, total, nil
}

// ListSupportingCurrency lists payment configs supporting a currency
func (r *paymentConfigRepository) ListSupportingCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return r.ListByCurrency(ctx, currency, limit, offset)
}

// ListByMethod lists payment configs by payment method
func (r *paymentConfigRepository) ListByMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	// TODO: Implement based on how methods are stored in the config
	return nil, 0, fmt.Errorf("ListByMethod not implemented - requires methods field structure")
}

// ListSupportingMethod lists payment configs supporting a method
func (r *paymentConfigRepository) ListSupportingMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return r.ListByMethod(ctx, paymentMethod, limit, offset)
}

// GetEnabledByMethod gets enabled configs supporting a method
func (r *paymentConfigRepository) GetEnabledByMethod(ctx context.Context, paymentMethod string) ([]*entities.PaymentConfig, error) {
	// TODO: Implement based on how methods are stored
	return nil, fmt.Errorf("GetEnabledByMethod not implemented")
}

// GetEnabledByCurrencyAndMethod gets enabled configs by currency and method
func (r *paymentConfigRepository) GetEnabledByCurrencyAndMethod(ctx context.Context, currency, paymentMethod string) ([]*entities.PaymentConfig, error) {
	// TODO: Implement based on methods structure
	return nil, fmt.Errorf("GetEnabledByCurrencyAndMethod not implemented")
}

// ListByAmountRange lists configs by amount range
func (r *paymentConfigRepository) ListByAmountRange(ctx context.Context, minAmount, maxAmount float64, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	condition := "min_amount <= ? AND max_amount >= ?"
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where(condition, maxAmount, minAmount).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count configs by amount range: %w", err)
	}
	if err := r.db.WithContext(ctx).Where(condition, maxAmount, minAmount).Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list configs by amount range: %w", err)
	}
	return configs, total, nil
}

// ListByFeeRange lists configs by fee range
func (r *paymentConfigRepository) ListByFeeRange(ctx context.Context, minFee, maxFee float64, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	// TODO: Implement based on fee structure
	return nil, 0, fmt.Errorf("ListByFeeRange not implemented - requires fee field structure")
}

// GetLowestFeeConfig gets config with lowest fee
func (r *paymentConfigRepository) GetLowestFeeConfig(ctx context.Context, currency string, amount float64) (*entities.PaymentConfig, error) {
	// TODO: Implement based on fee calculation
	return nil, fmt.Errorf("GetLowestFeeConfig not implemented")
}

// UpdateSortOrder updates sort order of a config
func (r *paymentConfigRepository) UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error {
	result := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("id = ?", id).Update("sort_order", sortOrder)
	if result.Error != nil {
		return fmt.Errorf("failed to update sort order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment config not found")
	}
	return nil
}

// UpdateConfig updates config string
func (r *paymentConfigRepository) UpdateConfig(ctx context.Context, id uint, config string) error {
	result := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where("id = ?", id).Update("config", config)
	if result.Error != nil {
		return fmt.Errorf("failed to update config: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("payment config not found")
	}
	return nil
}

// UpdateMethods updates payment methods
func (r *paymentConfigRepository) UpdateMethods(ctx context.Context, id uint, methods []entities.Method) error {
	// TODO: Implement based on methods field structure
	return fmt.Errorf("UpdateMethods not implemented")
}

// Search searches configs by query
func (r *paymentConfigRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	searchQuery := "%" + strings.ToLower(query) + "%"
	whereClause := "LOWER(gateway) LIKE ? OR LOWER(name) LIKE ?"
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where(whereClause, searchQuery, searchQuery).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}
	if err := r.db.WithContext(ctx).Where(whereClause, searchQuery, searchQuery).Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search configs: %w", err)
	}
	return configs, total, nil
}

// ListPublic lists public payment configs
func (r *paymentConfigRepository) ListPublic(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return r.ListEnabled(ctx, limit, offset)
}

// ListPublicByCurrency lists public configs by currency
func (r *paymentConfigRepository) ListPublicByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	condition := "is_enabled = ? AND (supported_currencies = ? OR supported_currencies = 'ALL' OR supported_currencies = '*')"
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Where(condition, true, currency).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count public configs by currency: %w", err)
	}
	if err := r.db.WithContext(ctx).Where(condition, true, currency).Order("sort_order ASC").Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list public configs by currency: %w", err)
	}
	return configs, total, nil
}

// GetPublicEnabledConfigs gets public enabled configs
func (r *paymentConfigRepository) GetPublicEnabledConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error) {
	return r.GetEnabledByCurrency(ctx, currency)
}

// GetOrderedConfigs gets configs ordered by specified field
func (r *paymentConfigRepository) GetOrderedConfigs(ctx context.Context, orderBy string, ascending bool, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	orderClause := orderBy
	if !ascending {
		orderClause += " DESC"
	}
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count configs: %w", err)
	}
	if err := r.db.WithContext(ctx).Order(orderClause).Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get ordered configs: %w", err)
	}
	return configs, total, nil
}

// ListBySortOrder lists configs by sort order
func (r *paymentConfigRepository) ListBySortOrder(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	return r.GetOrderedConfigs(ctx, "sort_order", true, limit, offset)
}

// ListByGatewayType lists configs by gateway type
func (r *paymentConfigRepository) ListByGatewayType(ctx context.Context, gatewayType string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	// TODO: Implement based on gateway type field
	return nil, 0, fmt.Errorf("ListByGatewayType not implemented")
}

// GetSupportedGateways gets list of supported gateways
func (r *paymentConfigRepository) GetSupportedGateways(ctx context.Context) ([]string, error) {
	var gateways []string
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).Distinct("gateway").Pluck("gateway", &gateways).Error; err != nil {
		return nil, fmt.Errorf("failed to get supported gateways: %w", err)
	}
	return gateways, nil
}

// GetSupportedMethods gets list of supported methods
func (r *paymentConfigRepository) GetSupportedMethods(ctx context.Context) ([]string, error) {
	// TODO: Implement based on methods structure
	return nil, fmt.Errorf("GetSupportedMethods not implemented")
}

// ValidateGatewayConfig validates gateway configuration
func (r *paymentConfigRepository) ValidateGatewayConfig(ctx context.Context, gateway string, config string) (bool, error) {
	// TODO: Implement validation logic
	return true, nil
}

// TestGatewayConnection tests gateway connection
func (r *paymentConfigRepository) TestGatewayConnection(ctx context.Context, id uint) (bool, error) {
	// TODO: Implement connection test
	return true, nil
}

// GetMethodByCode gets method by code
func (r *paymentConfigRepository) GetMethodByCode(ctx context.Context, configID uint, methodCode string) (*entities.Method, error) {
	// TODO: Implement based on methods structure
	return nil, fmt.Errorf("GetMethodByCode not implemented")
}

// UpdateMethodStatus updates method status
func (r *paymentConfigRepository) UpdateMethodStatus(ctx context.Context, configID uint, methodCode string, isEnabled bool) error {
	// TODO: Implement based on methods structure
	return fmt.Errorf("UpdateMethodStatus not implemented")
}

// GetEnabledMethods gets enabled methods for a config
func (r *paymentConfigRepository) GetEnabledMethods(ctx context.Context, configID uint) ([]entities.Method, error) {
	// TODO: Implement based on methods structure
	return nil, fmt.Errorf("GetEnabledMethods not implemented")
}

// GetCurrencyStats gets currency statistics
func (r *paymentConfigRepository) GetCurrencyStats(ctx context.Context) (map[string]int64, error) {
	// TODO: Implement stats aggregation
	return make(map[string]int64), fmt.Errorf("GetCurrencyStats not implemented")
}

// GetMethodStats gets method statistics
func (r *paymentConfigRepository) GetMethodStats(ctx context.Context) (map[string]int64, error) {
	// TODO: Implement stats aggregation
	return make(map[string]int64), fmt.Errorf("GetMethodStats not implemented")
}

// GetGatewayStats gets gateway statistics
func (r *paymentConfigRepository) GetGatewayStats(ctx context.Context) (map[string]int64, error) {
	var result map[string]int64 = make(map[string]int64)
	
	type GatewayCount struct {
		Gateway string
		Count   int64
	}
	
	var gatewayCounts []GatewayCount
	if err := r.db.WithContext(ctx).Model(&entities.PaymentConfig{}).
		Select("gateway, COUNT(*) as count").
		Group("gateway").
		Scan(&gatewayCounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get gateway stats: %w", err)
	}
	
	for _, gc := range gatewayCounts {
		result[gc.Gateway] = gc.Count
	}
	
	return result, nil
}

// ListWithFilters lists configs with dynamic filters
func (r *paymentConfigRepository) ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	var configs []*entities.PaymentConfig
	var total int64
	db := r.db.WithContext(ctx).Model(&entities.PaymentConfig{})
	for key, value := range filters {
		db = db.Where(key+" = ?", value)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count configs with filters: %w", err)
	}
	if err := db.Limit(limit).Offset(offset).Find(&configs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list configs with filters: %w", err)
	}
	return configs, total, nil
}

// ImportConfigs imports payment configs
func (r *paymentConfigRepository) ImportConfigs(ctx context.Context, configs []*entities.PaymentConfig) error {
	if len(configs) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Create(&configs).Error; err != nil {
		return fmt.Errorf("failed to import configs: %w", err)
	}
	return nil
}

// ListByEnvironment lists configs by environment
func (r *paymentConfigRepository) ListByEnvironment(ctx context.Context, environment string, limit, offset int) ([]*entities.PaymentConfig, int64, error) {
	// TODO: Implement based on environment field
	return nil, 0, fmt.Errorf("ListByEnvironment not implemented")
}

// GetProductionConfigs gets production configs
func (r *paymentConfigRepository) GetProductionConfigs(ctx context.Context) ([]*entities.PaymentConfig, error) {
	// TODO: Implement based on environment field
	return nil, fmt.Errorf("GetProductionConfigs not implemented")
}

// GetTestConfigs gets test configs
func (r *paymentConfigRepository) GetTestConfigs(ctx context.Context) ([]*entities.PaymentConfig, error) {
	// TODO: Implement based on environment field  
	return nil, fmt.Errorf("GetTestConfigs not implemented")
}