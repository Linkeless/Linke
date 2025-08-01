package query

import (
	"context"
	"fmt"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/repository"
	"linke/internal/payment/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// PaymentQueryHandler handles payment queries
type PaymentQueryHandler struct {
	paymentRepo repository.PaymentRepository
}

// NewPaymentQueryHandler creates a new PaymentQueryHandler
func NewPaymentQueryHandler(paymentRepo repository.PaymentRepository) *PaymentQueryHandler {
	return &PaymentQueryHandler{
		paymentRepo: paymentRepo,
	}
}

// GetPayment handles the GetPaymentQuery
func (h *PaymentQueryHandler) GetPayment(ctx context.Context, query GetPaymentQuery) (*PaymentDTO, error) {
	paymentID, err := valueobject.NewPaymentID(query.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment ID: %w", err)
	}

	payment, err := h.paymentRepo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return h.toPaymentDTO(payment), nil
}

// GetPaymentByNumber handles the GetPaymentByNumberQuery
func (h *PaymentQueryHandler) GetPaymentByNumber(ctx context.Context, query GetPaymentByNumberQuery) (*PaymentDTO, error) {
	paymentNumber, err := valueobject.NewPaymentNumber(query.PaymentNumber)
	if err != nil {
		return nil, fmt.Errorf("invalid payment number: %w", err)
	}

	payment, err := h.paymentRepo.FindByPaymentNumber(ctx, paymentNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	return h.toPaymentDTO(payment), nil
}

// ListPayments handles the ListPaymentsQuery
func (h *PaymentQueryHandler) ListPayments(ctx context.Context, query ListPaymentsQuery) (*PaymentListResult, error) {
	filters := h.buildPaymentFilters(query)

	payments, totalCount, err := h.paymentRepo.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}

	paymentDTOs := make([]PaymentDTO, 0, len(payments))
	for _, payment := range payments {
		paymentDTOs = append(paymentDTOs, *h.toPaymentDTO(payment))
	}

	hasMore := false
	if query.Limit > 0 {
		hasMore = int64(query.Offset+query.Limit) < totalCount
	}

	return &PaymentListResult{
		Payments:   paymentDTOs,
		TotalCount: totalCount,
		Limit:      query.Limit,
		Offset:     query.Offset,
		HasMore:    hasMore,
	}, nil
}

// GetPaymentsByUser handles the GetPaymentsByUserQuery
func (h *PaymentQueryHandler) GetPaymentsByUser(ctx context.Context, query GetPaymentsByUserQuery) (*PaymentListResult, error) {
	userID, err := sharedvo.NewUserIDFromUint(query.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	payments, err := h.paymentRepo.FindByUserID(ctx, userID, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments by user: %w", err)
	}

	totalCount, err := h.paymentRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count payments by user: %w", err)
	}

	paymentDTOs := make([]PaymentDTO, 0, len(payments))
	for _, payment := range payments {
		paymentDTOs = append(paymentDTOs, *h.toPaymentDTO(payment))
	}

	hasMore := false
	if query.Limit > 0 {
		hasMore = int64(query.Offset+query.Limit) < totalCount
	}

	return &PaymentListResult{
		Payments:   paymentDTOs,
		TotalCount: totalCount,
		Limit:      query.Limit,
		Offset:     query.Offset,
		HasMore:    hasMore,
	}, nil
}

// GetPaymentsByInvoice handles the GetPaymentsByInvoiceQuery
func (h *PaymentQueryHandler) GetPaymentsByInvoice(ctx context.Context, query GetPaymentsByInvoiceQuery) ([]PaymentDTO, error) {
	// Create shared invoice ID directly
	invoiceID, err := sharedvo.NewInvoiceID(query.InvoiceID)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice ID: %w", err)
	}

	payments, err := h.paymentRepo.FindByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments by invoice: %w", err)
	}

	paymentDTOs := make([]PaymentDTO, 0, len(payments))
	for _, payment := range payments {
		paymentDTOs = append(paymentDTOs, *h.toPaymentDTO(payment))
	}

	return paymentDTOs, nil
}

// GetPaymentsByStatus handles the GetPaymentsByStatusQuery
func (h *PaymentQueryHandler) GetPaymentsByStatus(ctx context.Context, query GetPaymentsByStatusQuery) (*PaymentListResult, error) {
	status, err := valueobject.NewPaymentStatus(query.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid payment status: %w", err)
	}

	payments, err := h.paymentRepo.FindByStatus(ctx, status, query.Limit, query.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments by status: %w", err)
	}

	totalCount, err := h.paymentRepo.CountByStatus(ctx, status)
	if err != nil {
		return nil, fmt.Errorf("failed to count payments by status: %w", err)
	}

	paymentDTOs := make([]PaymentDTO, 0, len(payments))
	for _, payment := range payments {
		paymentDTOs = append(paymentDTOs, *h.toPaymentDTO(payment))
	}

	hasMore := false
	if query.Limit > 0 {
		hasMore = int64(query.Offset+query.Limit) < totalCount
	}

	return &PaymentListResult{
		Payments:   paymentDTOs,
		TotalCount: totalCount,
		Limit:      query.Limit,
		Offset:     query.Offset,
		HasMore:    hasMore,
	}, nil
}

// GetExpiredPayments handles the GetExpiredPaymentsQuery
func (h *PaymentQueryHandler) GetExpiredPayments(ctx context.Context, query GetExpiredPaymentsQuery) ([]PaymentDTO, error) {
	payments, err := h.paymentRepo.FindExpiredPayments(ctx, query.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find expired payments: %w", err)
	}

	paymentDTOs := make([]PaymentDTO, 0, len(payments))
	for _, payment := range payments {
		paymentDTOs = append(paymentDTOs, *h.toPaymentDTO(payment))
	}

	return paymentDTOs, nil
}

// GetPaymentStats handles the GetPaymentStatsQuery
func (h *PaymentQueryHandler) GetPaymentStats(ctx context.Context, query GetPaymentStatsQuery) (*PaymentStatsResult, error) {
	filters := repository.PaymentFilters{
		DateRange: &repository.DateRange{
			Start: &query.StartDate,
			End:   &query.EndDate,
		},
	}

	if query.UserID != nil {
		userID, err := sharedvo.NewUserIDFromUint(*query.UserID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %w", err)
		}
		filters.UserID = &userID
	}

	payments, _, err := h.paymentRepo.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments for stats: %w", err)
	}

	return h.calculateStats(payments, query.GroupBy), nil
}

// Helper methods

// buildPaymentFilters builds repository filters from query
func (h *PaymentQueryHandler) buildPaymentFilters(query ListPaymentsQuery) repository.PaymentFilters {
	filters := repository.PaymentFilters{
		Search:    query.Search,
		SortBy:    query.SortBy,
		SortOrder: query.SortOrder,
		Limit:     query.Limit,
		Offset:    query.Offset,
	}

	if query.UserID != nil {
		if userID, err := sharedvo.NewUserIDFromUint(*query.UserID); err == nil {
			filters.UserID = &userID
		}
	}

	if query.InvoiceID != nil {
		if invoiceID, err := sharedvo.NewInvoiceID(*query.InvoiceID); err == nil {
			filters.InvoiceID = &invoiceID
		}
	}

	if query.Status != "" {
		if status, err := valueobject.NewPaymentStatus(query.Status); err == nil {
			filters.Status = &status
		}
	}

	if query.PaymentGateway != "" {
		if gateway, err := valueobject.NewPaymentGateway(query.PaymentGateway); err == nil {
			filters.PaymentGateway = &gateway
		}
	}

	if query.PaymentMethod != "" {
		if method, err := valueobject.NewPaymentMethod(query.PaymentMethod); err == nil {
			filters.PaymentMethod = &method
		}
	}

	if query.Currency != "" {
		if currency, err := valueobject.NewCurrency(query.Currency); err == nil {
			filters.Currency = &currency
		}
	}

	if query.MinAmount != nil || query.MaxAmount != nil {
		amountRange := &repository.AmountRange{}
		
		if query.MinAmount != nil && query.Currency != "" {
			if currency, err := valueobject.NewCurrency(query.Currency); err == nil {
				if minAmount, err := valueobject.NewMoney(*query.MinAmount, currency); err == nil {
					amountRange.Min = &minAmount
				}
			}
		}
		
		if query.MaxAmount != nil && query.Currency != "" {
			if currency, err := valueobject.NewCurrency(query.Currency); err == nil {
				if maxAmount, err := valueobject.NewMoney(*query.MaxAmount, currency); err == nil {
					amountRange.Max = &maxAmount
				}
			}
		}
		
		filters.AmountRange = amountRange
	}

	if query.StartDate != "" || query.EndDate != "" {
		filters.DateRange = &repository.DateRange{
			Start: &query.StartDate,
			End:   &query.EndDate,
		}
	}

	return filters
}

// toPaymentDTO converts Payment aggregate to PaymentDTO
func (h *PaymentQueryHandler) toPaymentDTO(payment *aggregate.Payment) *PaymentDTO {
	refundableAmount, _ := payment.GetRefundableAmount()
	netAmount, _ := payment.GetNetAmount()

	return &PaymentDTO{
		ID:               payment.ID().Value(),
		PaymentNumber:    payment.PaymentNumber().Value(),
		InvoiceID:        payment.InvoiceID().Value(),
		UserID:           payment.UserID().ToUint(),
		Status:           payment.Status().Value(),
		Amount:           payment.Amount().Amount(),
		Currency:         payment.Amount().Currency().Code(),
		PaymentMethod:    payment.PaymentMethod().Value(),
		PaymentGateway:   payment.PaymentGateway().Value(),
		PaymentIntentID:  payment.PaymentIntentID(),
		GatewayTransactionID: payment.GatewayTransactionID(),
		GatewayFee:       payment.GatewayFee().Amount(),
		PaymentURL:       payment.PaymentURL(),
		QRCodeURL:        payment.QRCodeURL(),
		RedirectURL:      payment.RedirectURL(),
		ExpiresAt:        payment.ExpiresAt(),
		ProcessedAt:      payment.ProcessedAt(),
		CompletedAt:      payment.CompletedAt(),
		RefundAmount:     payment.RefundAmount().Amount(),
		RefundedAt:       payment.RefundedAt(),
		RefundReason:     payment.RefundReason(),
		RefundReference:  payment.RefundReference().Value(),
		NotificationCount: payment.NotificationCount(),
		LastNotificationAt: payment.LastNotificationAt(),
		Notes:            payment.Notes(),
		Metadata:         payment.Metadata(),
		CreatedAt:        payment.CreatedAt(),
		UpdatedAt:        payment.UpdatedAt(),
		
		// Computed fields
		IsExpired:        payment.IsExpired(),
		CanRefund:        payment.CanBeRefunded(),
		RefundableAmount: refundableAmount.Amount(),
		NetAmount:        netAmount.Amount(),
		StatusDisplay:    h.getStatusDisplay(payment.Status()),
		MethodDisplay:    payment.PaymentMethod().GetDisplayName(),
		GatewayDisplay:   payment.PaymentGateway().GetDisplayName(),
	}
}

// getStatusDisplay returns a human-readable status display
func (h *PaymentQueryHandler) getStatusDisplay(status valueobject.PaymentStatus) string {
	statusDisplayMap := map[string]string{
		"pending":    "待支付",
		"processing": "处理中",
		"completed":  "已完成",
		"failed":     "失败",
		"cancelled":  "已取消",
	}
	
	if display, exists := statusDisplayMap[status.Value()]; exists {
		return display
	}
	
	return status.Value()
}

// calculateStats calculates payment statistics
func (h *PaymentQueryHandler) calculateStats(payments []*aggregate.Payment, groupBy string) *PaymentStatsResult {
	stats := &PaymentStatsResult{
		StatusBreakdown:   make(map[string]int64),
		MethodBreakdown:   make(map[string]int64),
		GatewayBreakdown:  make(map[string]int64),
		CurrencyBreakdown: make(map[string]float64),
	}

	for _, payment := range payments {
		stats.TotalPayments++
		stats.TotalAmount += payment.Amount().Amount()
		stats.TotalRefunded += payment.RefundAmount().Amount()
		
		// Status breakdown
		stats.StatusBreakdown[payment.Status().Value()]++
		
		// Method breakdown
		stats.MethodBreakdown[payment.PaymentMethod().Value()]++
		
		// Gateway breakdown
		stats.GatewayBreakdown[payment.PaymentGateway().Value()]++
		
		// Currency breakdown
		currency := payment.Amount().Currency().Code()
		stats.CurrencyBreakdown[currency] += payment.Amount().Amount()
	}

	stats.NetAmount = stats.TotalAmount - stats.TotalRefunded

	// TODO: Implement time series data based on groupBy parameter
	// This would require grouping payments by time periods (day, week, month, year)

	return stats
}