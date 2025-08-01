package query

import (
	"time"

	"linke/internal/invoice/domain/model"
	"linke/internal/invoice/domain/valueobject"
)

// GetInvoiceQuery represents a query to get an invoice by ID
type GetInvoiceQuery struct {
	InvoiceID string `json:"invoice_id"`
}

// GetUserInvoiceQuery represents a query to get an invoice for a specific user
type GetUserInvoiceQuery struct {
	InvoiceID string `json:"invoice_id"`
	UserID    uint   `json:"user_id"`
}

// ListInvoicesQuery represents a query to list invoices with filters
type ListInvoicesQuery struct {
	Page      int        `json:"page"`
	Size      int        `json:"size"`
	UserID    *uint      `json:"user_id"`
	Status    []string   `json:"status"`
	Type      *string    `json:"type"`
	DateFrom  *time.Time `json:"date_from"`
	DateTo    *time.Time `json:"date_to"`
	Search    string     `json:"search"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order"`
}

// ListInvoicesResult represents the result of listing invoices
type ListInvoicesResult struct {
	Items      []*model.Invoice `json:"items"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Size       int              `json:"size"`
	TotalPages int              `json:"total_pages"`
	HasNext    bool             `json:"has_next"`
	HasPrev    bool             `json:"has_prev"`
}

// GetInvoiceStatsQuery represents a query to get invoice statistics
type GetInvoiceStatsQuery struct {
	UserID   *uint      `json:"user_id"`
	DateFrom *time.Time `json:"date_from"`
	DateTo   *time.Time `json:"date_to"`
}

// InvoiceStatsResult represents invoice statistics result
type InvoiceStatsResult struct {
	TotalInvoices     int64             `json:"total_invoices"`
	DraftCount        int64             `json:"draft_count"`
	SentCount         int64             `json:"sent_count"`
	PaidCount         int64             `json:"paid_count"`
	OverdueCount      int64             `json:"overdue_count"`
	VoidedCount       int64             `json:"voided_count"`
	TotalAmount       valueobject.Money `json:"total_amount"`
	PaidAmount        valueobject.Money `json:"paid_amount"`
	OutstandingAmount valueobject.Money `json:"outstanding_amount"`
	OverdueAmount     valueobject.Money `json:"overdue_amount"`
}

// GetInvoicePaymentHistoryQuery represents a query to get payment history for an invoice
type GetInvoicePaymentHistoryQuery struct {
	InvoiceID string `json:"invoice_id"`
	UserID    uint   `json:"user_id"`
}

// PaymentHistoryResult represents a payment history entry
type PaymentHistoryResult struct {
	ID          string            `json:"id"`
	Amount      valueobject.Money `json:"amount"`
	PaymentRef  string            `json:"payment_ref"`
	PaymentDate time.Time         `json:"payment_date"`
	Notes       string            `json:"notes"`
}