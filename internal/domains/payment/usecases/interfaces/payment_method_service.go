package interfaces

import (
	"context"

	"linke/internal/domains/payment/entities"
)

// PaymentMethodService defines the interface for payment method business logic
type PaymentMethodService interface {
	// CreatePaymentMethod creates a new payment method for a user
	CreatePaymentMethod(ctx context.Context, userID uint, req *entities.CreatePaymentMethodRequest) (*entities.PaymentMethodResponse, error)

	// GetPaymentMethod retrieves a payment method by ID for a specific user
	GetPaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*entities.PaymentMethodResponse, error)

	// ListPaymentMethods retrieves all payment methods for a user
	ListPaymentMethods(ctx context.Context, userID uint) (*entities.PaymentMethodListResponse, error)

	// ListActivePaymentMethods retrieves all active payment methods for a user
	ListActivePaymentMethods(ctx context.Context, userID uint) (*entities.PaymentMethodListResponse, error)

	// ListPaymentMethodsByGateway retrieves payment methods for a user by gateway
	ListPaymentMethodsByGateway(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethodListResponse, error)

	// UpdatePaymentMethod updates an existing payment method
	UpdatePaymentMethod(ctx context.Context, userID, paymentMethodID uint, req *entities.UpdatePaymentMethodRequest) (*entities.PaymentMethodResponse, error)

	// SetDefaultPaymentMethod sets a payment method as the default for the user
	SetDefaultPaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*entities.PaymentMethodResponse, error)

	// DeletePaymentMethod soft deletes a payment method
	DeletePaymentMethod(ctx context.Context, userID, paymentMethodID uint) error

	// GetDefaultPaymentMethod retrieves the default payment method for a user
	GetDefaultPaymentMethod(ctx context.Context, userID uint) (*entities.PaymentMethodResponse, error)

	// GetDefaultPaymentMethodByGateway retrieves the default payment method for a user and gateway
	GetDefaultPaymentMethodByGateway(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethodResponse, error)

	// ValidatePaymentMethod validates a payment method with the gateway
	ValidatePaymentMethod(ctx context.Context, userID, paymentMethodID uint) (*entities.PaymentMethodResponse, error)

	// ProcessPaymentWithMethod processes a payment using a specific payment method
	ProcessPaymentWithMethod(ctx context.Context, userID, paymentMethodID uint, amount float64, currency string) (*entities.PaymentRecord, error)

	// ProcessPaymentWithDefaultMethod processes a payment using the default payment method
	ProcessPaymentWithDefaultMethod(ctx context.Context, userID uint, gateway string, amount float64, currency string) (*entities.PaymentRecord, error)

	// RetryFailedPayment retries a failed payment with alternative payment methods
	RetryFailedPayment(ctx context.Context, userID uint, paymentRecordID uint) (*entities.PaymentRecord, error)

	// GetPaymentMethodUsageStats retrieves usage statistics for a payment method
	GetPaymentMethodUsageStats(ctx context.Context, userID, paymentMethodID uint) (*PaymentMethodUsageStats, error)

	// RefreshExpiredMethods marks expired payment methods as expired
	RefreshExpiredMethods(ctx context.Context) error

	// RevalidatePaymentMethods revalidates payment methods that need validation
	RevalidatePaymentMethods(ctx context.Context) error

	// GetUserPaymentMethodsCount returns the count of payment methods for a user
	GetUserPaymentMethodsCount(ctx context.Context, userID uint) (int64, error)

	// IsPaymentMethodLimitReached checks if user has reached payment method limit
	IsPaymentMethodLimitReached(ctx context.Context, userID uint) (bool, error)

	// GetRecommendedPaymentMethod returns the best payment method for a user based on history
	GetRecommendedPaymentMethod(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethodResponse, error)
}

// PaymentMethodUsageStats represents usage statistics for a payment method
type PaymentMethodUsageStats struct {
	PaymentMethodID  uint    `json:"payment_method_id"`
	TotalUses        int     `json:"total_uses"`
	SuccessfulUses   int     `json:"successful_uses"`
	FailedUses       int     `json:"failed_uses"`
	SuccessRate      float64 `json:"success_rate"`
	FailureRate      float64 `json:"failure_rate"`
	LastUsed         *string `json:"last_used"`
	AverageAmount    float64 `json:"average_amount"`
	TotalAmount      float64 `json:"total_amount"`
	RecentUses30Days int     `json:"recent_uses_30_days"`
}

// PaymentMethodValidationResult represents the result of payment method validation
type PaymentMethodValidationResult struct {
	IsValid       bool   `json:"is_valid"`
	ErrorMessage  string `json:"error_message,omitempty"`
	GatewayCode   string `json:"gateway_code,omitempty"`
	LastValidated string `json:"last_validated"`
}
