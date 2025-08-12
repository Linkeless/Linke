package dto

import (
	"fmt"
	"strings"
)

// PaymentMethodSchemas defines the configuration schemas for different payment methods
var PaymentMethodSchemas = map[string]PaymentMethodConfigSchema{
	"epay": {
		Method:      "epay",
		DisplayName: "EPay支付",
		Description: "EPay支付网关，支持支付宝、微信、银联等多种支付方式",
		RequiredFields: []FieldDefinition{
			{
				Name:        "url",
				DisplayName: "API接口地址",
				Type:        "url",
				Required:    true,
				Description: "EPay支付网关的API接口地址",
				Placeholder: "https://pay.bayspay.com/submit.php",
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
	"crypto": {
		Method:      "crypto",
		DisplayName: "加密货币支付",
		Description: "支持USDT、BTC、ETH等主流加密货币支付",
		RequiredFields: []FieldDefinition{
			{
				Name:        "network",
				DisplayName: "区块链网络",
				Type:        "select",
				Required:    true,
				Description: "选择区块链网络类型",
				Options:     []string{"trc", "erc", "bep", "btc"},
				Validation:  "required|in:trc,erc,bep,btc",
				Sensitive:   false,
			},
			{
				Name:        "currency",
				DisplayName: "币种",
				Type:        "string",
				Required:    true,
				Description: "加密货币币种代码",
				Placeholder: "USDT",
				Validation:  "required|alpha|min:2|max:10",
				Sensitive:   false,
			},
			{
				Name:        "wallet_address",
				DisplayName: "钱包地址",
				Type:        "string",
				Required:    true,
				Description: "收款钱包地址",
				Placeholder: "TXXXxxxXXXxxxXXXxxxXXXxxxXXX",
				Validation:  "required|min:25|max:64",
				Sensitive:   false,
			},
		},
		OptionalFields: []FieldDefinition{
			{
				Name:        "api_endpoint",
				DisplayName: "区块链API接口",
				Type:        "url",
				Required:    false,
				Description: "用于查询交易状态的区块链API接口",
				Placeholder: "https://api.trongrid.io",
				Validation:  "url",
				Sensitive:   false,
			},
			{
				Name:        "api_key",
				DisplayName: "API密钥",
				Type:        "string",
				Required:    false,
				Description: "区块链API访问密钥（如需要）",
				Placeholder: "请输入API密钥",
				Validation:  "max:128",
				Sensitive:   true,
			},
			{
				Name:         "contract_address",
				DisplayName:  "合约地址",
				Type:         "string",
				Required:     false,
				Description:  "代币合约地址（ERC20/TRC20等代币需要）",
				Placeholder:  "TR7NHqjeKQxGTCI...",
				Validation:   "min:25|max:64",
				DefaultValue: "",
				Sensitive:    false,
			},
			{
				Name:         "min_confirmations",
				DisplayName:  "最小确认数",
				Type:         "number",
				Required:     false,
				Description:  "交易确认所需的最小区块确认数",
				DefaultValue: 1,
				Validation:   "numeric|min:1|max:100",
				Sensitive:    false,
			},
		},
		SupportedCurrencies: []string{"USDT", "BTC", "ETH", "TRX"},
		DefaultConfig: map[string]interface{}{
			"api_endpoint":      "",
			"api_key":           "",
			"contract_address":  "",
			"min_confirmations": 1,
		},
		ValidationRules: map[string]string{
			"network":         "required|in:trc,erc,bep,btc",
			"currency":        "required|alpha|min:2|max:10",
			"wallet_address":  "required|min:25|max:64",
			"api_endpoint":    "url",
			"contract_address": "min:25|max:64",
		},
	},
	"alipay": {
		Method:      "alipay",
		DisplayName: "支付宝",
		Description: "支付宝官方接口，支持网页支付、手机支付等",
		RequiredFields: []FieldDefinition{
			{
				Name:        "app_id",
				DisplayName: "应用ID",
				Type:        "string",
				Required:    true,
				Description: "支付宝开放平台分配的应用ID",
				Placeholder: "2021000000000000",
				Validation:  "required|numeric|digits:16",
				Sensitive:   false,
			},
			{
				Name:        "private_key",
				DisplayName: "应用私钥",
				Type:        "textarea",
				Required:    true,
				Description: "应用私钥，用于签名验证",
				Placeholder: "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
				Validation:  "required|min:100",
				Sensitive:   true,
			},
			{
				Name:        "alipay_public_key",
				DisplayName: "支付宝公钥",
				Type:        "textarea",
				Required:    true,
				Description: "支付宝公钥，用于验证支付宝返回的数据",
				Placeholder: "-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----",
				Validation:  "required|min:100",
				Sensitive:   false,
			},
		},
		OptionalFields: []FieldDefinition{
			{
				Name:         "gateway_url",
				DisplayName:  "网关地址",
				Type:         "url",
				Required:     false,
				Description:  "支付宝网关地址",
				DefaultValue: "https://openapi.alipay.com/gateway.do",
				Validation:   "url",
				Sensitive:    false,
			},
			{
				Name:         "charset",
				DisplayName:  "字符编码",
				Type:         "select",
				Required:     false,
				Description:  "请求和返回数据的字符编码",
				Options:      []string{"UTF-8", "GBK"},
				DefaultValue: "UTF-8",
				Validation:   "in:UTF-8,GBK",
				Sensitive:    false,
			},
			{
				Name:         "sign_type",
				DisplayName:  "签名算法",
				Type:         "select",
				Required:     false,
				Description:  "签名算法类型",
				Options:      []string{"RSA2", "RSA"},
				DefaultValue: "RSA2",
				Validation:   "in:RSA2,RSA",
				Sensitive:    false,
			},
		},
		SupportedCurrencies: []string{"CNY"},
		DefaultConfig: map[string]interface{}{
			"gateway_url": "https://openapi.alipay.com/gateway.do",
			"charset":     "UTF-8",
			"sign_type":   "RSA2",
		},
		ValidationRules: map[string]string{
			"app_id":            "required|numeric|digits:16",
			"private_key":       "required|min:100",
			"alipay_public_key": "required|min:100",
		},
	},
	"wechat": {
		Method:      "wechat",
		DisplayName: "微信支付",
		Description: "微信支付官方接口，支持公众号支付、H5支付、扫码支付等",
		RequiredFields: []FieldDefinition{
			{
				Name:        "mch_id",
				DisplayName: "商户号",
				Type:        "string",
				Required:    true,
				Description: "微信支付分配的商户号",
				Placeholder: "1234567890",
				Validation:  "required|numeric|digits:10",
				Sensitive:   false,
			},
			{
				Name:        "app_id",
				DisplayName: "应用ID",
				Type:        "string",
				Required:    true,
				Description: "微信公众号或应用的AppID",
				Placeholder: "wx1234567890abcdef",
				Validation:  "required|size:18|starts_with:wx",
				Sensitive:   false,
			},
			{
				Name:        "api_key",
				DisplayName: "API密钥",
				Type:        "string",
				Required:    true,
				Description: "微信支付API密钥",
				Placeholder: "请输入32位API密钥",
				Validation:  "required|size:32",
				Sensitive:   true,
			},
		},
		OptionalFields: []FieldDefinition{
			{
				Name:         "api_cert_path",
				DisplayName:  "API证书路径",
				Type:         "string",
				Required:     false,
				Description:  "API客户端证书路径（退款等高级功能需要）",
				Placeholder:  "/path/to/apiclient_cert.pem",
				Validation:   "max:255",
				DefaultValue: "",
				Sensitive:    false,
			},
			{
				Name:         "api_key_path",
				DisplayName:  "API密钥路径",
				Type:         "string",
				Required:     false,
				Description:  "API客户端密钥路径（退款等高级功能需要）",
				Placeholder:  "/path/to/apiclient_key.pem",
				Validation:   "max:255",
				DefaultValue: "",
				Sensitive:    false,
			},
			{
				Name:         "gateway_url",
				DisplayName:  "网关地址",
				Type:         "url",
				Required:     false,
				Description:  "微信支付网关地址",
				DefaultValue: "https://api.mch.weixin.qq.com",
				Validation:   "url",
				Sensitive:    false,
			},
		},
		SupportedCurrencies: []string{"CNY"},
		DefaultConfig: map[string]interface{}{
			"api_cert_path": "",
			"api_key_path":  "",
			"gateway_url":   "https://api.mch.weixin.qq.com",
		},
		ValidationRules: map[string]string{
			"mch_id":  "required|numeric|digits:10",
			"app_id":  "required|size:18|starts_with:wx",
			"api_key": "required|size:32",
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