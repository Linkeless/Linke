package dto

import (
	"time"

	"linke/internal/domains/payment/entities"
)

// ==================== Payment Agent DTOs ====================

// PaymentConfigAgent defines interface for handling complex payment config operations
type PaymentConfigAgent interface {
	// Batch operations
	BatchCreateConfigs(configs []*BatchCreateConfigRequest) (*BatchOperationResult, error)
	BatchUpdateConfigs(updates []*BatchUpdateConfigRequest) (*BatchOperationResult, error)
	BatchToggleConfigs(configIDs []uint, enable bool) (*BatchOperationResult, error)
	
	// Migration operations
	MigrateFromLegacyConfig(legacyConfigs []*LegacyConfigData) (*BatchOperationResult, error)
	SyncConfigsWithProvider(provider string) (*SyncOperationResult, error)
	
	// Health check operations
	BatchHealthCheck(configIDs []uint) (*HealthCheckResult, error)
	AutoRepairConfigs(configIDs []uint) (*RepairOperationResult, error)
	
	// Complex configuration tasks
	SetupEPayConfiguration(ePayRequest *EPaySetupRequest) (*EPaySetupResult, error)
	SetupCryptoConfiguration(cryptoRequest *CryptoSetupRequest) (*CryptoSetupResult, error)
}

// ==================== Agent Request DTOs ====================

// BatchCreateConfigRequest represents a single config creation in batch
type BatchCreateConfigRequest struct {
	Method              string            `json:"method" example:"epay"`
	Name                string            `json:"name" example:"EPay Payment Method"`
	URL                 string            `json:"url" example:"https://pay.bayspay.com/submit.php"`
	PID                 string            `json:"pid" example:"merchant123"`
	Key                 string            `json:"key" example:"secret123"`
	NotifyURL           string            `json:"notify_url,omitempty"`
	ReturnURL           string            `json:"return_url,omitempty"`
	SupportedCurrencies string            `json:"supported_currencies" example:"CNY"`
	Methods             []entities.Method `json:"methods,omitempty"`
	Tags                []string          `json:"tags,omitempty"` // For batch organization
}

// BatchUpdateConfigRequest represents a single config update in batch
type BatchUpdateConfigRequest struct {
	ConfigID            uint              `json:"config_id"`
	Name                *string           `json:"name,omitempty"`
	URL                 *string           `json:"url,omitempty"`
	PID                 *string           `json:"pid,omitempty"`
	Key                 *string           `json:"key,omitempty"`
	NotifyURL           *string           `json:"notify_url,omitempty"`
	ReturnURL           *string           `json:"return_url,omitempty"`
	SupportedCurrencies *string           `json:"supported_currencies,omitempty"`
	Methods             []entities.Method `json:"methods,omitempty"`
	Tags                []string          `json:"tags,omitempty"`
}

// LegacyConfigData represents old JSON-based configuration for migration
type LegacyConfigData struct {
	Gateway string `json:"gateway"`
	Name    string `json:"name"`
	Config  string `json:"config"` // Old JSON config
}

// EPaySetupRequest represents EPay-specific configuration setup
type EPaySetupRequest struct {
	MerchantID   string            `json:"merchant_id" binding:"required" example:"12345"`
	MerchantKey  string            `json:"merchant_key" binding:"required" example:"secret123"`
	NotifyURL    string            `json:"notify_url" example:"https://example.com/webhook"`
	ReturnURL    string            `json:"return_url" example:"https://example.com/return"`
	Methods      []entities.Method `json:"methods,omitempty"` // Alipay, WeChat, etc.
}

// CryptoSetupRequest represents crypto wallet configuration setup
type CryptoSetupRequest struct {
	Network         string `json:"network" binding:"required" example:"trc"`
	Currency        string `json:"currency" binding:"required" example:"USDT"`
	WalletAddress   string `json:"wallet_address" binding:"required"`
	APIEndpoint     string `json:"api_endpoint" example:"https://api.trongrid.io"`
	APIKey          string `json:"api_key,omitempty"`
	ContractAddress string `json:"contract_address,omitempty"`
}

// ==================== Agent Response DTOs ====================

// BatchOperationResult represents the result of batch operations
type BatchOperationResult struct {
	TotalRequests    int                        `json:"total_requests"`
	SuccessfulCount  int                        `json:"successful_count"`
	FailedCount      int                        `json:"failed_count"`
	Results          []*SingleOperationResult   `json:"results"`
	Summary          *OperationSummary          `json:"summary"`
	Warnings         []string                   `json:"warnings,omitempty"`
	ExecutionTime    time.Duration              `json:"execution_time"`
	ProcessedAt      time.Time                  `json:"processed_at"`
}

// SingleOperationResult represents the result of a single operation
type SingleOperationResult struct {
	Index       int    `json:"index"`
	ConfigID    *uint  `json:"config_id,omitempty"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	Method      string `json:"method,omitempty"`
	Action      string `json:"action"` // create, update, delete, toggle
	Duration    int64  `json:"duration_ms"`
}

// OperationSummary provides a summary of batch operations
type OperationSummary struct {
	CreatedConfigs  []uint            `json:"created_configs"`
	UpdatedConfigs  []uint            `json:"updated_configs"`
	FailedMethods   []string          `json:"failed_methods"`
	ValidationErrors map[string]string `json:"validation_errors,omitempty"`
	Recommendations []string          `json:"recommendations,omitempty"`
}

// SyncOperationResult represents the result of synchronization operations
type SyncOperationResult struct {
	Provider         string                   `json:"provider"`
	SyncedConfigs    []uint                   `json:"synced_configs"`
	NewConfigs       []uint                   `json:"new_configs"`
	UpdatedConfigs   []uint                   `json:"updated_configs"`
	RemovedConfigs   []uint                   `json:"removed_configs"`
	SyncErrors       []*SyncError             `json:"sync_errors,omitempty"`
	LastSyncAt       time.Time                `json:"last_sync_at"`
	NextSyncAt       time.Time                `json:"next_sync_at"`
}

// SyncError represents synchronization errors
type SyncError struct {
	ConfigID uint   `json:"config_id"`
	Method   string `json:"method"`
	Error    string `json:"error"`
	Severity string `json:"severity"` // warning, error, critical
}

// HealthCheckResult represents the result of health check operations
type HealthCheckResult struct {
	TotalConfigs    int                     `json:"total_configs"`
	HealthyConfigs  int                     `json:"healthy_configs"`
	UnhealthyConfigs int                    `json:"unhealthy_configs"`
	CheckResults    []*ConfigHealthStatus   `json:"check_results"`
	OverallStatus   string                  `json:"overall_status"` // healthy, degraded, critical
	CheckedAt       time.Time               `json:"checked_at"`
	Recommendations []string                `json:"recommendations,omitempty"`
}

// ConfigHealthStatus represents the health status of a single config
type ConfigHealthStatus struct {
	ConfigID      uint                   `json:"config_id"`
	Method        string                 `json:"method"`
	Status        string                 `json:"status"` // healthy, unhealthy, unknown
	ResponseTime  int64                  `json:"response_time_ms"`
	LastChecked   time.Time              `json:"last_checked"`
	Issues        []string               `json:"issues,omitempty"`
	Metrics       map[string]interface{} `json:"metrics,omitempty"`
}

// RepairOperationResult represents the result of repair operations
type RepairOperationResult struct {
	TotalConfigs   int                      `json:"total_configs"`
	RepairedConfigs int                     `json:"repaired_configs"`
	FailedRepairs  int                      `json:"failed_repairs"`
	RepairResults  []*ConfigRepairResult    `json:"repair_results"`
	RepairSummary  *RepairSummary           `json:"repair_summary"`
	RepairedAt     time.Time                `json:"repaired_at"`
}

// ConfigRepairResult represents the repair result for a single config
type ConfigRepairResult struct {
	ConfigID      uint     `json:"config_id"`
	Method        string   `json:"method"`
	Success       bool     `json:"success"`
	ActionsToken  []string `json:"actions_taken"`
	RemainingIssues []string `json:"remaining_issues,omitempty"`
	Error         string   `json:"error,omitempty"`
}

// RepairSummary provides a summary of repair operations
type RepairSummary struct {
	CommonIssues      []string          `json:"common_issues"`
	RepairActions     map[string]int    `json:"repair_actions"` // action -> count
	SystemImprovements []string         `json:"system_improvements"`
}

// EPaySetupResult represents the result of EPay setup
type EPaySetupResult struct {
	ConfigID        uint                   `json:"config_id"`
	Method          string                 `json:"method"`
	SetupSuccess    bool                   `json:"setup_success"`
	TestResult      *EPayTestResult        `json:"test_result,omitempty"`
	Configuration   map[string]interface{} `json:"configuration"`
	SetupErrors     []string               `json:"setup_errors,omitempty"`
	Recommendations []string               `json:"recommendations,omitempty"`
	SetupAt         time.Time              `json:"setup_at"`
}

// EPayTestResult represents EPay connection test result
type EPayTestResult struct {
	ConnectionSuccess bool   `json:"connection_success"`
	ResponseTime      int64  `json:"response_time_ms"`
	APIVersion        string `json:"api_version,omitempty"`
	SupportedMethods  []string `json:"supported_methods,omitempty"`
	TestError         string `json:"test_error,omitempty"`
}

// CryptoSetupResult represents the result of crypto setup
type CryptoSetupResult struct {
	ConfigID          uint                   `json:"config_id"`
	WalletConfigID    uint                   `json:"wallet_config_id"` // Link to CryptoWalletConfig
	Method            string                 `json:"method"`
	SetupSuccess      bool                   `json:"setup_success"`
	AddressValidation *AddressValidationResult `json:"address_validation,omitempty"`
	Configuration     map[string]interface{} `json:"configuration"`
	SetupErrors       []string               `json:"setup_errors,omitempty"`
	Recommendations   []string               `json:"recommendations,omitempty"`
	SetupAt           time.Time              `json:"setup_at"`
}

// AddressValidationResult represents wallet address validation result
type AddressValidationResult struct {
	IsValid       bool   `json:"is_valid"`
	AddressFormat string `json:"address_format"`
	Network       string `json:"network"`
	Balance       string `json:"balance,omitempty"`
	ValidationError string `json:"validation_error,omitempty"`
}