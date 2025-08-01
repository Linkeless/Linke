package persistence

import (
	"time"

	"gorm.io/gorm"
)

// PaymentPO represents the payment persistent object for database storage
type PaymentPO struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	InvoiceID uint `json:"invoice_id" gorm:"not null;index"`
	UserID    uint `json:"user_id" gorm:"not null;index"`

	// Payment Information
	PaymentNumber      string `json:"payment_number" gorm:"uniqueIndex;size:32;not null"`
	PaymentIntentID    string `json:"payment_intent_id,omitempty" gorm:"size:255;index"`
	Status             string `json:"status" gorm:"size:20;not null;default:'pending';index"`

	// Amount Information
	Amount       float64 `json:"amount" gorm:"type:decimal(10,2);not null"`
	Currency     string  `json:"currency" gorm:"size:3;not null;default:'USD'"`
	ExchangeRate float64 `json:"exchange_rate" gorm:"type:decimal(10,6);default:1.0"`

	// Payment Channel Information
	PaymentMethod        string  `json:"payment_method" gorm:"size:50;not null;index"`
	PaymentGateway       string  `json:"payment_gateway" gorm:"size:50;not null;index"`
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

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for PaymentPO
func (PaymentPO) TableName() string {
	return "payments"
}