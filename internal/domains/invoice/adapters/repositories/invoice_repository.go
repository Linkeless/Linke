package repositories

import (
	"context"
	"fmt"
	"time"

	"linke/internal/domains/invoice/entities"
	"linke/internal/domains/invoice/usecases/interfaces"
	"linke/internal/shared/framework"
	"linke/internal/shared/repository"

	"gorm.io/gorm"
)

// invoiceRepository implements the InvoiceRepository interface
type invoiceRepository struct {
	*repository.UserScopedTimeBasedRepositoryImpl[entities.Invoice, uint]
}

// NewInvoiceRepository creates a new InvoiceRepository implementation
func NewInvoiceRepository(db *gorm.DB, frameworkLogger framework.Logger) interfaces.InvoiceRepository {
	return &invoiceRepository{
		UserScopedTimeBasedRepositoryImpl: repository.NewUserScopedTimeBasedRepository[entities.Invoice, uint](db, frameworkLogger),
	}
}

// Domain-specific implementations for InvoiceRepository

// GetByInvoiceNumber retrieves an invoice by its invoice number (domain-specific method)
func (r *invoiceRepository) GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := r.GetDB().WithContext(ctx).Where("invoice_number = ?", invoiceNumber).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice with number %s not found", invoiceNumber)
		}
		return nil, fmt.Errorf("failed to get invoice by number: %w", err)
	}
	return &invoice, nil
}

// GetBySubscriptionOrderID retrieves an invoice by subscription order ID
func (r *invoiceRepository) GetBySubscriptionOrderID(ctx context.Context, orderID uint) (*entities.Invoice, error) {
	var invoice entities.Invoice
	if err := r.GetDB().WithContext(ctx).Where("subscription_order_id = ?", orderID).First(&invoice).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invoice for order %d not found", orderID)
		}
		return nil, fmt.Errorf("failed to get invoice by order ID: %w", err)
	}
	return &invoice, nil
}

// ListBySubscriptionOrderIDs retrieves invoices by subscription order IDs
func (r *invoiceRepository) ListBySubscriptionOrderIDs(ctx context.Context, orderIDs []uint, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	// Count total records
	if err := r.GetDB().WithContext(ctx).Model(&entities.Invoice{}).
		Where("subscription_order_id IN ?", orderIDs).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices by order IDs: %w", err)
	}

	// Get paginated results
	if err := r.GetDB().WithContext(ctx).
		Where("subscription_order_id IN ?", orderIDs).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by order IDs: %w", err)
	}

	return invoices, total, nil
}

// GetUserInvoiceHistory retrieves invoice history for a user
func (r *invoiceRepository) GetUserInvoiceHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error) {
	return r.ListByUser(ctx, userID, limit, offset)
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

// Invoice-specific time-based operations

// ListByDueDate retrieves invoices due before a specific date
func (r *invoiceRepository) ListByDueDate(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.Invoice, int64, error) {
	var invoices []*entities.Invoice
	var total int64

	// Count total records
	if err := r.GetDB().WithContext(ctx).Model(&entities.Invoice{}).
		Where("due_date < ?", beforeDate).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices by due date: %w", err)
	}

	// Get paginated results
	if err := r.GetDB().WithContext(ctx).
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
	if err := r.GetDB().WithContext(ctx).Model(&entities.Invoice{}).
		Where("currency = ?", currency).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices by currency: %w", err)
	}

	// Get paginated results
	if err := r.GetDB().WithContext(ctx).Where("currency = ?", currency).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&invoices).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices by currency: %w", err)
	}

	return invoices, total, nil
}

// Invoice-specific statistics

// GetStatusStats returns statistics of invoices by status
func (r *invoiceRepository) GetStatusStats(ctx context.Context) (map[string]int64, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var results []StatusCount
	if err := r.GetDB().WithContext(ctx).
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
	if err := r.GetDB().WithContext(ctx).
		Model(&entities.Invoice{}).
		Select("COALESCE(SUM(total_amount), 0)").
		Where("currency = ? AND status = 'paid' AND created_at >= ?", currency, since).
		Scan(&revenue).Error; err != nil {
		return 0, fmt.Errorf("failed to get revenue stats: %w", err)
	}

	return revenue, nil
}

// Invoice-specific existence checks

// ExistsByInvoiceNumber checks if an invoice with the given number exists
func (r *invoiceRepository) ExistsByInvoiceNumber(ctx context.Context, invoiceNumber string) (bool, error) {
	var count int64
	if err := r.GetDB().WithContext(ctx).
		Model(&entities.Invoice{}).
		Where("invoice_number = ?", invoiceNumber).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check invoice number existence: %w", err)
	}
	return count > 0, nil
}
