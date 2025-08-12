package dto

import (
	"fmt"
	"strings"
)

// PaymentMethodSchemas defines the configuration schemas for epay payment method
var PaymentMethodSchemas = map[string]PaymentMethodConfigSchema{
	"epay": {
		Method:      "epay",
		DisplayName: "EPay支付",
		Description: "EPay支付网关，支持支付宝、微信等多种支付方式",
		RequiredFields: []FieldDefinition{
			{
				Name:        "url",
				DisplayName: "API接口地址",
				Type:        "url",
				Required:    true,
				Description: "EPay支付网关的API接口地址",
				Placeholder: "https://pay.example.com/submit.php",
				Validation:  "url",
				Sensitive:   false,
			},
			{
				Name:        "pid",
				DisplayName: "商户号",
				Type:        "string",
				Required:    true,
				Description: "EPay分配的商户号",
				Placeholder: "10001",
				Validation:  "required|min:3|max:20",
				Sensitive:   false,
			},
			{
				Name:        "key",
				DisplayName: "商户密钥",
				Type:        "string",
				Required:    true,
				Description: "EPay分配的商户密钥，用于签名验证",
				Placeholder: "请输入商户密钥",
				Validation:  "required|min:16",
				Sensitive:   true,
			},
		},
		OptionalFields: []FieldDefinition{
			{
				Name:         "notify_url",
				DisplayName:  "异步通知地址",
				Type:         "url",
				Required:     false,
				Description:  "支付结果异步通知回调地址",
				Placeholder:  "https://example.com/webhook/epay",
				Validation:   "url",
				DefaultValue: "",
				Sensitive:    false,
			},
			{
				Name:         "return_url",
				DisplayName:  "同步返回地址",
				Type:         "url",
				Required:     false,
				Description:  "支付完成后用户返回地址",
				Placeholder:  "https://example.com/payment/return",
				Validation:   "url",
				DefaultValue: "",
				Sensitive:    false,
			},
		},
		SupportedCurrencies: []string{"CNY"},
		DefaultConfig: map[string]interface{}{
			"notify_url": "",
			"return_url": "",
		},
		ValidationRules: map[string]string{
			"url": "required|url",
			"pid": "required|min:3|max:20|alphanum",
			"key": "required|min:16|max:128",
		},
	},
}

// GetPaymentMethodSchema returns the configuration schema for a specific payment method
func GetPaymentMethodSchema(method string) (PaymentMethodConfigSchema, bool) {
	schema, exists := PaymentMethodSchemas[method]
	return schema, exists
}

// GetAllPaymentMethodSchemas returns all available payment method schemas
func GetAllPaymentMethodSchemas() map[string]PaymentMethodConfigSchema {
	return PaymentMethodSchemas
}

// GetSupportedPaymentMethods returns a list of all supported payment method names
func GetSupportedPaymentMethods() []string {
	methods := make([]string, 0, len(PaymentMethodSchemas))
	for method := range PaymentMethodSchemas {
		methods = append(methods, method)
	}
	return methods
}

// ValidatePaymentMethodConfig validates configuration against method schema
func ValidatePaymentMethodConfig(method string, config map[string]interface{}) ([]string, error) {
	schema, exists := GetPaymentMethodSchema(method)
	if !exists {
		return nil, fmt.Errorf("unsupported payment method: %s", method)
	}
	
	var errors []string
	
	// Check required fields
	for _, field := range schema.RequiredFields {
		value, exists := config[field.Name]
		if !exists || value == nil || value == "" {
			errors = append(errors, fmt.Sprintf("required field '%s' is missing", field.Name))
			continue
		}
		
		// Type validation
		if err := validateFieldValue(field, value); err != nil {
			errors = append(errors, fmt.Sprintf("field '%s': %v", field.Name, err))
		}
	}
	
	// Validate optional fields if provided
	for _, field := range schema.OptionalFields {
		if value, exists := config[field.Name]; exists && value != nil && value != "" {
			if err := validateFieldValue(field, value); err != nil {
				errors = append(errors, fmt.Sprintf("field '%s': %v", field.Name, err))
			}
		}
	}
	
	return errors, nil
}

// validateFieldValue validates a single field value against its definition
func validateFieldValue(field FieldDefinition, value interface{}) error {
	// Basic type checks
	switch field.Type {
	case "string", "textarea":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number":
		switch value.(type) {
		case int, int64, float64:
			// OK
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "url":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for URL, got %T", value)
		}
		if str != "" { // Allow empty URLs for optional fields
			if !isValidURL(str) {
				return fmt.Errorf("invalid URL format")
			}
		}
	case "select":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for select, got %T", value)
		}
		if len(field.Options) > 0 {
			found := false
			for _, option := range field.Options {
				if str == option {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("invalid option '%s', must be one of: %v", str, field.Options)
			}
		}
	}
	
	return nil
}

// isValidURL checks if a string is a valid URL
func isValidURL(str string) bool {
	// Simple URL validation - in production, you might want to use a more robust method
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}