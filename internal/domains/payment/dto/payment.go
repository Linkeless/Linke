package dto

import (
	"encoding/json"
	"strings"
	"time"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/entities"
	"linke/internal/shared/dto"
	"linke/internal/shared/pool"
)

// Package dto provides data transfer objects for the payment domain.
//
// This package is organized into functional sections:
// 1. Payment Record DTOs - Payment transactions and records
// 2. Payment Method DTOs - Payment method management
// 3. Payment Config DTOs - Payment gateway configurations
// 4. Helper & Conversion Functions - Utilities and entity conversions

// =====================================================================
// PAYMENT RECORD DTOs - Payment transactions and records
// =====================================================================

// PaymentRecordResponse represents the payment record data structure for API responses
type PaymentRecordResponse struct {
	ID                  uint       `json:"id" example:"1"`                                          // Payment ID
	UserID              uint       `json:"user_id" example:"1"`                                     // User ID
	SubscriptionOrderID *uint      `json:"subscription_order_id,omitempty" example:"1"`             // Subscription order ID
	PaymentNo           string     `json:"payment_no" example:"PAY202401010001"`                    // Payment number
	OutTradeNo          string     `json:"out_trade_no" example:"ORDER202401010001"`                // Merchant order number
	TransactionID       string     `json:"transaction_id,omitempty" example:"TXN123456789"`         // Transaction ID
	Gateway             string     `json:"gateway" example:"epay"`                                  // Payment gateway
	PaymentMethod       string     `json:"payment_method" example:"alipay"`                         // Payment method
	Amount              float64    `json:"amount" example:"29.99"`                                  // Payment amount
	Currency            string     `json:"currency" example:"CNY"`                                  // Currency
	ExchangeRate        float64    `json:"exchange_rate" example:"1.0000"`                          // Exchange rate
	Status              string     `json:"status" example:"completed"`                              // Payment status
	PaymentStatus       string     `json:"payment_status,omitempty" example:"success"`              // Gateway payment status
	PaymentURL          string     `json:"payment_url,omitempty" example:"https://example.com/pay"` // Payment URL
	QRCodeURL           string     `json:"qr_code_url,omitempty" example:"https://example.com/qr"`  // QR code URL
	ExpiredAt           *time.Time `json:"expired_at,omitempty" example:"2024-01-01T01:00:00Z"`     // Expiration time
	PaidAt              *time.Time `json:"paid_at,omitempty" example:"2024-01-01T00:30:00Z"`        // Payment completion time
	NotifiedAt          *time.Time `json:"notified_at,omitempty" example:"2024-01-01T00:31:00Z"`    // Notification time
	RefundAmount        float64    `json:"refund_amount" example:"0"`                               // Refund amount
	RefundStatus        string     `json:"refund_status,omitempty" example:"none"`                  // Refund status
	RefundedAt          *time.Time `json:"refunded_at,omitempty" example:"2024-01-02T10:00:00Z"`    // Refund time
	RefundReason        string     `json:"refund_reason,omitempty" example:"User request"`          // Refund reason
	Remark              string     `json:"remark,omitempty" example:"Subscription payment"`         // Remark
	CreatedAt           time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`               // Creation time
	UpdatedAt           time.Time  `json:"updated_at" example:"2024-01-01T00:30:00Z"`               // Update time

	// Related data (to be populated at application layer)
	User              *dto.UserBasicDTO              `json:"user,omitempty"`               // User info
	SubscriptionOrder *dto.SubscriptionOrderBasicDTO `json:"subscription_order,omitempty"` // Subscription order info

	// Computed fields
	IsExpired        bool    `json:"is_expired"`        // Expiration status
	CanRefund        bool    `json:"can_refund"`        // Refundable status
	RefundableAmount float64 `json:"refundable_amount"` // Refundable amount
}

// CreatePaymentOrderRequest represents the unified request to create a payment order
type CreatePaymentOrderRequest struct {
	UserID              uint    `json:"user_id"`
	SubscriptionOrderID *uint   `json:"subscription_order_id,omitempty" example:"1"`
	InvoiceID           *uint   `json:"invoice_id,omitempty" example:"1"`
	Gateway             string  `json:"gateway" binding:"required" example:"epay"`
	PaymentMethod       string  `json:"payment_method" binding:"required" example:"alipay"`
	Amount              float64 `json:"amount" binding:"required,gt=0" example:"29.99"`
	Currency            string  `json:"currency" binding:"required" example:"CNY"`
	Subject             string  `json:"subject" binding:"required" example:"Premium Subscription"`
	Body                string  `json:"body,omitempty" example:"Monthly premium subscription payment"`
	ClientIP            string  `json:"client_ip"`  // Client IP
	NotifyURL           string  `json:"notify_url"` // Async notification URL
	ReturnURL           string  `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	ExpiredMinutes      int     `json:"expired_minutes,omitempty" example:"30"`
	Metadata            string  `json:"metadata,omitempty"` // Additional metadata
}

// CreatePaymentOrderResponse represents the unified response from payment order creation
type CreatePaymentOrderResponse struct {
	PaymentNo   string    `json:"payment_no"`             // Internal payment number
	PaymentURL  string    `json:"payment_url"`            // Payment URL
	QRCodeURL   string    `json:"qr_code_url"`            // QR code URL
	Amount      float64   `json:"amount"`                 // Payment amount
	Currency    string    `json:"currency"`               // Currency
	ExpiredAt   time.Time `json:"expired_at"`             // Expiration time
	GatewayData string    `json:"gateway_data,omitempty"` // Raw gateway response
}

// QueryPaymentOrderResponse represents the unified response from payment order query
type QueryPaymentOrderResponse struct {
	PaymentNo     string `json:"payment_no"`
	Status        string `json:"status"`
	PaidAmount    string `json:"paid_amount,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	PaidAt        string `json:"paid_at,omitempty"`
}

// NotifyData represents the unified notification data
type NotifyData struct {
	PaymentNo     string  `json:"payment_no"`
	OutTradeNo    string  `json:"out_trade_no"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	PaidAt        string  `json:"paid_at,omitempty"`
}

// =====================================================================
// PAYMENT METHOD DTOs - Payment method management
// =====================================================================

// PaymentMethodResponse represents the payment method data structure for API responses
type PaymentMethodResponse struct {
	ID              uint       `json:"id" example:"1"`                                             // Payment method ID
	UserID          uint       `json:"user_id" example:"1"`                                        // User ID
	Type            string     `json:"type" example:"card"`                                        // Payment method type
	Gateway         string     `json:"gateway" example:"epay"`                                     // Payment gateway
	Method          string     `json:"method" example:"alipay"`                                    // Payment method
	DisplayName     string     `json:"display_name" example:"My Alipay Account"`                   // User-friendly name
	MaskedInfo      string     `json:"masked_info" example:"ali***@example.com"`                   // Masked payment info
	Brand           string     `json:"brand,omitempty" example:"Alipay"`                           // Brand/provider
	ExpiryMonth     *int       `json:"expiry_month,omitempty" example:"12"`                        // Expiry month (cards only)
	ExpiryYear      *int       `json:"expiry_year,omitempty" example:"2025"`                       // Expiry year (cards only)
	IsDefault       bool       `json:"is_default" example:"true"`                                  // Is default payment method
	IsActive        bool       `json:"is_active" example:"true"`                                   // Is active
	Status          string     `json:"status" example:"active"`                                    // Status
	LastValidatedAt *time.Time `json:"last_validated_at,omitempty" example:"2024-01-01T00:00:00Z"` // Last validation
	BillingCountry  string     `json:"billing_country,omitempty" example:"CN"`                     // Billing country
	BillingPostcode string     `json:"billing_postcode,omitempty" example:"100000"`                // Billing postcode
	LastUsedAt      *time.Time `json:"last_used_at,omitempty" example:"2024-01-01T00:00:00Z"`      // Last used
	SuccessfulUses  int        `json:"successful_uses" example:"5"`                                // Successful payment count
	FailedUses      int        `json:"failed_uses" example:"1"`                                    // Failed payment count
	CreatedAt       time.Time  `json:"created_at" example:"2024-01-01T00:00:00Z"`                  // Creation time
	UpdatedAt       time.Time  `json:"updated_at" example:"2024-01-01T00:00:00Z"`                  // Update time

	// Computed fields
	IsExpired       bool    `json:"is_expired"`       // Expiration status
	CanBeUsed       bool    `json:"can_be_used"`      // Usability status
	FailureRate     float64 `json:"failure_rate"`     // Failure rate
	NeedsValidation bool    `json:"needs_validation"` // Validation requirement
}

// CreatePaymentMethodRequest represents the request to create a new payment method
type CreatePaymentMethodRequest struct {
	Type              string `json:"type" binding:"required,oneof=card bank_account digital_wallet" example:"card"`
	Gateway           string `json:"gateway" binding:"required" example:"epay"`
	Method            string `json:"method" binding:"required" example:"alipay"`
	DisplayName       string `json:"display_name" binding:"required,max=100" example:"My Alipay Account"`
	PaymentToken      string `json:"payment_token" binding:"required" example:"tok_1234567890"`
	GatewayCustomerID string `json:"gateway_customer_id,omitempty" example:"cus_1234567890"`
	MaskedInfo        string `json:"masked_info" binding:"max=100" example:"ali***@example.com"`
	Brand             string `json:"brand,omitempty" example:"Alipay"`
	ExpiryMonth       *int   `json:"expiry_month,omitempty" example:"12"`
	ExpiryYear        *int   `json:"expiry_year,omitempty" example:"2025"`
	BillingCountry    string `json:"billing_country,omitempty" example:"CN"`
	BillingPostcode   string `json:"billing_postcode,omitempty" example:"100000"`
	SetAsDefault      bool   `json:"set_as_default" example:"false"`
}

// UpdatePaymentMethodRequest represents the request to update a payment method
type UpdatePaymentMethodRequest struct {
	DisplayName     *string `json:"display_name,omitempty" binding:"omitempty,max=100" example:"Updated Payment Method"`
	ExpiryMonth     *int    `json:"expiry_month,omitempty" example:"12"`
	ExpiryYear      *int    `json:"expiry_year,omitempty" example:"2026"`
	BillingCountry  *string `json:"billing_country,omitempty" example:"CN"`
	BillingPostcode *string `json:"billing_postcode,omitempty" example:"100000"`
	Active          *bool   `json:"is_active,omitempty" example:"true"`
}

// SetDefaultPaymentMethodRequest represents the request to set a payment method as default
type SetDefaultPaymentMethodRequest struct {
	PaymentMethodID uint `json:"payment_method_id" binding:"required" example:"1"`
}

// PaymentMethodListResponse represents the response for listing payment methods
type PaymentMethodListResponse struct {
	PaymentMethods []PaymentMethodResponse `json:"payment_methods"`
	Total          int                     `json:"total" example:"5"`
	DefaultMethod  *PaymentMethodResponse  `json:"default_method,omitempty"`
}

// PaymentMethodUsageStats represents usage statistics for a payment method
type PaymentMethodUsageStats struct {
	PaymentMethodID  uint    `json:"payment_method_id"`
	TotalUses        int     `json:"total_uses"`
	SuccessfulUses   int     `json:"successful_uses"`
	FailedUses       int     `json:"failed_uses"`
	SuccessRate      float64 `json:"success_rate"`
	FailureRate      float64 `json:"failure_rate"`
	LastUsed         *string `json:"last_used"`
	AverageAmount    float64 `json:"average_amount"`
	TotalAmount      float64 `json:"total_amount"`
	RecentUses30Days int     `json:"recent_uses_30_days"`
}

// PaymentMethodValidationResult represents the result of payment method validation
type PaymentMethodValidationResult struct {
	IsValid       bool   `json:"is_valid"`
	ErrorMessage  string `json:"error_message,omitempty"`
	GatewayCode   string `json:"gateway_code,omitempty"`
	LastValidated string `json:"last_validated"`
}

// =====================================================================
// PAYMENT CONFIG DTOs - Payment gateway configurations
// =====================================================================

// PaymentConfigResponse represents the payment config data structure for API responses
type PaymentConfigResponse struct {
	ID                  uint              `json:"id" example:"1"`                                             // Config ID
	Method              string            `json:"method" example:"epay"`                                      // Payment method identifier
	Name                string            `json:"name" example:"EPay Payment Method"`                         // Display name
	URL                 string            `json:"url" example:"https://api.example.com"`                      // API endpoint URL
	PID                 string            `json:"pid" example:"partner123"`                                   // Partner/Merchant ID
	Key                 string            `json:"key" example:"secret123"`                                    // API key/secret
	NotifyURL           string            `json:"notify_url,omitempty" example:"https://example.com/webhook"` // Callback URL
	ReturnURL           string            `json:"return_url,omitempty" example:"https://example.com/return"`  // Return URL
	IsEnabled           bool              `json:"is_enabled" example:"true"`                                  // Enabled status
	SortOrder           int               `json:"sort_order" example:"1"`                                     // Sort order
	SupportedCurrencies string            `json:"supported_currencies" example:"CNY"`                         // Supported currencies
	Methods             []entities.Method `json:"methods,omitempty"`                                          // Payment methods
	MinAmount           float64           `json:"min_amount" example:"0.01"`                                  // Minimum amount
	MaxAmount           float64           `json:"max_amount" example:"99999.99"`                              // Maximum amount
	FixedFee            float64           `json:"fixed_fee" example:"0.00"`                                   // Fixed fee
	PercentageFee       float64           `json:"percentage_fee" example:"0.6"`                               // Percentage fee
	CreatedAt           time.Time         `json:"created_at" example:"2024-01-01T00:00:00Z"`                  // Creation time
	UpdatedAt           time.Time         `json:"updated_at" example:"2024-01-01T00:00:00Z"`                  // Update time
}

// CreatePaymentConfigRequest represents the request to create a payment config
type CreatePaymentConfigRequest struct {
	Method              string            `json:"method" binding:"required" example:"epay"`
	Name                string            `json:"name" binding:"required" example:"EPay Payment Method"`
	URL                 string            `json:"url" binding:"required" example:"https://api.example.com"`
	PID                 string            `json:"pid" binding:"required" example:"partner123"`
	Key                 string            `json:"key" binding:"required" example:"secret123"`
	NotifyURL           string            `json:"notify_url,omitempty" example:"https://example.com/webhook"`
	ReturnURL           string            `json:"return_url,omitempty" example:"https://example.com/return"`
	IsEnabled           *bool             `json:"is_enabled,omitempty" example:"true"`
	SortOrder           int               `json:"sort_order,omitempty" example:"1"`
	SupportedCurrencies string            `json:"supported_currencies,omitempty" example:"CNY"`
	Methods             []entities.Method `json:"methods,omitempty"`
	MinAmount           float64           `json:"min_amount,omitempty" example:"0.01"`
	MaxAmount           float64           `json:"max_amount,omitempty" example:"99999.99"`
	FixedFee            float64           `json:"fixed_fee,omitempty" example:"0.00"`
	PercentageFee       float64           `json:"percentage_fee,omitempty" example:"0.6"`
}

// UpdatePaymentConfigRequest represents the request to update a payment config
type UpdatePaymentConfigRequest struct {
	Name                *string           `json:"name,omitempty" example:"EPay Payment Method"`
	URL                 *string           `json:"url,omitempty" example:"https://api.example.com"`
	PID                 *string           `json:"pid,omitempty" example:"partner123"`
	Key                 *string           `json:"key,omitempty" example:"secret123"`
	NotifyURL           *string           `json:"notify_url,omitempty" example:"https://example.com/webhook"`
	ReturnURL           *string           `json:"return_url,omitempty" example:"https://example.com/return"`
	IsEnabled           *bool             `json:"is_enabled,omitempty" example:"true"`
	SortOrder           *int              `json:"sort_order,omitempty" example:"1"`
	SupportedCurrencies *string           `json:"supported_currencies,omitempty" example:"CNY"`
	Methods             []entities.Method `json:"methods,omitempty"`
	MinAmount           *float64          `json:"min_amount,omitempty" example:"0.01"`
	MaxAmount           *float64          `json:"max_amount,omitempty" example:"99999.99"`
	FixedFee            *float64          `json:"fixed_fee,omitempty" example:"0.00"`
	PercentageFee       *float64          `json:"percentage_fee,omitempty" example:"0.6"`
}

// GetPaymentConfigsRequest represents the request to get payment configs
type GetPaymentConfigsRequest struct {
	Method    string `form:"method,omitempty" example:"epay"`
	IsEnabled *bool  `form:"is_enabled,omitempty" example:"true"`
	Limit     int    `form:"limit,omitempty" example:"10"`
	Offset    int    `form:"offset,omitempty" example:"0"`
}

// =====================================================================
// HELPER & CONVERSION FUNCTIONS - Utilities and entity conversions
// =====================================================================

// ------- Validation Functions -------

// ValidateEpayConfig validates epay configuration fields
func ValidateEpayConfig(req *CreatePaymentConfigRequest) []string {
	var errors []string

	// Validate required fields for epay
	if req.Method == "epay" {
		if req.URL == "" {
			errors = append(errors, "URL is required for epay method")
		}
		if req.PID == "" {
			errors = append(errors, "PID (Partner ID) is required for epay method")
		}
		if req.Key == "" {
			errors = append(errors, "Key is required for epay method")
		}

		// Basic URL validation
		if req.URL != "" && !isValidURL(req.URL) {
			errors = append(errors, "Invalid URL format")
		}

		if req.NotifyURL != "" && !isValidURL(req.NotifyURL) {
			errors = append(errors, "Invalid notify URL format")
		}

		if req.ReturnURL != "" && !isValidURL(req.ReturnURL) {
			errors = append(errors, "Invalid return URL format")
		}
	}

	return errors
}

// ValidateEpayUpdateConfig validates epay update configuration fields
func ValidateEpayUpdateConfig(req *UpdatePaymentConfigRequest, existingMethod string) []string {
	var errors []string

	// Only validate if this is an epay configuration
	if existingMethod == "epay" {
		// Validate URLs if provided
		if req.URL != nil && *req.URL != "" && !isValidURL(*req.URL) {
			errors = append(errors, "Invalid URL format")
		}

		if req.NotifyURL != nil && *req.NotifyURL != "" && !isValidURL(*req.NotifyURL) {
			errors = append(errors, "Invalid notify URL format")
		}

		if req.ReturnURL != nil && *req.ReturnURL != "" && !isValidURL(*req.ReturnURL) {
			errors = append(errors, "Invalid return URL format")
		}
	}

	return errors
}

// isValidURL checks if a string is a valid URL
func isValidURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// ------- Security Functions -------

// maskTransactionID partially masks the transaction ID for security
func maskTransactionID(transactionID string) string {
	if transactionID == "" {
		return ""
	}

	// Test expectation: mask positions 4..= (len-4), keep first 3 and last 3 for 12+ length
	// For short IDs (<= 6), mask all
	id := transactionID
	n := len(id)
	if n <= 6 {
		return strings.Repeat("*", n)
	}
	// Keep first 3 and last 3
	prefix := id[:3]
	suffix := id[n-3:]
	return prefix + strings.Repeat("*", n-6) + suffix
}

// getSecurePaymentURL returns payment URL only if payment is still pending and not expired
func getSecurePaymentURL(pr *entities.PaymentRecord) string {
	if pr.Status == constants.PaymentRecordStatusPending && !pr.IsExpired() {
		return pr.PaymentURL
	}
	return ""
}

// getSecureQRCodeURL returns QR code URL only if payment is still pending and not expired
func getSecureQRCodeURL(pr *entities.PaymentRecord) string {
	if pr.Status == constants.PaymentRecordStatusPending && !pr.IsExpired() {
		return pr.QRCodeURL
	}
	return ""
}

// maskSensitiveKey masks sensitive key for API responses
func maskSensitiveKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

// ------- Parsing Functions -------

// parseSupportedMethods parses the SupportedMethods JSON string to []entities.Method
func parseSupportedMethods(supportedMethods string) []entities.Method {
	if supportedMethods == "" {
		return nil
	}

	var methods []entities.Method
	if err := json.Unmarshal([]byte(supportedMethods), &methods); err != nil {
		// If parsing fails, return empty slice
		return nil
	}

	return methods
}

// ------- Conversion Functions -------

// ToPaymentRecordUserResponse converts a PaymentRecord entity to PaymentRecordResponse for user-facing APIs
func ToPaymentRecordUserResponse(pr *entities.PaymentRecord) *PaymentRecordResponse {
	return &PaymentRecordResponse{
		ID:                  pr.ID,
		UserID:              pr.UserID,
		SubscriptionOrderID: pr.SubscriptionOrderID,
		PaymentNo:           pr.PaymentNo,
		OutTradeNo:          pr.OutTradeNo,
		TransactionID:       maskTransactionID(pr.TransactionID),
		Gateway:             pr.Gateway,
		PaymentMethod:       pr.PaymentMethod,
		Amount:              pr.Amount,
		Currency:            pr.Currency,
		ExchangeRate:        pr.ExchangeRate,
		Status:              pr.Status,
		PaymentStatus:       pr.PaymentStatus,
		PaymentURL:          getSecurePaymentURL(pr),
		QRCodeURL:           getSecureQRCodeURL(pr),
		ExpiredAt:           pr.ExpiredAt,
		PaidAt:              pr.PaidAt,
		NotifiedAt:          pr.NotifiedAt,
		RefundAmount:        pr.RefundAmount,
		RefundStatus:        pr.RefundStatus,
		RefundedAt:          pr.RefundedAt,
		RefundReason:        pr.RefundReason,
		Remark:              pr.Remark,
		CreatedAt:           pr.CreatedAt,
		UpdatedAt:           pr.UpdatedAt,
		IsExpired:           pr.IsExpired(),
		CanRefund:           pr.CanBeRefunded(),
		RefundableAmount:    pr.GetRefundableAmount(),
	}
}

// ToPaymentRecordSecureResponse converts a PaymentRecord entity to a secure response for tests
func ToPaymentRecordSecureResponse(pr *entities.PaymentRecord) *PaymentRecordResponse {
	// This is essentially the same as ToPaymentRecordUserResponse but can be extended
	// with additional security measures if needed
	return ToPaymentRecordUserResponse(pr)
}

// ToPaymentMethodResponse converts a PaymentMethod entity to PaymentMethodResponse
func ToPaymentMethodResponse(pm *entities.PaymentMethod) *PaymentMethodResponse {
	return &PaymentMethodResponse{
		ID:              pm.ID,
		UserID:          pm.UserID,
		Type:            pm.Type,
		Method:          pm.Method,
		DisplayName:     pm.DisplayName,
		MaskedInfo:      pm.MaskedInfo,
		Brand:           pm.Brand,
		ExpiryMonth:     pm.ExpiryMonth,
		ExpiryYear:      pm.ExpiryYear,
		Status:          pm.Status,
		IsDefault:       pm.IsDefault,
		SuccessfulUses:  pm.SuccessfulUses,
		FailedUses:      pm.FailedUses,
		LastUsedAt:      pm.LastUsedAt,
		BillingCountry:  pm.BillingCountry,
		BillingPostcode: pm.BillingPostcode,
		CreatedAt:       pm.CreatedAt,
		UpdatedAt:       pm.UpdatedAt,
		IsActive:        pm.IsActive(),
		CanBeUsed:       pm.CanBeUsedForPayment(),
		FailureRate:     pm.GetFailureRate(),
		NeedsValidation: pm.NeedsRevalidation(),
	}
}

// ToPaymentConfigResponse converts a PaymentConfig entity to PaymentConfigResponse for admin APIs
func ToPaymentConfigResponse(pc *entities.PaymentConfig) *PaymentConfigResponse {
	return &PaymentConfigResponse{
		ID:                  pc.ID,
		Method:              pc.Method,
		Name:                pc.Name,
		URL:                 pc.URL,
		PID:                 pc.PID,
		Key:                 maskSensitiveKey(pc.Key),
		NotifyURL:           pc.NotifyURL,
		ReturnURL:           pc.ReturnURL,
		SupportedCurrencies: pc.SupportedCurrencies,
		Methods:             parseSupportedMethods(pc.SupportedMethods),
		MinAmount:           pc.MinAmount,
		MaxAmount:           pc.MaxAmount,
		IsEnabled:           pc.IsEnabled,
		SortOrder:           pc.SortOrder,
		CreatedAt:           pc.CreatedAt,
		UpdatedAt:           pc.UpdatedAt,
	}
}

// ToPaymentConfigResponseWithFullKey converts PaymentConfig to response with full key (for debugging)
func ToPaymentConfigResponseWithFullKey(pc *entities.PaymentConfig) *PaymentConfigResponse {
	return &PaymentConfigResponse{
		ID:                  pc.ID,
		Method:              pc.Method,
		Name:                pc.Name,
		URL:                 pc.URL,
		PID:                 pc.PID,
		Key:                 pc.Key, // 不掩码，显示完整key
		NotifyURL:           pc.NotifyURL,
		ReturnURL:           pc.ReturnURL,
		SupportedCurrencies: pc.SupportedCurrencies,
		Methods:             parseSupportedMethods(pc.SupportedMethods),
		MinAmount:           pc.MinAmount,
		MaxAmount:           pc.MaxAmount,
		IsEnabled:           pc.IsEnabled,
		SortOrder:           pc.SortOrder,
		CreatedAt:           pc.CreatedAt,
		UpdatedAt:           pc.UpdatedAt,
	}
}

// ToPaymentConfigPublicResponse converts a PaymentConfig entity to public response (without sensitive data)
func ToPaymentConfigPublicResponse(pc *entities.PaymentConfig) *PaymentConfigResponse {
	return &PaymentConfigResponse{
		ID:                  pc.ID,
		Method:              pc.Method,
		Name:                pc.Name,
		SupportedCurrencies: pc.SupportedCurrencies,
		Methods:             parseSupportedMethods(pc.SupportedMethods),
		MinAmount:           pc.MinAmount,
		MaxAmount:           pc.MaxAmount,
		IsEnabled:           pc.IsEnabled,
		SortOrder:           pc.SortOrder,
		CreatedAt:           pc.CreatedAt,
		UpdatedAt:           pc.UpdatedAt,
	}
}

// =====================================================================
// OBJECT POOLS - Memory optimization for frequently used DTOs
// =====================================================================

var (
	// Pool for PaymentRecordResponse objects
	paymentRecordResponsePool = pool.NewPool(
		func() *PaymentRecordResponse {
			return &PaymentRecordResponse{}
		},
		func(resp *PaymentRecordResponse) {
			*resp = PaymentRecordResponse{}
		},
	)

	// Pool for CreatePaymentOrderRequest objects
	createPaymentOrderRequestPool = pool.NewPool(
		func() *CreatePaymentOrderRequest {
			return &CreatePaymentOrderRequest{}
		},
		func(req *CreatePaymentOrderRequest) {
			*req = CreatePaymentOrderRequest{}
		},
	)

	// Pool for NotifyData objects
	notifyDataPool = pool.NewPool(
		func() *NotifyData {
			return &NotifyData{}
		},
		func(data *NotifyData) {
			*data = NotifyData{}
		},
	)
)

// GetPaymentRecordResponse gets a PaymentRecordResponse from the pool
func GetPaymentRecordResponse() *PaymentRecordResponse {
	return paymentRecordResponsePool.Get()
}

// PutPaymentRecordResponse returns a PaymentRecordResponse to the pool
func PutPaymentRecordResponse(resp *PaymentRecordResponse) {
	if resp == nil {
		return
	}
	paymentRecordResponsePool.Put(resp)
}

// GetCreatePaymentOrderRequest gets a CreatePaymentOrderRequest from the pool
func GetCreatePaymentOrderRequest() *CreatePaymentOrderRequest {
	return createPaymentOrderRequestPool.Get()
}

// PutCreatePaymentOrderRequest returns a CreatePaymentOrderRequest to the pool
func PutCreatePaymentOrderRequest(req *CreatePaymentOrderRequest) {
	if req == nil {
		return
	}
	createPaymentOrderRequestPool.Put(req)
}

// GetNotifyData gets a NotifyData from the pool
func GetNotifyData() *NotifyData {
	return notifyDataPool.Get()
}

// PutNotifyData returns a NotifyData to the pool
func PutNotifyData(data *NotifyData) {
	if data == nil {
		return
	}
	notifyDataPool.Put(data)
}
