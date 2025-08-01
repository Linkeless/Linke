package query

import "time"

// GetPaymentConfigQuery represents a query to get a single payment config
type GetPaymentConfigQuery struct {
	ConfigID uint `json:"config_id" validate:"required"`
}

// GetPaymentConfigByGatewayQuery represents a query to get a payment config by gateway
type GetPaymentConfigByGatewayQuery struct {
	Gateway string `json:"gateway" validate:"required"`
}

// ListPaymentConfigsQuery represents a query to list payment configs with filters
type ListPaymentConfigsQuery struct {
	Gateway   string `form:"gateway"`
	IsEnabled *bool  `form:"is_enabled"`
	Currency  string `form:"currency"`
	Method    string `form:"method"`
	SortBy    string `form:"sort_by"`
	SortOrder string `form:"sort_order"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

// GetActivePaymentConfigsQuery represents a query to get active payment configs
type GetActivePaymentConfigsQuery struct {
	Currency string `form:"currency"`
}

// GetPaymentConfigsByCurrencyQuery represents a query to get configs supporting a currency
type GetPaymentConfigsByCurrencyQuery struct {
	Currency string `json:"currency" validate:"required"`
}

// GetPaymentConfigsByMethodQuery represents a query to get configs supporting a method
type GetPaymentConfigsByMethodQuery struct {
	Method string `json:"method" validate:"required"`
}

// PaymentMethodConfigDTO represents a payment method configuration DTO
type PaymentMethodConfigDTO struct {
	Method      string  `json:"method"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Icon        string  `json:"icon,omitempty"`
	IsEnabled   bool    `json:"is_enabled"`
	SortOrder   int     `json:"sort_order"`
	FeeType     string  `json:"fee_type"`
	FeeValue    float64 `json:"fee_value"`
	FeeMin      float64 `json:"fee_min"`
	FeeMax      float64 `json:"fee_max"`
	Environment string  `json:"environment"`
	
	// Computed fields
	DisplayName     string `json:"display_name"`
	ProcessingTime  int    `json:"processing_time_minutes"`
	RequiresKYC     bool   `json:"requires_kyc"`
	Category        string `json:"category"`
}

// PaymentConfigDTO represents a payment config data transfer object
type PaymentConfigDTO struct {
	ID                  uint                     `json:"id"`
	Gateway             string                   `json:"gateway"`
	Name                string                   `json:"name"`
	IsEnabled           bool                     `json:"is_enabled"`
	SortOrder           int                      `json:"sort_order"`
	SupportedCurrencies []string                 `json:"supported_currencies"`
	SupportedMethods    []PaymentMethodConfigDTO `json:"supported_methods"`
	MinAmount           float64                  `json:"min_amount"`
	MaxAmount           float64                  `json:"max_amount"`
	FixedFee           float64                  `json:"fixed_fee"`
	PercentageFee      float64                  `json:"percentage_fee"`
	CreatedAt          time.Time                `json:"created_at"`
	UpdatedAt          time.Time                `json:"updated_at"`
	
	// Computed fields
	IsActive        bool     `json:"is_active"`
	GatewayDisplay  string   `json:"gateway_display"`
	GatewayType     string   `json:"gateway_type"`
	MethodCount     int      `json:"method_count"`
	ActiveMethodCount int    `json:"active_method_count"`
	CurrencyCount   int      `json:"currency_count"`
}

// PublicPaymentConfigDTO represents a payment config DTO for public API (without sensitive data)
type PublicPaymentConfigDTO struct {
	ID                  uint                     `json:"id"`
	Gateway             string                   `json:"gateway"`
	Name                string                   `json:"name"`
	SortOrder           int                      `json:"sort_order"`
	SupportedCurrencies []string                 `json:"supported_currencies"`
	SupportedMethods    []PaymentMethodConfigDTO `json:"supported_methods"`
	MinAmount           float64                  `json:"min_amount"`
	MaxAmount           float64                  `json:"max_amount"`
	
	// Computed fields
	GatewayDisplay  string `json:"gateway_display"`
	GatewayType     string `json:"gateway_type"`
	MethodCount     int    `json:"method_count"`
}

// PaymentConfigListResult represents the result of a payment config list query
type PaymentConfigListResult struct {
	Configs    []PaymentConfigDTO `json:"configs"`
	TotalCount int64              `json:"total_count"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
	HasMore    bool               `json:"has_more"`
}

// PublicPaymentConfigListResult represents the result of a public payment config list query
type PublicPaymentConfigListResult struct {
	Configs []PublicPaymentConfigDTO `json:"configs"`
}

// PaymentConfigStatsResult represents payment config statistics
type PaymentConfigStatsResult struct {
	TotalConfigs         int64            `json:"total_configs"`
	ActiveConfigs        int64            `json:"active_configs"`
	InactiveConfigs      int64            `json:"inactive_configs"`
	GatewayBreakdown     map[string]int64 `json:"gateway_breakdown"`
	TypeBreakdown        map[string]int64 `json:"type_breakdown"`
	CurrencySupport      map[string]int64 `json:"currency_support"`
	MethodSupport        map[string]int64 `json:"method_support"`
}