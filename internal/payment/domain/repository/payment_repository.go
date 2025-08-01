package repository

import (
	"context"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/valueobject"
	sharedvo "linke/internal/shared/valueobject"
)

// PaymentRepository defines the interface for payment persistence operations
type PaymentRepository interface {
	// Save saves a payment aggregate
	Save(ctx context.Context, payment *aggregate.Payment) error
	
	// Update updates a payment aggregate
	Update(ctx context.Context, payment *aggregate.Payment) error
	
	// FindByID finds a payment by its ID
	FindByID(ctx context.Context, id valueobject.PaymentID) (*aggregate.Payment, error)
	
	// FindByPaymentNumber finds a payment by its payment number
	FindByPaymentNumber(ctx context.Context, paymentNumber valueobject.PaymentNumber) (*aggregate.Payment, error)
	
	// FindByInvoiceID finds payments by invoice ID
	FindByInvoiceID(ctx context.Context, invoiceID sharedvo.InvoiceID) ([]*aggregate.Payment, error)
	
	// FindByUserID finds payments by user ID with pagination
	FindByUserID(ctx context.Context, userID sharedvo.UserID, limit, offset int) ([]*aggregate.Payment, error)
	
	// FindByStatus finds payments by status with pagination
	FindByStatus(ctx context.Context, status valueobject.PaymentStatus, limit, offset int) ([]*aggregate.Payment, error)
	
	// FindByGatewayTransactionID finds a payment by gateway transaction ID
	FindByGatewayTransactionID(ctx context.Context, transactionID string) (*aggregate.Payment, error)
	
	// FindExpiredPayments finds payments that have expired
	FindExpiredPayments(ctx context.Context, limit int) ([]*aggregate.Payment, error)
	
	// FindWithFilters finds payments with complex filters
	FindWithFilters(ctx context.Context, filters PaymentFilters) ([]*aggregate.Payment, int64, error)
	
	// Delete soft deletes a payment
	Delete(ctx context.Context, id valueobject.PaymentID) error
	
	// Exists checks if a payment exists by ID
	Exists(ctx context.Context, id valueobject.PaymentID) (bool, error)
	
	// ExistsByPaymentNumber checks if a payment exists by payment number
	ExistsByPaymentNumber(ctx context.Context, paymentNumber valueobject.PaymentNumber) (bool, error)
	
	// Count returns the total count of payments
	Count(ctx context.Context) (int64, error)
	
	// CountByStatus returns the count of payments by status
	CountByStatus(ctx context.Context, status valueobject.PaymentStatus) (int64, error)
	
	// CountByUserID returns the count of payments by user ID
	CountByUserID(ctx context.Context, userID sharedvo.UserID) (int64, error)
}

// PaymentFilters represents the filters for payment queries
type PaymentFilters struct {
	UserID         *sharedvo.UserID
	InvoiceID      *sharedvo.InvoiceID
	Status         *valueobject.PaymentStatus
	PaymentGateway *valueobject.PaymentGateway
	PaymentMethod  *valueobject.PaymentMethod
	Currency       *valueobject.Currency
	AmountRange    *AmountRange
	DateRange      *DateRange
	Search         string
	SortBy         string
	SortOrder      string
	Limit          int
	Offset         int
}

// AmountRange represents an amount range filter
type AmountRange struct {
	Min *valueobject.Money
	Max *valueobject.Money
}

// DateRange represents a date range filter
type DateRange struct {
	Start *string // ISO 8601 date string
	End   *string // ISO 8601 date string
}