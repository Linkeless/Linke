package interfaces

import (
	"context"

	"linke/internal/domains/payment/entities"
	"linke/internal/shared/framework"
)

// PaymentMethodRepository defines the interface for payment method data access
type PaymentMethodRepository interface {
	framework.UserScopedRepository[entities.PaymentMethod, uint]

	// User-specific queries
	GetByUserID(ctx context.Context, userID uint) ([]*entities.PaymentMethod, error)
	GetActiveByUserID(ctx context.Context, userID uint) ([]*entities.PaymentMethod, error)
	GetByUserIDAndGateway(ctx context.Context, userID uint, gateway string) ([]*entities.PaymentMethod, error)
	GetDefaultByUserID(ctx context.Context, userID uint) (*entities.PaymentMethod, error)
	GetDefaultByUserIDAndGateway(ctx context.Context, userID uint, gateway string) (*entities.PaymentMethod, error)
	GetUserPaymentMethodCount(ctx context.Context, userID uint) (int64, error)

	// Token-based queries
	GetByPaymentToken(ctx context.Context, gateway, token string) (*entities.PaymentMethod, error)

	// Default management
	SetAsDefault(ctx context.Context, userID, paymentMethodID uint) error
	UnsetDefault(ctx context.Context, userID uint, gateway string) error

	// Expiry and validation
	GetExpiredMethods(ctx context.Context) ([]*entities.PaymentMethod, error)
	GetMethodsNeedingValidation(ctx context.Context) ([]*entities.PaymentMethod, error)

	// Usage tracking
	UpdateLastUsed(ctx context.Context, id uint, successful bool) error
	GetHighFailureRateMethods(ctx context.Context, threshold float64) ([]*entities.PaymentMethod, error)

	// Ownership validation
	ValidateOwnership(ctx context.Context, paymentMethodID, userID uint) (bool, error)
}
