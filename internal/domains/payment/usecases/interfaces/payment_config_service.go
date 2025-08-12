package interfaces

import (
	"context"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
)

// PaymentConfigService defines a simplified interface for payment configuration operations
// Consolidates validation and preparation logic
type PaymentConfigService interface {
	// Basic CRUD operations
	CreatePaymentConfig(ctx context.Context, req *dto.CreatePaymentConfigRequest) (*entities.PaymentConfig, error)
	GetPaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error)
	GetPaymentConfigByMethod(ctx context.Context, method string) (*entities.PaymentConfig, error)
	UpdatePaymentConfig(ctx context.Context, configID uint, req *dto.UpdatePaymentConfigRequest) (*entities.PaymentConfig, error)
	DeletePaymentConfig(ctx context.Context, configID uint) error

	// Config listing with flexible filtering
	GetPaymentConfigs(ctx context.Context, req *dto.GetPaymentConfigsRequest) ([]*entities.PaymentConfig, int64, error)
	GetActivePaymentConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error)
	GetPaymentConfigsByMethod(ctx context.Context, method string) ([]*entities.PaymentConfig, error)

	// Config management
	TogglePaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error)
	
	// Factory methods (for gateway initialization)
	GetEnabledConfigs() ([]*entities.PaymentConfig, error)
	GetConfigByMethod(method string) (*entities.PaymentConfig, error)
	
	// Validation - consolidated validation logic
	ValidateCreatePaymentConfig(ctx context.Context, req *dto.CreatePaymentConfigRequest) []string
	ValidateUpdatePaymentConfig(ctx context.Context, configID uint, req *dto.UpdatePaymentConfigRequest) []string
}

