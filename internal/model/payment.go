package model

import (
	"time"

	"gorm.io/gorm"
)

// Payment represents a payment (fund transfer record)
type Payment struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	InvoiceID uint `json:"invoice_id" gorm:"not null;index"`
	UserID    uint `json:"user_id" gorm:"not null;index"`

	// Payment Information
	PaymentNumber      string `json:"payment_number" gorm:"uniqueIndex;size:32;not null"`
	PaymentIntentID    string `json:"payment_intent_id,omitempty" gorm:"size:255;index"`
	Status             string `json:"status" gorm:"size:20;not null;default:'pending';index"` // pending, processing, completed, failed, cancelled

	// Amount Information
	Amount       float64 `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency     string  `json:"currency" gorm:"size:3;not null;default:'USD'"`
	ExchangeRate float64 `json:"exchange_rate" gorm:"type:decimal(10,6);default:1.0"`

	// Payment Channel Information
	PaymentMethod        string  `json:"payment_method" gorm:"size:50;not null;index"`        // alipay, wechat, credit_card, etc.
	PaymentGateway       string  `json:"payment_gateway" gorm:"size:50;not null;index"`       // stripe, epay, epusdt, etc.
	GatewayTransactionID string  `json:"gateway_transaction_id,omitempty" gorm:"size:255;index"`
	GatewayFee           float64 `json:"gateway_fee" gorm:"type:decimal(10,2);default:0"`

	// Payment Details
	PaymentURL  string `json:"payment_url,omitempty" gorm:"type:text"`
	QRCodeURL   string `json:"qr_code_url,omitempty" gorm:"type:text"`
	RedirectURL string `json:"redirect_url,omitempty" gorm:"type:text"`

	// Time Information
	ExpiresAt   *time.Time `json:"expires_at,omitempty" gorm:"index"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" gorm:"index"`
	CompletedAt *time.Time `json:"completed_at,omitempty" gorm:"index"`

	// Refund Information
	RefundAmount    float64    `json:"refund_amount" gorm:"type:decimal(10,2);default:0"`
	RefundedAt      *time.Time `json:"refunded_at,omitempty" gorm:"index"`
	RefundReason    string     `json:"refund_reason,omitempty" gorm:"type:text"`
	RefundReference string     `json:"refund_reference,omitempty" gorm:"size:255"`

	// Notification Information
	WebhookData          string     `json:"webhook_data,omitempty" gorm:"type:json"`
	NotificationCount    int        `json:"notification_count" gorm:"default:0"`
	LastNotificationAt   *time.Time `json:"last_notification_at,omitempty" gorm:"index"`

	// Business Fields
	Notes    string `json:"notes,omitempty" gorm:"type:text"`
	Metadata string `json:"metadata,omitempty" gorm:"type:json"`

	// Relationships (no foreign key constraints for performance)
	Invoice *Invoice `json:"invoice,omitempty" gorm:"-"`
	User    *User    `json:"user,omitempty" gorm:"-"`

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for Payment model
func (Payment) TableName() string {
	return "payments"
}

// Payment status constants
const (
	NewPaymentStatusPending    = "pending"
	NewPaymentStatusProcessing = "processing"
	NewPaymentStatusCompleted  = "completed"
	NewPaymentStatusFailed     = "failed"
	NewPaymentStatusCancelled  = "cancelled"
)

// Payment gateway constants
const (
	NewPaymentGatewayStripe = "stripe"
	NewPaymentGatewayEpay   = "epay"
	NewPaymentGatewayEPUSDT = "epusdt"
	NewPaymentGatewayPayPal = "paypal"
)

// Payment method constants
const (
	NewPaymentMethodCreditCard = "credit_card"
	NewPaymentMethodAlipay     = "alipay"
	NewPaymentMethodWechat     = "wechat"
	NewPaymentMethodQQPay      = "qqpay"
	NewPaymentMethodUnionPay   = "unionpay"
	NewPaymentMethodUSDT       = "usdt"
	NewPaymentMethodBTC        = "btc"
	NewPaymentMethodETH        = "eth"
	NewPaymentMethodPayPal     = "paypal"
	NewPaymentMethodBankWire   = "bank_wire"
)

// Currency constants
const (
	NewCurrencyUSD  = "USD"
	NewCurrencyCNY  = "CNY"
	NewCurrencyEUR  = "EUR"
	NewCurrencyGBP  = "GBP"
	NewCurrencyUSDT = "USDT"
)

// Business logic methods

// IsPending checks if the payment is pending
func (p *Payment) IsPending() bool {
	return p.Status == NewPaymentStatusPending
}

// IsProcessing checks if the payment is processing
func (p *Payment) IsProcessing() bool {
	return p.Status == NewPaymentStatusProcessing
}

// IsCompleted checks if the payment is completed
func (p *Payment) IsCompleted() bool {
	return p.Status == NewPaymentStatusCompleted
}

// IsFailed checks if the payment has failed
func (p *Payment) IsFailed() bool {
	return p.Status == NewPaymentStatusFailed
}

// IsCancelled checks if the payment is cancelled
func (p *Payment) IsCancelled() bool {
	return p.Status == NewPaymentStatusCancelled
}

// IsExpired checks if the payment has expired
func (p *Payment) IsExpired() bool {
	return p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now())
}

// CanBeRefunded checks if the payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.IsCompleted() && p.RefundAmount < p.Amount
}

// GetRefundableAmount returns the amount that can be refunded
func (p *Payment) GetRefundableAmount() float64 {
	if !p.CanBeRefunded() {
		return 0
	}
	return p.Amount - p.RefundAmount
}

// IsDeleted checks if the payment is soft deleted
func (p *Payment) IsDeleted() bool {
	return p.DeletedAt.Valid
}

// GetNetAmount returns the net amount after gateway fees
func (p *Payment) GetNetAmount() float64 {
	return p.Amount - p.GatewayFee
}

// PaymentResponse represents the payment data structure for API responses
type PaymentResponse struct {
	ID             uint       `json:"id" example:"1"`
	InvoiceID      uint       `json:"invoice_id" example:"1"`
	UserID         uint       `json:"user_id" example:"1"`
	PaymentNumber  string     `json:"payment_number" example:"PAY20240101001"`
	PaymentIntentID string    `json:"payment_intent_id,omitempty" example:"pi_1234567890"`
	Status         string     `json:"status" example:"completed"`
	
	// Amount Information
	Amount       float64 `json:"amount" example:"24.99"`
	Currency     string  `json:"currency" example:"USD"`
	ExchangeRate float64 `json:"exchange_rate" example:"1.0"`
	
	// Payment Channel Information
	PaymentMethod        string  `json:"payment_method" example:"credit_card"`
	PaymentGateway       string  `json:"payment_gateway" example:"stripe"`
	GatewayTransactionID string  `json:"gateway_transaction_id,omitempty" example:"txn_1234567890"`
	GatewayFee           float64 `json:"gateway_fee" example:"0.99"`
	
	// Payment Details (sensitive info removed for user responses)
	PaymentURL  string `json:"payment_url,omitempty" example:"https://checkout.stripe.com/pay/..."`
	QRCodeURL   string `json:"qr_code_url,omitempty" example:"https://api.qrserver.com/v1/create-qr-code/?data=..."`
	RedirectURL string `json:"redirect_url,omitempty" example:"https://example.com/payment/return"`
	
	// Time Information
	ExpiresAt   *time.Time `json:"expires_at,omitempty" example:"2024-01-01T01:00:00Z"`
	ProcessedAt *time.Time `json:"processed_at,omitempty" example:"2024-01-01T00:30:00Z"`
	CompletedAt *time.Time `json:"completed_at,omitempty" example:"2024-01-01T00:31:00Z"`
	
	// Refund Information
	RefundAmount    float64    `json:"refund_amount" example:"0"`
	RefundedAt      *time.Time `json:"refunded_at,omitempty"`
	RefundReason    string     `json:"refund_reason,omitempty"`
	RefundReference string     `json:"refund_reference,omitempty"`
	
	// Notification Information (admin only)
	NotificationCount  int        `json:"notification_count,omitempty" example:"3"`
	LastNotificationAt *time.Time `json:"last_notification_at,omitempty" example:"2024-01-01T00:31:00Z"`
	
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-01T00:31:00Z"`
	
	// Related data
	Invoice *InvoiceResponse `json:"invoice,omitempty"`
	User    *UserResponse    `json:"user,omitempty"`
	
	// Computed fields
	IsExpired        bool    `json:"is_expired" example:"false"`
	CanRefund        bool    `json:"can_refund" example:"true"`
	RefundableAmount float64 `json:"refundable_amount" example:"24.99"`
	NetAmount        float64 `json:"net_amount" example:"24.00"`
}

// ToResponse converts Payment to PaymentResponse
func (p *Payment) ToResponse() *PaymentResponse {
	resp := &PaymentResponse{
		ID:                   p.ID,
		InvoiceID:            p.InvoiceID,
		UserID:               p.UserID,
		PaymentNumber:        p.PaymentNumber,
		PaymentIntentID:      p.PaymentIntentID,
		Status:               p.Status,
		Amount:               p.Amount,
		Currency:             p.Currency,
		ExchangeRate:         p.ExchangeRate,
		PaymentMethod:        p.PaymentMethod,
		PaymentGateway:       p.PaymentGateway,
		GatewayTransactionID: p.GatewayTransactionID,
		GatewayFee:           p.GatewayFee,
		PaymentURL:           p.PaymentURL,
		QRCodeURL:            p.QRCodeURL,
		RedirectURL:          p.RedirectURL,
		ExpiresAt:            p.ExpiresAt,
		ProcessedAt:          p.ProcessedAt,
		CompletedAt:          p.CompletedAt,
		RefundAmount:         p.RefundAmount,
		RefundedAt:           p.RefundedAt,
		RefundReason:         p.RefundReason,
		RefundReference:      p.RefundReference,
		NotificationCount:    p.NotificationCount,
		LastNotificationAt:   p.LastNotificationAt,
		CreatedAt:            p.CreatedAt,
		UpdatedAt:            p.UpdatedAt,
		
		// Computed fields
		IsExpired:        p.IsExpired(),
		CanRefund:        p.CanBeRefunded(),
		RefundableAmount: p.GetRefundableAmount(),
		NetAmount:        p.GetNetAmount(),
	}
	
	// Include related data if loaded
	if p.Invoice != nil {
		resp.Invoice = p.Invoice.ToResponse()
	}
	if p.User != nil {
		resp.User = p.User.ToResponse()
	}
	
	return resp
}

// ToUserResponse converts Payment to a response suitable for the paying user
func (p *Payment) ToUserResponse() *PaymentResponse {
	resp := p.ToResponse()
	
	// Remove sensitive admin information for user responses
	resp.User = nil
	resp.NotificationCount = 0
	resp.LastNotificationAt = nil
	
	return resp
}