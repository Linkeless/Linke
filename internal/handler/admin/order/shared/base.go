package shared

import (
	"linke/internal/service"
)

// BaseHandler provides common dependencies for order handlers
type BaseHandler struct {
	SubscriptionOrderService *service.SubscriptionOrderService
	PaymentService           *service.PaymentService
	UserService              *service.UserService
	Validator                *OrderValidator
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(
	subscriptionOrderService *service.SubscriptionOrderService,
	paymentService *service.PaymentService,
	userService *service.UserService,
) *BaseHandler {
	return &BaseHandler{
		SubscriptionOrderService: subscriptionOrderService,
		PaymentService:           paymentService,
		UserService:              userService,
		Validator:                NewOrderValidator(),
	}
}