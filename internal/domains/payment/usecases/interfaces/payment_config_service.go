package interfaces

import (
	"context"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
)

// PaymentConfigService defines the interface for payment configuration operations
type PaymentConfigService interface {
	// Config CRUD operations
	CreatePaymentConfig(ctx context.Context, req *dto.CreatePaymentConfigRequest) (*entities.PaymentConfig, error)
	GetPaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error)
	GetPaymentConfigByMethod(ctx context.Context, method string) (*entities.PaymentConfig, error)
	UpdatePaymentConfig(ctx context.Context, configID uint, req *dto.UpdatePaymentConfigRequest) (*entities.PaymentConfig, error)
	DeletePaymentConfig(ctx context.Context, configID uint) error

	// Config listing and filtering
	GetPaymentConfigs(ctx context.Context, req *dto.GetPaymentConfigsRequest) ([]*entities.PaymentConfig, int64, error)
	GetActivePaymentConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error)
	GetPaymentConfigsByMethod(ctx context.Context, method string) ([]*entities.PaymentConfig, error)

	// Config management
	TogglePaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error)
}

