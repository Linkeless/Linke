package query

import (
	"time"
)

// GetPaymentQuery represents a query to get a single payment
type GetPaymentQuery struct {
	PaymentID uint `json:"payment_id" validate:"required"`
}

// GetPaymentByNumberQuery represents a query to get a payment by number
type GetPaymentByNumberQuery struct {
	PaymentNumber string `json:"payment_number" validate:"required"`
}

// ListPaymentsQuery represents a query to list payments with filters
type ListPaymentsQuery struct {
	UserID         *uint    `form:"user_id"`
	InvoiceID      *uint    `form:"invoice_id"`
	Status         string   `form:"status"`
	PaymentGateway string   `form:"payment_gateway"`
	PaymentMethod  string   `form:"payment_method"`
	Currency       string   `form:"currency"`
	MinAmount      *float64 `form:"min_amount"`
	MaxAmount      *float64 `form:"max_amount"`
	StartDate      string   `form:"start_date"`
	EndDate        string   `form:"end_date"`
	Search         string   `form:"search"`
	SortBy         string   `form:"sort_by"`
	SortOrder      string   `form:"sort_order"`
	Limit          int      `form:"limit"`
	Offset         int      `form:"offset"`
}

// GetPaymentsByUserQuery represents a query to get payments by user
type GetPaymentsByUserQuery struct {
	UserID uint `json:"user_id" validate:"required"`
	Limit  int  `form:"limit"`
	Offset int  `form:"offset"`
}

// GetPaymentsByInvoiceQuery represents a query to get payments by invoice
type GetPaymentsByInvoiceQuery struct {
	InvoiceID uint `json:"invoice_id" validate:"required"`
}

// GetPaymentsByStatusQuery represents a query to get payments by status
type GetPaymentsByStatusQuery struct {
	Status string `form:"status" validate:"required"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// GetExpiredPaymentsQuery represents a query to get expired payments
type GetExpiredPaymentsQuery struct {
	Limit int `form:"limit"`
}

// GetPaymentStatsQuery represents a query to get payment statistics
type GetPaymentStatsQuery struct {
	UserID    *uint  `form:"user_id"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	GroupBy   string `form:"group_by"` // day, week, month, year
}

// PaymentDTO represents a payment data transfer object
type PaymentDTO struct {
	ID                   uint       `json:"id"`
	PaymentNumber        string     `json:"payment_number"`
	InvoiceID            uint       `json:"invoice_id"`
	UserID               uint       `json:"user_id"`
	Status               string     `json:"status"`
	Amount               float64    `json:"amount"`
	Currency             string     `json:"currency"`
	PaymentMethod        string     `json:"payment_method"`
	PaymentGateway       string     `json:"payment_gateway"`
	PaymentIntentID      string     `json:"payment_intent_id,omitempty"`
	GatewayTransactionID string     `json:"gateway_transaction_id,omitempty"`
	GatewayFee           float64    `json:"gateway_fee"`
	PaymentURL           string     `json:"payment_url,omitempty"`
	QRCodeURL            string     `json:"qr_code_url,omitempty"`
	RedirectURL          string     `json:"redirect_url,omitempty"`
	ExpiresAt            *time.Time `json:"expires_at,omitempty"`
	ProcessedAt          *time.Time `json:"processed_at,omitempty"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	RefundAmount         float64    `json:"refund_amount"`
	RefundedAt           *time.Time `json:"refunded_at,omitempty"`
	RefundReason         string     `json:"refund_reason,omitempty"`
	RefundReference      string     `json:"refund_reference,omitempty"`
	NotificationCount    int        `json:"notification_count"`
	LastNotificationAt   *time.Time `json:"last_notification_at,omitempty"`
	Notes                string     `json:"notes,omitempty"`
	Metadata             string     `json:"metadata,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Computed fields
	IsExpired        bool    `json:"is_expired"`
	CanRefund        bool    `json:"can_refund"`
	RefundableAmount float64 `json:"refundable_amount"`
	NetAmount        float64 `json:"net_amount"`
	StatusDisplay    string  `json:"status_display"`
	MethodDisplay    string  `json:"method_display"`
	GatewayDisplay   string  `json:"gateway_display"`
}

// PaymentListResult represents the result of a payment list query
type PaymentListResult struct {
	Payments   []PaymentDTO `json:"payments"`
	TotalCount int64        `json:"total_count"`
	Limit      int          `json:"limit"`
	Offset     int          `json:"offset"`
	HasMore    bool         `json:"has_more"`
}

// PaymentStatsResult represents payment statistics
type PaymentStatsResult struct {
	TotalPayments     int64              `json:"total_payments"`
	TotalAmount       float64            `json:"total_amount"`
	TotalRefunded     float64            `json:"total_refunded"`
	NetAmount         float64            `json:"net_amount"`
	StatusBreakdown   map[string]int64   `json:"status_breakdown"`
	MethodBreakdown   map[string]int64   `json:"method_breakdown"`
	GatewayBreakdown  map[string]int64   `json:"gateway_breakdown"`
	CurrencyBreakdown map[string]float64 `json:"currency_breakdown"`
	TimeSeriesData    []TimeSeriesPoint  `json:"time_series_data,omitempty"`
}

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	PaymentCount int64     `json:"payment_count"`
	Amount       float64   `json:"amount"`
	RefundAmount float64   `json:"refund_amount"`
}
