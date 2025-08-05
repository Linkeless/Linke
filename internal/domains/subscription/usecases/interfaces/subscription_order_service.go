package interfaces

import (
	"context"
	"time"

	invoiceEntities "linke/internal/domains/invoice/entities"
	"linke/internal/domains/subscription/entities"
)

// SubscriptionOrderService defines the interface for subscription order operations
type SubscriptionOrderService interface {
	// Order creation and management
	CreateSubscriptionOrder(ctx context.Context, req *CreateSubscriptionOrderRequest) (*CreateSubscriptionOrderResponse, error)
	GetSubscriptionOrder(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error)
	GetSubscriptionOrderByNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error)

	// Order processing
	ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error
	CancelSubscriptionOrder(ctx context.Context, orderID uint, reason string) error

	// Order listing
	GetUserSubscriptionOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	GetSubscriptionOrders(ctx context.Context, req *GetSubscriptionOrdersRequest) ([]*entities.SubscriptionOrder, int64, error)

	// Order statistics
	GetOrderStatistics(ctx context.Context, fromDate, toDate time.Time) (map[string]any, error)

	// Quick Purchase
	QuickPurchase(ctx context.Context, req *QuickPurchaseRequest) (*QuickPurchaseResponse, error)
}

// CreateSubscriptionOrderRequest represents the request to create a subscription order
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

// CreateSubscriptionOrderResponse represents the response after creating a subscription order
type CreateSubscriptionOrderResponse struct {
	Order         *entities.SubscriptionOrderResponse `json:"order"`
	Invoice       *invoiceEntities.InvoiceResponse    `json:"invoice"`
	PaymentRecord any                                 `json:"payment_record" swaggertype:"object"`
	PaymentURL    string                              `json:"payment_url"`
	QRCodeURL     string                              `json:"qr_code_url,omitempty"`
	ExpiredAt     time.Time                           `json:"expired_at"`
}

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

// QuickPurchaseRequest represents the request for quick purchase
type QuickPurchaseRequest struct {
	UserID            uint   `json:"user_id" binding:"required" example:"1"`
	PlanID            uint   `json:"plan_id" binding:"required" example:"1"`
	PaymentGateway    string `json:"payment_gateway" binding:"required" example:"epay"`
	PaymentMethod     string `json:"payment_method" binding:"required" example:"alipay"`
	PaymentMethodID   *uint  `json:"payment_method_id,omitempty" example:"1"`       // Optional: Use saved payment method
	UseDefaultPayment bool   `json:"use_default_payment,omitempty" example:"false"` // Use user's default payment method
	CouponCode        string `json:"coupon_code,omitempty" example:"SAVE20"`
	ClientIP          string `json:"client_ip,omitempty" example:"192.168.1.1"`
	ReturnURL         string `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	Metadata          string `json:"metadata,omitempty"`
}

// QuickPurchaseResponse represents the response for quick purchase
type QuickPurchaseResponse struct {
	PaymentRecord any                                `json:"payment_record" swaggertype:"object"`
	PaymentURL    string                             `json:"payment_url"`
	QRCodeURL     string                             `json:"qr_code_url,omitempty"`
	ExpiredAt     time.Time                          `json:"expired_at"`
	PlanInfo      *entities.SubscriptionPlanResponse `json:"plan_info"`
	DiscountInfo  *QuickPurchaseDiscountInfo         `json:"discount_info,omitempty"`
}

// QuickPurchaseDiscountInfo represents discount information in quick purchase response
type QuickPurchaseDiscountInfo struct {
	CouponCode     string  `json:"coupon_code"`
	DiscountAmount float64 `json:"discount_amount"`
	OriginalAmount float64 `json:"original_amount"`
	FinalAmount    float64 `json:"final_amount"`
}
