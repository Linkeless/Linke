package interfaces

import (
	"context"

	"linke/internal/domains/payment/dto"
)

// PaymentMethodService defines a simplified interface for payment method business logic
// Focused on core payment method management operations
type PaymentMethodService interface {
	// Basic CRUD operations
	CreatePaymentMethod(ctx context.Context, userID uint, req *dto.CreatePaymentMethodRequest) (*dto.PaymentMethodResponse, error)
	GetPaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*dto.PaymentMethodResponse, error)
	UpdatePaymentMethod(ctx context.Context, userID, paymentMethodID uint, req *dto.UpdatePaymentMethodRequest) (*dto.PaymentMethodResponse, error)
	DeletePaymentMethod(ctx context.Context, userID, paymentMethodID uint) error

	// List operations
	ListPaymentMethods(ctx context.Context, userID uint) (*dto.PaymentMethodListResponse, error)
	ListActivePaymentMethods(ctx context.Context, userID uint) (*dto.PaymentMethodListResponse, error)
	ListPaymentMethodsByGateway(ctx context.Context, userID uint, gateway string) (*dto.PaymentMethodListResponse, error)

	// Default payment method management
	GetDefaultPaymentMethod(ctx context.Context, userID uint) (*dto.PaymentMethodResponse, error)
	GetDefaultPaymentMethodByGateway(ctx context.Context, userID uint, gateway string) (*dto.PaymentMethodResponse, error)
	SetDefaultPaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*dto.PaymentMethodResponse, error)

	// Validation
	ValidatePaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*dto.PaymentMethodResponse, error)

	// Statistics (keep for backward compatibility but simplified)
	GetPaymentMethodUsageStats(ctx context.Context, userID, paymentMethodID uint) (*dto.PaymentMethodUsageStats, error)

	// Business logic helpers
	GetUserPaymentMethodsCount(ctx context.Context, userID uint) (int64, error)
	IsPaymentMethodLimitReached(ctx context.Context, userID uint) (bool, error)

	// Maintenance operations
	RefreshExpiredMethods(ctx context.Context) error
	RevalidatePaymentMethods(ctx context.Context) error
}
