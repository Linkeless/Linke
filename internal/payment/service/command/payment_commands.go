package command

import (
	"time"

	"linke/internal/payment/domain/valueobject"
)

// CreatePaymentCommand represents the command to create a new payment
type CreatePaymentCommand struct {
	PaymentNumber  string  `json:"payment_number" validate:"required"`
	InvoiceID      uint    `json:"invoice_id" validate:"required"`
	UserID         uint    `json:"user_id" validate:"required"`
	Amount         float64 `json:"amount" validate:"required,min=0.01"`
	Currency       string  `json:"currency" validate:"required"`
	PaymentMethod  string  `json:"payment_method" validate:"required"`
	PaymentGateway string  `json:"payment_gateway" validate:"required"`
	Notes          string  `json:"notes,omitempty"`
	Metadata       string  `json:"metadata,omitempty"`
}

// UpdatePaymentDetailsCommand represents the command to update payment details
type UpdatePaymentDetailsCommand struct {
	PaymentID       uint   `json:"payment_id" validate:"required"`
	PaymentIntentID string `json:"payment_intent_id,omitempty"`
	PaymentURL      string `json:"payment_url,omitempty"`
	QRCodeURL       string `json:"qr_code_url,omitempty"`
	RedirectURL     string `json:"redirect_url,omitempty"`
	ExpiresAt       *time.Time `json:"expires_at,omitempty"`
}

// SetGatewayTransactionCommand represents the command to set gateway transaction details
type SetGatewayTransactionCommand struct {
	PaymentID       uint    `json:"payment_id" validate:"required"`
	TransactionID   string  `json:"transaction_id" validate:"required"`
	GatewayFee      float64 `json:"gateway_fee" validate:"min=0"`
	GatewayFeeCurrency string `json:"gateway_fee_currency,omitempty"`
}

// ProcessPaymentCommand represents the command to start payment processing
type ProcessPaymentCommand struct {
	PaymentID uint `json:"payment_id" validate:"required"`
}

// CompletePaymentCommand represents the command to complete a payment
type CompletePaymentCommand struct {
	PaymentID   uint   `json:"payment_id" validate:"required"`
	WebhookData string `json:"webhook_data,omitempty"`
}

// FailPaymentCommand represents the command to fail a payment
type FailPaymentCommand struct {
	PaymentID   uint   `json:"payment_id" validate:"required"`
	Reason      string `json:"reason" validate:"required"`
	WebhookData string `json:"webhook_data,omitempty"`
}

// CancelPaymentCommand represents the command to cancel a payment
type CancelPaymentCommand struct {
	PaymentID uint   `json:"payment_id" validate:"required"`
	Reason    string `json:"reason" validate:"required"`
}

// RefundPaymentCommand represents the command to refund a payment
type RefundPaymentCommand struct {
	PaymentID    uint    `json:"payment_id" validate:"required"`
	RefundAmount float64 `json:"refund_amount" validate:"required,min=0.01"`
	Reason       string  `json:"reason" validate:"required"`
}

// UpdatePaymentNotificationCommand represents the command to update payment notification
type UpdatePaymentNotificationCommand struct {
	PaymentID   uint   `json:"payment_id" validate:"required"`
	WebhookData string `json:"webhook_data,omitempty"`
}

// UpdatePaymentNotesCommand represents the command to update payment notes
type UpdatePaymentNotesCommand struct {
	PaymentID uint   `json:"payment_id" validate:"required"`
	Notes     string `json:"notes"`
}

// UpdatePaymentMetadataCommand represents the command to update payment metadata
type UpdatePaymentMetadataCommand struct {
	PaymentID uint   `json:"payment_id" validate:"required"`
	Metadata  string `json:"metadata"`
}

// DeletePaymentCommand represents the command to delete a payment
type DeletePaymentCommand struct {
	PaymentID uint `json:"payment_id" validate:"required"`
}

// CreatePaymentResult represents the result of creating a payment
type CreatePaymentResult struct {
	PaymentID     valueobject.PaymentID     `json:"payment_id"`
	PaymentNumber valueobject.PaymentNumber `json:"payment_number"`
	Amount        valueobject.Money         `json:"amount"`
	Status        valueobject.PaymentStatus `json:"status"`
	CreatedAt     time.Time                 `json:"created_at"`
}

// PaymentCommandResult represents a generic payment command result
type PaymentCommandResult struct {
	PaymentID     valueobject.PaymentID     `json:"payment_id"`
	PaymentNumber valueobject.PaymentNumber `json:"payment_number"`
	Status        valueobject.PaymentStatus `json:"status"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

// RefundPaymentResult represents the result of refunding a payment
type RefundPaymentResult struct {
	PaymentID       valueobject.PaymentID     `json:"payment_id"`
	PaymentNumber   valueobject.PaymentNumber `json:"payment_number"`
	RefundAmount    valueobject.Money         `json:"refund_amount"`
	TotalRefunded   valueobject.Money         `json:"total_refunded"`
	RefundReference valueobject.PaymentNumber `json:"refund_reference"`
	RefundedAt      time.Time                 `json:"refunded_at"`
}