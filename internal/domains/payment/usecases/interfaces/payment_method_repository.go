package interfaces

import (
	"context"

	"linke/internal/domains/payment/entities"
)

// PaymentMethodRepository defines the interface for payment method data access
type PaymentMethodRepository interface {
	// Create creates a new payment method for a user
	Create(ctx context.Context, paymentMethod *entities.PaymentMethod) error

	// GetByID retrieves a payment method by its ID
	GetByID(ctx context.Context, id uint) (*entities.PaymentMethod, error)

	// GetByUserID retrieves all payment methods for a user
	GetByUserID(ctx context.Context, userID uint) ([]*entities.PaymentMethod, error)

	// GetActiveByUserID retrieves all active payment methods for a user
	GetActiveByUserID(ctx context.Context, userID uint) ([]*entities.PaymentMethod, error)

	// GetByUserIDAndGateway retrieves payment methods for a user by gateway
	GetByUserIDAndGateway(ctx context.Context, userID uint, gateway string) ([]*entities.PaymentMethod, error)

	// GetDefaultByUserID retrieves the default payment method for a user
	GetDefaultByUserID(ctx context.Context, userID uint) (*entities.PaymentMethod, error)

	// GetDefaultByUserIDAndGateway retrieves the default payment method for a user and gateway
	GetDefaultByUserIDAndGateway(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethod, error)

	// GetByPaymentToken retrieves a payment method by its payment token and gateway
	GetByPaymentToken(ctx context.Context, gateway, token string) (*entities.PaymentMethod, error)

	// Update updates an existing payment method
	Update(ctx context.Context, paymentMethod *entities.PaymentMethod) error

	// UpdateStatus updates the status of a payment method
	UpdateStatus(ctx context.Context, id uint, status string) error

	// SetAsDefault sets a payment method as the default for the user and gateway
	SetAsDefault(ctx context.Context, userID, paymentMethodID uint) error

	// UnsetDefault removes default status from a payment method
	UnsetDefault(ctx context.Context, userID uint, gateway string) error

	// Delete soft deletes a payment method
	Delete(ctx context.Context, id uint) error

	// HardDelete permanently deletes a payment method
	HardDelete(ctx context.Context, id uint) error

	// GetExpiredMethods retrieves payment methods that have expired
	GetExpiredMethods(ctx context.Context) ([]*entities.PaymentMethod, error)

	// GetMethodsNeedingValidation retrieves payment methods that need revalidation
	GetMethodsNeedingValidation(ctx context.Context) ([]*entities.PaymentMethod, error)

	// UpdateLastUsed updates the last used timestamp and usage statistics
	UpdateLastUsed(ctx context.Context, id uint, successful bool) error

	// GetUserPaymentMethodCount returns the count of payment methods for a user
	GetUserPaymentMethodCount(ctx context.Context, userID uint) (int64, error)

	// GetHighFailureRateMethods returns payment methods with high failure rates
	GetHighFailureRateMethods(ctx context.Context, threshold float64) ([]*entities.PaymentMethod, error)

	// ValidateOwnership checks if a payment method belongs to a specific user
	ValidateOwnership(ctx context.Context, paymentMethodID, userID uint) (bool, error)
}
