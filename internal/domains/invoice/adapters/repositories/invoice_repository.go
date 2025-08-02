package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"linke/internal/domains/invoice/entities"
	"linke/internal/domains/invoice/usecases/interfaces"

	"gorm.io/gorm"
)

// invoiceRepository implements the InvoiceRepository interface
type invoiceRepository struct {
	db *gorm.DB
}

// NewInvoiceRepository creates a new InvoiceRepository implementation
func NewInvoiceRepository(db *gorm.DB) interfaces.InvoiceRepository {
	return &invoiceRepository{
		db: db,
	}
}

// Basic CRUD operations

// Create creates a new invoice in the database
func (r *invoiceRepository) Create(ctx context.Context, invoice *entities.Invoice) error {
	if err := r.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

// GetByID retrieves an invoice by its ID
func (r *invoiceRepository) GetByID(ctx context.Context, id uint) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := r.db.WithContext(ctx).First(&invoice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get invoice by id: %w", err)
	}
	return &invoice, nil
}

// GetByInvoiceNumber retrieves an invoice by its invoice number
func (r *invoiceRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := r.db.WithContext(ctx).Where("invoice_number = ?", invoiceNumber).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice with number %s not found", invoiceNumber)
		}
		return nil, fmt.Errorf("failed to get invoice by number: %w", err)
	}
	return &invoice, nil
}

// Update updates an existing invoice
func (r *invoiceRepository) Update(ctx context.Context, invoice *entities.Invoice) error {
	if err := r.db.WithContext(ctx).Save(invoice).Error; err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}
	return nil
}

// Delete soft deletes an invoice by its ID
func (r *invoiceRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&entities.Invoice{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}
	return nil
}

// User operations

// ListByUser retrieves invoices for a specific user with pagination
func (r *invoiceRepository) ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entities.Invoice{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user invoices: %w", err)
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list user invoices: %w", err)
	}

	return invoices, total, nil
}

// GetUserInvoiceHistory retrieves invoice history for a user
func (r *invoiceRepository) GetUserInvoiceHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error) {
	return r.ListByUser(ctx, userID, limit, offset)
}

// Status operations

// ListByStatus retrieves invoices by status with pagination
func (r *invoiceRepository) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entities.Invoice{}).Where("status = ?", status).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices by status: %w", err)
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by status: %w", err)
	}

	return invoices, total, nil
}

// ListPending retrieves pending invoices
func (r *invoiceRepository) ListPending(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error) {
	return r.ListByStatus(ctx, "pending", limit, offset)
}

// ListPaid retrieves paid invoices
func (r *invoiceRepository) ListPaid(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error) {
	return r.ListByStatus(ctx, "paid", limit, offset)
}

// ListOverdue retrieves overdue invoices
func (r *invoiceRepository) ListOverdue(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error) {
	return r.ListByStatus(ctx, "overdue", limit, offset)
}

// UpdateStatus updates the status of an invoice
func (r *invoiceRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	if err := r.db.WithContext(ctx).Model(&entities.Invoice{}).
		Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return fmt.Errorf("failed to update invoice status: %w", err)
	}
	return nil
}

// Time-based operations

// ListByDateRange retrieves invoices within a date range
func (r *invoiceRepository) ListByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entities.Invoice{}).
		Where("created_at BETWEEN ? AND ?", start, end).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices by date range: %w", err)
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by date range: %w", err)
	}

	return invoices, total, nil
}

// ListByDueDate retrieves invoices due before a specific date
func (r *invoiceRepository) ListByDueDate(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entities.Invoice{}).
		Where("due_date < ?", beforeDate).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices by due date: %w", err)
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Where("due_date < ?", beforeDate).
		Order("due_date ASC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by due date: %w", err)
	}

	return invoices, total, nil
}

// Currency operations

// ListByCurrency retrieves invoices by currency
func (r *invoiceRepository) ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entities.Invoice{}).
		Where("currency = ?", currency).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices by currency: %w", err)
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).Where("currency = ?", currency).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by currency: %w", err)
	}

	return invoices, total, nil
}

// List operations

// List retrieves all invoices with pagination
func (r *invoiceRepository) List(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&entities.Invoice{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count total invoices: %w", err)
	}

	// Get paginated results
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}

	return invoices, total, nil
}

// CountTotal returns the total number of invoices
func (r *invoiceRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.Invoice{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count total invoices: %w", err)
	}
	return count, nil
}

// Statistics

// GetStatusStats returns statistics of invoices by status
func (r *invoiceRepository) GetStatusStats(ctx context.Context) (map[string]int64, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var results []StatusCount
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get status stats: %w", err)
	}

	stats := make(map[string]int64)
	for _, result := range results {
		stats[result.Status] = result.Count
	}

	return stats, nil
}

// GetRevenueStats returns revenue statistics for a specific currency since a date
func (r *invoiceRepository) GetRevenueStats(ctx context.Context, currency string, since time.Time) (float64, error) {
	var revenue float64
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("currency = ? AND status = 'paid' AND created_at >= ?", currency, since).
		Scan(&revenue).Error; err != nil {
		return 0, fmt.Errorf("failed to get revenue stats: %w", err)
	}

	return revenue, nil
}

// Search operations

// Search searches invoices by various fields
func (r *invoiceRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	searchPattern := "%" + strings.ToLower(query) + "%"

	// Build search conditions
	searchQuery := r.db.WithContext(ctx).Where(
		"LOWER(invoice_number) LIKE ? OR LOWER(billing_name) LIKE ? OR LOWER(billing_email) LIKE ? OR LOWER(company_name) LIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	)

	// Count total records
	if err := searchQuery.Model(&entities.Invoice{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	// Get paginated results
	if err := searchQuery.
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to search invoices: %w", err)
	}

	return invoices, total, nil
}

// Existence checks

// ExistsByInvoiceNumber checks if an invoice with the given number exists
func (r *invoiceRepository) ExistsByInvoiceNumber(ctx context.Context, invoiceNumber string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("invoice_number = ?", invoiceNumber).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check invoice number existence: %w", err)
	}
	return count > 0, nil
}