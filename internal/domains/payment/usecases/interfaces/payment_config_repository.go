package interfaces

import (
	"context"
	"linke/internal/domains/payment/entities"
)

// PaymentConfigRepository defines the interface for payment config data access operations
type PaymentConfigRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, config *entities.PaymentConfig) error
	GetByID(ctx context.Context, id uint) (*entities.PaymentConfig, error)
	GetByGateway(ctx context.Context, gateway string) (*entities.PaymentConfig, error)
	Update(ctx context.Context, config *entities.PaymentConfig) error
	Delete(ctx context.Context, id uint) error

	// Soft delete operations
	SoftDelete(ctx context.Context, id uint) error
	Restore(ctx context.Context, id uint) error
	HardDelete(ctx context.Context, id uint) error

	// List operations with pagination
	List(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	ListActive(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	ListDeleted(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error)

	// Status filtering
	ListByStatus(ctx context.Context, isEnabled bool, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	ListEnabled(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	ListDisabled(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error)

	// Currency support operations
	ListByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	ListSupportingCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	GetEnabledByCurrency(ctx context.Context, currency string) ([]*entities.PaymentConfig, error)

	// Method support operations
	ListByMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	ListSupportingMethod(ctx context.Context, paymentMethod string, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	GetEnabledByMethod(ctx context.Context, paymentMethod string) ([]*entities.PaymentConfig, error)

	// Combined filtering
	GetEnabledByCurrencyAndMethod(ctx context.Context, currency, paymentMethod string) ([]*entities.PaymentConfig, error)
	GetAvailableForPayment(ctx context.Context, currency string, amount float64) ([]*entities.PaymentConfig, error)

	// Amount validation operations
	ListByAmountRange(ctx context.Context, minAmount, maxAmount float64, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	GetConfigsForAmount(ctx context.Context, amount float64, currency string) ([]*entities.PaymentConfig, error)

	// Fee operations
	ListByFeeRange(ctx context.Context, minFee, maxFee float64, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	GetLowestFeeConfig(ctx context.Context, currency string, amount float64) (*entities.PaymentConfig, error)

	// Status management
	UpdateStatus(ctx context.Context, id uint, isEnabled bool) error
	UpdateSortOrder(ctx context.Context, id uint, sortOrder int) error
	UpdateConfig(ctx context.Context, id uint, config string) error
	UpdateMethods(ctx context.Context, id uint, methods []entities.Method) error

	// Batch operations
	BatchUpdateStatus(ctx context.Context, ids []uint, isEnabled bool) (int, []uint, error)
	BatchUpdateSortOrder(ctx context.Context, updates map[uint]int) error
	BatchDelete(ctx context.Context, ids []uint) (int, []uint, error)

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.PaymentConfig, int64, error)

	// Statistics
	CountTotal(ctx context.Context) (int64, error)
	CountEnabled(ctx context.Context) (int64, error)
	CountDisabled(ctx context.Context) (int64, error)
	CountByGateway(ctx context.Context, gateway string) (int64, error)

	// Existence checks
	ExistsByGateway(ctx context.Context, gateway string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)

	// Public operations (for frontend)
	ListPublic(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	ListPublicByCurrency(ctx context.Context, currency string, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	GetPublicEnabledConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error)

	// Ordering operations
	GetOrderedConfigs(ctx context.Context, orderBy string, ascending bool, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	ListBySortOrder(ctx context.Context, limit, offset int) ([]*entities.PaymentConfig, int64, error)

	// Gateway management
	ListByGatewayType(ctx context.Context, gatewayType string, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	GetSupportedGateways(ctx context.Context) ([]string, error)
	GetSupportedMethods(ctx context.Context) ([]string, error)

	// Configuration validation
	ValidateGatewayConfig(ctx context.Context, gateway string, config string) (bool, error)
	TestGatewayConnection(ctx context.Context, id uint) (bool, error)

	// Method management within configs
	GetMethodByCode(ctx context.Context, configID uint, methodCode string) (*entities.Method, error)
	UpdateMethodStatus(ctx context.Context, configID uint, methodCode string, isEnabled bool) error
	GetEnabledMethods(ctx context.Context, configID uint) ([]entities.Method, error)

	// Currency and method stats
	GetCurrencyStats(ctx context.Context) (map[string]int64, error)
	GetMethodStats(ctx context.Context) (map[string]int64, error)
	GetGatewayStats(ctx context.Context) (map[string]int64, error)

	// Advanced filtering
	ListWithFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*entities.PaymentConfig, int64, error)

	// Configuration backup and restore
	ExportConfigs(ctx context.Context) ([]*entities.PaymentConfig, error)
	ImportConfigs(ctx context.Context, configs []*entities.PaymentConfig) error

	// Environment-specific operations
	ListByEnvironment(ctx context.Context, environment string, limit, offset int) ([]*entities.PaymentConfig, int64, error)
	GetProductionConfigs(ctx context.Context) ([]*entities.PaymentConfig, error)
	GetTestConfigs(ctx context.Context) ([]*entities.PaymentConfig, error)
}