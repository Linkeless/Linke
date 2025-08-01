package persistence

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/repository"
	"linke/internal/payment/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// PaymentGormRepository implements PaymentRepository using GORM
type PaymentGormRepository struct {
	db     *gorm.DB
	mapper *PaymentMapper
}

// NewPaymentGormRepository creates a new PaymentGormRepository
func NewPaymentGormRepository(db *gorm.DB) repository.PaymentRepository {
	return &PaymentGormRepository{
		db:     db,
		mapper: NewPaymentMapper(),
	}
}

// Save saves a payment aggregate
func (r *PaymentGormRepository) Save(ctx context.Context, payment *aggregate.Payment) error {
	po, err := r.mapper.ToPersistentObject(payment)
	if err != nil {
		return fmt.Errorf("failed to convert payment to persistent object: %w", err)
	}

	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	// Note: The aggregate ID should be set during aggregate creation
	// No need to update it after save as it's already set

	return nil
}

// Update updates a payment aggregate
func (r *PaymentGormRepository) Update(ctx context.Context, payment *aggregate.Payment) error {
	po, err := r.mapper.ToPersistentObject(payment)
	if err != nil {
		return fmt.Errorf("failed to convert payment to persistent object: %w", err)
	}

	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	return nil
}

// FindByID finds a payment by its ID
func (r *PaymentGormRepository) FindByID(ctx context.Context, id valueobject.PaymentID) (*aggregate.Payment, error) {
	var po PaymentPO
	if err := r.db.WithContext(ctx).First(&po, id.Value()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found with ID: %s", id.String())
		}
		return nil, fmt.Errorf("failed to find payment by ID: %w", err)
	}

	return r.mapper.ToAggregate(&po)
}

// FindByPaymentNumber finds a payment by its payment number
func (r *PaymentGormRepository) FindByPaymentNumber(ctx context.Context, paymentNumber valueobject.PaymentNumber) (*aggregate.Payment, error) {
	var po PaymentPO
	if err := r.db.WithContext(ctx).Where("payment_number = ?", paymentNumber.Value()).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found with number: %s", paymentNumber.String())
		}
		return nil, fmt.Errorf("failed to find payment by number: %w", err)
	}

	return r.mapper.ToAggregate(&po)
}

// FindByInvoiceID finds payments by invoice ID
func (r *PaymentGormRepository) FindByInvoiceID(ctx context.Context, invoiceID sharedvo.InvoiceID) ([]*aggregate.Payment, error) {
	var pos []*PaymentPO
	if err := r.db.WithContext(ctx).Where("invoice_id = ?", invoiceID.Value()).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find payments by invoice ID: %w", err)
	}

	return r.mapper.ToAggregateList(pos)
}

// FindByUserID finds payments by user ID with pagination
func (r *PaymentGormRepository) FindByUserID(ctx context.Context, userID sharedvo.UserID, limit, offset int) ([]*aggregate.Payment, error) {
	var pos []*PaymentPO
	query := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find payments by user ID: %w", err)
	}

	return r.mapper.ToAggregateList(pos)
}

// FindByStatus finds payments by status with pagination
func (r *PaymentGormRepository) FindByStatus(ctx context.Context, status valueobject.PaymentStatus, limit, offset int) ([]*aggregate.Payment, error) {
	var pos []*PaymentPO
	query := r.db.WithContext(ctx).Where("status = ?", status.Value()).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find payments by status: %w", err)
	}

	return r.mapper.ToAggregateList(pos)
}

// FindByGatewayTransactionID finds a payment by gateway transaction ID
func (r *PaymentGormRepository) FindByGatewayTransactionID(ctx context.Context, transactionID string) (*aggregate.Payment, error) {
	var po PaymentPO
	if err := r.db.WithContext(ctx).Where("gateway_transaction_id = ?", transactionID).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment not found with gateway transaction ID: %s", transactionID)
		}
		return nil, fmt.Errorf("failed to find payment by gateway transaction ID: %w", err)
	}

	return r.mapper.ToAggregate(&po)
}

// FindExpiredPayments finds payments that have expired
func (r *PaymentGormRepository) FindExpiredPayments(ctx context.Context, limit int) ([]*aggregate.Payment, error) {
	now := time.Now()
	var pos []*PaymentPO
	
	query := r.db.WithContext(ctx).Where("expires_at < ? AND status IN (?)", now, []string{"pending", "processing"}).Order("expires_at ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find expired payments: %w", err)
	}

	return r.mapper.ToAggregateList(pos)
}

// FindWithFilters finds payments with complex filters
func (r *PaymentGormRepository) FindWithFilters(ctx context.Context, filters repository.PaymentFilters) ([]*aggregate.Payment, int64, error) {
	query := r.db.WithContext(ctx).Model(&PaymentPO{})

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", filters.UserID.Value())
	}

	if filters.InvoiceID != nil {
		query = query.Where("invoice_id = ?", filters.InvoiceID.Value())
	}

	if filters.Status != nil {
		query = query.Where("status = ?", filters.Status.Value())
	}

	if filters.PaymentGateway != nil {
		query = query.Where("payment_gateway = ?", filters.PaymentGateway.Value())
	}

	if filters.PaymentMethod != nil {
		query = query.Where("payment_method = ?", filters.PaymentMethod.Value())
	}

	if filters.Currency != nil {
		query = query.Where("currency = ?", filters.Currency.Code())
	}

	// Amount range filtering
	if filters.AmountRange != nil {
		if filters.AmountRange.Min != nil {
			query = query.Where("amount >= ?", filters.AmountRange.Min.Amount())
		}
		if filters.AmountRange.Max != nil {
			query = query.Where("amount <= ?", filters.AmountRange.Max.Amount())
		}
	}

	// Date range filtering
	if filters.DateRange != nil {
		if filters.DateRange.Start != nil {
			if startDate, err := time.Parse("2006-01-02", *filters.DateRange.Start); err == nil {
				query = query.Where("created_at >= ?", startDate)
			}
		}
		if filters.DateRange.End != nil {
			if endDate, err := time.Parse("2006-01-02", *filters.DateRange.End); err == nil {
				endDate = endDate.Add(24 * time.Hour)
				query = query.Where("created_at < ?", endDate)
			}
		}
	}

	// Search functionality
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where(
			"payment_number LIKE ? OR gateway_transaction_id LIKE ? OR notes LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	validSortFields := map[string]bool{
		"created_at":        true,
		"updated_at":        true,
		"completed_at":      true,
		"amount":           true,
		"status":           true,
		"payment_number":   true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}

	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var pos []*PaymentPO
	if err := query.Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find payments: %w", err)
	}

	payments, err := r.mapper.ToAggregateList(pos)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to convert payments: %w", err)
	}

	return payments, totalCount, nil
}

// Delete soft deletes a payment
func (r *PaymentGormRepository) Delete(ctx context.Context, id valueobject.PaymentID) error {
	if err := r.db.WithContext(ctx).Delete(&PaymentPO{}, id.Value()).Error; err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	return nil
}

// Exists checks if a payment exists by ID
func (r *PaymentGormRepository) Exists(ctx context.Context, id valueobject.PaymentID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentPO{}).Where("id = ?", id.Value()).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment existence: %w", err)
	}

	return count > 0, nil
}

// ExistsByPaymentNumber checks if a payment exists by payment number
func (r *PaymentGormRepository) ExistsByPaymentNumber(ctx context.Context, paymentNumber valueobject.PaymentNumber) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentPO{}).Where("payment_number = ?", paymentNumber.Value()).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check payment number existence: %w", err)
	}

	return count > 0, nil
}

// Count returns the total count of payments
func (r *PaymentGormRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentPO{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count payments: %w", err)
	}

	return count, nil
}

// CountByStatus returns the count of payments by status
func (r *PaymentGormRepository) CountByStatus(ctx context.Context, status valueobject.PaymentStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentPO{}).Where("status = ?", status.Value()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count payments by status: %w", err)
	}

	return count, nil
}

// CountByUserID returns the count of payments by user ID
func (r *PaymentGormRepository) CountByUserID(ctx context.Context, userID sharedvo.UserID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&PaymentPO{}).Where("user_id = ?", userID.Value()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count payments by user ID: %w", err)
	}

	return count, nil
}