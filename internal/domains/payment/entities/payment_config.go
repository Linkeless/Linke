package entities

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// PaymentConfig represents payment gateway configuration
type PaymentConfig struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Gateway Information
	Gateway         string `json:"gateway" gorm:"unique;size:50;not null"`                         // Payment gateway (epay, epusdt)
	Name            string `json:"name" gorm:"size:100;not null"`                                  // Display name
	
	// Configuration
	Config          string `json:"config" gorm:"type:json;not null"`                               // JSON configuration
	
	// Status and Settings
	IsEnabled       bool   `json:"is_enabled" gorm:"not null;default:true;index"`                  // Whether enabled
	SortOrder       int    `json:"sort_order" gorm:"default:0;index"`                              // Sort order
	
	// Currency and Methods
	SupportedCurrencies string `json:"supported_currencies" gorm:"type:text"`                     // Supported currencies
	SupportedMethods    string `json:"supported_methods" gorm:"type:text"`                        // Supported methods (JSON)
	
	// Limits
	MinAmount       float64 `json:"min_amount" gorm:"type:decimal(10,2);default:0.01"`             // Minimum amount
	MaxAmount       float64 `json:"max_amount" gorm:"type:decimal(10,2);default:99999.99"`         // Maximum amount
	
	// Fee Information
	FixedFee        float64 `json:"fixed_fee" gorm:"type:decimal(10,2);default:0.00"`              // Fixed fee amount
	PercentageFee   float64 `json:"percentage_fee" gorm:"type:decimal(5,4);default:0.0000"`        // Percentage fee (0.0000-99.9999)

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Method represents a payment method within a gateway
type Method struct {
	Code         string  `json:"code"`                    // Method code (alipay, wechat, usdt, etc.)
	Name         string  `json:"name"`                    // Display name
	Description  string  `json:"description,omitempty"`  // Description
	Icon         string  `json:"icon,omitempty"`         // Icon URL
	IsEnabled    bool    `json:"is_enabled"`             // Whether enabled
	SortOrder    int     `json:"sort_order"`             // Sort order
	FeeType      string  `json:"fee_type"`               // none, fixed, percentage
	FeeValue     float64 `json:"fee_value"`              // Fee value
	FeeMin       float64 `json:"fee_min"`                // Minimum fee
	FeeMax       float64 `json:"fee_max"`                // Maximum fee
	Environment  string  `json:"environment"`            // production, sandbox, test
}

// TableName returns the table name for PaymentConfig model
func (PaymentConfig) TableName() string {
	return "payment_configs"
}

// Fee type constants
const (
	FeeTypeNone       = "none"
	FeeTypeFixed      = "fixed"
	FeeTypePercentage = "percentage"
)

// Environment constants
const (
	EnvironmentProduction = "production"
	EnvironmentSandbox    = "sandbox"
	EnvironmentTest       = "test"
)

// GetMethods returns the parsed methods from SupportedMethods JSON
func (pc *PaymentConfig) GetMethods() ([]Method, error) {
	if pc.SupportedMethods == "" {
		return []Method{}, nil
	}
	
	var methods []Method
	if err := json.Unmarshal([]byte(pc.SupportedMethods), &methods); err != nil {
		return nil, err
	}
	
	return methods, nil
}

// SetMethods sets the SupportedMethods field from a slice of Method structs
func (pc *PaymentConfig) SetMethods(methods []Method) error {
	data, err := json.Marshal(methods)
	if err != nil {
		return err
	}
	
	pc.SupportedMethods = string(data)
	return nil
}

// GetMethodByCode returns a specific method by its code
func (pc *PaymentConfig) GetMethodByCode(code string) (*Method, error) {
	methods, err := pc.GetMethods()
	if err != nil {
		return nil, err
	}
	
	for _, method := range methods {
		if method.Code == code {
			return &method, nil
		}
	}
	
	return nil, nil // Not found
}

// IsActive checks if the payment config is active
func (pc *PaymentConfig) IsActive() bool {
	return pc.IsEnabled && !pc.IsDeleted()
}

// IsDeleted checks if the payment config is soft deleted
func (pc *PaymentConfig) IsDeleted() bool {
	return pc.DeletedAt.Valid
}

// SupportsCurrency checks if the payment config supports a specific currency
func (pc *PaymentConfig) SupportsCurrency(currency string) bool {
	if pc.SupportedCurrencies == "" {
		return false
	}
	
	// Simple check - in production, you might want to use a more robust method
	return pc.SupportedCurrencies == currency || 
		   pc.SupportedCurrencies == "ALL" ||
		   pc.SupportedCurrencies == "*"
}

// CalculateFixedFee returns the fixed fee
func (pc *PaymentConfig) CalculateFixedFee() float64 {
	return pc.FixedFee
}

// CalculatePercentageFee calculates the percentage fee for a given amount
func (pc *PaymentConfig) CalculatePercentageFee(amount float64) float64 {
	return amount * pc.PercentageFee / 100
}

// IsAmountValid checks if the amount is within the valid range
func (pc *PaymentConfig) IsAmountValid(amount float64) bool {
	return amount >= pc.MinAmount && amount <= pc.MaxAmount
}

// PaymentConfigResponse represents the payment config data structure for API responses
type PaymentConfigResponse struct {
	ID                  uint      `json:"id" example:"1"`                                      // Config ID
	Gateway             string    `json:"gateway" example:"epay"`                             // Payment gateway
	Name                string    `json:"name" example:"EPay Gateway"`                        // Display name
	IsEnabled           bool      `json:"is_enabled" example:"true"`                          // Enabled status
	SortOrder           int       `json:"sort_order" example:"1"`                             // Sort order
	SupportedCurrencies string    `json:"supported_currencies" example:"CNY"`                // Supported currencies
	Methods             []Method  `json:"methods,omitempty"`                                  // Payment methods
	MinAmount           float64   `json:"min_amount" example:"0.01"`                          // Minimum amount
	MaxAmount           float64   `json:"max_amount" example:"99999.99"`                      // Maximum amount
	FixedFee            float64   `json:"fixed_fee" example:"0.00"`                           // Fixed fee
	PercentageFee       float64   `json:"percentage_fee" example:"0.6"`                       // Percentage fee
	CreatedAt           time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`          // Creation time
	UpdatedAt           time.Time `json:"updated_at" example:"2024-01-01T00:00:00Z"`          // Update time
}

// ToResponse converts PaymentConfig to PaymentConfigResponse
func (pc *PaymentConfig) ToResponse() *PaymentConfigResponse {
	response := &PaymentConfigResponse{
		ID:                  pc.ID,
		Gateway:             pc.Gateway,
		Name:                pc.Name,
		IsEnabled:           pc.IsEnabled,
		SortOrder:           pc.SortOrder,
		SupportedCurrencies: pc.SupportedCurrencies,
		MinAmount:           pc.MinAmount,
		MaxAmount:           pc.MaxAmount,
		FixedFee:            pc.FixedFee,
		PercentageFee:       pc.PercentageFee,
		CreatedAt:           pc.CreatedAt,
		UpdatedAt:           pc.UpdatedAt,
	}
	
	// Parse methods if available
	if methods, err := pc.GetMethods(); err == nil {
		response.Methods = methods
	}
	
	return response
}

// ToPublicResponse converts PaymentConfig to a public response (without sensitive config)
func (pc *PaymentConfig) ToPublicResponse() *PaymentConfigResponse {
	response := &PaymentConfigResponse{
		ID:                  pc.ID,
		Gateway:             pc.Gateway,
		Name:                pc.Name,
		SupportedCurrencies: pc.SupportedCurrencies,
		MinAmount:           pc.MinAmount,
		MaxAmount:           pc.MaxAmount,
		SortOrder:           pc.SortOrder,
	}
	
	// Parse methods if available and filter for public display
	if methods, err := pc.GetMethods(); err == nil {
		var publicMethods []Method
		for _, method := range methods {
			if method.IsEnabled {
				publicMethods = append(publicMethods, method)
			}
		}
		response.Methods = publicMethods
	}
	
	return response
}