package query

import (
	"linke/internal/handler/admin/coupon/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// CouponSearchHandler handles coupon search operations
type CouponSearchHandler struct {
	*shared.BaseHandler
}

// NewCouponSearchHandler creates a new coupon search handler
func NewCouponSearchHandler(couponService *service.CouponService) *CouponSearchHandler {
	return &CouponSearchHandler{
		BaseHandler: shared.NewBaseHandler(couponService),
	}
}

// GetCouponByCode godoc
// @Summary [Admin] Get coupon by code
// @Description Get coupon information by coupon code
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Coupon Code"
// @Success 200 {object} response.StandardResponse{data=model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/code/{code} [get]
func (h *CouponSearchHandler) GetCouponByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Coupon code is required")
		return
	}

	// Validate coupon code format
	if err := h.Validator.ValidateCouponCode(code); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get coupon by code
	coupon, err := h.CouponService.GetCouponByCode(c.Request.Context(), code)
	if err != nil {
		logger.Error("Failed to get coupon by code", logger.Error2("error", err), logger.String("code", code))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to get coupon", err.Error())
		}
		return
	}

	response.OK(c, "Coupon retrieved successfully", coupon.ToResponse())
}