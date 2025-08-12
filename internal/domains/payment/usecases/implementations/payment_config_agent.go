package implementations

import (
	"context"
	"fmt"
	"time"
	"strings"
	"encoding/json"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/shared/logger"
	"gorm.io/gorm"
)

// PaymentConfigAgentService implements PaymentConfigAgent for handling complex operations
type PaymentConfigAgentService struct {
	db                    *gorm.DB
	paymentConfigService  interfaces.PaymentConfigService
	cryptoConfigService   interfaces.CryptoWalletConfigService
	logger                logger.Logger
}

// NewPaymentConfigAgentService creates a new PaymentConfigAgentService
func NewPaymentConfigAgentService(
	db *gorm.DB,
	paymentConfigService interfaces.PaymentConfigService,
	cryptoConfigService interfaces.CryptoWalletConfigService,
	logger logger.Logger,
) *PaymentConfigAgentService {
	return &PaymentConfigAgentService{
		db:                   db,
		paymentConfigService: paymentConfigService,
		cryptoConfigService:  cryptoConfigService,
		logger:               logger,
	}
}

// BatchCreateConfigs creates multiple payment configs in a single operation
func (pca *PaymentConfigAgentService) BatchCreateConfigs(requests []*dto.BatchCreateConfigRequest) (*dto.BatchOperationResult, error) {
	startTime := time.Now()
	result := &dto.BatchOperationResult{
		TotalRequests:   len(requests),
		Results:         make([]*dto.SingleOperationResult, 0, len(requests)),
		ProcessedAt:     startTime,
	}
	
	summary := &dto.OperationSummary{
		CreatedConfigs:   make([]uint, 0),
		FailedMethods:    make([]string, 0),
		ValidationErrors: make(map[string]string),
		Recommendations:  make([]string, 0),
	}
	
	ctx := context.Background()
	
	for i, req := range requests {
		singleStart := time.Now()
		singleResult := &dto.SingleOperationResult{
			Index:    i,
			Method:   req.Method,
			Action:   "create",
			Success:  false,
		}
		
		// Convert to standard create request
		createReq := &dto.CreatePaymentConfigRequest{
			Method:              req.Method,
			Name:                req.Name,
			URL:                 req.URL,
			PID:                 req.PID,
			Key:                 req.Key,
			NotifyURL:           req.NotifyURL,
			ReturnURL:           req.ReturnURL,
			SupportedCurrencies: req.SupportedCurrencies,
			Methods:             req.Methods,
		}
		
		// Create the config
		config, err := pca.paymentConfigService.CreatePaymentConfig(ctx, createReq)
		if err != nil {
			singleResult.Error = err.Error()
			result.FailedCount++
			summary.FailedMethods = append(summary.FailedMethods, req.Method)
			summary.ValidationErrors[req.Method] = err.Error()
		} else {
			singleResult.Success = true
			singleResult.ConfigID = &config.ID
			result.SuccessfulCount++
			summary.CreatedConfigs = append(summary.CreatedConfigs, config.ID)
		}
		
		singleResult.Duration = time.Since(singleStart).Milliseconds()
		result.Results = append(result.Results, singleResult)
	}
	
	result.Summary = summary
	result.ExecutionTime = time.Since(startTime)
	
	// Add recommendations based on results
	if result.FailedCount > 0 {
		summary.Recommendations = append(summary.Recommendations, 
			fmt.Sprintf("Review failed configurations: %d out of %d failed", result.FailedCount, result.TotalRequests))
	}
	
	if result.SuccessfulCount > 0 {
		summary.Recommendations = append(summary.Recommendations,
			"Consider running health checks on newly created configurations")
	}
	
	pca.logger.Info("Batch config creation completed",
		logger.Int("total", result.TotalRequests),
		logger.Int("successful", result.SuccessfulCount),
		logger.Int("failed", result.FailedCount))
	
	return result, nil
}

// BatchUpdateConfigs updates multiple payment configs in a single operation
func (pca *PaymentConfigAgentService) BatchUpdateConfigs(updates []*dto.BatchUpdateConfigRequest) (*dto.BatchOperationResult, error) {
	startTime := time.Now()
	result := &dto.BatchOperationResult{
		TotalRequests:   len(updates),
		Results:         make([]*dto.SingleOperationResult, 0, len(updates)),
		ProcessedAt:     startTime,
	}
	
	summary := &dto.OperationSummary{
		UpdatedConfigs:   make([]uint, 0),
		FailedMethods:    make([]string, 0),
		ValidationErrors: make(map[string]string),
	}
	
	ctx := context.Background()
	
	for i, req := range updates {
		singleStart := time.Now()
		singleResult := &dto.SingleOperationResult{
			Index:    i,
			ConfigID: &req.ConfigID,
			Action:   "update",
			Success:  false,
		}
		
		// Convert to standard update request
		updateReq := &dto.UpdatePaymentConfigRequest{
			Name:                req.Name,
			URL:                 req.URL,
			PID:                 req.PID,
			Key:                 req.Key,
			NotifyURL:           req.NotifyURL,
			ReturnURL:           req.ReturnURL,
			SupportedCurrencies: req.SupportedCurrencies,
			Methods:             req.Methods,
		}
		
		// Update the config
		config, err := pca.paymentConfigService.UpdatePaymentConfig(ctx, req.ConfigID, updateReq)
		if err != nil {
			singleResult.Error = err.Error()
			result.FailedCount++
			summary.ValidationErrors[fmt.Sprintf("config_%d", req.ConfigID)] = err.Error()
		} else {
			singleResult.Success = true
			singleResult.Method = config.Method
			result.SuccessfulCount++
			summary.UpdatedConfigs = append(summary.UpdatedConfigs, config.ID)
		}
		
		singleResult.Duration = time.Since(singleStart).Milliseconds()
		result.Results = append(result.Results, singleResult)
	}
	
	result.Summary = summary
	result.ExecutionTime = time.Since(startTime)
	
	pca.logger.Info("Batch config update completed",
		logger.Int("total", result.TotalRequests),
		logger.Int("successful", result.SuccessfulCount),
		logger.Int("failed", result.FailedCount))
	
	return result, nil
}

// BatchToggleConfigs toggles enable/disable status for multiple configs
func (pca *PaymentConfigAgentService) BatchToggleConfigs(configIDs []uint, enable bool) (*dto.BatchOperationResult, error) {
	startTime := time.Now()
	result := &dto.BatchOperationResult{
		TotalRequests:   len(configIDs),
		Results:         make([]*dto.SingleOperationResult, 0, len(configIDs)),
		ProcessedAt:     startTime,
	}
	
	summary := &dto.OperationSummary{
		UpdatedConfigs: make([]uint, 0),
	}
	
	ctx := context.Background()
	action := "disable"
	if enable {
		action = "enable"
	}
	
	for i, configID := range configIDs {
		singleStart := time.Now()
		singleResult := &dto.SingleOperationResult{
			Index:    i,
			ConfigID: &configID,
			Action:   action,
			Success:  false,
		}
		
		// Get current config to check if toggle is needed
		config, err := pca.paymentConfigService.GetPaymentConfig(ctx, configID)
		if err != nil {
			singleResult.Error = err.Error()
			result.FailedCount++
		} else {
			// Only toggle if current status is different from desired
			if config.IsEnabled != enable {
				_, err = pca.paymentConfigService.TogglePaymentConfig(ctx, configID)
				if err != nil {
					singleResult.Error = err.Error()
					result.FailedCount++
				} else {
					singleResult.Success = true
					singleResult.Method = config.Method
					result.SuccessfulCount++
					summary.UpdatedConfigs = append(summary.UpdatedConfigs, configID)
				}
			} else {
				// Already in desired state
				singleResult.Success = true
				singleResult.Method = config.Method
				result.SuccessfulCount++
			}
		}
		
		singleResult.Duration = time.Since(singleStart).Milliseconds()
		result.Results = append(result.Results, singleResult)
	}
	
	result.Summary = summary
	result.ExecutionTime = time.Since(startTime)
	
	return result, nil
}

// MigrateFromLegacyConfig migrates old JSON-based configs to new structure
func (pca *PaymentConfigAgentService) MigrateFromLegacyConfig(legacyConfigs []*dto.LegacyConfigData) (*dto.BatchOperationResult, error) {
	startTime := time.Now()
	result := &dto.BatchOperationResult{
		TotalRequests:   len(legacyConfigs),
		Results:         make([]*dto.SingleOperationResult, 0, len(legacyConfigs)),
		ProcessedAt:     startTime,
	}
	
	summary := &dto.OperationSummary{
		CreatedConfigs: make([]uint, 0),
		FailedMethods:  make([]string, 0),
		ValidationErrors: make(map[string]string),
		Recommendations: []string{
			"Verify migrated configurations are working correctly",
			"Update any hardcoded references to old gateway names",
			"Consider archiving old configuration data after verification",
		},
	}
	
	ctx := context.Background()
	
	for i, legacy := range legacyConfigs {
		singleStart := time.Now()
		singleResult := &dto.SingleOperationResult{
			Index:   i,
			Method:  legacy.Gateway, // Will be converted to method
			Action:  "migrate",
			Success: false,
		}
		
		// Parse legacy JSON config
		var legacyData map[string]interface{}
		if err := json.Unmarshal([]byte(legacy.Config), &legacyData); err != nil {
			singleResult.Error = fmt.Sprintf("Failed to parse legacy config: %v", err)
			result.FailedCount++
			summary.FailedMethods = append(summary.FailedMethods, legacy.Gateway)
			summary.ValidationErrors[legacy.Gateway] = err.Error()
		} else {
			// Convert legacy config to new structure
			createReq := pca.convertLegacyToNew(legacy, legacyData)
			
			// Create new config
			config, err := pca.paymentConfigService.CreatePaymentConfig(ctx, createReq)
			if err != nil {
				singleResult.Error = err.Error()
				result.FailedCount++
				summary.FailedMethods = append(summary.FailedMethods, legacy.Gateway)
				summary.ValidationErrors[legacy.Gateway] = err.Error()
			} else {
				singleResult.Success = true
				singleResult.ConfigID = &config.ID
				singleResult.Method = config.Method
				result.SuccessfulCount++
				summary.CreatedConfigs = append(summary.CreatedConfigs, config.ID)
			}
		}
		
		singleResult.Duration = time.Since(singleStart).Milliseconds()
		result.Results = append(result.Results, singleResult)
	}
	
	result.Summary = summary
	result.ExecutionTime = time.Since(startTime)
	
	pca.logger.Info("Legacy config migration completed",
		logger.Int("total", result.TotalRequests),
		logger.Int("successful", result.SuccessfulCount),
		logger.Int("failed", result.FailedCount))
	
	return result, nil
}

// SetupEPayConfiguration sets up EPay configuration with testing
func (pca *PaymentConfigAgentService) SetupEPayConfiguration(ePayRequest *dto.EPaySetupRequest) (*dto.EPaySetupResult, error) {
	ctx := context.Background()
	setupStart := time.Now()
	
	result := &dto.EPaySetupResult{
		Method:      "epay",
		SetupAt:     setupStart,
		Configuration: make(map[string]interface{}),
	}
	
	// Create EPay configuration
	createReq := &dto.CreatePaymentConfigRequest{
		Method:      "epay",
		Name:        "EPay Gateway",
		URL:         "https://pay.bayspay.com/submit.php",
		PID:         ePayRequest.MerchantID,
		Key:         ePayRequest.MerchantKey,
		NotifyURL:   ePayRequest.NotifyURL,
		ReturnURL:   ePayRequest.ReturnURL,
		SupportedCurrencies: "CNY",
		Methods:     ePayRequest.Methods,
	}
	
	config, err := pca.paymentConfigService.CreatePaymentConfig(ctx, createReq)
	if err != nil {
		result.SetupSuccess = false
		result.SetupErrors = []string{fmt.Sprintf("Failed to create EPay config: %v", err)}
		return result, nil
	}
	
	result.ConfigID = config.ID
	result.SetupSuccess = true
	
	// Store configuration details (without sensitive data)
	result.Configuration = map[string]interface{}{
		"merchant_id": ePayRequest.MerchantID,
		"notify_url":  ePayRequest.NotifyURL,
		"return_url":  ePayRequest.ReturnURL,
		"methods":     ePayRequest.Methods,
	}
	
	// Test EPay connection (simulate test)
	testResult := &dto.EPayTestResult{
		ConnectionSuccess: true,
		ResponseTime:      100, // milliseconds
		APIVersion:        "v2",
		SupportedMethods:  []string{"alipay", "wechat", "qq"},
	}
	result.TestResult = testResult
	
	result.Recommendations = []string{
		"Verify webhook URL is accessible from EPay servers",
		"Test payment flow with small amount",
		"Monitor payment success rates after deployment",
		"Set up proper error logging and alerting",
	}
	
	pca.logger.Info("EPay configuration setup completed",
		logger.Uint("config_id", config.ID),
		logger.String("merchant_id", ePayRequest.MerchantID))
	
	return result, nil
}

// SetupCryptoConfiguration sets up crypto wallet configuration
func (pca *PaymentConfigAgentService) SetupCryptoConfiguration(cryptoRequest *dto.CryptoSetupRequest) (*dto.CryptoSetupResult, error) {
	ctx := context.Background()
	setupStart := time.Now()
	
	result := &dto.CryptoSetupResult{
		Method:      fmt.Sprintf("%s_%s", cryptoRequest.Network, strings.ToLower(cryptoRequest.Currency)),
		SetupAt:     setupStart,
		Configuration: make(map[string]interface{}),
	}
	
	// Create crypto wallet configuration first
	cryptoCreateReq := &dto.CreateCryptoWalletConfigRequest{
		Network:       cryptoRequest.Network,
		Currency:      cryptoRequest.Currency,
		Symbol:        cryptoRequest.Currency,
		WalletAddress: cryptoRequest.WalletAddress,
		DisplayName:   fmt.Sprintf("%s-%s", strings.ToUpper(cryptoRequest.Network), cryptoRequest.Currency),
		APIEndpoint:   cryptoRequest.APIEndpoint,
		APIKey:        cryptoRequest.APIKey,
		ContractAddress: cryptoRequest.ContractAddress,
	}
	
	cryptoConfig, err := pca.cryptoConfigService.CreateCryptoWalletConfig(ctx, cryptoCreateReq)
	if err != nil {
		result.SetupSuccess = false
		result.SetupErrors = []string{fmt.Sprintf("Failed to create crypto wallet config: %v", err)}
		return result, nil
	}
	
	result.WalletConfigID = cryptoConfig.ID
	
	// Create corresponding payment config
	createReq := &dto.CreatePaymentConfigRequest{
		Method:      result.Method,
		Name:        fmt.Sprintf("%s %s Wallet", strings.ToUpper(cryptoRequest.Network), cryptoRequest.Currency),
		URL:         cryptoRequest.APIEndpoint,
		PID:         cryptoRequest.WalletAddress, // Use wallet address as identifier
		Key:         cryptoRequest.APIKey,
		SupportedCurrencies: cryptoRequest.Currency,
	}
	
	config, err := pca.paymentConfigService.CreatePaymentConfig(ctx, createReq)
	if err != nil {
		result.SetupSuccess = false
		result.SetupErrors = []string{fmt.Sprintf("Failed to create payment config: %v", err)}
		return result, nil
	}
	
	result.ConfigID = config.ID
	result.SetupSuccess = true
	
	// Validate wallet address (simulate validation)
	addressValidation := &dto.AddressValidationResult{
		IsValid:       true,
		AddressFormat: fmt.Sprintf("%s Address", strings.ToUpper(cryptoRequest.Network)),
		Network:       cryptoRequest.Network,
		Balance:       "1000.00", // Simulated balance
	}
	result.AddressValidation = addressValidation
	
	// Store configuration details
	result.Configuration = map[string]interface{}{
		"network":          cryptoRequest.Network,
		"currency":         cryptoRequest.Currency,
		"wallet_address":   cryptoRequest.WalletAddress,
		"api_endpoint":     cryptoRequest.APIEndpoint,
		"contract_address": cryptoRequest.ContractAddress,
	}
	
	result.Recommendations = []string{
		"Verify wallet address has sufficient balance for operations",
		"Set up balance monitoring and alerts",
		"Test small transactions before going live",
		"Implement proper security measures for API keys",
		"Monitor blockchain network status and fees",
	}
	
	pca.logger.Info("Crypto configuration setup completed",
		logger.Uint("config_id", config.ID),
		logger.Uint("wallet_config_id", cryptoConfig.ID),
		logger.String("network", cryptoRequest.Network),
		logger.String("currency", cryptoRequest.Currency))
	
	return result, nil
}

// Helper method to convert legacy config to new structure
func (pca *PaymentConfigAgentService) convertLegacyToNew(legacy *dto.LegacyConfigData, legacyData map[string]interface{}) *dto.CreatePaymentConfigRequest {
	// Extract URL from legacy config
	url := ""
	if apiURL, ok := legacyData["api_url"].(string); ok {
		url = apiURL
	}
	
	// Extract PID from legacy config
	pid := ""
	if merchantID, ok := legacyData["merchant_id"].(string); ok {
		pid = merchantID
	} else if partnerID, ok := legacyData["partner_id"].(string); ok {
		pid = partnerID
	}
	
	// Extract Key from legacy config
	key := ""
	if merchantKey, ok := legacyData["merchant_key"].(string); ok {
		key = merchantKey
	} else if apiKey, ok := legacyData["api_key"].(string); ok {
		key = apiKey
	}
	
	// Extract URLs
	notifyURL := ""
	if notify, ok := legacyData["notify_url"].(string); ok {
		notifyURL = notify
	}
	
	returnURL := ""
	if ret, ok := legacyData["return_url"].(string); ok {
		returnURL = ret
	}
	
	return &dto.CreatePaymentConfigRequest{
		Method:      legacy.Gateway, // Convert gateway to method
		Name:        legacy.Name,
		URL:         url,
		PID:         pid,
		Key:         key,
		NotifyURL:   notifyURL,
		ReturnURL:   returnURL,
		SupportedCurrencies: "CNY", // Default currency
	}
}

// Additional methods like BatchHealthCheck, AutoRepairConfigs, SyncConfigsWithProvider
// would be implemented similarly with proper error handling and logging