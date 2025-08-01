package repository

import (
	"context"
	"time"

	"linke/internal/invoice/domain/model"
	"linke/internal/invoice/domain/valueobject"
)

// InvoiceRepository defines the interface for invoice persistence
type InvoiceRepository interface {
	// Save saves an invoice (create or update)
	Save(ctx context.Context, invoice *model.Invoice) error

	// FindByID finds an invoice by its ID
	FindByID(ctx context.Context, id valueobject.InvoiceID) (*model.Invoice, error)

	// FindByInvoiceNumber finds an invoice by its invoice number
	FindByInvoiceNumber(ctx context.Context, number valueobject.InvoiceNumber) (*model.Invoice, error)

	// FindByOrderID finds an invoice by order ID
	FindByOrderID(ctx context.Context, orderID uint) (*model.Invoice, error)

	// FindByUserID finds invoices for a specific user
	FindByUserID(ctx context.Context, userID uint, limit, offset int) ([]*model.Invoice, error)

	// ExistsByOrderID checks if an invoice exists for the given order ID
	ExistsByOrderID(ctx context.Context, orderID uint) (bool, error)

	// ExistsByInvoiceNumber checks if an invoice with the given number exists
	ExistsByInvoiceNumber(ctx context.Context, number valueobject.InvoiceNumber) (bool, error)

	// FindOverdueInvoices finds all overdue invoices
	FindOverdueInvoices(ctx context.Context, limit, offset int) ([]*model.Invoice, error)

	// FindInvoicesByStatus finds invoices by status
	FindInvoicesByStatus(ctx context.Context, status valueobject.InvoiceStatus, limit, offset int) ([]*model.Invoice, error)

	// List lists invoices with filtering and pagination
	List(ctx context.Context, filters InvoiceFilters) ([]*model.Invoice, int64, error)

	// Delete soft deletes an invoice
	Delete(ctx context.Context, id valueobject.InvoiceID) error

	// CountByStatus counts invoices by status
	CountByStatus(ctx context.Context, status valueobject.InvoiceStatus) (int64, error)

	// CountOverdue counts overdue invoices
	CountOverdue(ctx context.Context) (int64, error)

	// GetNextSequenceNumber gets the next sequence number for invoice number generation
	GetNextSequenceNumber(ctx context.Context, date string) (int, error)

	// GetPaymentHistory gets payment history for an invoice
	GetPaymentHistory(ctx context.Context, invoiceID valueobject.InvoiceID) ([]*PaymentHistoryEntry, error)
}

// InvoiceFilters represents filters for invoice listing
type InvoiceFilters struct {
	UserID        *uint                        `json:"user_id,omitempty"`
	OrderID       *uint                        `json:"order_id,omitempty"`
	Status        *valueobject.InvoiceStatus   `json:"status,omitempty"`
	InvoiceType   *valueobject.InvoiceType     `json:"invoice_type,omitempty"`
	Currency      *valueobject.Currency        `json:"currency,omitempty"`
	IsOverdue     *bool                        `json:"is_overdue,omitempty"`
	StartDate     *string                      `json:"start_date,omitempty"`  // YYYY-MM-DD format
	EndDate       *string                      `json:"end_date,omitempty"`    // YYYY-MM-DD format
	Search        *string                      `json:"search,omitempty"`      // Search in number, name, email
	SortBy        string                       `json:"sort_by,omitempty"`     // Field to sort by
	SortOrder     string                       `json:"sort_order,omitempty"`  // asc or desc
	Limit         int                          `json:"limit,omitempty"`
	Offset        int                          `json:"offset,omitempty"`
}

// PaymentHistoryEntry represents a payment history entry from the repository
type PaymentHistoryEntry struct {
	ID          string            `json:"id"`
	Amount      valueobject.Money `json:"amount"`
	PaymentRef  string            `json:"payment_ref"`
	PaymentDate time.Time         `json:"payment_date"`
	Notes       string            `json:"notes"`
}