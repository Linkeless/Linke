package shared

import (
	"linke/internal/service"
)

// BaseHandler provides common dependencies for coupon handlers
type BaseHandler struct {
	CouponService *service.CouponService
	Validator     *CouponValidator
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(couponService *service.CouponService) *BaseHandler {
	return &BaseHandler{
		CouponService: couponService,
		Validator:     NewCouponValidator(),
	}
}