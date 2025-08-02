package interfaces

import (
	"context"
	"linke/internal/domains/payment/entities"
)

// PaymentConfigService defines the interface for payment configuration operations
type PaymentConfigService interface {
	// Config CRUD operations
	CreatePaymentConfig(ctx context.Context, req *CreatePaymentConfigRequest) (*entities.PaymentConfig, error)
	GetPaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error)
	GetPaymentConfigByGateway(ctx context.Context, gateway string) (*entities.PaymentConfig, error)
	UpdatePaymentConfig(ctx context.Context, configID uint, req *UpdatePaymentConfigRequest) (*entities.PaymentConfig, error)
	DeletePaymentConfig(ctx context.Context, configID uint) error

	// Config listing and filtering
	GetPaymentConfigs(ctx context.Context, req *GetPaymentConfigsRequest) ([]*entities.PaymentConfig, int64, error)
	GetActivePaymentConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error)
	GetPaymentConfigsByGateway(ctx context.Context, gateway string) ([]*entities.PaymentConfig, error)

	// Config management
	TogglePaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error)
}

// CreatePaymentConfigRequest represents the request to create a payment config
type CreatePaymentConfigRequest struct {
	Gateway             string            `json:"gateway" binding:"required" example:"epay"`
	Name                string            `json:"name" binding:"required" example:"EPay Gateway"`
	Config              string            `json:"config" binding:"required" example:"{\"api_url\":\"...\"}"`
	IsEnabled           *bool             `json:"is_enabled,omitempty" example:"true"`
	SortOrder           int               `json:"sort_order,omitempty" example:"1"`
	SupportedCurrencies string            `json:"supported_currencies,omitempty" example:"CNY"`
	Methods             []entities.Method `json:"methods,omitempty"`
	MinAmount           float64           `json:"min_amount,omitempty" example:"0.01"`
	MaxAmount           float64           `json:"max_amount,omitempty" example:"99999.99"`
	FixedFee            float64           `json:"fixed_fee,omitempty" example:"0.00"`
	PercentageFee       float64           `json:"percentage_fee,omitempty" example:"0.6"`
}

// UpdatePaymentConfigRequest represents the request to update a payment config
type UpdatePaymentConfigRequest struct {
	Name                *string           `json:"name,omitempty" example:"EPay Gateway"`
	Config              *string           `json:"config,omitempty" example:"{\"api_url\":\"...\"}"`
	IsEnabled           *bool             `json:"is_enabled,omitempty" example:"true"`
	SortOrder           *int              `json:"sort_order,omitempty" example:"1"`
	SupportedCurrencies *string           `json:"supported_currencies,omitempty" example:"CNY"`
	Methods             []entities.Method `json:"methods,omitempty"`
	MinAmount           *float64          `json:"min_amount,omitempty" example:"0.01"`
	MaxAmount           *float64          `json:"max_amount,omitempty" example:"99999.99"`
	FixedFee            *float64          `json:"fixed_fee,omitempty" example:"0.00"`
	PercentageFee       *float64          `json:"percentage_fee,omitempty" example:"0.6"`
}

// GetPaymentConfigsRequest represents the request to get payment configs
type GetPaymentConfigsRequest struct {
	Gateway   string `form:"gateway,omitempty" example:"epay"`
	IsEnabled *bool  `form:"is_enabled,omitempty" example:"true"`
	Limit     int    `form:"limit,omitempty" example:"10"`
	Offset    int    `form:"offset,omitempty" example:"0"`
}