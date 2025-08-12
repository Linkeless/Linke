package entities

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"linke/internal/domains/payment/constants"
)

// PaymentConfig represents epay payment gateway configuration
type PaymentConfig struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Basic Information
	Name   string `json:"name" gorm:"size:100;not null"`         // Display name
	Method string `json:"method" gorm:"unique;size:50;not null"` // Payment method identifier

	// Epay Configuration (url + pid + key)
	URL string `json:"url" gorm:"column:url;size:255;not null"` // Epay API endpoint URL
	PID string `json:"pid" gorm:"column:pid;size:100;not null"` // Epay Partner/Merchant ID
	Key string `json:"key" gorm:"column:key;size:255;not null"` // Epay API key/secret

	// Status and Settings
	IsEnabled bool `json:"is_enabled" gorm:"column:is_enabled;not null;default:true;index"` // Whether enabled
	SortOrder int  `json:"sort_order" gorm:"column:sort_order;default:0;index"`             // Sort order

	// Currency and Methods
	SupportedCurrencies string `json:"supported_currencies" gorm:"column:supported_currencies;type:text"` // Supported currencies
	SupportedMethods    string `json:"supported_methods" gorm:"column:supported_methods;type:text"`       // Supported epay methods (JSON)

	// Limits
	MinAmount float64 `json:"min_amount" gorm:"column:min_amount;type:decimal(10,2);default:0.01"`     // Minimum amount
	MaxAmount float64 `json:"max_amount" gorm:"column:max_amount;type:decimal(10,2);default:99999.99"` // Maximum amount

	// Fee Information
	FixedFee      float64 `json:"fixed_fee" gorm:"column:fixed_fee;type:decimal(10,2);default:0.00"`           // Fixed fee amount
	PercentageFee float64 `json:"percentage_fee" gorm:"column:percentage_fee;type:decimal(5,4);default:0.0000"` // Percentage fee (0.0000-99.9999)

	// Additional Settings (optional)
	NotifyURL string `json:"notify_url" gorm:"column:notify_url;size:255"` // Callback/webhook URL
	ReturnURL string `json:"return_url" gorm:"column:return_url;size:255"` // Return URL after payment

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// Method represents an epay payment method
type Method struct {
	Code        string  `json:"code"`                  // Method code (alipay, wechat, qqpay)
	Name        string  `json:"name"`                  // Display name
	Description string  `json:"description,omitempty"` // Description
	Icon        string  `json:"icon,omitempty"`        // Icon URL
	IsEnabled   bool    `json:"is_enabled"`            // Whether enabled
	SortOrder   int     `json:"sort_order"`            // Sort order
	FeeType     string  `json:"fee_type"`              // constants.FeeTypeNone, constants.FeeTypeFixed, constants.FeeTypePercentage
	FeeValue    float64 `json:"fee_value"`             // Fee value
	FeeMin      float64 `json:"fee_min"`               // Minimum fee
	FeeMax      float64 `json:"fee_max"`               // Maximum fee
}

// TableName returns the table name for PaymentConfig model
func (PaymentConfig) TableName() string {
	return "payment_configs"
}

// GetMethods returns the parsed epay methods from SupportedMethods JSON
func (pc *PaymentConfig) GetMethods() ([]Method, error) {
	if pc.SupportedMethods == "" {
		return nil, nil
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

// GetMethodByCode returns a specific epay method by its code
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

// GetConfig returns epay configuration as a map for compatibility
func (pc *PaymentConfig) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"url":        pc.URL,
		"pid":        pc.PID,
		"key":        pc.Key,
		"notify_url": pc.NotifyURL,
		"return_url": pc.ReturnURL,
		"gateway":    constants.PaymentGatewayEpay,
	}
}

// SetConfig sets epay configuration from a map
func (pc *PaymentConfig) SetConfig(config map[string]interface{}) {
	if url, ok := config["url"].(string); ok {
		pc.URL = url
	}
	if pid, ok := config["pid"].(string); ok {
		pc.PID = pid
	}
	if key, ok := config["key"].(string); ok {
		pc.Key = key
	}
	if notifyURL, ok := config["notify_url"].(string); ok {
		pc.NotifyURL = notifyURL
	}
	if returnURL, ok := config["return_url"].(string); ok {
		pc.ReturnURL = returnURL
	}
}

// IsConfigValid checks if the epay configuration is valid
func (pc *PaymentConfig) IsConfigValid() bool {
	return pc.URL != "" && pc.PID != "" && pc.Key != ""
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

// IsEpayMethod checks if the method is a supported epay method
func (pc *PaymentConfig) IsEpayMethod(method string) bool {
	supportedMethods := []string{
		constants.PaymentMethodAlipay,
		constants.PaymentMethodWechat,
		constants.PaymentMethodQQ,
	}

	for _, supportedMethod := range supportedMethods {
		if method == supportedMethod {
			return true
		}
	}
	return false
}

// GetSupportedEpayMethods returns all supported epay payment methods
func (pc *PaymentConfig) GetSupportedEpayMethods() []string {
	return []string{
		constants.PaymentMethodAlipay,
		constants.PaymentMethodWechat,
		constants.PaymentMethodQQ,
	}
}

