package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	authInterfaces "linke/internal/domains/auth/usecases/interfaces"
	"linke/internal/domains/coupon/dto"
	"linke/internal/domains/coupon/usecases/interfaces"
	userEntities "linke/internal/domains/user/entities"
	"linke/internal/shared/logger"
	"linke/internal/shared/middleware"
	"linke/internal/shared/response"
)

// UserCouponHandler handles user coupon operations
type UserCouponHandler struct {
	couponService interfaces.CouponService
	authService   authInterfaces.AuthService
}

// NewUserCouponHandler creates a new UserCouponHandler
func NewUserCouponHandler(couponService interfaces.CouponService, authService authInterfaces.AuthService) *UserCouponHandler {
	return &UserCouponHandler{
		couponService: couponService,
		authService:   authService,
	}
}

// GetPublicCoupons godoc
// @Summary Get public coupons
// @Description Get list of public coupons available to all users (no authentication required)
// @Tags Coupon
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" minimum(1) maximum(50) default(10) example(10)
// @Success 200 {object} []dto.CouponResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /coupons [get]
func (h *UserCouponHandler) GetPublicCoupons(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Get public coupons
	coupons, err := h.couponService.GetPublicCoupons(c.Request.Context(), limit)
	if err != nil {
		logger.Error("Failed to get public coupons", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get public coupons")
		return
	}

	// Convert to public response format (without sensitive info) with preallocation
	couponResponses := make([]*dto.CouponResponse, 0, len(coupons))
	for _, coupon := range coupons {
		couponResponses = append(couponResponses, dto.ToPublicResponse(coupon))
	}

	response.OK(c, couponResponses)
}

// ValidateCoupon godoc
// @Summary Validate coupon code
// @Description Validate a coupon code for a specific user and order
// @Tags Coupon
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Coupon code to validate"
// @Param order_amount query number true "Order amount for validation"
// @Param plan_id query int true "Plan ID for validation"
// @Param currency query string false "Currency code (default: CNY)"
// @Success 200 {object} dto.ValidateCouponResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /coupons/{code}/validation [get]
func (h *UserCouponHandler) ValidateCoupon(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Get parameters
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "Coupon code is required")
		return
	}

	orderAmountStr := c.Query("order_amount")
	if orderAmountStr == "" {
		response.BadRequest(c, "Order amount is required")
		return
	}

	orderAmount, err := strconv.ParseFloat(orderAmountStr, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order amount")
		return
	}

	planIDStr := c.Query("plan_id")
	if planIDStr == "" {
		response.BadRequest(c, "Plan ID is required")
		return
	}

	planID, err := strconv.ParseUint(planIDStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID")
		return
	}

	currency := c.Query("currency")
	if currency == "" {
		currency = "CNY" // Default currency
	}

	// Build validation request
	req := dto.ValidateCouponRequest{
		Code:        code,
		UserID:      uint64(user.ID),
		OrderAmount: orderAmount,
		PlanID:      planID,
		Currency:    currency,
	}

	// Validate coupon
	validationResponse, err := h.couponService.ValidateCoupon(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to validate coupon",
			logger.String("code", req.Code),
			logger.Uint("user_id", uint(user.ID)),
			logger.ErrorField(err),
		)

		// Check for coupon not found
		if err.Error() == "coupon not found" || err.Error() == "record not found" {
			response.NotFound(c, "Coupon not found")
			return
		}

		response.InternalServerError(c, "Failed to validate coupon")
		return
	}

	logger.Info("Coupon validation completed",
		logger.String("code", req.Code),
		logger.Uint("user_id", uint(user.ID)),
		logger.Bool("valid", validationResponse.Valid),
	)

	response.OK(c, validationResponse)
}

// GetMyCouponUsage godoc
// @Summary Get my coupon usage history
// @Description Get current user's coupon usage records
// @Tags Coupon
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" minimum(1) default(1) example(1)
// @Param limit query int false "Items per page" minimum(1) maximum(100) default(10) example(10)
// @Success 200 {object} response.HALCollectionResponse
// @Failure 400 {object} response.BadRequestResponse
// @Failure 401 {object} response.UnauthorizedResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /coupons/my-usage [get]
func (h *UserCouponHandler) GetMyCouponUsage(c *gin.Context) {
	// Get current user from context
	userValue, exists := c.Get(middleware.AuthContextKey)
	if !exists {
		response.Unauthorized(c, "Authentication required")
		return
	}

	user, ok := userValue.(*userEntities.User)
	if !ok {
		response.Unauthorized(c, "Invalid user context")
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Get user's coupon usage
	usageRecords, total, err := h.couponService.GetUserCouponUsage(c.Request.Context(), uint64(user.ID), limit, offset)
	if err != nil {
		logger.Error("Failed to get user coupon usage",
			logger.Uint("user_id", uint(user.ID)),
			logger.ErrorField(err),
		)
		response.InternalServerError(c, "Failed to get coupon usage history")
		return
	}

	// Convert to responses
	var usageResponses []*dto.CouponUsageResponse
	for _, usage := range usageRecords {
		usageResponses = append(usageResponses, dto.CouponUsageToResponse(usage))
	}

	response.Paginated(c, "Coupon usage history retrieved successfully", usageResponses, page, limit, total, "/api/v1/coupons/usage/my")
}

// authServiceAdapter adapts the domain AuthService to middleware AuthService interface
type authServiceAdapter struct {
	authService authInterfaces.AuthService
}

func (a *authServiceAdapter) ValidateToken(token string) (any, error) {
	user, err := a.authService.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// RegisterRoutes registers all user coupon routes
func (h *UserCouponHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Create auth service adapter
	authAdapter := &authServiceAdapter{authService: h.authService}

	// Coupon routes
	couponGroup := router.Group("/coupons")
	{
		// Public route (no authentication required)
		couponGroup.GET("", h.GetPublicCoupons)

		// Authenticated routes - apply auth middleware
		couponGroup.GET("/:code/validation", middleware.AuthMiddleware(authAdapter), h.ValidateCoupon)
		couponGroup.GET("/my-usage", middleware.AuthMiddleware(authAdapter), h.GetMyCouponUsage)
	}
}
