package command

import (
	"context"
	"fmt"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/repository"
	"linke/internal/payment/domain/valueobject"
	"linke/internal/shared/domain"
)

// PaymentConfigCommandHandler handles payment config commands
type PaymentConfigCommandHandler struct {
	configRepo repository.PaymentConfigRepository
	eventBus   domain.EventBus
}

// NewPaymentConfigCommandHandler creates a new PaymentConfigCommandHandler
func NewPaymentConfigCommandHandler(
	configRepo repository.PaymentConfigRepository,
	eventBus domain.EventBus,
) *PaymentConfigCommandHandler {
	return &PaymentConfigCommandHandler{
		configRepo: configRepo,
		eventBus:   eventBus,
	}
}

// CreatePaymentConfig handles the CreatePaymentConfigCommand
func (h *PaymentConfigCommandHandler) CreatePaymentConfig(ctx context.Context, cmd CreatePaymentConfigCommand) (*CreatePaymentConfigResult, error) {
	// Create value objects
	gateway, err := valueobject.NewPaymentGateway(cmd.Gateway)
	if err != nil {
		return nil, fmt.Errorf("invalid payment gateway: %w", err)
	}

	// Check if config already exists for this gateway
	exists, err := h.configRepo.ExistsByGateway(ctx, gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to check gateway existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("payment config for gateway %s already exists", gateway.String())
	}

	// Convert supported currencies
	supportedCurrencies := make([]valueobject.Currency, 0, len(cmd.SupportedCurrencies))
	for _, currencyStr := range cmd.SupportedCurrencies {
		currency, err := valueobject.NewCurrency(currencyStr)
		if err != nil {
			return nil, fmt.Errorf("invalid currency %s: %w", currencyStr, err)
		}
		supportedCurrencies = append(supportedCurrencies, currency)
	}

	// Get base currency for amounts
	baseCurrency, err := valueobject.NewCurrency(cmd.BaseCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid base currency: %w", err)
	}

	// Create payment config aggregate
	config, err := aggregate.NewPaymentConfig(
		gateway,
		cmd.Name,
		cmd.Config,
		supportedCurrencies,
		baseCurrency,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment config: %w", err)
	}

	// Set optional fields
	if err := config.UpdateSortOrder(cmd.SortOrder); err != nil {
		return nil, fmt.Errorf("failed to set sort order: %w", err)
	}

	// Add methods if provided
	for _, methodReq := range cmd.Methods {
		methodConfig, err := methodReq.ToEntity()
		if err != nil {
			return nil, fmt.Errorf("failed to convert method config: %w", err)
		}
		if err := config.AddSupportedMethod(methodConfig); err != nil {
			return nil, fmt.Errorf("failed to add method %s: %w", methodReq.Method, err)
		}
	}

	// Set amount limits and fees
	minAmount, err := valueobject.NewMoney(cmd.MinAmount, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid min amount: %w", err)
	}

	maxAmount, err := valueobject.NewMoney(cmd.MaxAmount, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid max amount: %w", err)
	}

	if err := config.UpdateAmountLimits(minAmount, maxAmount); err != nil {
		return nil, fmt.Errorf("failed to set amount limits: %w", err)
	}

	fixedFee, err := valueobject.NewMoney(cmd.FixedFee, baseCurrency)
	if err != nil {
		return nil, fmt.Errorf("invalid fixed fee: %w", err)
	}

	if err := config.UpdateFeeSettings(fixedFee, cmd.PercentageFee); err != nil {
		return nil, fmt.Errorf("failed to set fee settings: %w", err)
	}

	// Save config
	if err := h.configRepo.Save(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to save payment config: %w", err)
	}

	// Publish domain events
	if err := h.eventBus.Publish(ctx, config.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}

	config.ClearDomainEvents()

	return &CreatePaymentConfigResult{
		ConfigID:            config.ID(),
		Gateway:             config.Gateway(),
		Name:                config.Name(),
		SupportedCurrencies: config.SupportedCurrencies(),
		IsEnabled:           config.IsEnabled(),
		CreatedAt:           config.CreatedAt(),
	}, nil
}

// UpdatePaymentConfig handles the UpdatePaymentConfigCommand
func (h *PaymentConfigCommandHandler) UpdatePaymentConfig(ctx context.Context, cmd UpdatePaymentConfigCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	// Update name if provided
	if cmd.Name != nil {
		if err := config.UpdateName(*cmd.Name); err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}

	// Update config if provided
	if cmd.Config != nil {
		if err := config.UpdateConfig(*cmd.Config); err != nil {
			return nil, fmt.Errorf("failed to update config: %w", err)
		}
	}

	// Update sort order if provided
	if cmd.SortOrder != nil {
		if err := config.UpdateSortOrder(*cmd.SortOrder); err != nil {
			return nil, fmt.Errorf("failed to update sort order: %w", err)
		}
	}

	// Update supported currencies if provided
	if len(cmd.SupportedCurrencies) > 0 {
		// Clear existing currencies (simplified approach)
		// In a real implementation, you might want to add/remove individual currencies
		for _, currency := range config.SupportedCurrencies() {
			if err := config.RemoveSupportedCurrency(currency); err != nil {
				// Ignore error if it's the last currency
			}
		}

		// Add new currencies
		for _, currencyStr := range cmd.SupportedCurrencies {
			currency, err := valueobject.NewCurrency(currencyStr)
			if err != nil {
				return nil, fmt.Errorf("invalid currency %s: %w", currencyStr, err)
			}
			if err := config.AddSupportedCurrency(currency); err != nil {
				return nil, fmt.Errorf("failed to add currency %s: %w", currencyStr, err)
			}
		}
	}

	// Update methods if provided
	if len(cmd.Methods) > 0 {
		// Clear existing methods (simplified approach)
		for _, method := range config.SupportedMethods() {
			if err := config.RemoveSupportedMethod(method.Method()); err != nil {
				// Ignore error if method doesn't exist
			}
		}

		// Add new methods
		for _, methodReq := range cmd.Methods {
			methodConfig, err := methodReq.ToEntity()
			if err != nil {
				return nil, fmt.Errorf("failed to convert method config: %w", err)
			}
			if err := config.AddSupportedMethod(methodConfig); err != nil {
				return nil, fmt.Errorf("failed to add method %s: %w", methodReq.Method, err)
			}
		}
	}

	// Update amount limits if provided
	if cmd.MinAmount != nil && cmd.MaxAmount != nil {
		// Use the first supported currency as base currency
		baseCurrency := config.SupportedCurrencies()[0]

		minAmount, err := valueobject.NewMoney(*cmd.MinAmount, baseCurrency)
		if err != nil {
			return nil, fmt.Errorf("invalid min amount: %w", err)
		}

		maxAmount, err := valueobject.NewMoney(*cmd.MaxAmount, baseCurrency)
		if err != nil {
			return nil, fmt.Errorf("invalid max amount: %w", err)
		}

		if err := config.UpdateAmountLimits(minAmount, maxAmount); err != nil {
			return nil, fmt.Errorf("failed to update amount limits: %w", err)
		}
	}

	// Update fee settings if provided
	if cmd.FixedFee != nil || cmd.PercentageFee != nil {
		baseCurrency := config.SupportedCurrencies()[0]
		
		fixedFee := config.FixedFee()
		if cmd.FixedFee != nil {
			var err error
			fixedFee, err = valueobject.NewMoney(*cmd.FixedFee, baseCurrency)
			if err != nil {
				return nil, fmt.Errorf("invalid fixed fee: %w", err)
			}
		}

		percentageFee := config.PercentageFee()
		if cmd.PercentageFee != nil {
			percentageFee = *cmd.PercentageFee
		}

		if err := config.UpdateFeeSettings(fixedFee, percentageFee); err != nil {
			return nil, fmt.Errorf("failed to update fee settings: %w", err)
		}
	}

	// Save config
	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	// Publish domain events
	if err := h.eventBus.Publish(ctx, config.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}

	config.ClearDomainEvents()

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}

// EnablePaymentConfig handles the EnablePaymentConfigCommand
func (h *PaymentConfigCommandHandler) EnablePaymentConfig(ctx context.Context, cmd EnablePaymentConfigCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	if err := config.Enable(); err != nil {
		return nil, fmt.Errorf("failed to enable payment config: %w", err)
	}

	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	// Publish domain events
	if err := h.eventBus.Publish(ctx, config.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}

	config.ClearDomainEvents()

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}

// DisablePaymentConfig handles the DisablePaymentConfigCommand
func (h *PaymentConfigCommandHandler) DisablePaymentConfig(ctx context.Context, cmd DisablePaymentConfigCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	if err := config.Disable(); err != nil {
		return nil, fmt.Errorf("failed to disable payment config: %w", err)
	}

	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	// Publish domain events
	if err := h.eventBus.Publish(ctx, config.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}

	config.ClearDomainEvents()

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}

// DeletePaymentConfig handles the DeletePaymentConfigCommand
func (h *PaymentConfigCommandHandler) DeletePaymentConfig(ctx context.Context, cmd DeletePaymentConfigCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	if err := config.SoftDelete(); err != nil {
		return nil, fmt.Errorf("failed to delete payment config: %w", err)
	}

	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	// Publish domain events
	if err := h.eventBus.Publish(ctx, config.DomainEvents()...); err != nil {
		return nil, fmt.Errorf("failed to publish domain events: %w", err)
	}

	config.ClearDomainEvents()

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}

// AddSupportedCurrency handles the AddSupportedCurrencyCommand
func (h *PaymentConfigCommandHandler) AddSupportedCurrency(ctx context.Context, cmd AddSupportedCurrencyCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	currency, err := valueobject.NewCurrency(cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	if err := config.AddSupportedCurrency(currency); err != nil {
		return nil, fmt.Errorf("failed to add supported currency: %w", err)
	}

	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}

// RemoveSupportedCurrency handles the RemoveSupportedCurrencyCommand
func (h *PaymentConfigCommandHandler) RemoveSupportedCurrency(ctx context.Context, cmd RemoveSupportedCurrencyCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	currency, err := valueobject.NewCurrency(cmd.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	if err := config.RemoveSupportedCurrency(currency); err != nil {
		return nil, fmt.Errorf("failed to remove supported currency: %w", err)
	}

	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}

// AddSupportedMethod handles the AddSupportedMethodCommand
func (h *PaymentConfigCommandHandler) AddSupportedMethod(ctx context.Context, cmd AddSupportedMethodCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	methodConfig, err := cmd.MethodConfig.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert method config: %w", err)
	}

	if err := config.AddSupportedMethod(methodConfig); err != nil {
		return nil, fmt.Errorf("failed to add supported method: %w", err)
	}

	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}

// UpdateSupportedMethod handles the UpdateSupportedMethodCommand
func (h *PaymentConfigCommandHandler) UpdateSupportedMethod(ctx context.Context, cmd UpdateSupportedMethodCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	method, err := valueobject.NewPaymentMethod(cmd.Method)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method: %w", err)
	}

	methodConfig, err := cmd.MethodConfig.ToEntity()
	if err != nil {
		return nil, fmt.Errorf("failed to convert method config: %w", err)
	}

	if err := config.UpdateSupportedMethod(method, methodConfig); err != nil {
		return nil, fmt.Errorf("failed to update supported method: %w", err)
	}

	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}

// RemoveSupportedMethod handles the RemoveSupportedMethodCommand
func (h *PaymentConfigCommandHandler) RemoveSupportedMethod(ctx context.Context, cmd RemoveSupportedMethodCommand) (*PaymentConfigCommandResult, error) {
	configID, err := valueobject.NewPaymentConfigID(cmd.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	method, err := valueobject.NewPaymentMethod(cmd.Method)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method: %w", err)
	}

	if err := config.RemoveSupportedMethod(method); err != nil {
		return nil, fmt.Errorf("failed to remove supported method: %w", err)
	}

	if err := h.configRepo.Update(ctx, config); err != nil {
		return nil, fmt.Errorf("failed to update payment config: %w", err)
	}

	return &PaymentConfigCommandResult{
		ConfigID:  config.ID(),
		Gateway:   config.Gateway(),
		Name:      config.Name(),
		IsEnabled: config.IsEnabled(),
		UpdatedAt: config.UpdatedAt(),
	}, nil
}