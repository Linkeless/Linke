package constants

// Order Type Constants
const (
	OrderTypeNew       = "new"
	OrderTypeRenewal   = "renewal"
	OrderTypeUpgrade   = "upgrade"
	OrderTypeDowngrade = "downgrade"
)

// Order Status Constants
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusPaid      = "paid"
	OrderStatusFailed    = "failed"
	OrderStatusCancelled = "cancelled"
	OrderStatusRefunded  = "refunded"
)

// Payment Status Constants
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
	PaymentStatusFailed     = "failed"
	PaymentStatusCancelled  = "cancelled"
	PaymentStatusRefunded   = "refunded"
)

// Invoice Status Constants
const (
	InvoiceStatusPending = "pending"
	InvoiceStatusSent    = "sent"
	InvoiceStatusPaid    = "paid"
	InvoiceStatusOverdue = "overdue"
	InvoiceStatusVoided  = "voided"
)

// Discount Type Constants
const (
	DiscountTypePercentage = "percentage"
	DiscountTypeFixed      = "fixed"
)
