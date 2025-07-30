package user

import (
	"strconv"

	"linke/internal/logger"
	"linke/internal/model"
	"linke/internal/response"
	"linke/internal/service"

	"github.com/gin-gonic/gin"
)

// UserCouponHandler handles user coupon operations
type UserCouponHandler struct {
	couponService *service.CouponService
}

// NewUserCouponHandler creates a new user coupon handler
func NewUserCouponHandler(couponService *service.CouponService) *UserCouponHandler {
	return &UserCouponHandler{
		couponService: couponService,
	}
}

// GetPublicCoupons is deprecated and disabled for security reasons
// Coupon codes should be distributed through targeted marketing channels
func (h *UserCouponHandler) GetPublicCoupons(c *gin.Context) {
	// This endpoint has been disabled for security reasons
	response.BadRequest(c, "This endpoint has been disabled for security reasons", "Coupon codes should be obtained through official marketing channels")
}

// ValidateCoupon godoc
// @Summary [User] Validate coupon code
// @Description Validate a coupon code for a specific plan and order amount
// @Tags User - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param validate_request body ValidateCouponRequest true "Coupon validation data"
// @Success 200 {object} response.StandardResponse{data=service.ValidateCouponResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /user/coupons/validate [post]
func (h *UserCouponHandler) ValidateCoupon(c *gin.Context) {
	// Get current user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Define the request structure locally
	var req struct {
		CouponCode  string  `json:"coupon_code" binding:"required" example:"SAVE20"`
		PlanID      uint    `json:"plan_id" binding:"required" example:"1"`
		OrderAmount float64 `json:"order_amount" binding:"required,min=0" example:"29.99"`
		Currency    string  `json:"currency" binding:"required,len=3" example:"USD"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid validate coupon request", logger.Error2("error", err))
		response.BadRequest(c, "Invalid request data", err.Error())
		return
	}

	// Create coupon validation request
	validateReq := &service.ValidateCouponRequest{
		Code:        req.CouponCode,
		UserID:      uint64(userID.(uint)),
		OrderAmount: req.OrderAmount,
		PlanID:      uint64(req.PlanID),
		Currency:    req.Currency,
	}

	// Validate coupon
	validateResp, err := h.couponService.ValidateCoupon(c.Request.Context(), validateReq)
	if err != nil {
		logger.Error("Failed to validate coupon", logger.Error2("error", err), 
			logger.String("coupon_code", req.CouponCode),
			logger.Uint("user_id", userID.(uint)))
		response.InternalServerError(c, "Failed to validate coupon", err.Error())
		return
	}

	response.OK(c, "Coupon validation completed", validateResp)
}

// GetMyCouponUsages godoc
// @Summary [User] Get my coupon usage history
// @Description Get a paginated list of current user's coupon usage records
// @Tags User - Coupons
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of items per page (1-100)" minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" minimum(0)
// @Success 200 {object} response.PaginatedResponse{data=[]model.CouponUsageResponse}
// @Failure 400 {object} response.StandardResponse
// @Failure 401 {object} response.StandardResponse
// @Failure 500 {object} response.StandardResponse
// @Router /user/coupons/usages [get]
func (h *UserCouponHandler) GetMyCouponUsages(c *gin.Context) {
	// Get current user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Error("User ID not found in context")
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Parse query parameters
	limit := 20 // default
	offset := 0 // default

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get user's coupon usages
	userID64 := uint64(userID.(uint))
	usages, totalCount, err := h.couponService.GetCouponUsages(c.Request.Context(), nil, &userID64, limit, offset)
	if err != nil {
		logger.Error("Failed to get user coupon usages", logger.Error2("error", err), logger.Uint("user_id", userID.(uint)))
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