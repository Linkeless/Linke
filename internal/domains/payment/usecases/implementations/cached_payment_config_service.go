package implementations

import (
	"context"
	"encoding/json"
	"fmt"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
	"linke/internal/shared/cache"
	"linke/internal/shared/logger"
)

// CachedPaymentConfigService wraps PaymentConfigService with caching capabilities
type CachedPaymentConfigService struct {
	base         *PaymentConfigService
	cacheManager cache.CacheManager
	cacheKeys    *cache.PaymentCacheKeys
}

// NewCachedPaymentConfigService creates a new cached payment config service
func NewCachedPaymentConfigService(
	base *PaymentConfigService,
	cacheManager cache.CacheManager,
	allKeys *cache.AllCacheKeys,
) *CachedPaymentConfigService {
	return &CachedPaymentConfigService{
		base:         base,
		cacheManager: cacheManager,
		cacheKeys:    allKeys.Payment,
	}
}

// CreatePaymentConfig creates a new payment config and invalidates cache
func (ccs *CachedPaymentConfigService) CreatePaymentConfig(ctx context.Context, req *dto.CreatePaymentConfigRequest) (*entities.PaymentConfig, error) {
	config, err := ccs.base.CreatePaymentConfig(ctx, req)
	if err != nil {
		return nil, err
	}

	// Invalidate relevant caches since we added a new config
	ccs.invalidatePaymentConfigCaches(ctx)

	return config, nil
}

// GetPaymentConfig gets a payment config by ID with caching
func (ccs *CachedPaymentConfigService) GetPaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error) {
	cacheKey := ccs.buildConfigCacheKey("id", fmt.Sprintf("%d", configID))

	// Try to get from cache first
	cached, err := ccs.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var config entities.PaymentConfig
		if err := json.Unmarshal(cached, &config); err == nil {
			return &config, nil
		}
		// If unmarshal fails, continue to fetch from database
		logger.Warn("Failed to unmarshal cached payment config",
			logger.Uint("config_id", configID),
			logger.ErrorField(err))
	}

	// Fetch from database
	config, err := ccs.base.GetPaymentConfig(ctx, configID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	ccs.cachePaymentConfig(ctx, config, cacheKey)

	return config, nil
}

// GetPaymentConfigByMethod gets a payment config by method with caching
func (ccs *CachedPaymentConfigService) GetPaymentConfigByMethod(ctx context.Context, method string) (*entities.PaymentConfig, error) {
	cacheKey := ccs.buildConfigCacheKey("method", method)

	// Try to get from cache first
	cached, err := ccs.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var config entities.PaymentConfig
		if err := json.Unmarshal(cached, &config); err == nil {
			return &config, nil
		}
		// If unmarshal fails, continue to fetch from database
		logger.Warn("Failed to unmarshal cached payment config",
			logger.String("method", method),
			logger.ErrorField(err))
	}

	// Fetch from database
	config, err := ccs.base.GetPaymentConfigByMethod(ctx, method)
	if err != nil {
		return nil, err
	}

	// Cache the result
	ccs.cachePaymentConfig(ctx, config, cacheKey)

	return config, nil
}

// GetPaymentConfigs gets payment configs with filtering and pagination (no caching due to dynamic filters)
func (ccs *CachedPaymentConfigService) GetPaymentConfigs(ctx context.Context, req *dto.GetPaymentConfigsRequest) ([]*entities.PaymentConfig, int64, error) {
	// Don't cache paginated/filtered results due to complexity
	return ccs.base.GetPaymentConfigs(ctx, req)
}

// GetActivePaymentConfigs gets active payment configs with caching
func (ccs *CachedPaymentConfigService) GetActivePaymentConfigs(ctx context.Context, currency string) ([]*entities.PaymentConfig, error) {
	cacheKey := ccs.buildConfigCacheKey("active", currency)

	// Try to get from cache first
	cached, err := ccs.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var configs []*entities.PaymentConfig
		if err := json.Unmarshal(cached, &configs); err == nil {
			return configs, nil
		}
		// If unmarshal fails, continue to fetch from database
		logger.Warn("Failed to unmarshal cached active payment configs",
			logger.String("currency", currency),
			logger.ErrorField(err))
	}

	// Fetch from database
	configs, err := ccs.base.GetActivePaymentConfigs(ctx, currency)
	if err != nil {
		return nil, err
	}

	// Cache the result with long TTL since active configs don't change frequently
	if data, err := json.Marshal(configs); err == nil {
		_ = ccs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.LongCacheTTL)
	}

	return configs, nil
}

// UpdatePaymentConfig updates a payment config and invalidates cache
func (ccs *CachedPaymentConfigService) UpdatePaymentConfig(ctx context.Context, configID uint, req *dto.UpdatePaymentConfigRequest) (*entities.PaymentConfig, error) {
	config, err := ccs.base.UpdatePaymentConfig(ctx, configID, req)
	if err != nil {
		return nil, err
	}

	// Invalidate caches for this config and related queries
	ccs.invalidatePaymentConfigCaches(ctx)
	ccs.invalidateConfigCache(ctx, configID, config.Method)

	return config, nil
}

// DeletePaymentConfig deletes a payment config and invalidates cache
func (ccs *CachedPaymentConfigService) DeletePaymentConfig(ctx context.Context, configID uint) error {
	// Get config first to know which gateway to invalidate
	config, err := ccs.base.GetPaymentConfig(ctx, configID)
	if err != nil {
		return err
	}

	err = ccs.base.DeletePaymentConfig(ctx, configID)
	if err != nil {
		return err
	}

	// Invalidate caches
	ccs.invalidatePaymentConfigCaches(ctx)
	ccs.invalidateConfigCache(ctx, configID, config.Method)

	return nil
}

// TogglePaymentConfig toggles a payment config and invalidates cache
func (ccs *CachedPaymentConfigService) TogglePaymentConfig(ctx context.Context, configID uint) (*entities.PaymentConfig, error) {
	config, err := ccs.base.TogglePaymentConfig(ctx, configID)
	if err != nil {
		return nil, err
	}

	// Invalidate caches since enabled status changed
	ccs.invalidatePaymentConfigCaches(ctx)
	ccs.invalidateConfigCache(ctx, configID, config.Method)

	return config, nil
}

// GetPaymentConfigsByMethod gets configs for a specific method with caching
func (ccs *CachedPaymentConfigService) GetPaymentConfigsByMethod(ctx context.Context, method string) ([]*entities.PaymentConfig, error) {
	cacheKey := ccs.buildConfigCacheKey("method_all", method)

	// Try to get from cache first
	cached, err := ccs.cacheManager.GetCache().Get(ctx, cacheKey)
	if err == nil && cached != nil {
		var configs []*entities.PaymentConfig
		if err := json.Unmarshal(cached, &configs); err == nil {
			return configs, nil
		}
		// If unmarshal fails, continue to fetch from database
		logger.Warn("Failed to unmarshal cached payment configs by method",
			logger.String("method", method),
			logger.ErrorField(err))
	}

	// Fetch from database
	configs, err := ccs.base.GetPaymentConfigsByMethod(ctx, method)
	if err != nil {
		return nil, err
	}

	// Cache the result with long TTL
	if data, err := json.Marshal(configs); err == nil {
		_ = ccs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.LongCacheTTL)
	}

	return configs, nil
}

// Helper methods

// buildConfigCacheKey builds a cache key for payment config operations
func (ccs *CachedPaymentConfigService) buildConfigCacheKey(operation, identifier string) string {
	return cache.CachePrefixConfig + "payment:" + operation + ":" + identifier
}

// cachePaymentConfig caches a payment config
func (ccs *CachedPaymentConfigService) cachePaymentConfig(ctx context.Context, config *entities.PaymentConfig, cacheKey string) {
	if config == nil {
		return
	}

	// Create a sanitized version for caching (remove sensitive data)
	cachedConfig := &entities.PaymentConfig{
		ID:                  config.ID,
		Method:              config.Method,
		Name:                config.Name,
		URL:                 config.URL,
		PID:                 config.PID,
		NotifyURL:           config.NotifyURL,
		ReturnURL:           config.ReturnURL,
		IsEnabled:           config.IsEnabled,
		SortOrder:           config.SortOrder,
		SupportedCurrencies: config.SupportedCurrencies,
		SupportedMethods:    config.SupportedMethods,
		MinAmount:           config.MinAmount,
		MaxAmount:           config.MaxAmount,
		FixedFee:            config.FixedFee,
		PercentageFee:       config.PercentageFee,
		CreatedAt:           config.CreatedAt,
		UpdatedAt:           config.UpdatedAt,
		// Deliberately exclude sensitive Key field
	}

	if data, err := json.Marshal(cachedConfig); err == nil {
		// Use long TTL for config data since it doesn't change frequently
		_ = ccs.cacheManager.GetCache().Set(ctx, cacheKey, data, cache.LongCacheTTL)
	}
}

// invalidateConfigCache invalidates cache entries for a specific config
func (ccs *CachedPaymentConfigService) invalidateConfigCache(ctx context.Context, configID uint, method string) {
	// Invalidate specific config caches
	idCacheKey := ccs.buildConfigCacheKey("id", fmt.Sprintf("%d", configID))
	methodCacheKey := ccs.buildConfigCacheKey("method", method)
	methodAllCacheKey := ccs.buildConfigCacheKey("method_all", method)

	_ = ccs.cacheManager.GetCache().Delete(ctx, idCacheKey)
	_ = ccs.cacheManager.GetCache().Delete(ctx, methodCacheKey)
	_ = ccs.cacheManager.GetCache().Delete(ctx, methodAllCacheKey)
}

// invalidatePaymentConfigCaches invalidates general payment config caches
func (ccs *CachedPaymentConfigService) invalidatePaymentConfigCaches(ctx context.Context) {
	// Invalidate active config caches (with wildcard pattern)
	activePattern := cache.CachePrefixConfig + "payment:active:*"
	_ = ccs.cacheManager.GetCache().DeleteByPattern(ctx, activePattern)

	// Invalidate payment methods cache as it depends on config
	paymentMethodsKey := ccs.cacheKeys.PaymentMethods()
	_ = ccs.cacheManager.GetCache().Delete(ctx, paymentMethodsKey)
}
