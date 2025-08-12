package interfaces

import (
	"context"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
)

// CryptoWalletConfigService defines the interface for crypto wallet configuration operations
type CryptoWalletConfigService interface {
	// Config CRUD operations
	CreateCryptoWalletConfig(ctx context.Context, req *dto.CreateCryptoWalletConfigRequest) (*entities.CryptoWalletConfig, error)
	GetCryptoWalletConfig(ctx context.Context, configID uint) (*entities.CryptoWalletConfig, error)
	GetCryptoWalletConfigByAddress(ctx context.Context, address string) (*entities.CryptoWalletConfig, error)
	UpdateCryptoWalletConfig(ctx context.Context, configID uint, req *dto.UpdateCryptoWalletConfigRequest) (*entities.CryptoWalletConfig, error)
	DeleteCryptoWalletConfig(ctx context.Context, configID uint) error

	// Config listing and filtering
	GetCryptoWalletConfigs(ctx context.Context, req *dto.GetCryptoWalletConfigsRequest) ([]*entities.CryptoWalletConfig, int64, error)
	GetActiveCryptoWalletConfigs(ctx context.Context, network, currency string) ([]*entities.CryptoWalletConfig, error)
	GetCryptoWalletConfigsByNetwork(ctx context.Context, network string) ([]*entities.CryptoWalletConfig, error)
	GetCryptoWalletConfigsByCurrency(ctx context.Context, currency string) ([]*entities.CryptoWalletConfig, error)
	GetCryptoWalletConfigsByPaymentMethod(ctx context.Context, paymentMethod string) ([]*entities.CryptoWalletConfig, error)

	// Config management
	ToggleCryptoWalletConfig(ctx context.Context, configID uint) (*entities.CryptoWalletConfig, error)
	ActivateCryptoWalletConfig(ctx context.Context, configID uint) (*entities.CryptoWalletConfig, error)
	DeactivateCryptoWalletConfig(ctx context.Context, configID uint) (*entities.CryptoWalletConfig, error)
	UpdateConfigSortOrder(ctx context.Context, configID uint, sortOrder int) (*entities.CryptoWalletConfig, error)

	// Address validation and management
	ValidateWalletAddress(ctx context.Context, req *dto.ValidateCryptoWalletAddressRequest) (*dto.ValidateCryptoWalletAddressResponse, error)

	// Payment processing support
	GetAvailableConfigsForPayment(ctx context.Context, network, currency string, amount float64) ([]*entities.CryptoWalletConfig, error)
	GetBestConfigForPayment(ctx context.Context, network, currency string, amount float64) (*entities.CryptoWalletConfig, error)
}