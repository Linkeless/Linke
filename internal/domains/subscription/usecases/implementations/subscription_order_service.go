package implementations

import (
	"context"
	"time"

	"gorm.io/gorm"

	couponInterfaces "linke/internal/domains/coupon/usecases/interfaces"
	invoiceInterfaces "linke/internal/domains/invoice/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/domains/subscription/dto"
	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/events"
)

// SubscriptionOrderService combines all order-related functionality using composition
type SubscriptionOrderService struct {
	// Base dependencies
	base *SubscriptionOrderServiceBase

	// Specialized services
	creationService   *SubscriptionOrderCreationService
	managementService *SubscriptionOrderManagementService
	operationsService *SubscriptionOrderOperationsService
	analyticsService  *SubscriptionOrderAnalyticsService
	utilsService      *SubscriptionOrderUtilsService
}

// NewSubscriptionOrderService creates a new subscription order service with all functionality
func NewSubscriptionOrderService(
	db *gorm.DB,
	subscriptionPlanService interfaces.SubscriptionPlanService,
	userSubscriptionService interfaces.UserSubscriptionService,
	paymentService paymentInterfaces.PaymentService,
	paymentMethodService paymentInterfaces.PaymentMethodService,
	couponService couponInterfaces.CouponService,
	invoiceService invoiceInterfaces.InvoiceService,
	eventPublisher events.Publisher,
) *SubscriptionOrderService {
	// Create base service with all dependencies
	base := NewSubscriptionOrderServiceBase(
		db,
		subscriptionPlanService,
		userSubscriptionService,
		paymentService,
		paymentMethodService,
		couponService,
		invoiceService,
		eventPublisher,
	)

	// Create utils service first (needed by creation service)
	utilsService := NewSubscriptionOrderUtilsService(base)

	// Create specialized services
	creationService := NewSubscriptionOrderCreationService(base, utilsService)
	managementService := NewSubscriptionOrderManagementService(base)
	operationsService := NewSubscriptionOrderOperationsService(base)
	analyticsService := NewSubscriptionOrderAnalyticsService(base)

	return &SubscriptionOrderService{
		base:              base,
		creationService:   creationService,
		managementService: managementService,
		operationsService: operationsService,
		analyticsService:  analyticsService,
		utilsService:      utilsService,
	}
}

// ===============================================================================
// ORDER CREATION METHODS - Delegated to SubscriptionOrderCreationService
// ===============================================================================

// CreateOrderWithInvoice creates a complete order -> invoice -> payment flow
func (sos *SubscriptionOrderService) CreateOrderWithInvoice(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
	return sos.creationService.CreateOrderWithInvoice(ctx, req)
}

// GenerateInvoiceForOrder creates an invoice for an existing order
func (sos *SubscriptionOrderService) GenerateInvoiceForOrder(ctx context.Context, orderID uint) (any, error) {
	return sos.creationService.GenerateInvoiceForOrder(ctx, orderID)
}

// CreatePaymentForOrder creates a payment record for an existing order
func (sos *SubscriptionOrderService) CreatePaymentForOrder(ctx context.Context, req *dto.PayOrderRequest) (*dto.PayOrderResponse, error) {
	return sos.creationService.CreatePaymentForOrder(ctx, req)
}

// CreateSubscriptionOrder creates a new subscription order with payment
func (sos *SubscriptionOrderService) CreateSubscriptionOrder(ctx context.Context, req *dto.CreateSubscriptionOrderRequest) (*dto.CreateSubscriptionOrderResponse, error) {
	return sos.creationService.CreateSubscriptionOrder(ctx, req)
}

// CreateUpgradeDowngradeOrder creates an order for subscription upgrade or downgrade
func (sos *SubscriptionOrderService) CreateUpgradeDowngradeOrder(ctx context.Context, userID uint, req *UpgradeDowngradeRequest) (*dto.CreateSubscriptionOrderResponse, error) {
	return sos.creationService.CreateUpgradeDowngradeOrder(ctx, userID, req)
}

// ===============================================================================
// ORDER MANAGEMENT METHODS - Delegated to SubscriptionOrderManagementService
// ===============================================================================

// GetSubscriptionOrderSummary aggregates order + latest payment + latest invoice
func (sos *SubscriptionOrderService) GetSubscriptionOrderSummary(ctx context.Context, orderID uint) (map[string]any, error) {
	return sos.managementService.GetSubscriptionOrderSummary(ctx, orderID)
}

// GetSubscriptionOrder gets a subscription order by ID
func (sos *SubscriptionOrderService) GetSubscriptionOrder(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error) {
	return sos.managementService.GetSubscriptionOrder(ctx, orderID)
}

// GetSubscriptionOrderByNumber gets a subscription order by order number
func (sos *SubscriptionOrderService) GetSubscriptionOrderByNumber(ctx context.Context, orderNumber string) (*entities.SubscriptionOrder, error) {
	return sos.managementService.GetSubscriptionOrderByNumber(ctx, orderNumber)
}

// GetUserSubscriptionOrders gets subscription orders for a user with pagination
func (sos *SubscriptionOrderService) GetUserSubscriptionOrders(ctx context.Context, userID uint, limit, offset int) ([]*entities.SubscriptionOrder, int64, error) {
	return sos.managementService.GetUserSubscriptionOrders(ctx, userID, limit, offset)
}

// GetOrdersWithFiltering gets orders with advanced filtering and search
func (sos *SubscriptionOrderService) GetOrdersWithFiltering(ctx context.Context, req *GetOrdersRequest) ([]*entities.SubscriptionOrder, int64, error) {
	return sos.managementService.GetOrdersWithFiltering(ctx, req)
}

// GetOrderWithRelations gets an order with all related data
func (sos *SubscriptionOrderService) GetOrderWithRelations(ctx context.Context, orderID uint) (*entities.SubscriptionOrder, error) {
	return sos.managementService.GetOrderWithRelations(ctx, orderID)
}

// GetSubscriptionOrders gets subscription orders with filtering
func (sos *SubscriptionOrderService) GetSubscriptionOrders(ctx context.Context, req *dto.GetSubscriptionOrdersRequest) ([]*entities.SubscriptionOrder, int64, error) {
	return sos.managementService.GetSubscriptionOrders(ctx, req)
}

// ===============================================================================
// ORDER OPERATIONS METHODS - Delegated to SubscriptionOrderOperationsService
// ===============================================================================

// ProcessOrderPaymentSuccess processes successful payment for an order
func (sos *SubscriptionOrderService) ProcessOrderPaymentSuccess(ctx context.Context, orderID uint) error {
	return sos.operationsService.ProcessOrderPaymentSuccess(ctx, orderID)
}

// UpdateOrderStatus updates order status with admin notes and payment verification
func (sos *SubscriptionOrderService) UpdateOrderStatus(ctx context.Context, orderID uint, status string, adminID uint, notes, reason string) (*entities.SubscriptionOrder, error) {
	return sos.operationsService.UpdateOrderStatus(ctx, orderID, status, adminID, notes, reason)
}

// UpdateOrderStatusWithEvidence updates order status with payment evidence verification
func (sos *SubscriptionOrderService) UpdateOrderStatusWithEvidence(ctx context.Context, orderID uint, req *UpdateOrderStatusRequest, adminID uint) (*entities.SubscriptionOrder, error) {
	return sos.operationsService.UpdateOrderStatusWithEvidence(ctx, orderID, req, adminID)
}

// ProcessRefund processes a refund for an order
func (sos *SubscriptionOrderService) ProcessRefund(ctx context.Context, req *ProcessRefundRequest) (*entities.SubscriptionOrder, error) {
	return sos.operationsService.ProcessRefund(ctx, req)
}

// CancelSubscriptionOrder cancels a subscription order
func (sos *SubscriptionOrderService) CancelSubscriptionOrder(ctx context.Context, orderID uint, reason string) error {
	return sos.operationsService.CancelSubscriptionOrder(ctx, orderID, reason)
}

// ===============================================================================
// ANALYTICS METHODS - Delegated to SubscriptionOrderAnalyticsService
// ===============================================================================

// GetOrderStatistics gets comprehensive order statistics
func (sos *SubscriptionOrderService) GetOrderStatistics(ctx context.Context, fromDate, toDate time.Time) (map[string]any, error) {
	return sos.analyticsService.GetOrderStatistics(ctx, fromDate, toDate)
}

// BulkUpdateOrders performs bulk operations on orders
func (sos *SubscriptionOrderService) BulkUpdateOrders(ctx context.Context, req *BulkUpdateOrdersRequest) (*BulkUpdateResult, error) {
	return sos.analyticsService.BulkUpdateOrders(ctx, req)
}

// ===============================================================================
// UTILITY METHODS - Direct access to utils service methods
// ===============================================================================

// generateOrderNumber generates a unique order number with enhanced security
func (sos *SubscriptionOrderService) generateOrderNumber() string {
	return sos.utilsService.generateOrderNumber()
}

// checkDuplicateOrders prevents users from creating multiple pending orders for the same plan
func (sos *SubscriptionOrderService) checkDuplicateOrders(ctx context.Context, userID, planID uint, orderType string) error {
	return sos.utilsService.checkDuplicateOrders(ctx, userID, planID, orderType)
}
