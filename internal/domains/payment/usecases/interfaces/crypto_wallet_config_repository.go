package interfaces

import (
	"context"
	"linke/internal/domains/payment/entities"
)

// CryptoWalletConfigRepository defines the interface for crypto wallet config data access operations
type CryptoWalletConfigRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, config *entities.CryptoWalletConfig) error
	GetByID(ctx context.Context, id uint) (*entities.CryptoWalletConfig, error)
	GetByWalletAddress(ctx context.Context, address string) (*entities.CryptoWalletConfig, error)
	Update(ctx context.Context, config *entities.CryptoWalletConfig) error
	Delete(ctx context.Context, id uint) error

	// Soft delete operations
	SoftDelete(ctx context.Context, id uint) error

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.CryptoWalletConfig, int64, error)
	ListActive(ctx context.Context, limit, offset int) ([]*entities.CryptoWalletConfig, int64, error)

	// Status filtering
	ListByStatus(ctx context.Context, isEnabled bool, limit, offset int) ([]*entities.CryptoWalletConfig, int64, error)

	// Network and currency operations
	ListByNetwork(ctx context.Context, network string, limit, offset int) ([]*entities.CryptoWalletConfig, int64, error)
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.CryptoWalletConfig, int64, error)
	ListByNetworkAndCurrency(ctx context.Context, network, currency string, limit, offset int) ([]*entities.CryptoWalletConfig, int64, error)

	// Get enabled configs for payment processing
	GetEnabledByNetwork(ctx context.Context, network string) ([]*entities.CryptoWalletConfig, error)
	GetEnabledByCurrency(ctx context.Context, currency string) ([]*entities.CryptoWalletConfig, error)
	GetEnabledByNetworkAndCurrency(ctx context.Context, network, currency string) ([]*entities.CryptoWalletConfig, error)

	// Payment method operations
	GetByPaymentMethod(ctx context.Context, paymentMethod string) ([]*entities.CryptoWalletConfig, error)
	GetEnabledByPaymentMethod(ctx context.Context, paymentMethod string) ([]*entities.CryptoWalletConfig, error)

	// Amount validation operations
	GetAvailableForPayment(ctx context.Context, network, currency string, amount float64) ([]*entities.CryptoWalletConfig, error)

	// Status management
	UpdateStatus(ctx context.Context, id uint, isEnabled bool) error
	UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountEnabled(ctx context.Context) (int64, error)
	CountDisabled(ctx context.Context) (int64, error)

	// Existence checks
	ExistsByWalletAddress(ctx context.Context, address string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
}