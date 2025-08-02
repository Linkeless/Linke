package implementations

import (
	"context"
	"fmt"
	"strings"

	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"

	"gorm.io/gorm"
)

type PaymentConfigService struct {
	db *gorm.DB
}

func NewPaymentConfigService(db *gorm.DB) *PaymentConfigService {
	return &PaymentConfigService{
		db: db,
	}
}


// CreatePaymentConfig creates a new payment config
func (pcs *PaymentConfigService) CreatePaymentConfig(ctx context.Context, req *interfaces.CreatePaymentConfigRequest) (*entities.PaymentConfig, error) {
	// Check if config already exists
	var existingConfig entities.PaymentConfig
	if err := pcs.db.WithContext(ctx).Where("gateway = ?", req.Gateway).First(&existingConfig).Error; err == nil {
		return nil, fmt.Errorf("payment config for gateway '%s' already exists", req.Gateway)
	} else if err != gorm.ErrRecordNotFound {
		logger.Error("Failed to check existing payment config", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to check existing payment config: %w", err)
	}

	// Set defaults
	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	supportedCurrencies := "CNY"
	if req.SupportedCurrencies != "" {
		supportedCurrencies = req.SupportedCurrencies
	}

	minAmount := 0.01
	if req.MinAmount > 0 {
		minAmount = req.MinAmount
	}

	maxAmount := 99999.99
	if req.MaxAmount > 0 {
		maxAmount = req.MaxAmount
	}

	// Create the config
	config := &entities.PaymentConfig{
		Gateway:             req.Gateway,
		Name:                req.Name,
		Config:              req.Config,
		IsEnabled:           isEnabled,
		SortOrder:           req.SortOrder,
		SupportedCurrencies: supportedCurrencies,
		MinAmount:           minAmount,
		MaxAmount:           maxAmount,
		FixedFee:            req.FixedFee,
		PercentageFee:       req.PercentageFee,
	}

	// Set methods if provided
	if len(req.Methods) > 0 {
		if err := config.SetMethods(req.Methods); err != nil {
			logger.Error("Failed to set methods for payment config", logger.Error2("error", err))
			return nil, fmt.Errorf("failed to set methods: %w", err)
		}
	}

	if err := pcs.db.WithContext(ctx).Create(config).Error; err != nil {
		logger.Error("Failed to create payment config", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to create payment config: %w", err)
	}

	logger.Info("Payment config created successfully",
		logger.Uint("config_id", config.ID),
		logger.String("gateway", config.Gateway))

	return config, nil
}

// GetPaymentConfig gets a payment config by ID
func (pcs *PaymentConfigService) GetPaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error) {
	var config entities.PaymentConfig
	if err := pcs.db.WithContext(ctx).First(&config, configID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment config not found")
		}
		logger.Error("Failed to get payment config", logger.Error2("error", err), logger.Uint("config_id", configID))
		return nil, fmt.Errorf("failed to get payment config: %w", err)
	}

	return &config, nil
}

// GetPaymentConfigByGateway gets a payment config by gateway
func (pcs *PaymentConfigService) GetPaymentConfigByGateway(ctx context.Context, gateway string) (*entities.PaymentConfig, error) {
	var config entities.PaymentConfig
	if err := pcs.db.WithContext(ctx).Where("gateway = ?", gateway).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("payment config not found")
		}
		logger.Error("Failed to get payment config by gateway", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get payment config: %w", err)
	}

	return &config, nil
}

// GetPaymentConfigs gets payment configs with filtering and pagination
func (pcs *PaymentConfigService) GetPaymentConfigs(ctx context.Context, req *interfaces.GetPaymentConfigsRequest) ([]*entities.PaymentConfig, int64, error) {
	query := pcs.db.WithContext(ctx).Model(&entities.PaymentConfig{})

	// Apply filters
	if req.Gateway != "" {
		query = query.Where("gateway = ?", req.Gateway)
	}

	if req.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *req.IsEnabled)
	}

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		logger.Error("Failed to count payment configs", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to count payment configs: %w", err)
	}

	// Apply pagination and ordering
	query = query.Order("sort_order ASC, created_at ASC")

	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}

	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	var configs []*entities.PaymentConfig
	if err := query.Find(&configs).Error; err != nil {
		logger.Error("Failed to get payment configs", logger.Error2("error", err))
		return nil, 0, fmt.Errorf("failed to get payment configs: %w", err)
	}

	return configs, totalCount, nil
}

// GetActivePaymentConfigs gets active payment configs for public display
func (pcs *PaymentConfigService) GetActivePaymentConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error) {
	query := pcs.db.WithContext(ctx).Model(&entities.PaymentConfig{}).
		Where("is_enabled = ?", true)

	if currency != "" {
		// Check if currency is supported (simple check)
		query = query.Where("supported_currencies = ? OR supported_currencies = 'ALL' OR supported_currencies = '*'", strings.ToUpper(currency))
	}

	var configs []*entities.PaymentConfig
	if err := query.Order("sort_order ASC, created_at ASC").Find(&configs).Error; err != nil {
		logger.Error("Failed to get active payment configs", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get active payment configs: %w", err)
	}

	return configs, nil
}

// UpdatePaymentConfig updates a payment config
func (pcs *PaymentConfigService) UpdatePaymentConfig(ctx context.Context, configID uint, req *interfaces.UpdatePaymentConfigRequest) (*entities.PaymentConfig, error) {
	// Get existing config
	config, err := pcs.GetPaymentConfig(ctx, configID)
	if err != nil {
		return nil, err
	}

	// Prepare updates
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Config != nil {
		updates["config"] = *req.Config
	}

	if req.IsEnabled != nil {
		updates["is_enabled"] = *req.IsEnabled
	}

	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}

	if req.SupportedCurrencies != nil {
		updates["supported_currencies"] = *req.SupportedCurrencies
	}

	if req.MinAmount != nil {
		updates["min_amount"] = *req.MinAmount
	}

	if req.MaxAmount != nil {
		updates["max_amount"] = *req.MaxAmount
	}

	if req.FixedFee != nil {
		updates["fixed_fee"] = *req.FixedFee
	}

	if req.PercentageFee != nil {
		updates["percentage_fee"] = *req.PercentageFee
	}

	// Handle methods update
	if len(req.Methods) > 0 {
		if err := config.SetMethods(req.Methods); err != nil {
			logger.Error("Failed to set methods for payment config", logger.Error2("error", err))
			return nil, fmt.Errorf("failed to set methods: %w", err)
		}
		updates["supported_methods"] = config.SupportedMethods
	}

	// Update the config
	if len(updates) > 0 {
		if err := pcs.db.WithContext(ctx).Model(config).Updates(updates).Error; err != nil {
			logger.Error("Failed to update payment config", logger.Error2("error", err), logger.Uint("config_id", configID))
			return nil, fmt.Errorf("failed to update payment config: %w", err)
		}
	}

	// Reload the config
	if err := pcs.db.WithContext(ctx).First(config, configID).Error; err != nil {
		logger.Error("Failed to reload updated payment config", logger.Error2("error", err), logger.Uint("config_id", configID))
		return nil, fmt.Errorf("failed to reload updated payment config: %w", err)
	}

	logger.Info("Payment config updated successfully", logger.Uint("config_id", config.ID))

	return config, nil
}

// DeletePaymentConfig soft deletes a payment config
func (pcs *PaymentConfigService) DeletePaymentConfig(ctx context.Context, configID uint) error {
	// Check if config exists
	config, err := pcs.GetPaymentConfig(ctx, configID)
	if err != nil {
		return err
	}

	// Soft delete the config
	if err := pcs.db.WithContext(ctx).Delete(config).Error; err != nil {
		logger.Error("Failed to delete payment config", logger.Error2("error", err), logger.Uint("config_id", configID))
		return fmt.Errorf("failed to delete payment config: %w", err)
	}

	logger.Info("Payment config deleted successfully", logger.Uint("config_id", configID))

	return nil
}

// TogglePaymentConfig toggles the enabled status of a payment config
func (pcs *PaymentConfigService) TogglePaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error) {
	// Get existing config
	config, err := pcs.GetPaymentConfig(ctx, configID)
	if err != nil {
		return nil, err
	}

	// Toggle enabled status
	newStatus := !config.IsEnabled
	if err := pcs.db.WithContext(ctx).Model(config).Update("is_enabled", newStatus).Error; err != nil {
		logger.Error("Failed to toggle payment config status", logger.Error2("error", err), logger.Uint("config_id", configID))
		return nil, fmt.Errorf("failed to toggle payment config status: %w", err)
	}

	config.IsEnabled = newStatus
	logger.Info("Payment config status toggled",
		logger.Uint("config_id", configID),
		logger.Any("enabled", newStatus))

	return config, nil
}

// GetPaymentConfigsByGateway gets all configs for a specific gateway
func (pcs *PaymentConfigService) GetPaymentConfigsByGateway(ctx context.Context, gateway string) ([]*entities.PaymentConfig, error) {
	var configs []*entities.PaymentConfig
	if err := pcs.db.WithContext(ctx).Where("gateway = ?", gateway).
		Order("sort_order ASC, created_at ASC").Find(&configs).Error; err != nil {
		logger.Error("Failed to get payment configs by gateway", logger.Error2("error", err))
		return nil, fmt.Errorf("failed to get payment configs by gateway: %w", err)
	}

	return configs, nil
}
