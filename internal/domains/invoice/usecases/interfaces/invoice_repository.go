package interfaces

import (
	"context"
	"linke/internal/domains/invoice/entities"
	"linke/internal/shared/framework"
	"time"
)

// InvoiceRepository defines the interface for invoice data access operations
// It extends UserScopedRepository and TimeBasedRepository with Invoice-specific methods
type InvoiceRepository interface {
	framework.UserScopedRepository[entities.Invoice, uint]
	framework.TimeBasedRepository[entities.Invoice, uint]
	
	// Invoice-specific query methods
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error)

	// User operations (extending UserScopedRepository)
	GetUserInvoiceHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error)

	// Invoice-specific status operations
	ListPending(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error)
	ListPaid(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error)
	ListOverdue(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error)

	// Invoice-specific time-based operations (extending TimeBasedRepository)
	ListByDueDate(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.Invoice, int64, error)

	// Currency operations
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.Invoice, int64, error)

	// Invoice-specific statistics
	GetStatusStats(ctx context.Context) (map[string]int64, error)
	GetRevenueStats(ctx context.Context, currency string, since time.Time) (float64, error)

	// Invoice-specific existence checks
	ExistsByInvoiceNumber(ctx context.Context, invoiceNumber string) (bool, error)
}
