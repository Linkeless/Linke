package repository

import (
	"context"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/valueobject"
)

// PaymentConfigRepository defines the interface for payment config persistence operations
type PaymentConfigRepository interface {
	// Save saves a payment config aggregate
	Save(ctx context.Context, config *aggregate.PaymentConfig) error
	
	// Update updates a payment config aggregate
	Update(ctx context.Context, config *aggregate.PaymentConfig) error
	
	// FindByID finds a payment config by its ID
	FindByID(ctx context.Context, id valueobject.PaymentConfigID) (*aggregate.PaymentConfig, error)
	
	// FindByGateway finds a payment config by gateway
	FindByGateway(ctx context.Context, gateway valueobject.PaymentGateway) (*aggregate.PaymentConfig, error)
	
	// FindAll finds all payment configs
	FindAll(ctx context.Context) ([]*aggregate.PaymentConfig, error)
	
	// FindActive finds all active payment configs
	FindActive(ctx context.Context) ([]*aggregate.PaymentConfig, error)
	
	// FindByCurrency finds payment configs that support a specific currency
	FindByCurrency(ctx context.Context, currency valueobject.Currency) ([]*aggregate.PaymentConfig, error)
	
	// FindByMethod finds payment configs that support a specific payment method
	FindByMethod(ctx context.Context, method valueobject.PaymentMethod) ([]*aggregate.PaymentConfig, error)
	
	// FindWithFilters finds payment configs with filters
	FindWithFilters(ctx context.Context, filters PaymentConfigFilters) ([]*aggregate.PaymentConfig, int64, error)
	
	// Delete soft deletes a payment config
	Delete(ctx context.Context, id valueobject.PaymentConfigID) error
	
	// Exists checks if a payment config exists by ID
	Exists(ctx context.Context, id valueobject.PaymentConfigID) (bool, error)
	
	// ExistsByGateway checks if a payment config exists by gateway
	ExistsByGateway(ctx context.Context, gateway valueobject.PaymentGateway) (bool, error)
	
	// Count returns the total count of payment configs
	Count(ctx context.Context) (int64, error)
	
	// CountActive returns the count of active payment configs
	CountActive(ctx context.Context) (int64, error)
}

// PaymentConfigFilters represents the filters for payment config queries
type PaymentConfigFilters struct {
	Gateway   *valueobject.PaymentGateway
	IsEnabled *bool
	Currency  *valueobject.Currency
	Method    *valueobject.PaymentMethod
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}