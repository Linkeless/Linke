package coupon

import (
	couponshared "linke/internal/handler/user/coupon/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// CouponOperationHandler handles coupon operation-related requests
type CouponOperationHandler struct {
	*couponshared.BaseCouponHandler
	validator *couponshared.CouponValidator
}

// NewCouponOperationHandler creates a new coupon operation handler
func NewCouponOperationHandler(couponService *service.CouponService) *CouponOperationHandler {
	return &CouponOperationHandler{
		BaseCouponHandler: couponshared.NewBaseCouponHandler(couponService),
		validator:         couponshared.NewCouponValidator(),
	}
}

// ValidateCoupon godoc
// @Summary [User] Validate coupon code
// @Description Validate a coupon code for a specific plan and order amount
// @Tags User - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param validate_request body couponshared.ValidateCouponRequest true "Coupon validation data"
// @Success 200 {object} response.StandardResponse{data=service.ValidateCouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /user/coupons/validate [post]
func (h *CouponOperationHandler) ValidateCoupon(c *gin.Context) {
	// Get current user ID from context
	userID, valid := h.validator.GetUserIDFromContext(c)
	if !valid {
		return
	}

	// Bind and validate request
	req, valid := h.validator.BindAndValidateCouponRequest(c)
	if !valid {
		return
	}

	// Create coupon validation request
	validateReq := &service.ValidateCouponRequest{
		Code:        req.CouponCode,
		UserID:      uint64(userID),
		OrderAmount: req.OrderAmount,  
		PlanID:      uint64(req.PlanID),
		Currency:    req.Currency,
	}

	// Validate coupon
	validateResp, err := h.BaseCouponHandler.CouponService.ValidateCoupon(c.Request.Context(), validateReq)
	if err != nil {
		logger.Error("Failed to validate coupon", logger.Error2("error", err), 
			logger.String("coupon_code", req.CouponCode),
			logger.Uint("user_id", userID))
		response.InternalServerError(c, "Failed to validate coupon", err.Error())
		return
	}

	response.OK(c, "Coupon validation completed", validateResp)
}

// GetPublicCoupons is deprecated and disabled for security reasons
// Coupon codes should be distributed through targeted marketing channels
func (h *CouponOperationHandler) GetPublicCoupons(c *gin.Context) {
	// This endpoint has been disabled for security reasons
	response.BadRequest(c, "This endpoint has been disabled for security reasons", "Coupon codes should be obtained through official marketing channels")
}