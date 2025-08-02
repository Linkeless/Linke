package interfaces

import (
	"context"
	paymentEntities "linke/internal/domains/payment/entities"
	"linke/internal/domains/subscription/entities"
	"time"
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
	GetOrderStatistics(ctx context.Context, fromDate, toDate time.Time) (map[string]interface{}, error)
}

// CreateSubscriptionOrderRequest represents the request to create a subscription order
type CreateSubscriptionOrderRequest struct {
	UserID             uint   `json:"user_id" binding:"required" example:"1"`
	SubscriptionPlanID uint   `json:"subscription_plan_id" binding:"required" example:"1"`
	OrderType          string `json:"order_type" binding:"required,oneof=new renewal upgrade downgrade" example:"new"`
	CouponCode         string `json:"coupon_code,omitempty" example:"SAVE20"`
	PaymentGateway     string `json:"payment_gateway" binding:"required" example:"epay"`
	PaymentMethod      string `json:"payment_method" binding:"required" example:"alipay"`
	ReturnURL          string `json:"return_url,omitempty" example:"https://example.com/payment/return"`
	Metadata           string `json:"metadata,omitempty"`
}

// CreateSubscriptionOrderResponse represents the response after creating a subscription order
type CreateSubscriptionOrderResponse struct {
	Order         *entities.SubscriptionOrderResponse    `json:"order"`
	PaymentRecord *paymentEntities.PaymentRecordResponse `json:"payment_record"`
	PaymentURL    string                                 `json:"payment_url"`
	QRCodeURL     string                                 `json:"qr_code_url,omitempty"`
	ExpiredAt     time.Time                              `json:"expired_at"`
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
