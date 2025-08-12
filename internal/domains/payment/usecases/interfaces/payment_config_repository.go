package interfaces

import (
	"context"
	"linke/internal/domains/payment/entities"
	"linke/internal/shared/framework"
)

// PaymentConfigFilter provides flexible filtering options for payment configs
type PaymentConfigFilter struct {
	Enabled    *bool
	Currency   string
	Method     string
	Gateway    string
	MinAmount  *float64
	MaxAmount  *float64
	SearchQuery string
}

// PaymentConfigRepository defines a simplified interface for payment config data access
// Following YAGNI principle - only include methods that are actually needed
type PaymentConfigRepository interface {
	framework.GenericRepository[entities.PaymentConfig, uint]

	// Core business queries
	GetByMethod(ctx context.Context, method string) (*entities.PaymentConfig, error)
	ListWithFilter(ctx context.Context, filter PaymentConfigFilter, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	GetEnabledConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error)

	// Existence checks - extends built-in ExistsByID
	ExistsByMethod(ctx context.Context, method string) (bool, error)
}
