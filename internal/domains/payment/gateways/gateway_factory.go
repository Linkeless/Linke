package gateways

import (
	"fmt"
	"sync"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/entities"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"
)

// GatewayFactory manages the creation and registration of payment gateways
type GatewayFactory struct {
	mu                   sync.RWMutex
	gateways             map[string]interfaces.PaymentGateway
	configs              map[string]*entities.PaymentConfig
	paymentConfigService interfaces.PaymentConfigService
}

// NewGatewayFactory creates a new gateway factory instance
func NewGatewayFactory(paymentConfigService interfaces.PaymentConfigService) *GatewayFactory {
	return &GatewayFactory{
		gateways:             make(map[string]interfaces.PaymentGateway),
		configs:              make(map[string]*entities.PaymentConfig),
		paymentConfigService: paymentConfigService,
	}
}

// RegisterGateway registers a payment gateway with the factory
func (gf *GatewayFactory) RegisterGateway(name string, gateway interfaces.PaymentGateway) error {
	gf.mu.Lock()
	defer gf.mu.Unlock()

	// Validate gateway configuration before registration
	if err := gateway.ValidateConfig(); err != nil {
		return fmt.Errorf("gateway validation failed for %s: %w", name, err)
	}

	// Test gateway connection
	if err := gateway.TestConnection(); err != nil {
		logger.Warn("Gateway connection test failed during registration",
			logger.String("gateway", name),
			logger.ErrorField(err))
		// Don't fail registration on connection test failure, just log it
	}

	gf.gateways[name] = gateway
	logger.Info("Payment gateway registered successfully", logger.String("gateway", name))

	return nil
}

// GetGateway retrieves a registered payment gateway
func (gf *GatewayFactory) GetGateway(name string) (interfaces.PaymentGateway, error) {
	gf.mu.RLock()
	defer gf.mu.RUnlock()

	gateway, exists := gf.gateways[name]
	if !exists {
		return nil, fmt.Errorf("gateway '%s' not found", name)
	}

	return gateway, nil
}

// GetAvailableGateways returns a list of all registered gateway names
func (gf *GatewayFactory) GetAvailableGateways() []string {
	gf.mu.RLock()
	defer gf.mu.RUnlock()

	gateways := make([]string, 0, len(gf.gateways))
	for name := range gf.gateways {
		gateways = append(gateways, name)
	}

	return gateways
}

// CreateGatewayFromConfig creates a gateway instance from payment configuration
func (gf *GatewayFactory) CreateGatewayFromConfig(config *entities.PaymentConfig) (interfaces.PaymentGateway, error) {
	if config == nil {
		return nil, fmt.Errorf("payment config is nil")
	}

	switch config.Method {
	case constants.PaymentGatewayEpay:
		return NewEpayGateway(config), nil
	default:
		return nil, fmt.Errorf("unsupported gateway type: %s", config.Method)
	}
}

// LoadGatewaysFromConfig loads and registers gateways from database configurations
func (gf *GatewayFactory) LoadGatewaysFromConfig() error {
	logger.Info("Loading payment gateways from configuration")

	// Get all enabled payment configurations
	configs, err := gf.paymentConfigService.GetEnabledConfigs()
	if err != nil {
		return fmt.Errorf("failed to load payment configurations: %w", err)
	}

	successCount := 0
	failureCount := 0

	for _, config := range configs {
		// Create gateway from configuration
		gateway, err := gf.CreateGatewayFromConfig(config)
		if err != nil {
			logger.Error("Failed to create gateway from config",
				logger.String("method", config.Method),
				logger.Uint("config_id", config.ID),
				logger.ErrorField(err))
			failureCount++
			continue
		}

		// Register the gateway
		if err := gf.RegisterGateway(config.Method, gateway); err != nil {
			logger.Error("Failed to register gateway",
				logger.String("method", config.Method),
				logger.Uint("config_id", config.ID),
				logger.ErrorField(err))
			failureCount++
			continue
		}

		// Store config for future reference
		gf.mu.Lock()
		gf.configs[config.Method] = config
		gf.mu.Unlock()

		successCount++
	}

	logger.Info("Gateway loading completed",
		logger.Int("success_count", successCount),
		logger.Int("failure_count", failureCount))

	if successCount == 0 {
		return fmt.Errorf("no gateways could be loaded successfully")
	}

	return nil
}

// ReloadGateway reloads a specific gateway configuration
func (gf *GatewayFactory) ReloadGateway(gatewayName string) error {
	logger.Info("Reloading gateway configuration", logger.String("gateway", gatewayName))

	// Get updated configuration from database
	config, err := gf.paymentConfigService.GetConfigByMethod(gatewayName)
	if err != nil {
		return fmt.Errorf("failed to get config for gateway %s: %w", gatewayName, err)
	}

	if config == nil {
		return fmt.Errorf("no configuration found for gateway %s", gatewayName)
	}

	// Check if gateway is enabled
	if !config.IsEnabled {
		// Unregister disabled gateway
		gf.mu.Lock()
		delete(gf.gateways, gatewayName)
		delete(gf.configs, gatewayName)
		gf.mu.Unlock()

		logger.Info("Gateway unregistered (disabled)", logger.String("gateway", gatewayName))
		return nil
	}

	// Create new gateway instance
	gateway, err := gf.CreateGatewayFromConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create gateway %s: %w", gatewayName, err)
	}

	// Register the updated gateway
	if err := gf.RegisterGateway(gatewayName, gateway); err != nil {
		return fmt.Errorf("failed to register updated gateway %s: %w", gatewayName, err)
	}

	// Update stored config
	gf.mu.Lock()
	gf.configs[gatewayName] = config
	gf.mu.Unlock()

	logger.Info("Gateway reloaded successfully", logger.String("gateway", gatewayName))
	return nil
}

// GetGatewayConfig retrieves the configuration for a specific gateway
func (gf *GatewayFactory) GetGatewayConfig(gatewayName string) (*entities.PaymentConfig, error) {
	gf.mu.RLock()
	defer gf.mu.RUnlock()

	config, exists := gf.configs[gatewayName]
	if !exists {
		return nil, fmt.Errorf("no configuration found for gateway %s", gatewayName)
	}

	return config, nil
}

// ValidateAllGateways validates the configuration of all registered gateways
func (gf *GatewayFactory) ValidateAllGateways() error {
	gf.mu.RLock()
	defer gf.mu.RUnlock()

	var validationErrors []string

	for name, gateway := range gf.gateways {
		if err := gateway.ValidateConfig(); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("%s: %v", name, err))
		}
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("gateway validation errors: %v", validationErrors)
	}

	return nil
}

// TestAllGatewayConnections tests the connection to all registered gateways
func (gf *GatewayFactory) TestAllGatewayConnections() map[string]error {
	gf.mu.RLock()
	defer gf.mu.RUnlock()

	results := make(map[string]error)

	for name, gateway := range gf.gateways {
		if err := gateway.TestConnection(); err != nil {
			results[name] = err
			logger.Warn("Gateway connection test failed",
				logger.String("gateway", name),
				logger.ErrorField(err))
		} else {
			results[name] = nil
			logger.Info("Gateway connection test successful", logger.String("gateway", name))
		}
	}

	return results
}

// GetSupportedPaymentMethods returns all supported payment methods across all gateways
func (gf *GatewayFactory) GetSupportedPaymentMethods() map[string][]string {
	gf.mu.RLock()
	defer gf.mu.RUnlock()

	methods := make(map[string][]string)

	for name, gateway := range gf.gateways {
		methods[name] = gateway.GetSupportedPaymentMethods()
	}

	return methods
}

// CreateDefaultEpayGateway creates a default epay gateway with minimal configuration
// This is useful for testing or when no database configuration is available
func (gf *GatewayFactory) CreateDefaultEpayGateway(url, pid, key string) error {
	if url == "" || pid == "" || key == "" {
		return fmt.Errorf("url, pid, and key are required for epay gateway")
	}

	// Create default configuration
	config := &entities.PaymentConfig{
		Method:              constants.PaymentGatewayEpay,
		Name:                "Default Epay Gateway",
		URL:                 url,
		PID:                 pid,
		Key:                 key,
		IsEnabled:           true,
		SupportedCurrencies: constants.CurrencyCNY,
		MinAmount:           0.01,
		MaxAmount:           99999.99,
		FixedFee:            0.00,
		PercentageFee:       0.6,
	}

	// Create and register gateway
	gateway := NewEpayGateway(config)
	if err := gf.RegisterGateway(constants.PaymentGatewayEpay, gateway); err != nil {
		return fmt.Errorf("failed to register default epay gateway: %w", err)
	}

	// Store config
	gf.mu.Lock()
	gf.configs[constants.PaymentGatewayEpay] = config
	gf.mu.Unlock()

	logger.Info("Default epay gateway created and registered successfully")
	return nil
}

// UnregisterGateway removes a gateway from the factory
func (gf *GatewayFactory) UnregisterGateway(name string) {
	gf.mu.Lock()
	defer gf.mu.Unlock()

	delete(gf.gateways, name)
	delete(gf.configs, name)

	logger.Info("Gateway unregistered", logger.String("gateway", name))
}

// GetGatewayCount returns the number of registered gateways
func (gf *GatewayFactory) GetGatewayCount() int {
	gf.mu.RLock()
	defer gf.mu.RUnlock()

	return len(gf.gateways)
}

// IsGatewayRegistered checks if a gateway is registered
func (gf *GatewayFactory) IsGatewayRegistered(name string) bool {
	gf.mu.RLock()
	defer gf.mu.RUnlock()

	_, exists := gf.gateways[name]
	return exists
}
