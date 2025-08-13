package interfaces

import (
	"context"
	"time"

	"linke/internal/domains/subscription/dto"
	"linke/internal/domains/subscription/entities"
)

// SubscriptionOrderService defines the interface for subscription order operations
type SubscriptionOrderService interface {
	// ==============================================================================
	// NEW STANDARDIZED ORDER CREATION FLOW
	// ==============================================================================
	
	// CreateOrderWithInvoice creates a complete order -> invoice -> payment flow
	CreateOrderWithInvoice(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error)
	
	// GenerateInvoiceForOrder creates an invoice for an existing order
	GenerateInvoiceForOrder(ctx context.Context, orderID uint) (any, error) // returns InvoiceResponse
	
	// CreatePaymentForOrder creates a payment record for an existing order
	CreatePaymentForOrder(ctx context.Context, req *dto.PayOrderRequest) (*dto.PayOrderResponse, error)

	// ==============================================================================
	// LEGACY ORDER MANAGEMENT (for backward compatibility)
	// ==============================================================================
	
	// CreateSubscriptionOrder creates a subscription order (legacy method)
	CreateSubscriptionOrder(ctx context.Context, req *dto.CreateSubscriptionOrderRequest) (*dto.CreateSubscriptionOrderResponse, error)
	GetSubscriptionOrder(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error)
	GetSubscriptionOrderByNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error)

	// Order processing
	ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error
	CancelSubscriptionOrder(ctx context.Context, orderID uint, reason string) error

	// Order listing
	GetUserSubscriptionOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error)
	GetSubscriptionOrders(ctx context.Context, req *dto.GetSubscriptionOrdersRequest) ([]*entities.SubscriptionOrder, int64, error)

	// Order statistics
	GetOrderStatistics(ctx context.Context, fromDate, toDate time.Time) (map[string]any, error)
	
	// Aggregation: Order + latest payment + latest invoice (business-layer aggregation only)
	GetSubscriptionOrderSummary(ctx context.Context, orderID uint) (map[string]any, error)
}
