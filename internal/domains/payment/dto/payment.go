package dto

import (
	"strings"
	"time"

	"linke/internal/domains/payment/constants"
	"linke/internal/domains/payment/entities"
	"linke/internal/shared/dto"
)

// ==================== Payment Record DTOs ====================

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
	ClientIP            string  `json:"client_ip"`           // Client IP
	NotifyURL           string  `json:"notify_url"`          // Async notification URL
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

// ==================== Payment Method DTOs ====================

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
	Type              string `json:"type" binding:"required,oneof=card bank_account digital_wallet crypto" example:"card"`
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

// ==================== Payment Config DTOs ====================

// PaymentConfigResponse represents the payment config data structure for API responses
type PaymentConfigResponse struct {
	ID                  uint                 `json:"id" example:"1"`                            // Config ID
	Gateway             string               `json:"gateway" example:"epay"`                    // Payment gateway
	Name                string               `json:"name" example:"EPay Gateway"`               // Display name
	IsEnabled           bool                 `json:"is_enabled" example:"true"`                 // Enabled status
	SortOrder           int                  `json:"sort_order" example:"1"`                    // Sort order
	SupportedCurrencies string               `json:"supported_currencies" example:"CNY"`        // Supported currencies
	Methods             []entities.Method    `json:"methods,omitempty"`                         // Payment methods
	MinAmount           float64              `json:"min_amount" example:"0.01"`                 // Minimum amount
	MaxAmount           float64              `json:"max_amount" example:"99999.99"`             // Maximum amount
	FixedFee            float64              `json:"fixed_fee" example:"0.00"`                  // Fixed fee
	PercentageFee       float64              `json:"percentage_fee" example:"0.6"`              // Percentage fee
	CreatedAt           time.Time            `json:"created_at" example:"2024-01-01T00:00:00Z"` // Creation time
	UpdatedAt           time.Time            `json:"updated_at" example:"2024-01-01T00:00:00Z"` // Update time
}

// CreatePaymentConfigRequest represents the request to create a payment config
type CreatePaymentConfigRequest struct {
	Gateway             string            `json:"gateway" binding:"required" example:"epay"`
	Name                string            `json:"name" binding:"required" example:"EPay Gateway"`
	Config              string            `json:"config" binding:"required" example:"{\"api_url\":\"...\"}"`
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
	Name                *string           `json:"name,omitempty" example:"EPay Gateway"`
	Config              *string           `json:"config,omitempty" example:"{\"api_url\":\"...\"}"`
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
	Gateway   string `form:"gateway,omitempty" example:"epay"`
	IsEnabled *bool  `form:"is_enabled,omitempty" example:"true"`
	Limit     int    `form:"limit,omitempty" example:"10"`
	Offset    int    `form:"offset,omitempty" example:"0"`
}

// ==================== Payment Retry DTOs ====================

// PaymentRetryResponse represents the payment retry data structure for API responses
type PaymentRetryResponse struct {
	ID              uint       `json:"id"`
	PaymentRecordID uint       `json:"payment_record_id"`
	AttemptNumber   int        `json:"attempt_number"`
	MaxAttempts     int        `json:"max_attempts"`
	NextRetryAt     time.Time  `json:"next_retry_at" swaggertype:"string" format:"date-time" example:"2024-01-01T12:00:00Z"`
	LastAttemptAt   time.Time  `json:"last_attempt_at" swaggertype:"string" format:"date-time" example:"2024-01-01T11:00:00Z"`
	RetryStrategy   string     `json:"retry_strategy"`
	Status          string     `json:"status"`
	FailureType     string     `json:"failure_type,omitempty"`
	LastFailureCode string     `json:"last_failure_code,omitempty"`
	TotalDelayTime  int        `json:"total_delay_time"`
	CompletedAt     *time.Time `json:"completed_at,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-01T13:00:00Z"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-01T12:30:00Z"`
	SuccessfulAt    *time.Time `json:"successful_at,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-01T12:15:00Z"`
	Notes           string     `json:"notes,omitempty"`
	CreatedAt       time.Time  `json:"created_at" swaggertype:"string" format:"date-time" example:"2024-01-01T10:00:00Z"`
	UpdatedAt       time.Time  `json:"updated_at" swaggertype:"string" format:"date-time" example:"2024-01-01T10:00:00Z"`
}

// PaymentRetryHistoryResponse represents the retry history response
type PaymentRetryHistoryResponse struct {
	ID                uint       `json:"id"`
	AttemptNumber     int        `json:"attempt_number"`
	AttemptedAt       time.Time  `json:"attempted_at" swaggertype:"string" format:"date-time" example:"2024-01-01T12:00:00Z"`
	Duration          int        `json:"duration"`
	Status            string     `json:"status"`
	ResponseCode      string     `json:"response_code,omitempty"`
	ResponseMessage   string     `json:"response_message,omitempty"`
	ErrorType         string     `json:"error_type,omitempty"`
	FailureReason     string     `json:"failure_reason,omitempty"`
	NextRetryAt       *time.Time `json:"next_retry_at,omitempty" swaggertype:"string" format:"date-time" example:"2024-01-01T13:00:00Z"`
	DelayFromPrevious int        `json:"delay_from_previous"`
	CreatedAt         time.Time  `json:"created_at" swaggertype:"string" format:"date-time" example:"2024-01-01T12:00:00Z"`
}

// RetryResult represents the result of a retry attempt
type RetryResult struct {
	Success       bool                        `json:"success"`
	RetryID       uint                        `json:"retry_id"`
	AttemptNumber int                         `json:"attempt_number"`
	Status        string                      `json:"status"`
	NextRetryAt   *time.Time                  `json:"next_retry_at,omitempty"`
	CompletedAt   *time.Time                  `json:"completed_at,omitempty"`
	ErrorMessage  string                      `json:"error_message,omitempty"`
	PaymentResult *PaymentProcessResult       `json:"payment_result,omitempty"`
	History       *PaymentRetryHistoryResponse `json:"history,omitempty"`
}

// PaymentProcessResult represents the result of payment processing
type PaymentProcessResult struct {
	PaymentRecordID uint       `json:"payment_record_id"`
	Status          string     `json:"status"`
	TransactionID   string     `json:"transaction_id,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	ErrorCode       string     `json:"error_code,omitempty"`
	ErrorMessage    string     `json:"error_message,omitempty"`
}

// RetryWithHistory represents a retry with its complete history
type RetryWithHistory struct {
	Retry   *PaymentRetryResponse           `json:"retry"`
	History []*PaymentRetryHistoryResponse  `json:"history"`
}

// RetryFilters represents filters for retry queries
type RetryFilters struct {
	UserID        *uint      `json:"user_id,omitempty"`
	Gateway       *string    `json:"gateway,omitempty"`
	PaymentMethod *string    `json:"payment_method,omitempty"`
	Status        *string    `json:"status,omitempty"`
	FailureType   *string    `json:"failure_type,omitempty"`
	FromDate      *time.Time `json:"from_date,omitempty"`
	ToDate        *time.Time `json:"to_date,omitempty"`
	MinAttempts   *int       `json:"min_attempts,omitempty"`
	MaxAttempts   *int       `json:"max_attempts,omitempty"`
}

// FailureAnalysis represents analysis of payment failures
type FailureAnalysis struct {
	Gateway             string            `json:"gateway"`
	TotalFailures       int64             `json:"total_failures"`
	TemporaryFailures   int64             `json:"temporary_failures"`
	PermanentFailures   int64             `json:"permanent_failures"`
	NetworkFailures     int64             `json:"network_failures"`
	GatewayFailures     int64             `json:"gateway_failures"`
	BusinessFailures    int64             `json:"business_failures"`
	TopFailureReasons   []*FailureReason  `json:"top_failure_reasons"`
	FailurePatterns     []*FailurePattern `json:"failure_patterns"`
	RecoveryRate        float64           `json:"recovery_rate"`
	AverageRecoveryTime float64           `json:"average_recovery_time"`
	RecommendedActions  []string          `json:"recommended_actions"`
}

// FailureReason represents a common failure reason
type FailureReason struct {
	Reason           string  `json:"reason"`
	Count            int64   `json:"count"`
	Percentage       float64 `json:"percentage"`
	RetrySuccessRate float64 `json:"retry_success_rate"`
}

// FailurePattern represents a failure pattern
type FailurePattern struct {
	Pattern         string  `json:"pattern"`
	Occurrences     int64   `json:"occurrences"`
	Percentage      float64 `json:"percentage"`
	// Additional fields for compatibility
	ErrorType       string  `json:"error_type"`
	FailureReason   string  `json:"failure_reason"`
	Count           int64   `json:"count"`
	SuccessRate     float64 `json:"success_rate"`
	AverageAttempts float64 `json:"average_attempts"`
}

// RetryStatistics represents retry statistics for admin interface
type RetryStatistics struct {
	Gateway                 string                   `json:"gateway"`
	PaymentMethod           string                   `json:"payment_method,omitempty"`
	DateRange              string                   `json:"date_range"`
	TotalRetries           int64                    `json:"total_retries"`
	SuccessfulRetries      int64                    `json:"successful_retries"`
	FailedRetries          int64                    `json:"failed_retries"`
	CancelledRetries       int64                    `json:"cancelled_retries"`
	SuccessRate            float64                  `json:"success_rate"`
	AverageAttempts        float64                  `json:"average_attempts"`
	AverageRetryDelay      float64                  `json:"average_retry_delay"`
	AverageDelayTime       float64                  `json:"average_delay_time"`
	FromDate               time.Time                `json:"from_date"`
	ToDate                 time.Time                `json:"to_date"`
	PaymentMethodStats     []*PaymentMethodStats    `json:"payment_method_stats"`
	FailureAnalysis        *FailureAnalysis         `json:"failure_analysis"`
	TrendData              []*RetryTrendData        `json:"trend_data"`
}

// PaymentMethodStats represents statistics for a specific payment method
type PaymentMethodStats struct {
	PaymentMethod   string  `json:"payment_method"`
	TotalRetries    int64   `json:"total_retries"`
	SuccessfulRetries int64 `json:"successful_retries"`
	SuccessRate     float64 `json:"success_rate"`
	AverageAttempts float64 `json:"average_attempts"`
}

// RetryTrendData represents trend data for a specific time period
type RetryTrendData struct {
	Date          string  `json:"date"`
	TotalRetries  int64   `json:"total_retries"`
	SuccessRate   float64 `json:"success_rate"`
	FailureRate   float64 `json:"failure_rate"`
}

// RetryHealthMetrics represents overall retry system health
type RetryHealthMetrics struct {
	TotalActiveRetries    int64                  `json:"total_active_retries"`
	OverdueRetries        int64                  `json:"overdue_retries"`
	SuccessRate24h        float64                `json:"success_rate_24h"`
	SuccessRate7d         float64                `json:"success_rate_7d"`
	AverageRetryDelay     float64                `json:"average_retry_delay"`
	GatewayHealth         []*GatewayHealthMetric `json:"gateway_health"`
	AlertsTriggered       []string               `json:"alerts_triggered"`
	SystemRecommendations []string               `json:"system_recommendations"`
}

// GatewayHealthMetric represents health metrics for a specific gateway
type GatewayHealthMetric struct {
	Gateway         string  `json:"gateway"`
	ActiveRetries   int64   `json:"active_retries"`
	SuccessRate     float64 `json:"success_rate"`
	AverageAttempts float64 `json:"average_attempts"`
	QueueDepth      int64   `json:"queue_depth"`
	ProcessingRate  float64 `json:"processing_rate"`
	HealthStatus    string  `json:"health_status"` // healthy, degraded, critical
}

// AdminRetryFilters represents filters for admin retry queries
type AdminRetryFilters struct {
	*RetryFilters
	IncludeHistory   bool     `json:"include_history,omitempty"`
	SortBy           string   `json:"sort_by,omitempty"`    // created_at, next_retry_at, attempt_number
	SortOrder        string   `json:"sort_order,omitempty"` // asc, desc
	PaymentRecordIDs []uint   `json:"payment_record_ids,omitempty"`
	UserIDs          []uint   `json:"user_ids,omitempty"`
	RetryStatuses    []string `json:"retry_statuses,omitempty"`
	FailureTypes     []string `json:"failure_types,omitempty"`
	Gateways         []string `json:"gateways,omitempty"`
	PaymentMethods   []string `json:"payment_methods,omitempty"`
	Limit            int      `json:"limit,omitempty"`
	Offset           int      `json:"offset,omitempty"`
}

// AdminRetryResponse represents response for admin retry queries
type AdminRetryResponse struct {
	Retries    []*PaymentRetryResponse `json:"retries"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
	TotalPages int                     `json:"total_pages"`
	Statistics *AdminRetryStatistics   `json:"statistics,omitempty"`
}

// AdminRetryStatistics represents statistics for admin interface
type AdminRetryStatistics struct {
	TotalRetries       int64   `json:"total_retries"`
	PendingRetries     int64   `json:"pending_retries"`
	InProgressRetries  int64   `json:"in_progress_retries"`
	CompletedRetries   int64   `json:"completed_retries"`
	FailedRetries      int64   `json:"failed_retries"`
	CancelledRetries   int64   `json:"cancelled_retries"`
	OverallSuccessRate float64 `json:"overall_success_rate"`
	AverageAttempts    float64 `json:"average_attempts"`
	AverageDelayTime   float64 `json:"average_delay_time"`
}

// ==================== Handler-specific Request DTOs ====================

// CancelRetryRequest represents the request to cancel a retry
type CancelRetryRequest struct {
	Reason string `json:"reason" binding:"required" example:"Manual cancellation by admin"`
}

// BulkRetryActionRequest represents the request for bulk retry actions
type BulkRetryActionRequest struct {
	RetryIDs []uint `json:"retry_ids" binding:"required"`
	Reason   string `json:"reason,omitempty" example:"Bulk operation by admin"`
}

// ==================== Retry Configuration DTOs ====================

// RetryConfiguration represents system-wide retry configuration
type RetryConfiguration struct {
	Enabled                    bool                                     `json:"enabled"`
	MaxConcurrentRetries       int                                      `json:"max_concurrent_retries"`
	ProcessingInterval         string                                   `json:"processing_interval" swaggertype:"string"`
	HealthCheckInterval        string                                   `json:"health_check_interval" swaggertype:"string"`
	DefaultStrategy            *entities.RetryStrategyConfig            `json:"default_strategy"`
	GatewayStrategies          map[string]*entities.RetryStrategyConfig `json:"gateway_strategies"`
	FailureClassificationRules map[string]*FailureClassificationRule    `json:"failure_classification_rules"`
	NotificationSettings       *RetryNotificationSettings               `json:"notification_settings"`
	MonitoringSettings         *RetryMonitoringSettings                 `json:"monitoring_settings"`
}

// FailureClassificationRule represents rules for classifying payment failures
type FailureClassificationRule struct {
	Gateway       string   `json:"gateway"`
	PaymentMethod string   `json:"payment_method,omitempty"`
	ErrorCodes    []string `json:"error_codes"`
	ErrorPatterns []string `json:"error_patterns"`
	FailureType   string   `json:"failure_type"`
	ShouldRetry   bool     `json:"should_retry"`
	Priority      int      `json:"priority"`
}

// RetryNotificationSettings represents notification settings for retries
type RetryNotificationSettings struct {
	Enabled             bool     `json:"enabled"`
	NotifyOnFailure     bool     `json:"notify_on_failure"`
	NotifyOnSuccess     bool     `json:"notify_on_success"`
	NotifyOnMaxAttempts bool     `json:"notify_on_max_attempts"`
	EmailTemplates      []string `json:"email_templates"`
	SMSTemplates        []string `json:"sms_templates"`
	WebhookURLs         []string `json:"webhook_urls"`
}

// RetryMonitoringSettings represents monitoring settings for retry system
type RetryMonitoringSettings struct {
	MetricsEnabled     bool             `json:"metrics_enabled"`
	AlertsEnabled      bool             `json:"alerts_enabled"`
	HealthCheckEnabled bool             `json:"health_check_enabled"`
	LogLevel           string           `json:"log_level"`
	RetentionPeriod    string           `json:"retention_period" swaggertype:"string"`
	AlertThresholds    *AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds represents alert thresholds for retry monitoring
type AlertThresholds struct {
	MaxPendingRetries  int     `json:"max_pending_retries"`
	MaxOverdueRetries  int     `json:"max_overdue_retries"`
	MinSuccessRate     float64 `json:"min_success_rate"`
	MaxAverageAttempts float64 `json:"max_average_attempts"`
	MaxProcessingDelay int     `json:"max_processing_delay"` // seconds
}

// DailyRetryStats represents daily retry statistics
type DailyRetryStats struct {
	Date              time.Time `json:"date"`
	TotalRetries      int64     `json:"total_retries"`
	SuccessfulRetries int64     `json:"successful_retries"`
	FailedRetries     int64     `json:"failed_retries"`
	SuccessRate       float64   `json:"success_rate"`
}

// AttemptStatistics represents statistics for retry attempts
type AttemptStatistics struct {
	RetryID         uint    `json:"retry_id"`
	TotalAttempts   int     `json:"total_attempts"`
	SuccessfulCount int     `json:"successful_count"`
	FailedCount     int     `json:"failed_count"`
	TimeoutCount    int     `json:"timeout_count"`
	ErrorCount      int     `json:"error_count"`
	AverageDuration float64 `json:"average_duration"`
	TotalDuration   int     `json:"total_duration"`
}

// ==================== Conversion Functions ====================

// ToPaymentRecordResponse converts PaymentRecord entity to PaymentRecordResponse DTO
func ToPaymentRecordResponse(pr *entities.PaymentRecord) *PaymentRecordResponse {
	if pr == nil {
		return nil
	}

	return &PaymentRecordResponse{
		ID:                  pr.ID,
		UserID:              pr.UserID,
		SubscriptionOrderID: pr.SubscriptionOrderID,
		PaymentNo:           pr.PaymentNo,
		OutTradeNo:          pr.OutTradeNo,
		TransactionID:       pr.TransactionID,
		Gateway:             pr.Gateway,
		PaymentMethod:       pr.PaymentMethod,
		Amount:              pr.Amount,
		Currency:            pr.Currency,
		ExchangeRate:        pr.ExchangeRate,
		Status:              pr.Status,
		PaymentStatus:       pr.PaymentStatus,
		PaymentURL:          pr.PaymentURL,
		QRCodeURL:           pr.QRCodeURL,
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

		// Computed fields
		IsExpired:        pr.IsExpired(),
		CanRefund:        pr.CanBeRefunded(),
		RefundableAmount: pr.GetRefundableAmount(),
	}
}

// ToPaymentRecordUserResponse converts PaymentRecord entity to user-safe PaymentRecordResponse DTO
func ToPaymentRecordUserResponse(pr *entities.PaymentRecord) *PaymentRecordResponse {
	if pr == nil {
		return nil
	}

	resp := ToPaymentRecordResponse(pr)
	// Remove sensitive information for user
	resp.User = nil
	resp.NotifiedAt = nil
	return resp
}

// ToPaymentRecordSecureResponse converts PaymentRecord entity to secure PaymentRecordResponse DTO
func ToPaymentRecordSecureResponse(pr *entities.PaymentRecord) *PaymentRecordResponse {
	if pr == nil {
		return nil
	}

	resp := &PaymentRecordResponse{
		ID:                  pr.ID,
		UserID:              pr.UserID,
		SubscriptionOrderID: pr.SubscriptionOrderID,
		PaymentNo:           pr.PaymentNo,
		TransactionID:       maskTransactionID(pr.TransactionID),
		Gateway:             pr.Gateway,
		PaymentMethod:       pr.PaymentMethod,
		Amount:              pr.Amount,
		Currency:            pr.Currency,
		Status:              pr.Status,
		PaymentURL:          getSecurePaymentURL(pr),
		QRCodeURL:           getSecureQRCodeURL(pr),
		ExpiredAt:           pr.ExpiredAt,
		PaidAt:              pr.PaidAt,
		RefundAmount:        pr.RefundAmount,
		RefundStatus:        pr.RefundStatus,
		RefundedAt:          pr.RefundedAt,
		RefundReason:        pr.RefundReason,
		CreatedAt:           pr.CreatedAt,
		UpdatedAt:           pr.UpdatedAt,

		// Computed fields
		IsExpired:        pr.IsExpired(),
		CanRefund:        pr.CanBeRefunded(),
		RefundableAmount: pr.GetRefundableAmount(),
	}

	return resp
}

// ToPaymentMethodResponse converts PaymentMethod entity to PaymentMethodResponse DTO
func ToPaymentMethodResponse(pm *entities.PaymentMethod) *PaymentMethodResponse {
	if pm == nil {
		return nil
	}

	return &PaymentMethodResponse{
		ID:              pm.ID,
		UserID:          pm.UserID,
		Type:            pm.Type,
		Gateway:         pm.Gateway,
		Method:          pm.Method,
		DisplayName:     pm.DisplayName,
		MaskedInfo:      pm.MaskedInfo,
		Brand:           pm.Brand,
		ExpiryMonth:     pm.ExpiryMonth,
		ExpiryYear:      pm.ExpiryYear,
		IsDefault:       pm.IsDefault,
		IsActive:        pm.Active,
		Status:          pm.Status,
		LastValidatedAt: pm.LastValidatedAt,
		BillingCountry:  pm.BillingCountry,
		BillingPostcode: pm.BillingPostcode,
		LastUsedAt:      pm.LastUsedAt,
		SuccessfulUses:  pm.SuccessfulUses,
		FailedUses:      pm.FailedUses,
		CreatedAt:       pm.CreatedAt,
		UpdatedAt:       pm.UpdatedAt,

		// Computed fields
		IsExpired:       pm.IsExpired(),
		CanBeUsed:       pm.CanBeUsedForPayment(),
		FailureRate:     pm.GetFailureRate(),
		NeedsValidation: pm.NeedsRevalidation(),
	}
}

// ToPaymentMethodSecureResponse converts PaymentMethod entity to secure PaymentMethodResponse DTO
func ToPaymentMethodSecureResponse(pm *entities.PaymentMethod) *PaymentMethodResponse {
	if pm == nil {
		return nil
	}

	resp := ToPaymentMethodResponse(pm)
	// Remove sensitive information
	resp.UserID = 0 // Remove user ID for external APIs
	return resp
}

// ToPaymentConfigResponse converts PaymentConfig entity to PaymentConfigResponse DTO
func ToPaymentConfigResponse(pc *entities.PaymentConfig) *PaymentConfigResponse {
	if pc == nil {
		return nil
	}

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

// ToPaymentConfigPublicResponse converts PaymentConfig entity to public PaymentConfigResponse DTO
func ToPaymentConfigPublicResponse(pc *entities.PaymentConfig) *PaymentConfigResponse {
	if pc == nil {
		return nil
	}

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
		var publicMethods []entities.Method
		for _, method := range methods {
			if method.IsEnabled {
				publicMethods = append(publicMethods, method)
			}
		}
		response.Methods = publicMethods
	}

	return response
}

// ToPaymentRetryResponse converts PaymentRetry entity to PaymentRetryResponse DTO
func ToPaymentRetryResponse(pr *entities.PaymentRetry) *PaymentRetryResponse {
	if pr == nil {
		return nil
	}

	return &PaymentRetryResponse{
		ID:              pr.ID,
		PaymentRecordID: pr.PaymentRecordID,
		AttemptNumber:   pr.AttemptNumber,
		MaxAttempts:     pr.MaxAttempts,
		NextRetryAt:     pr.NextRetryAt,
		LastAttemptAt:   pr.LastAttemptAt,
		RetryStrategy:   pr.RetryStrategy,
		Status:          pr.Status,
		FailureType:     pr.FailureType,
		LastFailureCode: pr.LastFailureCode,
		TotalDelayTime:  pr.TotalDelayTime,
		CompletedAt:     pr.CompletedAt,
		CancelledAt:     pr.CancelledAt,
		SuccessfulAt:    pr.SuccessfulAt,
		Notes:           pr.Notes,
		CreatedAt:       pr.CreatedAt,
		UpdatedAt:       pr.UpdatedAt,
	}
}

// ToPaymentRetryHistoryResponse converts PaymentRetryHistory entity to PaymentRetryHistoryResponse DTO
func ToPaymentRetryHistoryResponse(prh *entities.PaymentRetryHistory) *PaymentRetryHistoryResponse {
	if prh == nil {
		return nil
	}

	return &PaymentRetryHistoryResponse{
		ID:                prh.ID,
		AttemptNumber:     prh.AttemptNumber,
		AttemptedAt:       prh.AttemptedAt,
		Duration:          prh.Duration,
		Status:            prh.Status,
		ResponseCode:      prh.ResponseCode,
		ResponseMessage:   prh.ResponseMessage,
		ErrorType:         prh.ErrorType,
		FailureReason:     prh.FailureReason,
		NextRetryAt:       prh.NextRetryAt,
		DelayFromPrevious: prh.DelayFromPrevious,
		CreatedAt:         prh.CreatedAt,
	}
}

// ==================== Helper Functions ====================

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