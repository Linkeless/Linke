package handlers

import (
	"strconv"

	"linke/internal/domains/subscription/entities"
	"linke/internal/domains/subscription/usecases/interfaces"
	"linke/internal/shared/logger"
	"linke/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// SubscriptionPlanHandler handles public subscription plan operations
type SubscriptionPlanHandler struct {
	subscriptionPlanService interfaces.SubscriptionPlanService
}

// NewSubscriptionPlanHandler creates a new subscription plan handler
func NewSubscriptionPlanHandler(subscriptionPlanService interfaces.SubscriptionPlanService) *SubscriptionPlanHandler {
	return &SubscriptionPlanHandler{
		subscriptionPlanService: subscriptionPlanService,
	}
}

// GetSubscriptionPlans godoc
// @Summary Get available subscription plans
// @Description Get a list of visible subscription plans for purchase
// @Tags Public-Subscription-Plans
// @Accept json
// @Produce json
// @Param currency query string false "Filter by currency" example("USD")
// @Param limit query int false "Limit results" minimum(1) maximum(100) example(10)
// @Param offset query int false "Offset results" minimum(0) example(0)
// @Success 200 {object} response.PaginatedResponse{data=[]entities.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription/plans [get]
func (h *SubscriptionPlanHandler) GetSubscriptionPlans(c *gin.Context) {
	// Parse query parameters
	var req interfaces.GetSubscriptionPlansRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Force visible to true for public endpoint
	visible := true
	req.Visible = &visible
	// Force active status for public endpoint
	req.Status = "active"

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Get subscription plans
	plans, totalCount, err := h.subscriptionPlanService.GetSubscriptionPlans(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to get subscription plans", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get subscription plans", err.Error())
		return
	}

	// Convert to public response format (hides internal details)
	var planResponses []*entities.SubscriptionPlanResponse
	for _, plan := range plans {
		planResponses = append(planResponses, plan.ToPublicResponse())
	}

	response.OKPaginated(c, "Subscription plans retrieved successfully", planResponses, totalCount, req.Limit, req.Offset)
}

// GetPopularSubscriptionPlans godoc
// @Summary Get popular subscription plans
// @Description Get a list of popular/recommended subscription plans
// @Tags Public-Subscription-Plans
// @Accept json
// @Produce json
// @Param limit query int false "Limit results" minimum(1) maximum(20) example(5)
// @Success 200 {object} response.StandardResponse{data=[]entities.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription/plans/popular [get]
func (h *SubscriptionPlanHandler) GetPopularSubscriptionPlans(c *gin.Context) {
	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 20 {
		limit = 5
	}

	// Get popular subscription plans
	plans, err := h.subscriptionPlanService.GetPopularSubscriptionPlans(c.Request.Context(), limit)
	if err != nil {
		logger.Error("Failed to get popular subscription plans", logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get popular subscription plans", err.Error())
		return
	}

	// Convert to public response format (hides internal details)
	var planResponses []*entities.SubscriptionPlanResponse
	for _, plan := range plans {
		planResponses = append(planResponses, plan.ToPublicResponse())
	}

	response.OK(c, "Popular subscription plans retrieved successfully", planResponses)
}

// GetSubscriptionPlan godoc
// @Summary Get subscription plan details
// @Description Get detailed information about a specific subscription plan
// @Tags Public-Subscription-Plans
// @Accept json
// @Produce json
// @Param id path int true "Plan ID"
// @Success 200 {object} response.StandardResponse{data=entities.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription/plans/{id} [get]
func (h *SubscriptionPlanHandler) GetSubscriptionPlan(c *gin.Context) {
	// Parse plan ID
	planIDStr := c.Param("id")
	planID, err := strconv.ParseUint(planIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid plan ID", "Plan ID must be a valid number")
		return
	}

	// Get subscription plan
	plan, err := h.subscriptionPlanService.GetSubscriptionPlan(c.Request.Context(), uint(planID))
	if err != nil {
		if err.Error() == "subscription plan not found" {
			response.NotFound(c, "Subscription plan not found")
			return
		}
		logger.Error("Failed to get subscription plan", logger.ErrorField(err), logger.Uint("plan_id", uint(planID)))
		response.InternalServerError(c, "Failed to get subscription plan", err.Error())
		return
	}

	// Check if plan is available for public viewing
	if !plan.IsAvailableForPurchase() {
		response.NotFound(c, "Subscription plan not found")
		return
	}

	// Convert to public response format (hides internal details)
	response.OK(c, "Subscription plan retrieved successfully", plan.ToPublicResponse())
}

// GetSubscriptionPlanByCode godoc
// @Summary Get subscription plan by code
// @Description Get detailed information about a subscription plan using its unique code
// @Tags Public-Subscription-Plans
// @Accept json
// @Produce json
// @Param code path string true "Plan Code" example("premium-monthly")
// @Success 200 {object} response.StandardResponse{data=entities.SubscriptionPlanResponse}
// @Failure 400 {object} response.BadRequestResponse
// @Failure 404 {object} response.NotFoundResponse
// @Failure 500 {object} response.InternalServerErrorResponse
// @Router /subscription/plans/code/{code} [get]
func (h *SubscriptionPlanHandler) GetSubscriptionPlanByCode(c *gin.Context) {
	// Parse plan code
	planCode := c.Param("code")
	if planCode == "" {
		response.BadRequest(c, "Invalid plan code", "Plan code cannot be empty")
		return
	}

	// Get subscription plan by code
	plan, err := h.subscriptionPlanService.GetSubscriptionPlanByCode(c.Request.Context(), planCode)
	if err != nil {
		if err.Error() == "subscription plan not found" {
			response.NotFound(c, "Subscription plan not found")
			return
		}
		logger.Error("Failed to get subscription plan by code", logger.String("plan_code", planCode), logger.ErrorField(err))
		response.InternalServerError(c, "Failed to get subscription plan", err.Error())
		return
	}

	// Check if plan is available for public viewing
	if !plan.IsAvailableForPurchase() {
		response.NotFound(c, "Subscription plan not found")
		return
	}

	// Convert to public response format (hides internal details)
	response.OK(c, "Subscription plan retrieved successfully", plan.ToPublicResponse())
}

// RegisterRoutes registers all subscription plan routes
func (h *SubscriptionPlanHandler) RegisterRoutes(router *gin.RouterGroup) {
	// Public subscription plan routes - no authentication required
	planGroup := router.Group("/subscription/plans")
	{
		planGroup.GET("", h.GetSubscriptionPlans)
		planGroup.GET("/popular", h.GetPopularSubscriptionPlans)
		planGroup.GET("/:id", h.GetSubscriptionPlan)
		planGroup.GET("/code/:code", h.GetSubscriptionPlanByCode)
	}
}
