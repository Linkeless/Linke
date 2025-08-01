package query

import (
	"context"
	"fmt"

	"linke/internal/payment/domain/aggregate"
	"linke/internal/payment/domain/repository"
	"linke/internal/payment/domain/valueobject"
)

// PaymentConfigQueryHandler handles payment config queries
type PaymentConfigQueryHandler struct {
	configRepo repository.PaymentConfigRepository
}

// NewPaymentConfigQueryHandler creates a new PaymentConfigQueryHandler
func NewPaymentConfigQueryHandler(configRepo repository.PaymentConfigRepository) *PaymentConfigQueryHandler {
	return &PaymentConfigQueryHandler{
		configRepo: configRepo,
	}
}

// GetPaymentConfig handles the GetPaymentConfigQuery
func (h *PaymentConfigQueryHandler) GetPaymentConfig(ctx context.Context, query GetPaymentConfigQuery) (*PaymentConfigDTO, error) {
	configID, err := valueobject.NewPaymentConfigID(query.ConfigID)
	if err != nil {
		return nil, fmt.Errorf("invalid payment config ID: %w", err)
	}

	config, err := h.configRepo.FindByID(ctx, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	return h.toPaymentConfigDTO(config), nil
}

// GetPaymentConfigByGateway handles the GetPaymentConfigByGatewayQuery
func (h *PaymentConfigQueryHandler) GetPaymentConfigByGateway(ctx context.Context, query GetPaymentConfigByGatewayQuery) (*PaymentConfigDTO, error) {
	gateway, err := valueobject.NewPaymentGateway(query.Gateway)
	if err != nil {
		return nil, fmt.Errorf("invalid payment gateway: %w", err)
	}

	config, err := h.configRepo.FindByGateway(ctx, gateway)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment config: %w", err)
	}

	return h.toPaymentConfigDTO(config), nil
}

// ListPaymentConfigs handles the ListPaymentConfigsQuery
func (h *PaymentConfigQueryHandler) ListPaymentConfigs(ctx context.Context, query ListPaymentConfigsQuery) (*PaymentConfigListResult, error) {
	filters := h.buildPaymentConfigFilters(query)

	configs, totalCount, err := h.configRepo.FindWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list payment configs: %w", err)
	}

	configDTOs := make([]PaymentConfigDTO, 0, len(configs))
	for _, config := range configs {
		configDTOs = append(configDTOs, *h.toPaymentConfigDTO(config))
	}

	hasMore := false
	if query.Limit > 0 {
		hasMore = int64(query.Offset+query.Limit) < totalCount
	}

	return &PaymentConfigListResult{
		Configs:    configDTOs,
		TotalCount: totalCount,
		Limit:      query.Limit,
		Offset:     query.Offset,
		HasMore:    hasMore,
	}, nil
}

// GetActivePaymentConfigs handles the GetActivePaymentConfigsQuery
func (h *PaymentConfigQueryHandler) GetActivePaymentConfigs(ctx context.Context, query GetActivePaymentConfigsQuery) (*PublicPaymentConfigListResult, error) {
	configs, err := h.configRepo.FindActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find active payment configs: %w", err)
	}

	// Filter by currency if specified
	if query.Currency != "" {
		currency, err := valueobject.NewCurrency(query.Currency)
		if err != nil {
			return nil, fmt.Errorf("invalid currency: %w", err)
		}

		var filteredConfigs []*aggregate.PaymentConfig
		for _, config := range configs {
			if config.SupportsCurrency(currency) {
				filteredConfigs = append(filteredConfigs, config)
			}
		}
		configs = filteredConfigs
	}

	publicConfigDTOs := make([]PublicPaymentConfigDTO, 0, len(configs))
	for _, config := range configs {
		publicConfigDTOs = append(publicConfigDTOs, *h.toPublicPaymentConfigDTO(config))
	}

	return &PublicPaymentConfigListResult{
		Configs: publicConfigDTOs,
	}, nil
}

// GetPaymentConfigsByCurrency handles the GetPaymentConfigsByCurrencyQuery
func (h *PaymentConfigQueryHandler) GetPaymentConfigsByCurrency(ctx context.Context, query GetPaymentConfigsByCurrencyQuery) ([]PaymentConfigDTO, error) {
	currency, err := valueobject.NewCurrency(query.Currency)
	if err != nil {
		return nil, fmt.Errorf("invalid currency: %w", err)
	}

	configs, err := h.configRepo.FindByCurrency(ctx, currency)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment configs by currency: %w", err)
	}

	configDTOs := make([]PaymentConfigDTO, 0, len(configs))
	for _, config := range configs {
		configDTOs = append(configDTOs, *h.toPaymentConfigDTO(config))
	}

	return configDTOs, nil
}

// GetPaymentConfigsByMethod handles the GetPaymentConfigsByMethodQuery
func (h *PaymentConfigQueryHandler) GetPaymentConfigsByMethod(ctx context.Context, query GetPaymentConfigsByMethodQuery) ([]PaymentConfigDTO, error) {
	method, err := valueobject.NewPaymentMethod(query.Method)
	if err != nil {
		return nil, fmt.Errorf("invalid payment method: %w", err)
	}

	configs, err := h.configRepo.FindByMethod(ctx, method)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment configs by method: %w", err)
	}

	configDTOs := make([]PaymentConfigDTO, 0, len(configs))
	for _, config := range configs {
		configDTOs = append(configDTOs, *h.toPaymentConfigDTO(config))
	}

	return configDTOs, nil
}

// GetPaymentConfigStats gets payment config statistics
func (h *PaymentConfigQueryHandler) GetPaymentConfigStats(ctx context.Context) (*PaymentConfigStatsResult, error) {
	configs, err := h.configRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all payment configs: %w", err)
	}

	return h.calculateConfigStats(configs), nil
}

// Helper methods

// buildPaymentConfigFilters builds repository filters from query
func (h *PaymentConfigQueryHandler) buildPaymentConfigFilters(query ListPaymentConfigsQuery) repository.PaymentConfigFilters {
	filters := repository.PaymentConfigFilters{
		IsEnabled: query.IsEnabled,
		SortBy:    query.SortBy,
		SortOrder: query.SortOrder,
		Limit:     query.Limit,
		Offset:    query.Offset,
	}

	if query.Gateway != "" {
		if gateway, err := valueobject.NewPaymentGateway(query.Gateway); err == nil {
			filters.Gateway = &gateway
		}
	}

	if query.Currency != "" {
		if currency, err := valueobject.NewCurrency(query.Currency); err == nil {
			filters.Currency = &currency
		}
	}

	if query.Method != "" {
		if method, err := valueobject.NewPaymentMethod(query.Method); err == nil {
			filters.Method = &method
		}
	}

	return filters
}

// toPaymentConfigDTO converts PaymentConfig aggregate to PaymentConfigDTO
func (h *PaymentConfigQueryHandler) toPaymentConfigDTO(config *aggregate.PaymentConfig) *PaymentConfigDTO {
	// Convert supported currencies
	currencies := make([]string, 0, len(config.SupportedCurrencies()))
	for _, currency := range config.SupportedCurrencies() {
		currencies = append(currencies, currency.Code())
	}

	// Convert supported methods
	methods := make([]PaymentMethodConfigDTO, 0, len(config.SupportedMethods()))
	activeMethodCount := 0
	for _, methodConfig := range config.SupportedMethods() {
		methodDTO := PaymentMethodConfigDTO{
			Method:      methodConfig.Method().Value(),
			Name:        methodConfig.Name(),
			Description: methodConfig.Description(),
			Icon:        methodConfig.Icon(),
			IsEnabled:   methodConfig.IsEnabled(),
			SortOrder:   methodConfig.SortOrder(),
			FeeType:     string(methodConfig.FeeType()),
			FeeValue:    methodConfig.FeeValue(),
			FeeMin:      methodConfig.FeeMin(),
			FeeMax:      methodConfig.FeeMax(),
			Environment: string(methodConfig.Environment()),
			
			// Computed fields
			DisplayName:    methodConfig.Method().GetDisplayName(),
			ProcessingTime: methodConfig.Method().GetProcessingTime(),
			RequiresKYC:    methodConfig.Method().RequiresKYC(),
			Category:       methodConfig.Method().GetCategory(),
		}
		methods = append(methods, methodDTO)
		
		if methodConfig.IsEnabled() {
			activeMethodCount++
		}
	}

	return &PaymentConfigDTO{
		ID:                  config.ID().Value(),
		Gateway:             config.Gateway().Value(),
		Name:                config.Name(),
		IsEnabled:           config.IsEnabled(),
		SortOrder:           config.SortOrder(),
		SupportedCurrencies: currencies,
		SupportedMethods:    methods,
		MinAmount:           config.MinAmount().Amount(),
		MaxAmount:           config.MaxAmount().Amount(),
		FixedFee:           config.FixedFee().Amount(),
		PercentageFee:      config.PercentageFee(),
		CreatedAt:          config.CreatedAt(),
		UpdatedAt:          config.UpdatedAt(),
		
		// Computed fields
		IsActive:          config.IsActive(),
		GatewayDisplay:    config.Gateway().GetDisplayName(),
		GatewayType:       config.Gateway().GetType(),
		MethodCount:       len(methods),
		ActiveMethodCount: activeMethodCount,
		CurrencyCount:     len(currencies),
	}
}

// toPublicPaymentConfigDTO converts PaymentConfig aggregate to PublicPaymentConfigDTO
func (h *PaymentConfigQueryHandler) toPublicPaymentConfigDTO(config *aggregate.PaymentConfig) *PublicPaymentConfigDTO {
	// Convert supported currencies
	currencies := make([]string, 0, len(config.SupportedCurrencies()))
	for _, currency := range config.SupportedCurrencies() {
		currencies = append(currencies, currency.Code())
	}

	// Convert only enabled methods for public API
	methods := make([]PaymentMethodConfigDTO, 0)
	for _, methodConfig := range config.SupportedMethods() {
		if methodConfig.IsEnabled() {
			methodDTO := PaymentMethodConfigDTO{
				Method:      methodConfig.Method().Value(),
				Name:        methodConfig.Name(),
				Description: methodConfig.Description(),
				Icon:        methodConfig.Icon(),
				IsEnabled:   true,
				SortOrder:   methodConfig.SortOrder(),
				FeeType:     string(methodConfig.FeeType()),
				FeeValue:    methodConfig.FeeValue(),
				FeeMin:      methodConfig.FeeMin(),
				FeeMax:      methodConfig.FeeMax(),
				Environment: string(methodConfig.Environment()),
				
				// Computed fields
				DisplayName:    methodConfig.Method().GetDisplayName(),
				ProcessingTime: methodConfig.Method().GetProcessingTime(),
				RequiresKYC:    methodConfig.Method().RequiresKYC(),
				Category:       methodConfig.Method().GetCategory(),
			}
			methods = append(methods, methodDTO)
		}
	}

	return &PublicPaymentConfigDTO{
		ID:                  config.ID().Value(),
		Gateway:             config.Gateway().Value(),
		Name:                config.Name(),
		SortOrder:           config.SortOrder(),
		SupportedCurrencies: currencies,
		SupportedMethods:    methods,
		MinAmount:           config.MinAmount().Amount(),
		MaxAmount:           config.MaxAmount().Amount(),
		
		// Computed fields
		GatewayDisplay: config.Gateway().GetDisplayName(),
		GatewayType:    config.Gateway().GetType(),
		MethodCount:    len(methods),
	}
}

// calculateConfigStats calculates payment config statistics
func (h *PaymentConfigQueryHandler) calculateConfigStats(configs []*aggregate.PaymentConfig) *PaymentConfigStatsResult {
	stats := &PaymentConfigStatsResult{
		GatewayBreakdown: make(map[string]int64),
		TypeBreakdown:    make(map[string]int64),
		CurrencySupport:  make(map[string]int64),
		MethodSupport:    make(map[string]int64),
	}

	for _, config := range configs {
		stats.TotalConfigs++
		
		if config.IsActive() {
			stats.ActiveConfigs++
		} else {
			stats.InactiveConfigs++
		}
		
		// Gateway breakdown
		stats.GatewayBreakdown[config.Gateway().Value()]++
		
		// Type breakdown
		stats.TypeBreakdown[config.Gateway().GetType()]++
		
		// Currency support
		for _, currency := range config.SupportedCurrencies() {
			stats.CurrencySupport[currency.Code()]++
		}
		
		// Method support
		for _, method := range config.SupportedMethods() {
			if method.IsEnabled() {
				stats.MethodSupport[method.Method().Value()]++
			}
		}
	}

	return stats
}