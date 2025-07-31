package statistics

import (
	"linke/internal/handler/admin/coupon/shared"
	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// CouponUsageHandler handles coupon usage statistics operations
type CouponUsageHandler struct {
	*shared.BaseHandler
}

// NewCouponUsageHandler creates a new coupon usage handler
func NewCouponUsageHandler(couponService *service.CouponService) *CouponUsageHandler {
	return &CouponUsageHandler{
		BaseHandler: shared.NewBaseHandler(couponService),
	}
}

// GetCouponUsages godoc
// @Summary [Admin] Get coupon usage history
// @Description Get a paginated list of coupon usage records
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Param limit query int false "Number of items per page (1-100)" minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" minimum(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.CouponUsageResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/{id}/usages [get]
func (h *CouponUsageHandler) GetCouponUsages(c *gin.Context) {
	couponID, err := h.Validator.ValidateCouponID(c)
	if err != nil {
		return // Response already handled by validator
	}

	// Validate pagination parameters
	limit, offset := h.Validator.ValidatePaginationParams(c)

	// Check if coupon exists
	_, err = h.CouponService.GetCoupon(c.Request.Context(), couponID)
	if err != nil {
		logger.Error("Failed to check coupon existence", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to check coupon", err.Error())
		}
		return
	}

	// Get coupon usages
	usages, totalCount, err := h.CouponService.GetCouponUsages(c.Request.Context(), &couponID, nil, limit, offset)
	if err != nil {
		logger.Error("Failed to get coupon usages", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		response.InternalServerError(c, "Failed to get coupon usages", err.Error())
		return
	}

	// Convert to response format
	usageResponses := make([]*model.CouponUsageResponse, len(usages))
	for i, usage := range usages {
		usageResponses[i] = usage.ToResponse()
	}

	response.OKPaginated(c, "Coupon usages retrieved successfully", usageResponses, totalCount, limit, offset)
}