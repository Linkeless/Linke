package persistence

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"linke/internal/invoice/domain/model"
	"linke/internal/invoice/domain/repository"
	"linke/internal/invoice/domain/valueobject"
)

// InvoiceGormRepository implements invoice repository using GORM
type InvoiceGormRepository struct {
	db     *gorm.DB
	mapper *InvoiceMapper
}

// NewInvoiceGormRepository creates a new GORM invoice repository
func NewInvoiceGormRepository(db *gorm.DB) repository.InvoiceRepository {
	return &InvoiceGormRepository{
		db:     db,
		mapper: NewInvoiceMapper(),
	}
}

// Save saves an invoice (create or update)
func (r *InvoiceGormRepository) Save(ctx context.Context, invoice *model.Invoice) error {
	po := r.mapper.ToPersistence(invoice)

	// Use GORM's Save method which handles both create and update
	if err := r.db.WithContext(ctx).Save(po).Error; err != nil {
		return fmt.Errorf("failed to save invoice: %w", err)
	}

	return nil
}

// FindByID finds an invoice by its ID
func (r *InvoiceGormRepository) FindByID(ctx context.Context, id valueobject.InvoiceID) (*model.Invoice, error) {
	var po InvoicePO
	if err := r.db.WithContext(ctx).First(&po, id.Value()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found with ID %s", id.String())
		}
		return nil, fmt.Errorf("failed to find invoice by ID: %w", err)
	}

	return r.mapper.ToModel(&po)
}

// FindByInvoiceNumber finds an invoice by its invoice number
func (r *InvoiceGormRepository) FindByInvoiceNumber(ctx context.Context, number valueobject.InvoiceNumber) (*model.Invoice, error) {
	var po InvoicePO
	if err := r.db.WithContext(ctx).Where("invoice_number = ?", number.Value()).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found with number %s", number.String())
		}
		return nil, fmt.Errorf("failed to find invoice by number: %w", err)
	}

	return r.mapper.ToModel(&po)
}

// FindByOrderID finds an invoice by order ID
func (r *InvoiceGormRepository) FindByOrderID(ctx context.Context, orderID uint) (*model.Invoice, error) {
	var po InvoicePO
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice not found for order ID %d", orderID)
		}
		return nil, fmt.Errorf("failed to find invoice by order ID: %w", err)
	}

	return r.mapper.ToModel(&po)
}

// FindByUserID finds invoices for a specific user
func (r *InvoiceGormRepository) FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*model.Invoice, error) {
	var pos []InvoicePO
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find invoices by user ID: %w", err)
	}

	invoices := make([]*model.Invoice, len(pos))
	for i, po := range pos {
		invoice, err := r.mapper.ToModel(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to convert invoice to model: %w", err)
		}
		invoices[i] = invoice
	}

	return invoices, nil
}

// ExistsByOrderID checks if an invoice exists for the given order ID
func (r *InvoiceGormRepository) ExistsByOrderID(ctx context.Context, orderID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&InvoicePO{}).Where("order_id = ?", orderID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check invoice existence by order ID: %w", err)
	}
	return count > 0, nil
}

// ExistsByInvoiceNumber checks if an invoice with the given number exists
func (r *InvoiceGormRepository) ExistsByInvoiceNumber(ctx context.Context, number valueobject.InvoiceNumber) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&InvoicePO{}).Where("invoice_number = ?", number.Value()).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check invoice existence by number: %w", err)
	}
	return count > 0, nil
}

// FindOverdueInvoices finds all overdue invoices
func (r *InvoiceGormRepository) FindOverdueInvoices(ctx context.Context, limit, offset int) ([]*model.Invoice, error) {
	var pos []InvoicePO
	now := time.Now()
	
	query := r.db.WithContext(ctx).
		Where("due_at < ? AND status IN (?)", now, []string{"sent", "overdue"}).
		Where("paid_amount < total_amount").
		Order("due_at ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find overdue invoices: %w", err)
	}

	invoices := make([]*model.Invoice, len(pos))
	for i, po := range pos {
		invoice, err := r.mapper.ToModel(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to convert overdue invoice to model: %w", err)
		}
		invoices[i] = invoice
	}

	return invoices, nil
}

// FindInvoicesByStatus finds invoices by status
func (r *InvoiceGormRepository) FindInvoicesByStatus(ctx context.Context, status valueobject.InvoiceStatus, limit, offset int) ([]*model.Invoice, error) {
	var pos []InvoicePO
	query := r.db.WithContext(ctx).Where("status = ?", status.Value()).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find invoices by status: %w", err)
	}

	invoices := make([]*model.Invoice, len(pos))
	for i, po := range pos {
		invoice, err := r.mapper.ToModel(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to convert invoice to model: %w", err)
		}
		invoices[i] = invoice
	}

	return invoices, nil
}

// List lists invoices with filtering and pagination
func (r *InvoiceGormRepository) List(ctx context.Context, filters repository.InvoiceFilters) ([]*model.Invoice, int64, error) {
	query := r.db.WithContext(ctx).Model(&InvoicePO{})

	// Apply filters
	query = r.applyFilters(query, filters)

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Apply sorting
	query = r.applySorting(query, filters)

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var pos []InvoicePO
	if err := query.Find(&pos).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}

	invoices := make([]*model.Invoice, len(pos))
	for i, po := range pos {
		invoice, err := r.mapper.ToModel(&po)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert invoice to model: %w", err)
		}
		invoices[i] = invoice
	}

	return invoices, totalCount, nil
}

// Delete soft deletes an invoice
func (r *InvoiceGormRepository) Delete(ctx context.Context, id valueobject.InvoiceID) error {
	if err := r.db.WithContext(ctx).Delete(&InvoicePO{}, id.Value()).Error; err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}
	return nil
}

// CountByStatus counts invoices by status
func (r *InvoiceGormRepository) CountByStatus(ctx context.Context, status valueobject.InvoiceStatus) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&InvoicePO{}).Where("status = ?", status.Value()).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count invoices by status: %w", err)
	}
	return count, nil
}

// CountOverdue counts overdue invoices
func (r *InvoiceGormRepository) CountOverdue(ctx context.Context) (int64, error) {
	var count int64
	now := time.Now()
	
	if err := r.db.WithContext(ctx).Model(&InvoicePO{}).
		Where("due_at < ? AND status IN (?)", now, []string{"sent", "overdue"}).
		Where("paid_amount < total_amount").
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count overdue invoices: %w", err)
	}
	return count, nil
}

// GetNextSequenceNumber gets the next sequence number for invoice number generation
func (r *InvoiceGormRepository) GetNextSequenceNumber(ctx context.Context, date string) (int, error) {
	var count int64
	startOfDay, err := time.Parse("20060102", date)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %w", err)
	}
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.WithContext(ctx).Model(&InvoicePO{}).
		Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get sequence number: %w", err)
	}

	return int(count) + 1, nil
}

// GetPaymentHistory gets payment history for an invoice
func (r *InvoiceGormRepository) GetPaymentHistory(ctx context.Context, invoiceID valueobject.InvoiceID) ([]*repository.PaymentHistoryEntry, error) {
	// For now, return empty payment history as we don't have a payment table yet
	// In a real implementation, this would query a payment_records table
	return []*repository.PaymentHistoryEntry{}, nil
}

// 批量操作支持

// SaveBatch saves multiple invoices in a single transaction
func (r *InvoiceGormRepository) SaveBatch(ctx context.Context, invoices []*model.Invoice) error {
	if len(invoices) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, invoice := range invoices {
			po := r.mapper.ToPersistence(invoice)
			if err := tx.Save(po).Error; err != nil {
				return fmt.Errorf("failed to save invoice in batch: %w", err)
			}
		}
		return nil
	})
}

// FindByIDsBatch finds multiple invoices by their IDs in a single query
func (r *InvoiceGormRepository) FindByIDsBatch(ctx context.Context, ids []valueobject.InvoiceID) ([]*model.Invoice, error) {
	if len(ids) == 0 {
		return []*model.Invoice{}, nil
	}

	// Convert to uint slice for GORM
	uintIDs := make([]uint, len(ids))
	for i, id := range ids {
		uintIDs[i] = id.Value()
	}

	var pos []InvoicePO
	if err := r.db.WithContext(ctx).Where("id IN ?", uintIDs).Find(&pos).Error; err != nil {
		return nil, fmt.Errorf("failed to find invoices by IDs: %w", err)
	}

	invoices := make([]*model.Invoice, len(pos))
	for i, po := range pos {
		invoice, err := r.mapper.ToModel(&po)
		if err != nil {
			return nil, fmt.Errorf("failed to convert invoice to model: %w", err)
		}
		invoices[i] = invoice
	}

	return invoices, nil
}

// applyFilters applies filters to the query
func (r *InvoiceGormRepository) applyFilters(query *gorm.DB, filters repository.InvoiceFilters) *gorm.DB {
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}

	if filters.OrderID != nil {
		query = query.Where("order_id = ?", *filters.OrderID)
	}

	if filters.Status != nil {
		query = query.Where("status = ?", filters.Status.Value())
	}

	if filters.InvoiceType != nil {
		query = query.Where("invoice_type = ?", filters.InvoiceType.Value())
	}

	if filters.Currency != nil {
		query = query.Where("currency = ?", filters.Currency.Code())
	}

	// Date range filtering
	if filters.StartDate != nil {
		if startDate, err := time.Parse("2006-01-02", *filters.StartDate); err == nil {
			query = query.Where("issued_at >= ?", startDate)
		}
	}

	if filters.EndDate != nil {
		if endDate, err := time.Parse("2006-01-02", *filters.EndDate); err == nil {
			endDate = endDate.Add(24 * time.Hour)
			query = query.Where("issued_at < ?", endDate)
		}
	}

	// Overdue filter
	if filters.IsOverdue != nil {
		now := time.Now()
		if *filters.IsOverdue {
			query = query.Where("due_at < ? AND status != ?", now, "paid")
		} else {
			query = query.Where("due_at >= ? OR status = ?", now, "paid")
		}
	}

	// Search functionality
	if filters.Search != nil && *filters.Search != "" {
		searchPattern := "%" + *filters.Search + "%"
		query = query.Where(
			"invoice_number LIKE ? OR billing_name LIKE ? OR billing_email LIKE ? OR company_name LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	return query
}

// applySorting applies sorting to the query
func (r *InvoiceGormRepository) applySorting(query *gorm.DB, filters repository.InvoiceFilters) *gorm.DB {
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}

	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}

	validSortFields := map[string]bool{
		"created_at":     true,
		"updated_at":     true,
		"issued_at":      true,
		"due_at":         true,
		"paid_at":        true,
		"total_amount":   true,
		"status":         true,
		"invoice_number": true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	return query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
}
// 性能优化的查询方法

// CountOverdueInvoicesOptimized counts overdue invoices with optimized query
func (r *InvoiceGormRepository) CountOverdueInvoicesOptimized(ctx context.Context) (int64, error) {
	var count int64
	
	// 使用索引优化的查询
	if err := r.db.WithContext(ctx).Model(&InvoicePO{}).
		Where("status = ? AND due_at < ? AND paid_amount < total_amount", "overdue", time.Now()).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count overdue invoices: %w", err)
	}
	
	return count, nil
}

// UpdateStatusBatch updates status for multiple invoices efficiently
func (r *InvoiceGormRepository) UpdateStatusBatch(ctx context.Context, ids []valueobject.InvoiceID, status valueobject.InvoiceStatus) error {
	if len(ids) == 0 {
		return nil
	}

	// Convert to uint slice
	uintIDs := make([]uint, len(ids))
	for i, id := range ids {
		uintIDs[i] = id.Value()
	}

	// 批量更新状态
	if err := r.db.WithContext(ctx).Model(&InvoicePO{}).
		Where("id IN ?", uintIDs).
		Updates(map[string]interface{}{
			"status":     status.Value(),
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update invoice status in batch: %w", err)
	}

	return nil
}

// FindInvoicesSummary finds invoices with only essential fields for performance
func (r *InvoiceGormRepository) FindInvoicesSummary(ctx context.Context, filters repository.InvoiceFilters) ([]InvoiceSummary, error) {
	query := r.db.WithContext(ctx).Model(&InvoicePO{}).
		Select("id, invoice_number, user_id, status, total_amount, currency, due_at, created_at")
	
	// Apply filters
	query = r.applyFilters(query, filters)
	
	// Apply sorting and pagination
	query = r.applySorting(query, filters)
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	var summaries []InvoiceSummary
	if err := query.Scan(&summaries).Error; err != nil {
		return nil, fmt.Errorf("failed to find invoice summaries: %w", err)
	}

	return summaries, nil
}

// InvoiceSummary represents a lightweight invoice summary
type InvoiceSummary struct {
	ID            uint      `json:"id"`
	InvoiceNumber string    `json:"invoice_number"`
	UserID        uint      `json:"user_id"`
	Status        string    `json:"status"`
	TotalAmount   float64   `json:"total_amount"`
	Currency      string    `json:"currency"`
	DueAt         *time.Time `json:"due_at"`
	CreatedAt     time.Time `json:"created_at"`
}
