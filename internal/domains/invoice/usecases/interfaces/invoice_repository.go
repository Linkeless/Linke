package interfaces

import (
	"context"
	"time"
	"linke/internal/domains/invoice/entities"
)

// InvoiceRepository defines the interface for invoice data access operations
type InvoiceRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, invoice *entities.Invoice) error
	GetByID(ctx context.Context, id uint) (*entities.Invoice, error)
	GetByInvoiceNumber(ctx context.Context, invoiceNumber string) (*entities.Invoice, error)
	Update(ctx context.Context, invoice *entities.Invoice) error
	Delete(ctx context.Context, id uint) error

	// User operations
	ListByUser(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error)
	GetUserInvoiceHistory(ctx context.Context, userID uint, limit, offset int) ([]*entities.Invoice, int64, error)

	// Status operations
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*entities.Invoice, int64, error)
	ListPending(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error)
	ListPaid(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error)
	ListOverdue(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error)
	UpdateStatus(ctx context.Context, id uint, status string) error

	// Time-based operations
	ListByDateRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*entities.Invoice, int64, error)
	ListByDueDate(ctx context.Context, beforeDate time.Time, limit, offset int) ([]*entities.Invoice, int64, error)

	// Currency operations
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.Invoice, int64, error)

	// List operations
	List(ctx context.Context, limit, offset int) ([]*entities.Invoice, int64, error)
	CountTotal(ctx context.Context) (int64, error)

	// Statistics
	GetStatusStats(ctx context.Context) (map[string]int64, error)
	GetRevenueStats(ctx context.Context, currency string, since time.Time) (float64, error)

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Invoice, int64, error)

	// Existence checks
	ExistsByInvoiceNumber(ctx context.Context, invoiceNumber string) (bool, error)
}