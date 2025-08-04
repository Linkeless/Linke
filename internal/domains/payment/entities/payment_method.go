package entities

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"gorm.io/gorm"
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
	Gateway     string `json:"gateway" gorm:"size:50;not null;index"` // epay, epusdt
	Method      string `json:"method" gorm:"size:50;not null;index"`  // alipay, wechat, usdt, etc.
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

// Payment method type constants
const (
	PaymentMethodTypeCard          = "card"
	PaymentMethodTypeBankAccount   = "bank_account"
	PaymentMethodTypeDigitalWallet = "digital_wallet"
	PaymentMethodTypeCrypto        = "crypto"
)

// Payment method status constants
const (
	PaymentMethodStatusActive   = "active"
	PaymentMethodStatusInactive = "inactive"
	PaymentMethodStatusExpired  = "expired"
	PaymentMethodStatusInvalid  = "invalid"
)

// IsActive checks if the payment method is active and can be used
func (pm *PaymentMethod) IsActive() bool {
	return pm.Active && pm.Status == PaymentMethodStatusActive && !pm.IsExpired()
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

// ToResponse converts PaymentMethod to PaymentMethodResponse
func (pm *PaymentMethod) ToResponse() *PaymentMethodResponse {
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

// ToSecureResponse converts PaymentMethod to a secure response that excludes sensitive fields
func (pm *PaymentMethod) ToSecureResponse() *PaymentMethodResponse {
	resp := pm.ToResponse()

	// Remove sensitive information
	resp.UserID = 0 // Remove user ID for external APIs

	return resp
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
