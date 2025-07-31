package management

import (
	"linke/internal/handler/admin/coupon/shared"
	"linke/internal/logger"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// CouponCRUDHandler handles coupon CRUD operations
type CouponCRUDHandler struct {
	*shared.BaseHandler
}

// NewCouponCRUDHandler creates a new coupon CRUD handler
func NewCouponCRUDHandler(couponService *service.CouponService) *CouponCRUDHandler {
	return &CouponCRUDHandler{
		BaseHandler: shared.NewBaseHandler(couponService),
	}
}

// CreateCoupon godoc
// @Summary [Admin] Create coupon
// @Description Create a new coupon with specified parameters
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param create_request body service.CreateCouponRequest true "Coupon creation data"
// @Success 201 {object} response.StandardResponse{data=model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons [post]
func (h *CouponCRUDHandler) CreateCoupon(c *gin.Context) {
	userID, err := h.Validator.ValidateUserID(c)
	if err != nil {
		return // Response already handled by validator
	}

	var req service.CreateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid create coupon request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Additional validation if needed
	if err := h.Validator.ValidateCreateCouponRequest(c, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Create coupon
	coupon, err := h.CouponService.CreateCoupon(c.Request.Context(), userID, &req)
	if err != nil {
		logger.Error("Failed to create coupon", logger.Error2("error", err))
		response.InternalServerError(c, "Failed to create coupon", err.Error())
		return
	}

	logger.Info("Coupon created successfully by admin", 
		logger.Any("coupon_id", coupon.ID),
		logger.String("coupon_code", coupon.Code),
		logger.Any("admin_id", userID))

	response.CreatedWithMessage(c, "Coupon created successfully", coupon.ToResponse())
}

// GetCoupon godoc
// @Summary [Admin] Get coupon details
// @Description Get detailed information about a specific coupon
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Success 200 {object} response.StandardResponse{data=model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/{id} [get]
func (h *CouponCRUDHandler) GetCoupon(c *gin.Context) {
	couponID, err := h.Validator.ValidateCouponID(c)
	if err != nil {
		return // Response already handled by validator
	}

	// Get coupon
	coupon, err := h.CouponService.GetCoupon(c.Request.Context(), couponID)
	if err != nil {
		logger.Error("Failed to get coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to get coupon", err.Error())
		}
		return
	}

	response.OK(c, "Coupon retrieved successfully", coupon.ToResponse())
}

// UpdateCoupon godoc
// @Summary [Admin] Update coupon
// @Description Update an existing coupon's information
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Param update_request body service.UpdateCouponRequest true "Coupon update data"
// @Success 200 {object} response.StandardResponse{data=model.CouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/{id} [put]
func (h *CouponCRUDHandler) UpdateCoupon(c *gin.Context) {
	couponID, err := h.Validator.ValidateCouponID(c)
	if err != nil {
		return // Response already handled by validator
	}

	var req service.UpdateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid update coupon request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Additional validation if needed
	if err := h.Validator.ValidateUpdateCouponRequest(c, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Update coupon
	coupon, err := h.CouponService.UpdateCoupon(c.Request.Context(), couponID, &req)
	if err != nil {
		logger.Error("Failed to update coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to update coupon", err.Error())
		}
		return
	}

	// Get current user ID for logging
	userID, _ := h.Validator.ValidateUserID(c)
	logger.Info("Coupon updated successfully by admin", 
		logger.Any("coupon_id", coupon.ID),
		logger.Any("admin_id", userID))

	response.OK(c, "Coupon updated successfully", coupon.ToResponse())
}

// DeleteCoupon godoc
// @Summary [Admin] Delete coupon
// @Description Soft delete a coupon (it will no longer be usable)
// @Tags Admin - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Coupon ID"
// @Success 200 {object} response.StandardResponse
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 403 {object} response.StandardResponse
// @Failure 404 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /admin/coupons/{id} [delete]
func (h *CouponCRUDHandler) DeleteCoupon(c *gin.Context) {
	couponID, err := h.Validator.ValidateCouponID(c)
	if err != nil {
		return // Response already handled by validator
	}

	// Delete coupon
	err = h.CouponService.DeleteCoupon(c.Request.Context(), couponID)
	if err != nil {
		logger.Error("Failed to delete coupon", logger.Error2("error", err), logger.Any("coupon_id", couponID))
		if err.Error() == "coupon not found" {
			response.NotFound(c, "Coupon not found")
		} else {
			response.InternalServerError(c, "Failed to delete coupon", err.Error())
		}
		return
	}

	// Get current user ID for logging
	userID, _ := h.Validator.ValidateUserID(c)
	logger.Info("Coupon deleted successfully by admin", 
		logger.Any("coupon_id", couponID),
		logger.Any("admin_id", userID))

	response.OK(c, "Coupon deleted successfully", nil)
}