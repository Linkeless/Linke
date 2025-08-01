package query

import (
	"context"
	"fmt"

	"linke/internal/invoice/domain/model"
	"linke/internal/invoice/domain/repository"
	"linke/internal/invoice/domain/valueobject"
)

// InvoiceQueryHandler handles invoice queries
type InvoiceQueryHandler struct {
	repository repository.InvoiceRepository
}

// NewInvoiceQueryHandler creates a new invoice query handler
func NewInvoiceQueryHandler(repository repository.InvoiceRepository) *InvoiceQueryHandler {
	return &InvoiceQueryHandler{
		repository: repository,
	}
}

// GetInvoice handles get invoice query
func (h *InvoiceQueryHandler) GetInvoice(ctx context.Context, query GetInvoiceQuery) (*model.Invoice, error) {
	invoiceID, err := valueobject.ParseInvoiceID(query.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	return invoice, nil
}

// GetUserInvoice handles get user invoice query (ensures user owns the invoice)
func (h *InvoiceQueryHandler) GetUserInvoice(ctx context.Context, query GetUserInvoiceQuery) (*model.Invoice, error) {
	invoiceID, err := valueobject.ParseInvoiceID(query.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	// Verify ownership
	if invoice.UserID() != query.UserID {
		return nil, fmt.Errorf("invoice not found or access denied")
	}

	return invoice, nil
}


// ListInvoices handles list invoices query
func (h *InvoiceQueryHandler) ListInvoices(ctx context.Context, query ListInvoicesQuery) (*ListInvoicesResult, error) {
	filters := h.buildNewFilters(query)
	
	invoices, total, err := h.repository.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list invoices: %w", err)
	}

	// Calculate pagination info
	totalPages := int(total) / query.Size
	if int(total)%query.Size != 0 {
		totalPages++
	}

	hasNext := query.Page < totalPages
	hasPrev := query.Page > 1

	return &ListInvoicesResult{
		Items:      invoices,
		Total:      total,
		Page:       query.Page,
		Size:       query.Size,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}


// GetInvoiceStats handles get invoice stats query
func (h *InvoiceQueryHandler) GetInvoiceStats(ctx context.Context, query GetInvoiceStatsQuery) (*InvoiceStatsResult, error) {
	// Build filters for the stats query  
	var startDate, endDate *string
	if query.DateFrom != nil {
		dateStr := query.DateFrom.Format("2006-01-02")
		startDate = &dateStr
	}
	if query.DateTo != nil {
		dateStr := query.DateTo.Format("2006-01-02")
		endDate = &dateStr
	}

	filters := repository.InvoiceFilters{
		UserID:    query.UserID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Get all invoices matching the filters
	invoices, _, err := h.repository.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices for stats: %w", err)
	}

	// Initialize stats with zero values
	usdCurrency, _ := valueobject.NewCurrency("USD")
	stats := &InvoiceStatsResult{
		TotalInvoices:     0,
		DraftCount:        0,
		SentCount:         0,
		PaidCount:         0,
		OverdueCount:      0,
		VoidedCount:       0,
		TotalAmount:       valueobject.Zero(usdCurrency),
		PaidAmount:        valueobject.Zero(usdCurrency),
		OutstandingAmount: valueobject.Zero(usdCurrency),
		OverdueAmount:     valueobject.Zero(usdCurrency),
	}

	// Process each invoice
	for _, invoice := range invoices {
		stats.TotalInvoices++

		// Count by status
		switch {
		case invoice.Status().IsDraft():
			stats.DraftCount++
		case invoice.Status().IsSent():
			stats.SentCount++
		case invoice.Status().IsPaid():
			stats.PaidCount++
		case invoice.Status().IsOverdue():
			stats.OverdueCount++
		case invoice.Status().IsVoided():
			stats.VoidedCount++
		}

		// For aggregated amounts, we'll only count USD for simplicity
		// In a real application, you'd need currency conversion
		if invoice.TotalAmount().Currency().Code() == "USD" {
			stats.TotalAmount, _ = stats.TotalAmount.Add(invoice.TotalAmount())
			stats.PaidAmount, _ = stats.PaidAmount.Add(invoice.PaidAmount())

			if !invoice.Status().IsPaid() {
				remaining, _ := invoice.RemainingAmount()
				stats.OutstandingAmount, _ = stats.OutstandingAmount.Add(remaining)

				if invoice.IsOverdue() {
					stats.OverdueAmount, _ = stats.OverdueAmount.Add(remaining)
				}
			}
		}
	}

	return stats, nil
}


// buildNewFilters converts new query structure to repository filters
func (h *InvoiceQueryHandler) buildNewFilters(query ListInvoicesQuery) repository.InvoiceFilters {
	var searchPtr *string
	if query.Search != "" {
		searchPtr = &query.Search
	}
	
	filters := repository.InvoiceFilters{
		UserID:    query.UserID,
		Search:    searchPtr,
		SortBy:    query.SortBy,
		SortOrder: query.SortOrder,
		Limit:     query.Size,
		Offset:    (query.Page - 1) * query.Size,
	}

	// Handle status array
	if len(query.Status) > 0 {
		// For now, take the first status. In a real implementation,
		// you might want to support multiple statuses
		if status, err := valueobject.NewInvoiceStatus(query.Status[0]); err == nil {
			filters.Status = &status
		}
	}

	// Parse invoice type if provided
	if query.Type != nil {
		if invoiceType, err := valueobject.NewInvoiceType(*query.Type); err == nil {
			filters.InvoiceType = &invoiceType
		}
	}

	// Convert time.Time to string for compatibility
	if query.DateFrom != nil {
		dateStr := query.DateFrom.Format("2006-01-02")
		filters.StartDate = &dateStr
	}
	if query.DateTo != nil {
		dateStr := query.DateTo.Format("2006-01-02")
		filters.EndDate = &dateStr
	}

	return filters
}

// GetInvoicePaymentHistory handles get invoice payment history query
func (h *InvoiceQueryHandler) GetInvoicePaymentHistory(ctx context.Context, query GetInvoicePaymentHistoryQuery) ([]*PaymentHistoryResult, error) {
	// Parse invoice ID
	invoiceID, err := valueobject.ParseInvoiceID(query.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	// Find invoice to verify ownership
	invoice, err := h.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find invoice: %w", err)
	}

	// Verify ownership
	if invoice.UserID() != query.UserID {
		return nil, fmt.Errorf("invoice not found or access denied")
	}

	// Get payment history from repository
	entries, err := h.repository.GetPaymentHistory(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment history: %w", err)
	}

	// Convert to query result format
	history := make([]*PaymentHistoryResult, len(entries))
	for i, entry := range entries {
		history[i] = &PaymentHistoryResult{
			ID:          entry.ID,
			Amount:      entry.Amount,
			PaymentRef:  entry.PaymentRef,
			PaymentDate: entry.PaymentDate,
			Notes:       entry.Notes,
		}
	}

	return history, nil
}