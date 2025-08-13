package dto

import (
	"time"

	invoiceDto "linke/internal/domains/invoice/dto"
	"linke/internal/domains/subscription/entities"
)

// ==============================================================================
// ORDER CREATION AND MANAGEMENT DTOs
// ==============================================================================

// CreateOrderRequest represents the request to create a subscription order with invoice and payment
type CreateOrderRequest struct {
	UserID             uint   `json:"user_id" binding:"required" example:"1"`
	SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required" example:"1"`
	OrderType          string `json:"order_type" binding:"required,oneof=new renewal upgrade downgrade" example:"new"`
	CouponCode         string `json:"coupon_code,omitempty" example:"SAVE20"`
	PaymentGateway     string `json:"payment_gateway" binding:"required" example:"epay"`
	PaymentMethod      string `json:"payment_method" binding:"required" example:"alipay"`
	PaymentMethodID    *uint  `json:"payment_method_id,omitempty" example:"1"`       // Optional: Use saved payment method
	UseDefaultPayment  bool   `json:"use_default_payment,omitempty" example:"false"` // Use user's default payment method
	ReturnURL          string `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	Metadata           string `json:"metadata,omitempty"`
}

// CreateOrderResponse represents the response after creating a complete order flow
type CreateOrderResponse struct {
	Order         *entities.SubscriptionOrderResponse `json:"order"`
	Invoice       *invoiceDto.InvoiceResponse         `json:"invoice"`
	PaymentRecord any                                 `json:"payment_record" swaggertype:"object"`
	PaymentURL    string                              `json:"payment_url"`
	QRCodeURL     string                              `json:"qr_code_url,omitempty"`
	ExpiredAt     time.Time                           `json:"expired_at"`
}

// PayOrderRequest represents the request to create payment for an existing order
type PayOrderRequest struct {
	OrderID           uint   `json:"order_id" binding:"required" example:"1"`
	PaymentGateway    string `json:"payment_gateway" binding:"required" example:"epay"`
	PaymentMethod     string `json:"payment_method" binding:"required" example:"alipay"`
	PaymentMethodID   *uint  `json:"payment_method_id,omitempty" example:"1"`       // Optional: Use saved payment method
	UseDefaultPayment bool   `json:"use_default_payment,omitempty" example:"false"` // Use user's default payment method
	ClientIP          string `json:"client_ip,omitempty" example:"192.168.1.1"`
	ReturnURL         string `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	Metadata          string `json:"metadata,omitempty"`
}

// PayOrderResponse represents the response after creating payment for order
type PayOrderResponse struct {
	Order         *entities.SubscriptionOrderResponse `json:"order"`
	Invoice       *invoiceDto.InvoiceResponse         `json:"invoice,omitempty"`
	PaymentRecord any                                 `json:"payment_record" swaggertype:"object"`
	PaymentURL    string                              `json:"payment_url"`
	QRCodeURL     string                              `json:"qr_code_url,omitempty"`
	ExpiredAt     time.Time                           `json:"expired_at"`
}

// ==============================================================================
// BACKWARD COMPATIBILITY DTOs (legacy)
// ==============================================================================

// CreateSubscriptionOrderRequest represents the request to create a subscription order (legacy)
type CreateSubscriptionOrderRequest struct {
	UserID             uint   `json:"user_id" binding:"required" example:"1"`
	SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required" example:"1"`
	OrderType          string `json:"order_type" binding:"required,oneof=new renewal upgrade downgrade" example:"new"`
	CouponCode         string `json:"coupon_code,omitempty" example:"SAVE20"`
	PaymentGateway     string `json:"payment_gateway" binding:"required" example:"epay"`
	PaymentMethod      string `json:"payment_method" binding:"required" example:"alipay"`
	PaymentMethodID    *uint  `json:"payment_method_id,omitempty" example:"1"`       // Optional: Use saved payment method
	UseDefaultPayment  bool   `json:"use_default_payment,omitempty" example:"false"` // Use user's default payment method
	ReturnURL          string `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	Metadata           string `json:"metadata,omitempty"`
}

// CreateSubscriptionOrderResponse represents the response after creating a subscription order (legacy)
type CreateSubscriptionOrderResponse struct {
	Order         *entities.SubscriptionOrderResponse `json:"order"`
	Invoice       *invoiceDto.InvoiceResponse         `json:"invoice"`
	PaymentRecord any                                 `json:"payment_record" swaggertype:"object"`
	PaymentURL    string                              `json:"payment_url"`
	QRCodeURL     string                              `json:"qr_code_url,omitempty"`
	ExpiredAt     time.Time                           `json:"expired_at"`
}

// ==============================================================================
// QUERY AND LISTING DTOs
// ==============================================================================

// GetSubscriptionOrdersRequest represents the request to get subscription orders
type GetSubscriptionOrdersRequest struct {
	UserID    uint   `form:"user_id,omitempty" example:"1"`
	Status    string `form:"status,omitempty" example:"pending"`
	OrderType string `form:"order_type,omitempty" example:"new"`
	DateFrom  string `form:"date_from,omitempty" example:"2024-01-01"`
	DateTo    string `form:"date_to,omitempty" example:"2024-12-31"`
	Limit     int    `form:"limit,omitempty" example:"10"`
	Offset    int    `form:"offset,omitempty" example:"0"`
}

// ==============================================================================
// HELPER CONVERSION FUNCTIONS
// ==============================================================================

// ToCreateOrderRequest converts legacy CreateSubscriptionOrderRequest to new CreateOrderRequest
func (r *CreateSubscriptionOrderRequest) ToCreateOrderRequest() *CreateOrderRequest {
	return &CreateOrderRequest{
		UserID:             r.UserID,
		SubscriptionPlanID: r.SubscriptionPlanID,
		OrderType:          r.OrderType,
		CouponCode:         r.CouponCode,
		PaymentGateway:     r.PaymentGateway,
		PaymentMethod:      r.PaymentMethod,
		PaymentMethodID:    r.PaymentMethodID,
		UseDefaultPayment:  r.UseDefaultPayment,
		ReturnURL:          r.ReturnURL,
		Metadata:           r.Metadata,
	}
}

// ToCreateSubscriptionOrderResponse converts new CreateOrderResponse to legacy response
func (r *CreateOrderResponse) ToCreateSubscriptionOrderResponse() *CreateSubscriptionOrderResponse {
	return &CreateSubscriptionOrderResponse{
		Order:         r.Order,
		Invoice:       r.Invoice,
		PaymentRecord: r.PaymentRecord,
		PaymentURL:    r.PaymentURL,
		QRCodeURL:     r.QRCodeURL,
		ExpiredAt:     r.ExpiredAt,
	}
}
