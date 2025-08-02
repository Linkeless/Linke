package interfaces

import (
	"context"
	"linke/internal/domains/payment/entities"
	"time"
)

// PaymentService defines the interface for payment service operations
type PaymentService interface {
	// Gateway management
	RegisterGateway(name string, gateway PaymentGateway) error
	GetGateway(name string) (PaymentGateway, error)

	// Payment order operations
	CreatePaymentOrder(ctx context.Context, req *CreatePaymentOrderRequest) (*entities.PaymentRecord, error)
	GetPaymentRecord(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error)
	GetPaymentRecordByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error)

	// Payment processing
	UpdatePaymentStatus(ctx context.Context, paymentNo string, status string, transactionID string, paidAt *time.Time) error
	ProcessNotification(ctx context.Context, gateway string, data map[string]interface{}) error

	// User payment records
	GetUserPaymentRecords(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error)
	GetAvailablePaymentMethods(ctx context.Context) (map[string][]string, error)

	// Service dependencies
	SetSubscriptionOrderService(subscriptionOrderService SubscriptionOrderServiceInterface)

	// Utility
	GeneratePaymentNo() (string, error)
}

// PaymentGateway interface defines the methods that all payment gateways must implement
type PaymentGateway interface {
	CreatePaymentOrder(req *CreatePaymentOrderRequest) (*CreatePaymentOrderResponse, error)
	QueryPaymentOrder(outTradeNo string) (*QueryPaymentOrderResponse, error)
	VerifyPaymentNotify(data map[string]interface{}) (bool, *NotifyData)
	IsPaymentCompleted(status string) bool
	GetSupportedPaymentMethods() []string
	GetPaymentMethodName(method string) string
	ValidateConfig() error
	TestConnection() error
}

// SubscriptionOrderServiceInterface defines the interface for subscription order service
type SubscriptionOrderServiceInterface interface {
	ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error
}

// CreatePaymentOrderRequest represents the unified request to create a payment order
type CreatePaymentOrderRequest struct {
	UserID              uint    `json:"user_id"`
	SubscriptionOrderID *uint   `json:"subscription_order_id,omitempty"`
	Gateway             string  `json:"gateway"`            // epay, epusdt
	PaymentMethod       string  `json:"payment_method"`     // alipay, wechat, usdt, etc.
	Amount              float64 `json:"amount"`             // Amount in specified currency
	Currency            string  `json:"currency"`           // CNY, USD, USDT
	Subject             string  `json:"subject"`            // Order subject
	Body                string  `json:"body"`               // Order description
	ClientIP            string  `json:"client_ip"`          // Client IP
	NotifyURL           string  `json:"notify_url"`         // Async notification URL
	ReturnURL           string  `json:"return_url"`         // Sync return URL
	ExpiredMinutes      int     `json:"expired_minutes"`    // Expiration time in minutes
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
