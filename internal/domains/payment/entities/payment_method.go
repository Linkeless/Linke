package entities

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"

	"linke/internal/domains/payment/constants"
)

// PaymentMethod represents a user's stored payment method
// Security note: This entity NEVER stores raw payment data
// All sensitive data is tokenized via payment gateways
type PaymentMethod struct {
	// Primary Key
	ID uint `json:"id" gorm:"primaryKey"`

	// Foreign Keys
	UserID uint `json:"user_id" gorm:"not null;index"`

	// Basic Information
	Type        string `json:"type" gorm:"size:50;not null;index"`    // card, bank_account, digital_wallet
	Gateway     string `json:"gateway" gorm:"size:50;not null;index"` // epay
	Method      string `json:"method" gorm:"size:50;not null;index"`  // alipay, wechat, qqpay
	DisplayName string `json:"display_name" gorm:"size:100;not null"` // User-friendly name

	// Tokenized Payment Data (PCI Compliant)
	// These tokens are provided by payment gateways, not raw payment data
	PaymentToken      string `json:"payment_token" gorm:"size:255;not null;index"`  // Gateway payment token
	GatewayCustomerID string `json:"gateway_customer_id,omitempty" gorm:"size:255"` // Gateway customer ID

	// Masked Display Information (Safe to show to users)
	MaskedInfo  string `json:"masked_info" gorm:"size:100"`    // e.g., "**** 1234", "ali***@example.com"
	Brand       string `json:"brand,omitempty" gorm:"size:50"` // e.g., "Visa", "Alipay"
	ExpiryMonth *int   `json:"expiry_month,omitempty"`         // For cards only
	ExpiryYear  *int   `json:"expiry_year,omitempty"`          // For cards only

	// Status and Configuration
	IsDefault bool   `json:"is_default" gorm:"default:false;index"`           // Primary payment method
	Active    bool   `json:"is_active" gorm:"default:true;index"`             // Can be used for payments
	Status    string `json:"status" gorm:"size:20;not null;default:'active'"` // active, inactive, expired, invalid

	// Security and Validation
	LastValidatedAt *time.Time `json:"last_validated_at,omitempty" gorm:"index"` // Last validation timestamp
	ValidationHash  string     `json:"validation_hash,omitempty" gorm:"size:64"` // Hash for integrity checking

	// Billing Information (Optional, for cards)
	BillingCountry  string `json:"billing_country,omitempty" gorm:"size:10"`
	BillingPostcode string `json:"billing_postcode,omitempty" gorm:"size:20"`

	// Gateway-Specific Metadata
	GatewayMetadata string `json:"gateway_metadata,omitempty" gorm:"type:text"` // JSON metadata from gateway

	// Usage Statistics
	LastUsedAt     *time.Time `json:"last_used_at,omitempty" gorm:"index"` // Last successful payment
	SuccessfulUses int        `json:"successful_uses" gorm:"default:0"`    // Number of successful payments
	FailedUses     int        `json:"failed_uses" gorm:"default:0"`        // Number of failed payments

	// Security Tracking
	CreatedFromIP string `json:"created_from_ip,omitempty" gorm:"size:45"` // IP when method was added
	LastUpdateIP  string `json:"last_update_ip,omitempty" gorm:"size:45"`  // IP of last update

	// Timestamp Fields
	CreatedAt time.Time      `json:"created_at" gorm:"not null;index"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName returns the table name for PaymentMethod model
func (PaymentMethod) TableName() string {
	return "payment_methods"
}


// IsActive checks if the payment method is active and can be used
func (pm *PaymentMethod) IsActive() bool {
	return pm.Active && pm.Status == constants.PaymentMethodStatusActive && !pm.IsExpired()
}

// IsExpired checks if the payment method has expired (for cards)
func (pm *PaymentMethod) IsExpired() bool {
	if pm.ExpiryMonth == nil || pm.ExpiryYear == nil {
		return false
	}

	now := time.Now()
	expiryTime := time.Date(*pm.ExpiryYear, time.Month(*pm.ExpiryMonth), 1, 0, 0, 0, 0, time.UTC)
	// Add one month and subtract one day to get the last day of the expiry month
	expiryTime = expiryTime.AddDate(0, 1, -1)

	return now.After(expiryTime)
}

// CanBeUsedForPayment checks if payment method can be used for payments
func (pm *PaymentMethod) CanBeUsedForPayment() bool {
	return pm.IsActive() && !pm.IsDeleted()
}

// IsDeleted checks if the payment method is soft deleted
func (pm *PaymentMethod) IsDeleted() bool {
	return pm.DeletedAt.Valid
}

// NeedsRevalidation checks if the payment method needs revalidation
func (pm *PaymentMethod) NeedsRevalidation() bool {
	if pm.LastValidatedAt == nil {
		return true
	}

	// Revalidate every 30 days
	return time.Since(*pm.LastValidatedAt) > 30*24*time.Hour
}

// GenerateValidationHash generates a hash for integrity checking
func (pm *PaymentMethod) GenerateValidationHash() error {
	// Generate a random hash for validation
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	pm.ValidationHash = hex.EncodeToString(bytes)
	return nil
}

// UpdateLastUsed updates the last used timestamp and increments usage counters
func (pm *PaymentMethod) UpdateLastUsed(successful bool) {
	now := time.Now()
	pm.LastUsedAt = &now

	if successful {
		pm.SuccessfulUses++
	} else {
		pm.FailedUses++
	}
}

// GetFailureRate calculates the failure rate of this payment method
func (pm *PaymentMethod) GetFailureRate() float64 {
	total := pm.SuccessfulUses + pm.FailedUses
	if total == 0 {
		return 0
	}
	return float64(pm.FailedUses) / float64(total)
}

