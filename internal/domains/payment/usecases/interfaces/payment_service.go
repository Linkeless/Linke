package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/payment/dto"
	"linke/internal/domains/payment/entities"
)

// PaymentService defines payment-specific operations
type PaymentService interface {
	// Gateway management (domain-specific)
	RegisterGateway(name string, gateway PaymentGateway) error
	GetGateway(name string) (PaymentGateway, error)

	// Payment-specific operations that don't fit generic patterns
	CreatePaymentOrder(ctx context.Context, req *dto.CreatePaymentOrderRequest) (*entities.PaymentRecord, error)
	GetPaymentRecord(ctx context.Context, paymentNo string) (*entities.PaymentRecord, error)
	GetPaymentRecordByOutTradeNo(ctx context.Context, outTradeNo string) (*entities.PaymentRecord, error)
	UpdatePaymentStatus(ctx context.Context, paymentNo, status, transactionID string, paidAt *time.Time) error
	ProcessNotification(ctx context.Context, gateway string, data map[string]any) error
	GetAvailablePaymentMethods(ctx context.Context) (map[string][]string, error)
	GeneratePaymentNo() (string, error)

	// Service dependencies
	SetSubscriptionOrderService(subscriptionOrderService SubscriptionOrderServiceInterface)

	// Legacy method support for backward compatibility
	GetUserPaymentRecords(ctx context.Context, userID uint, limit, offset int) ([]*entities.PaymentRecord, int64, error)
}

// PaymentGateway interface defines the methods that all payment gateways must implement
type PaymentGateway interface {
	CreatePaymentOrder(req *dto.CreatePaymentOrderRequest) (*dto.CreatePaymentOrderResponse, error)
	QueryPaymentOrder(outTradeNo string) (*dto.QueryPaymentOrderResponse, error)
	VerifyPaymentNotify(data map[string]any) (bool, *dto.NotifyData)
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
