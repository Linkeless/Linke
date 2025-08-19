package implementations

import (
	"gorm.io/gorm"

	couponInterfaces "linke/internal/domains/coupon/usecases/interfaces"
	invoiceInterfaces "linke/internal/domains/invoice/usecases/interfaces"
	paymentInterfaces "linke/internal/domains/payment/usecases/interfaces"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/events"
)

// SubscriptionOrderServiceBase contains all shared dependencies
type SubscriptionOrderServiceBase struct {
	db                      *gorm.DB
	subscriptionPlanService interfaces.SubscriptionPlanService
	userSubscriptionService interfaces.UserSubscriptionService
	paymentService          paymentInterfaces.PaymentService
	paymentMethodService    paymentInterfaces.PaymentMethodService
	couponService           couponInterfaces.CouponService
	invoiceService          invoiceInterfaces.InvoiceService
	eventPublisher          events.Publisher
}

// NewSubscriptionOrderServiceBase creates a new base service with all dependencies
func NewSubscriptionOrderServiceBase(
	db *gorm.DB,
	subscriptionPlanService interfaces.SubscriptionPlanService,
	userSubscriptionService interfaces.UserSubscriptionService,
	paymentService paymentInterfaces.PaymentService,
	paymentMethodService paymentInterfaces.PaymentMethodService,
	couponService couponInterfaces.CouponService,
	invoiceService invoiceInterfaces.InvoiceService,
	eventPublisher events.Publisher,
) *SubscriptionOrderServiceBase {
	return &SubscriptionOrderServiceBase{
		db:                      db,
		subscriptionPlanService: subscriptionPlanService,
		userSubscriptionService: userSubscriptionService,
		paymentService:          paymentService,
		paymentMethodService:    paymentMethodService,
		couponService:           couponService,
		invoiceService:          invoiceService,
		eventPublisher:          eventPublisher,
	}
}
