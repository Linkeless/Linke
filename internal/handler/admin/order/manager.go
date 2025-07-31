package order

import (
	"linke/internal/handler/admin/order/operation"
	"linke/internal/handler/admin/order/query"
	"linke/internal/handler/admin/order/statistics"
	"linke/internal/handler/admin/order/status"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// AdminOrderManager manages all order-related admin handlers
type AdminOrderManager struct {
	// Sub-handlers for different order management aspects
	List       *query.OrderListHandler
	Detail     *query.OrderDetailHandler
	Status     *status.OrderStatusHandler
	Refund     *operation.OrderRefundHandler
	Bulk       *operation.OrderBulkHandler
	Stats      *statistics.OrderStatsHandler
}

// NewAdminOrderManager creates a new admin order manager with all sub-handlers
func NewAdminOrderManager(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *AdminOrderManager {
	return &AdminOrderManager{
		List:   query.NewOrderListHandler(subscriptionOrderService, paymentService, userService),
		Detail: query.NewOrderDetailHandler(subscriptionOrderService, paymentService, userService),
		Status: status.NewOrderStatusHandler(subscriptionOrderService, paymentService, userService),
		Refund: operation.NewOrderRefundHandler(subscriptionOrderService, paymentService, userService),
		Bulk:   operation.NewOrderBulkHandler(subscriptionOrderService, paymentService, userService),
		Stats:  statistics.NewOrderStatsHandler(subscriptionOrderService, paymentService, userService),
	}
}

// Legacy compatibility layer - maintains the same interface as the original AdminOrderHandler
// This allows existing code to continue working without changes while using the modular structure internally

// ListOrders delegates to the list module
func (m *AdminOrderManager) ListOrders(c *gin.Context) {
	m.List.ListOrders(c)
}

// GetOrder delegates to the detail module
func (m *AdminOrderManager) GetOrder(c *gin.Context) {
	m.Detail.GetOrder(c)
}

// UpdateOrderStatus delegates to the status module
func (m *AdminOrderManager) UpdateOrderStatus(c *gin.Context) {
	m.Status.UpdateOrderStatus(c)
}

// ProcessRefund delegates to the refund module
func (m *AdminOrderManager) ProcessRefund(c *gin.Context) {
	m.Refund.ProcessRefund(c)
}

// GetOrderStats delegates to the stats module
func (m *AdminOrderManager) GetOrderStats(c *gin.Context) {
	m.Stats.GetOrderStats(c)
}

// BulkUpdate delegates to the bulk module
func (m *AdminOrderManager) BulkUpdate(c *gin.Context) {
	m.Bulk.BulkUpdate(c)
}